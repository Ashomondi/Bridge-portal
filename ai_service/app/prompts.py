from decimal import Decimal, InvalidOperation
from typing import Iterable, Optional

from .schemas import AttendanceSummary, FeeBalance, SchoolProfile

PROMPT_VERSION = "r1-1"

ASSISTANT_SYSTEM_PROMPT = (
    "You are the Ikonex Academy parent assistant. You answer only about school fees "
    "and attendance using the exact facts provided to you. Never invent or alter any "
    "number, fee balance, date or name. Be brief, polite and clear. Reply in English. "
    "The current year is 2026."
)

REPORT_TERM_SYSTEM_PROMPT = (
    "You are an experienced CBC class teacher in Kenya drafting a report card comment. "
    "Write 2 to 3 constructive, encouraging sentences for one learner based only on the "
    "provided performance levels (EE, ME, AE, BE), trend and attendance. Never rank the "
    "learner, never compare them to classmates, never give positions or medical labels. "
    "Reply in English with the comment text only."
)

REPORT_RERELEASE_SYSTEM_PROMPT = (
    "You are a Kenyan class teacher writing a short, professional note to explain a "
    "corrected subject mark before results are re-released. Be factual and apologetic, "
    "avoid any learner ranking or medical language, and reply in English with the note only."
)

FORBIDDEN_TERMS = [
    r"\brank(?:ed|ing|s)?\b",
    r"\bposition\b",
    r"\b1st\b",
    r"\bfirst\b",
    r"\blast\b",
    r"\btop\b",
    r"\bbottom\b",
    r"\bbeat\b",
    r"\bbest\b",
    r"\bworst\b",
    r"\bclassmates?\b",
    r"\badhd\b",
    r"\bdyslexi(?:a|c)\b",
    r"\bautis(?:m|tic)\b",
    r"\bdiagnos(?:ed|is|tic)\b",
    r"\bdisorder\b",
    r"\bclinical\b",
]


def _to_decimal(value) -> Optional[Decimal]:
    if value is None:
        return None
    try:
        return Decimal(str(value).replace(",", "").strip())
    except (InvalidOperation, ValueError):
        return None


def format_kes(value) -> str:
    amount = _to_decimal(value)
    if amount is None:
        return f"KES {value}"
    return f"KES {amount:,.2f}"


def allowed_amount_strings(values: Iterable) -> set:
    allowed = set()
    for value in values:
        amount = _to_decimal(value)
        if amount is None:
            continue
        allowed.add(f"{amount:.2f}")
        allowed.add(f"{amount:,.2f}")
        allowed.add(str(int(amount)) if amount == amount.to_integral_value() else f"{amount}")
    return allowed


def fee_reply(fee: FeeBalance, child_name: str, admission: str) -> str:
    name = child_name or fee.full_name or "your child"
    adm = admission or fee.admission_number
    adm_text = f" (ADM: {adm})" if adm else ""
    due = fee.due_date or "not set"
    return (
        f"The current fee balance for {name}{adm_text} for {fee.term_name or 'this term'} "
        f"is {format_kes(fee.current_balance)} "
        f"(Billed: {format_kes(fee.billed_amount)}, Paid: {format_kes(fee.paid_amount)}, "
        f"Credit: {format_kes(fee.credit_balance)}). Due date: {due}. "
        "Would you like to make a payment via M-Pesa?"
    )


def attendance_reply(attendance: AttendanceSummary, child_name: str) -> str:
    name = child_name or "your child"
    return (
        f"For {attendance.term_name or 'the current term'}, {name} was present on "
        f"{attendance.days_present} of {attendance.total_school_days} school days "
        f"({attendance.days_absent} absent)."
    )


def results_blocked_reply(portal_link: str) -> str:
    return (
        "For privacy and security, exam results, grades and teacher comments are not "
        "shared on WhatsApp. Please open the Parent Portal to view them: "
        f"{portal_link}"
    )


def results_not_released_reply() -> str:
    return (
        "This term's results have not been released yet. Please check back on the Parent "
        "Portal once your child's class teacher releases them."
    )


def _contact_block(profile: Optional[SchoolProfile]) -> str:
    if profile is None:
        return "Please contact the school office for further help."
    parts = []
    if profile.name:
        parts.append(profile.name)
    if profile.contact_phone:
        parts.append(f"Phone: {profile.contact_phone}")
    if profile.contact_email:
        parts.append(f"Email: {profile.contact_email}")
    if profile.office_hours:
        parts.append(f"Office hours: {profile.office_hours}")
    if not parts:
        return "Please contact the school office for further help."
    return "Please contact the school: " + ", ".join(parts) + "."


def out_of_scope_reply(profile: Optional[SchoolProfile]) -> str:
    return (
        "I'm sorry, I can only help with school fee balances and attendance questions. "
        + _contact_block(profile)
    )


def lookup_limit_reply(profile: Optional[SchoolProfile]) -> str:
    return (
        "You have reached the limit of lookups for this session. Please wait a while "
        "before trying again, or " + _contact_block(profile).lower()
    )


def payment_confirmation_prompt(
    child_name: str, admission: str, amount: str, phone: str
) -> str:
    adm_text = f" (Adm: {admission})" if admission else ""
    return (
        f"Please confirm: You are about to initiate an M-Pesa STK push of "
        f"{format_kes(amount)} for {child_name}{adm_text} to M-Pesa number {phone}. "
        "Reply YES to confirm or NO to cancel."
    )


def payment_ask_missing_reply(missing: list) -> str:
    return "To continue with the M-Pesa payment, please provide: " + ", ".join(missing) + "."


def payment_success_reply(child_name: str, amount: str, phone: str) -> str:
    return (
        f"An M-Pesa payment request of {format_kes(amount)} for {child_name} has been "
        f"sent to {phone}. Please enter your M-Pesa PIN on your phone to complete it."
    )


def payment_cancelled_reply() -> str:
    return "No problem, the payment has been cancelled. Nothing was charged."


def payment_paused_reply(profile: Optional[SchoolProfile]) -> str:
    return (
        "Chat payments are paused for 24 hours after a registered phone number change, "
        "for security. " + _contact_block(profile)
    )


def payment_failed_reply(profile: Optional[SchoolProfile]) -> str:
    return (
        "Sorry, we could not start the M-Pesa payment right now. "
        + _contact_block(profile)
    )


def term_comment_fallback(marks, attendance: Optional[AttendanceSummary], child_name: str) -> str:
    name = child_name or "The learner"
    levels = []
    subjects = getattr(marks, "subjects", []) or []
    for subject in subjects:
        subject_name = subject.get("subject_name") or subject.get("name") or ""
        level = subject.get("level") or subject.get("performance_level") or ""
        if subject_name and level:
            levels.append(f"{level} in {subject_name}")
    performance = (
        f" demonstrated {', '.join(levels)}" if levels else " shown steady engagement"
    )
    attendance_text = ""
    if attendance and attendance.total_school_days:
        attendance_text = (
            f" Attendance was {attendance.days_present} of {attendance.total_school_days} "
            "school days."
        )
    return (
        f"{name}{performance} this term.{attendance_text} Continued focus and practice "
        "will support further progress next term."
    )


def rerelease_note_fallback(reason: Optional[str], subject_name: Optional[str]) -> str:
    subject = subject_name or "a subject"
    reason_text = reason.strip().rstrip(".") if reason else "a mark was reviewed"
    return (
        f"Please note that results for {subject} have been reviewed and corrected: "
        f"{reason_text}. The updated results have been re-released. We apologise for any "
        "inconvenience and appreciate your understanding."
    )
