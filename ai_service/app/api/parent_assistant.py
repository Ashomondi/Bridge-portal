from dataclasses import dataclass, field
from typing import List, Optional

from fastapi import APIRouter, Depends

from .. import prompts
from ..clients.backend_client import BackendClient, BackendError
from ..clients.llm_client import LLMClient, LLMError
from ..config import Settings, get_settings
from ..deps import get_backend_client, get_llm_client
from ..schemas import (
    AuditLogRequest,
    Intent,
    ParentQueryData,
    ParentQueryRequest,
    ParentQueryResponse,
    PaymentState,
    UserContext,
)
from ..security import AuthContext, require_auth
from ..state import SessionLimitError, SessionStore, get_session_store
from ..utils import (
    amounts_match,
    classify_intent,
    counts_present,
    gate_read,
    resolve_child,
    safe_profile,
)
from .chat_payment import CHILDREN_ENDPOINT, handle_payment_turn

router = APIRouter()


@dataclass
class QueryResult:
    reply: str
    intent: Intent
    payment_state: PaymentState
    action_taken: str
    portal_link: Optional[str] = None
    endpoints_read: List[str] = field(default_factory=list)


async def _audit(
    backend: BackendClient,
    ctx: UserContext,
    feature: str,
    intent: Intent,
    result: QueryResult,
) -> None:
    await backend.audit_log(
        AuditLogRequest(
            feature=feature,
            session_id=ctx.session_id,
            channel=ctx.channel,
            intent=intent.value,
            endpoints_read=result.endpoints_read,
            action_taken=result.action_taken,
        )
    )


def _clarify_reply(children) -> str:
    names = ", ".join(c.full_name for c in children if c.full_name)
    return f"Which child do you mean? You can choose: {names}."


def _fee_fact_prompt(student_name: str, fee) -> str:
    return (
        f"Student: {student_name} (Admission {fee.admission_number}).\n"
        f"Term: {fee.term_name}. Due date: {fee.due_date}.\n"
        f"Current balance: {prompts.format_kes(fee.current_balance)}.\n"
        f"Billed: {prompts.format_kes(fee.billed_amount)}.\n"
        f"Paid: {prompts.format_kes(fee.paid_amount)}.\n"
        f"Credit: {prompts.format_kes(fee.credit_balance)}.\n"
        "Write a short reply to the parent asking how much they owe, using ONLY these "
        "exact figures. Do not change any number. End by offering to pay via M-Pesa."
    )


async def _compose_amount_safe(llm, fallback, user_prompt, allowed_values) -> str:
    try:
        text = await llm.complete_text(prompts.ASSISTANT_SYSTEM_PROMPT, user_prompt)
    except LLMError:
        return fallback
    if not text.strip() or not amounts_match(text, allowed_values):
        return fallback
    return text


async def _handle_fee(
    req, ctx, backend, llm, store, settings
) -> QueryResult:
    child, children = await resolve_child(
        req.student_id, req.message, ctx.user_id, backend, store, ctx.session_id, settings
    )
    endpoints = [CHILDREN_ENDPOINT.format(user_id=ctx.user_id)]
    if child is None:
        return QueryResult(
            reply=_clarify_reply(children),
            intent=Intent.FEE_INQUIRY,
            payment_state=PaymentState.NONE,
            action_taken="FEE_CLARIFY_CHILD",
            endpoints_read=endpoints,
        )

    await gate_read(store, ctx.session_id, backend, settings)
    endpoints.append(f"/api/v1/internal/students/{child.student_id}/fee-balance")
    fee = await backend.get_fee_balance(child.student_id)

    fallback = prompts.fee_reply(fee, child.full_name, child.admission_number)
    allowed_values = [
        fee.current_balance,
        fee.billed_amount,
        fee.paid_amount,
        fee.credit_balance,
        fee.opening_balance,
    ]
    reply = await _compose_amount_safe(
        llm, fallback, _fee_fact_prompt(child.full_name, fee), allowed_values
    )
    return QueryResult(
        reply=reply,
        intent=Intent.FEE_INQUIRY,
        payment_state=PaymentState.NONE,
        action_taken="READ_FEE_BALANCE",
        endpoints_read=endpoints,
    )


async def _handle_attendance(req, ctx, backend, llm, store, settings) -> QueryResult:
    child, children = await resolve_child(
        req.student_id, req.message, ctx.user_id, backend, store, ctx.session_id, settings
    )
    endpoints = [CHILDREN_ENDPOINT.format(user_id=ctx.user_id)]
    if child is None:
        return QueryResult(
            reply=_clarify_reply(children),
            intent=Intent.ATTENDANCE_INQUIRY,
            payment_state=PaymentState.NONE,
            action_taken="ATTENDANCE_CLARIFY_CHILD",
            endpoints_read=endpoints,
        )

    await gate_read(store, ctx.session_id, backend, settings)
    endpoints.append(f"/api/v1/internal/students/{child.student_id}/attendance-summary")
    attendance = await backend.get_attendance_summary(child.student_id)

    fallback = prompts.attendance_reply(attendance, child.full_name)
    user_prompt = (
        f"Student: {child.full_name}.\n"
        f"Term: {attendance.term_name}.\n"
        f"School days: {attendance.total_school_days}.\n"
        f"Days present: {attendance.days_present}.\n"
        f"Days absent: {attendance.days_absent}.\n"
        "Write a short, clear attendance summary to the parent using ONLY these exact numbers."
    )
    try:
        text = await llm.complete_text(prompts.ASSISTANT_SYSTEM_PROMPT, user_prompt)
        reply = (
            text
            if counts_present(
                text,
                [
                    attendance.total_school_days,
                    attendance.days_present,
                    attendance.days_absent,
                ],
            )
            else fallback
        )
    except LLMError:
        reply = fallback

    return QueryResult(
        reply=reply,
        intent=Intent.ATTENDANCE_INQUIRY,
        payment_state=PaymentState.NONE,
        action_taken="READ_ATTENDANCE",
        endpoints_read=endpoints,
    )


