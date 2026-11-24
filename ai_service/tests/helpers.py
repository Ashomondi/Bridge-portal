import base64
import hashlib
import hmac
import json
import time
from typing import Optional

from app.clients.backend_client import BackendError
from app.clients.llm_client import LLMError
from app.schemas import (
    AttendanceSummary,
    Child,
    FeeBalance,
    MarksSummary,
    SchoolProfile,
    StkPushRequest,
)


def make_headers(
    settings,
    channel: str = "whatsapp",
    role: str = "parent",
    session_id: str = "session-1",
    exp: Optional[int] = None,
    secret: Optional[str] = None,
    overrides: Optional[dict] = None,
    delegated_session: Optional[str] = "delegated-session-token",
    use_bearer: bool = False,
) -> dict:
    payload = {
        "user_id": "user-1",
        "school_id": "school-1",
        "membership_id": "membership-1",
        "role": role,
        "channel": channel,
        "session_id": session_id,
        "exp": exp if exp is not None else int(time.time()) + 3600,
    }
    if overrides:
        payload.update(overrides)
    payload_b64 = (
        base64.urlsafe_b64encode(json.dumps(payload).encode("utf-8"))
        .rstrip(b"=")
        .decode("ascii")
    )
    token = secret or settings.ai_service_secret_token
    signature = hmac.new(
        settings.context_signing_key.encode("utf-8"),
        payload_b64.encode("utf-8"),
        hashlib.sha256,
    ).hexdigest()
    headers = {
        "X-User-Context": f"{payload_b64}.{signature}",
    }
    if use_bearer:
        headers["Authorization"] = f"Bearer {token}"
    else:
        headers["X-Internal-Service-Token"] = token
    if delegated_session:
        headers["X-Delegated-Session"] = delegated_session
    return headers


class FakeBackend:
    def __init__(self):
        self.profile = SchoolProfile(
            school_id="school-1",
            name="Ikonex High School",
            contact_phone="+254700111222",
            contact_email="office@ikonex.ac.ke",
            office_hours="Mon-Fri 8am-4pm",
        )
        self.children = [
            Child(
                student_id="stu-1",
                admission_number="ADM-1024",
                full_name="Brian Otieno",
                grade_name="Grade 8",
                stream_name="East",
                relationship="mother",
            )
        ]
        self.fee = FeeBalance(
            student_id="stu-1",
            student_name="Brian Otieno",
            admission_number="ADM-1024",
            due_date="2026-01-15",
            opening_balance="0.00",
            billed_amount="35000.00",
            paid_amount="22500.00",
            current_balance="12500.00",
            credit_balance="0.00",
        )
        self.attendance = AttendanceSummary(
            student_id="stu-1",
            term_id="term-1",
            present_days=62,
            absent_days=1,
            late_days=0,
            total_recorded=63,
            last_roll_call_day="2026-02-10",
        )
        self.marks = MarksSummary(
            student_id="stu-1",
            term_id="term-1",
            released=True,
            subjects=[
                {"subject": "Mathematics", "performance_level": "EE"},
                {"subject": "English", "performance_level": "ME"},
            ],
        )
        self.audit_calls = []
        self.stk_calls = []
        self.children_calls = 0
        self.marks_calls = 0
        self.stk_error = None

    async def get_school_profile(self):
        return self.profile

    async def get_children(self, user_id):
        self.children_calls += 1
        return self.children

    async def get_fee_balance(self, student_id):
        return self.fee

    async def get_attendance_summary(self, student_id):
        return self.attendance

    async def get_marks_summary(self, student_id, term_id=None):
        self.marks_calls += 1
        return self.marks

    async def stk_push(self, payload: StkPushRequest):
        self.stk_calls.append(payload)
        if self.stk_error is not None:
            raise self.stk_error
        return {"status": "success"}

    async def audit_log(self, payload):
        self.audit_calls.append(payload)


class FakeLLM:
    def __init__(self):
        self.text = ""
        self.error = None

    async def complete_text(self, system_prompt, user_prompt):
        if self.error is not None:
            raise self.error
        return self.text


__all__ = ["BackendError", "LLMError", "FakeBackend", "FakeLLM", "make_headers"]
