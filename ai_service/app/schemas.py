from enum import Enum
from typing import Any, Dict, List, Optional, Union

from pydantic import BaseModel, ConfigDict, Field


class Channel(str, Enum):
    in_app = "in_app"
    whatsapp = "whatsapp"


class Role(str, Enum):
    parent = "parent"
    teacher = "teacher"
    dean_of_studies = "dean_of_studies"
    principal = "principal"
    deputy = "deputy"


class Intent(str, Enum):
    FEE_INQUIRY = "FEE_INQUIRY"
    ATTENDANCE_INQUIRY = "ATTENDANCE_INQUIRY"
    RESULTS_INQUIRY = "RESULTS_INQUIRY"
    PAYMENT_FLOW = "PAYMENT_FLOW"
    OUT_OF_SCOPE = "OUT_OF_SCOPE"
    UNSUPPORTED = "UNSUPPORTED"


class PaymentState(str, Enum):
    NONE = "NONE"
    AWAITING_SLOTS = "AWAITING_SLOTS"
    AWAITING_CONFIRMATION = "AWAITING_CONFIRMATION"
    STK_SENT = "STK_SENT"


class UserContext(BaseModel):
    user_id: str
    school_id: str
    membership_id: str
    role: Role
    channel: Channel
    session_id: str
    exp: int


class ParentQueryRequest(BaseModel):
    session_id: str
    channel: Channel
    message: str
    student_id: Optional[str] = None


class ParentQueryData(BaseModel):
    reply: str
    intent: Intent
    channel: Channel
    payment_state: PaymentState
    action_taken: str
    portal_link: Optional[str] = None


class ParentQueryResponse(BaseModel):
    status: str = "success"
    data: ParentQueryData


class ReportDraftRequest(BaseModel):
    mode: str
    student_id: str
    term_id: str
    teacher_reason: Optional[str] = None
    subject_name: Optional[str] = None


class ReportDraftData(BaseModel):
    draft_comment: str
    requires_teacher_approval: bool = True
    model_used: str
    fallback_used: bool


class ReportDraftResponse(BaseModel):
    status: str = "success"
    data: ReportDraftData


Amount = Union[str, int, float]


class SchoolProfile(BaseModel):
    model_config = ConfigDict(extra="allow")

    school_id: str = ""
    name: str = ""
    contact_phone: str = ""
    contact_email: str = ""
    office_hours: str = ""
    portal_url: Optional[str] = None


class Child(BaseModel):
    model_config = ConfigDict(extra="allow")

    student_id: str
    admission_number: str = ""
    full_name: str = ""
    class_name: str = ""
    grade_level: str = ""


class FeeBalance(BaseModel):
    model_config = ConfigDict(extra="allow")

    student_id: str = ""
    full_name: str = ""
    admission_number: str = ""
    term_name: str = ""
    due_date: Optional[str] = None
    opening_balance: Amount = "0"
    billed_amount: Amount = "0"
    paid_amount: Amount = "0"
    current_balance: Amount = "0"
    credit_balance: Amount = "0"
    chat_payment_blocked_until: Optional[str] = None


class AttendanceSummary(BaseModel):
    model_config = ConfigDict(extra="allow")

    student_id: str = ""
    term_name: str = ""
    total_school_days: int = 0
    days_present: int = 0
    days_absent: int = 0
    recent_absences: List[str] = Field(default_factory=list)


class MarksSummary(BaseModel):
    model_config = ConfigDict(extra="allow")

    student_id: str = ""
    full_name: str = ""
    term_name: str = ""
    is_released: bool = False
    subjects: List[Dict[str, Any]] = Field(default_factory=list)
    previous_term_levels: Dict[str, str] = Field(default_factory=dict)


class StkPushRequest(BaseModel):
    student_id: str
    amount: str
    mpesa_phone: str
    session_id: str
    parent_confirmed: bool
    confirmation_message_text: str


class AuditLogRequest(BaseModel):
    feature: str
    session_id: str
    channel: Channel
    intent: Optional[str] = None
    endpoints_read: List[str] = Field(default_factory=list)
    action_taken: str = ""
    metadata: Dict[str, Any] = Field(default_factory=dict)