def _summarize_marks(marks) -> str:
    parts = []
    for subject in marks.subjects or []:
        name = subject.get("subject_name") or subject.get("name")
        level = subject.get("level") or subject.get("performance_level")
        if name and level:
            parts.append(f"{name}: {level}")
    if not parts:
        return (
            "The results are released. Please open the Performance Graph tab on the "
            "Parent Portal to view the subject performance levels."
        )
    return (
        f"For {marks.term_name or 'this term'}, the released performance levels are: "
        + "; ".join(parts)
        + ". To view the full performance graph, open the Performance Graph tab in the Parent Portal."
    )


async def _handle_results(req, ctx, backend, store, settings) -> QueryResult:
    if ctx.channel.value == "whatsapp":
        return QueryResult(
            reply=prompts.results_blocked_reply(settings.portal_results_url),
            intent=Intent.RESULTS_INQUIRY,
            payment_state=PaymentState.NONE,
            action_taken="RESULTS_BLOCKED_WHATSAPP",
            portal_link=settings.portal_results_url,
        )

    child, children = await resolve_child(
        req.student_id, req.message, ctx.user_id, backend, store, ctx.session_id, settings
    )
    endpoints = [CHILDREN_ENDPOINT.format(user_id=ctx.user_id)]
    if child is None:
        return QueryResult(
            reply=_clarify_reply(children),
            intent=Intent.RESULTS_INQUIRY,
            payment_state=PaymentState.NONE,
            action_taken="RESULTS_CLARIFY_CHILD",
            endpoints_read=endpoints,
        )

    await gate_read(store, ctx.session_id, backend, settings)
    endpoints.append(f"/api/v1/internal/students/{child.student_id}/marks-summary")
    marks = await backend.get_marks_summary(child.student_id)
    if not marks.is_released:
        return QueryResult(
            reply=prompts.results_not_released_reply(),
            intent=Intent.RESULTS_INQUIRY,
            payment_state=PaymentState.NONE,
            action_taken="RESULTS_NOT_RELEASED",
            endpoints_read=endpoints,
        )
    return QueryResult(
        reply=_summarize_marks(marks),
        intent=Intent.RESULTS_INQUIRY,
        payment_state=PaymentState.NONE,
        action_taken="READ_MARKS",
        endpoints_read=endpoints,
    )


async def _handle_out_of_scope(ctx, backend, store, settings) -> QueryResult:
    await gate_read(store, ctx.session_id, backend, settings)
    profile = await backend.get_school_profile()
    return QueryResult(
        reply=prompts.out_of_scope_reply(profile),
        intent=Intent.OUT_OF_SCOPE,
        payment_state=PaymentState.NONE,
        action_taken="OUT_OF_SCOPE_REFUSAL",
        endpoints_read=["/api/v1/internal/school-profile"],
    )


@router.post("/parent-query", response_model=ParentQueryResponse)
async def parent_query(
    req: ParentQueryRequest,
    auth: AuthContext = Depends(require_auth),
    backend: BackendClient = Depends(get_backend_client),
    llm: LLMClient = Depends(get_llm_client),
    store: SessionStore = Depends(get_session_store),
    settings: Settings = Depends(get_settings),
):
    ctx = auth.user
    intent = classify_intent(req.message)
    payment_state = store.get_payment_state(ctx.session_id)

    try:
        if payment_state in (
            PaymentState.AWAITING_CONFIRMATION,
            PaymentState.AWAITING_SLOTS,
        ) or intent == Intent.PAYMENT_FLOW:
            turn = await handle_payment_turn(
                req.message, req.student_id, ctx, store, backend, settings
            )
            result = QueryResult(
                reply=turn.reply,
                intent=Intent.PAYMENT_FLOW,
                payment_state=turn.payment_state,
                action_taken=turn.action_taken,
                endpoints_read=turn.endpoints_read,
            )
        elif intent == Intent.FEE_INQUIRY:
            result = await _handle_fee(req, ctx, backend, llm, store, settings)
        elif intent == Intent.ATTENDANCE_INQUIRY:
            result = await _handle_attendance(req, ctx, backend, llm, store, settings)
        elif intent == Intent.RESULTS_INQUIRY:
            result = await _handle_results(req, ctx, backend, store, settings)
        else:
            result = await _handle_out_of_scope(ctx, backend, store, settings)
    except SessionLimitError as exc:
        await _audit(
            backend,
            ctx,
            "chat_payment" if intent == Intent.PAYMENT_FLOW else "parent_assistant",
            intent,
            QueryResult(
                reply=exc.reply,
                intent=intent,
                payment_state=store.get_payment_state(ctx.session_id),
                action_taken=exc.code,
            ),
        )
        raise
    except BackendError:
        profile = await safe_profile(backend)
        result = QueryResult(
            reply=prompts.payment_failed_reply(profile),
            intent=intent,
            payment_state=store.get_payment_state(ctx.session_id),
            action_taken="BACKEND_ERROR",
        )

    feature = "chat_payment" if intent == Intent.PAYMENT_FLOW else "parent_assistant"
    await _audit(backend, ctx, feature, result.intent, result)

    data = ParentQueryData(
        reply=result.reply,
        intent=result.intent,
        channel=ctx.channel,
        payment_state=result.payment_state,
        action_taken=result.action_taken,
        portal_link=result.portal_link,
    )
    return ParentQueryResponse(data=data)
