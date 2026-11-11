from fastapi import APIRouter, Depends, HTTPException

from .. import prompts
from ..clients.backend_client import BackendClient, BackendError
from ..clients.llm_client import LLMClient, LLMError
from ..config import Settings, get_settings
from ..deps import get_backend_client, get_llm_client
from ..schemas import (
    AuditLogRequest,
    ReportDraftData,
    ReportDraftRequest,
    ReportDraftResponse,
    Role,
)
from ..security import AuthContext, require_auth
from ..utils import sanitize_forbidden_terms

router = APIRouter()

ALLOWED_ROLES = {Role.teacher, Role.dean_of_studies, Role.principal, Role.deputy}


def _subjects_summary(marks) -> str:
    lines = []
    for subject in marks.subjects or []:
        name = subject.get("subject_name") or subject.get("name") or ""
        level = subject.get("level") or subject.get("performance_level") or ""
        if name:
            lines.append(f"- {name}: {level}".rstrip(": "))
    return "\n".join(lines) if lines else "- (no subject levels provided)"


@router.post("/report-comments/draft", response_model=ReportDraftResponse)
async def draft_report_comment(
    req: ReportDraftRequest,
    auth: AuthContext = Depends(require_auth),
    backend: BackendClient = Depends(get_backend_client),
    llm: LLMClient = Depends(get_llm_client),
    settings: Settings = Depends(get_settings),
):
    ctx = auth.user
    if ctx.role not in ALLOWED_ROLES:
        raise HTTPException(
            status_code=403,
            detail={"code": "FORBIDDEN_ROLE", "message": "Role may not draft report comments"},
        )
    if req.mode not in ("term_comment", "rerelease_note"):
        raise HTTPException(
            status_code=400,
            detail={"code": "INVALID_MODE", "message": "Unknown drafting mode"},
        )
    if req.mode == "rerelease_note" and not (req.teacher_reason or "").strip():
        raise HTTPException(
            status_code=400,
            detail={"code": "TEACHER_REASON_REQUIRED", "message": "teacher_reason is required"},
        )

    endpoints_read = []
    fallback_used = False
    draft = ""

    if req.mode == "term_comment":
        marks = None
        attendance = None
        try:
            marks = await backend.get_marks_summary(req.student_id, req.term_id)
            endpoints_read.append(f"/api/v1/internal/students/{req.student_id}/marks-summary")
        except BackendError:
            marks = None
        try:
            attendance = await backend.get_attendance_summary(req.student_id)
            endpoints_read.append(
                f"/api/v1/internal/students/{req.student_id}/attendance-summary"
            )
        except BackendError:
            attendance = None

        child_name = getattr(marks, "full_name", "") if marks else ""
        fallback = prompts.term_comment_fallback(marks, attendance, child_name)
        user_prompt = (
            f"Student name: {child_name or 'the learner'}.\n"
            f"Term: {getattr(marks, 'term_name', '') if marks else ''}.\n"
            "Subject performance levels:\n"
            f"{_subjects_summary(marks) if marks else '- (none)'}\n"
        )
        if attendance:
            user_prompt += (
                f"Attendance: {attendance.days_present} of {attendance.total_school_days} "
                f"school days. {attendance.days_absent} absent.\n"
            )
        user_prompt += "Draft the report card comment now."
        try:
            draft = sanitize_forbidden_terms(
                await llm.complete_text(prompts.REPORT_TERM_SYSTEM_PROMPT, user_prompt)
            )
        except LLMError:
            draft = ""
        if not draft:
            draft = sanitize_forbidden_terms(fallback)
            fallback_used = True
    else:
        fallback = prompts.rerelease_note_fallback(req.teacher_reason, req.subject_name)
        user_prompt = (
            f"Subject: {req.subject_name or 'the subject'}.\n"
            f"Teacher's correction reason: {req.teacher_reason}.\n"
            "Draft the re-release note now."
        )
        try:
            draft = sanitize_forbidden_terms(
                await llm.complete_text(prompts.REPORT_RERELEASE_SYSTEM_PROMPT, user_prompt)
            )
        except LLMError:
            draft = ""
        if not draft:
            draft = sanitize_forbidden_terms(fallback)
            fallback_used = True

    if not draft:
        draft = prompts.rerelease_note_fallback(req.teacher_reason, req.subject_name)

    await backend.audit_log(
        AuditLogRequest(
            feature="report_drafter",
            session_id=ctx.session_id,
            channel=ctx.channel,
            intent="REPORT_COMMENT_DRAFT",
            endpoints_read=endpoints_read,
            action_taken="DRAFT_REPORT_COMMENT",
            metadata={"mode": req.mode, "fallback_used": fallback_used},
        )
    )

    return ReportDraftResponse(
        data=ReportDraftData(
            draft_comment=draft,
            requires_teacher_approval=True,
            model_used=settings.mistral_model_name,
            fallback_used=fallback_used,
        )
    )
