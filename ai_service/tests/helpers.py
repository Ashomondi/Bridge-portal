import base64
import hashlib
import hmac
import json
import time
from typing import Optional

from app.schemas import (
    AttendanceSummary,
    Child,
    FeeBalance,
    MarksSummary,
    SchoolProfile,
)


def make_headers(
    settings,
    channel: str = "whatsapp",
    role: str = "parent",
    session_id: str = "session-1",
    exp: Optional[int] = None,
    secret: Optional[str] = None,
    overrides: Optional[dict] = None,
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
        token.encode("utf-8"), payload_b64.encode("utf-8"), hashlib.sha256
    ).hexdigest()
    return {
        "X-Internal-Service-Token": token,
        "X-User-Context": f"{payload_b64}.{signature}",
    }


class FakeBackend:
    def __init__(self):
        self.profile = SchoolProfile(
            school_id="school-1",
            name="Ikonex High School",
            contact_phone="+254700111222",
            contact_email="office@ikonex.ac.ke",
            office_hours="Mon-Fri 8am-4pm",
            portal_url="https://portal.ikonex.ac.ke",
        )
        self.children = [
            Child(
                student_id="stu-1",
                admission_number="ADM-1024",
                full_name="Brian Otieno",
                class_name="Grade 8",
                grade_level="8",
            )
        ]
        self.fee = FeeBalance(
            student_id="stu-1",
            full_name="Brian Otieno",
            admission_number="ADM-1024",
            term_name="Term 1 2026",
            due_date="2026-01-15",
            opening_balance="0",
            billed_amount="35000.00",
            paid_amount="22500.00",
            current_balance="12500.00",
            credit_balance="0.00",
        )
        self.attendance = AttendanceSummary(
            student_id="stu-1",
            term_name="Term 1 2026",
            total_school_days=63,
            days_present=62,
            days_absent=1,
            recent_absences=["2026-02-10"],
        )
        self.marks = MarksSummary(
            student_id="stu-1",
            full_name="Brian Otieno",
            term_name="Term 1 2026",
            is_released=True,
            subjects=[
                {"subject_name": "Mathematics", "level": "EE"},
                {"subject_name": "English", "level": "ME"},
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

    async def stk_push(self, payload):
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

    async def complete_json(self, system_prompt, user_prompt):
        if self.error is not None:
            raise self.error
        return {}
