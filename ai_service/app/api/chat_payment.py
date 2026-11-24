from dataclasses import dataclass, field
from typing import List, Optional

from .. import prompts
from ..clients.backend_client import BackendClient, BackendError
from ..config import Settings
from ..schemas import PaymentState, StkPushRequest, UserContext
from ..state import PendingPayment, SessionLimitError, SessionStore
from ..utils import (
    display_phone,
    extract_amount,
    extract_phone,
    is_affirmation,
    is_negative,
    resolve_child,
    safe_profile,
)

STK_ENDPOINT = "/api/v1/internal/payments/stk-push"
CHILDREN_ENDPOINT = "/api/v1/internal/parents/{user_id}/children"


@dataclass
class PaymentTurnResult:
    reply: str
    payment_state: PaymentState
    action_taken: str
    endpoints_read: List[str] = field(default_factory=list)


async def _execute_payment(
    pending: PendingPayment,
    message: str,
    ctx: UserContext,
    store: SessionStore,
    backend: BackendClient,
    settings: Settings,
) -> PaymentTurnResult:
    if not await store.register_payment_init(ctx.session_id):
        profile = await safe_profile(backend)
        raise SessionLimitError(
            "PAYMENT_INIT_LIMIT_EXCEEDED", prompts.lookup_limit_reply(profile)
        )

    payload = StkPushRequest(
        student_id=pending.student_id or "",
        amount=pending.amount or "",
        phone_number=pending.mpesa_phone or "",
        source="chat",
        parent_confirmed=True,
    )
    try:
        await backend.stk_push(payload)
    except BackendError as exc:
        await store.set_pending(ctx.session_id, None, PaymentState.NONE)
        if exc.status_code == 403 and exc.code == "CHAT_PAYMENT_PAUSED_PHONE_CHANGE":
            profile = await safe_profile(backend)
            return PaymentTurnResult(
                reply=prompts.payment_paused_reply(profile),
                payment_state=PaymentState.NONE,
                action_taken="PAYMENT_PAUSED_PHONE_CHANGE",
                endpoints_read=[STK_ENDPOINT],
            )
        profile = await safe_profile(backend)
        return PaymentTurnResult(
            reply=prompts.payment_failed_reply(profile),
            payment_state=PaymentState.NONE,
            action_taken="PAYMENT_FAILED",
            endpoints_read=[STK_ENDPOINT],
        )

    await store.set_pending(ctx.session_id, None, PaymentState.STK_SENT)
    return PaymentTurnResult(
        reply=prompts.payment_success_reply(
            pending.child_name or "your child", pending.amount or "", display_phone(pending.mpesa_phone or "")
        ),
        payment_state=PaymentState.STK_SENT,
        action_taken="STK_PUSH_INITIATED",
        endpoints_read=[STK_ENDPOINT],
    )


async def handle_payment_turn(
    message: str,
    student_id_hint: Optional[str],
    ctx: UserContext,
    store: SessionStore,
    backend: BackendClient,
    settings: Settings,
) -> PaymentTurnResult:
    session_id = ctx.session_id
    state = await store.get_payment_state(session_id)
    pending = await store.get_pending(session_id)

    if state == PaymentState.AWAITING_CONFIRMATION and pending and pending.is_complete():
        if is_affirmation(message):
            return await _execute_payment(pending, message, ctx, store, backend, settings)
        if is_negative(message):
            await store.set_pending(session_id, None, PaymentState.NONE)
            return PaymentTurnResult(
                reply=prompts.payment_cancelled_reply(),
                payment_state=PaymentState.NONE,
                action_taken="PAYMENT_CANCELLED",
            )
        await store.clear_pending(session_id)
        pending = None

    prior = pending if pending is not None else PendingPayment()

    amount = extract_amount(message) or prior.amount
    phone = extract_phone(message) or prior.mpesa_phone
    target_id = prior.student_id or student_id_hint

    endpoints_read: List[str] = []
    child = None
    admission = prior.admission_number
    child_name = prior.child_name

    if target_id or not (prior.student_id and amount and phone):
        child, children = await resolve_child(
            target_id, message, ctx.user_id, backend, store, session_id, settings
        )
        endpoints_read.append(CHILDREN_ENDPOINT.format(user_id=ctx.user_id))
        if child is None:
            if not children:
                profile = await safe_profile(backend)
                return PaymentTurnResult(
                    reply=prompts.out_of_scope_reply(profile),
                    payment_state=PaymentState.NONE,
                    action_taken="PAYMENT_NO_LINKED_CHILD",
                    endpoints_read=endpoints_read,
                )
            names = ", ".join(c.full_name for c in children if c.full_name)
            await store.set_pending(
                session_id,
                PendingPayment(
                    student_id=None,
                    amount=amount,
                    mpesa_phone=phone,
                    awaiting_child=True,
                ),
                PaymentState.AWAITING_SLOTS,
            )
            return PaymentTurnResult(
                reply=f"Which child would you like to pay for? {names}.",
                payment_state=PaymentState.AWAITING_SLOTS,
                action_taken="PAYMENT_AWAITING_CHILD",
                endpoints_read=endpoints_read,
            )
        target_id = child.student_id
        child_name = child.full_name or child_name
        admission = child.admission_number or admission

    updated = PendingPayment(
        student_id=target_id,
        amount=amount,
        mpesa_phone=phone,
        child_name=child_name,
        admission_number=admission,
        awaiting_child=False,
    )

    missing = []
    if not updated.student_id:
        missing.append("which child (name or admission number)")
    if not updated.amount:
        missing.append("the amount to pay")
    if not updated.mpesa_phone:
        missing.append("the M-Pesa phone number")

    if missing:
        await store.set_pending(session_id, updated, PaymentState.AWAITING_SLOTS)
        return PaymentTurnResult(
            reply=prompts.payment_ask_missing_reply(missing),
            payment_state=PaymentState.AWAITING_SLOTS,
            action_taken="PAYMENT_AWAITING_SLOTS",
            endpoints_read=endpoints_read,
        )

    await store.set_pending(session_id, updated, PaymentState.AWAITING_CONFIRMATION)
    return PaymentTurnResult(
        reply=prompts.payment_confirmation_prompt(
            updated.child_name or "your child",
            updated.admission_number or "",
            updated.amount or "",
            display_phone(updated.mpesa_phone or ""),
        ),
        payment_state=PaymentState.AWAITING_CONFIRMATION,
        action_taken="PAYMENT_AWAITING_CONFIRMATION",
        endpoints_read=endpoints_read,
    )
