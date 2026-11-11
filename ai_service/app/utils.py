import re
from decimal import Decimal, InvalidOperation
from typing import List, Optional, Set, Tuple

from .clients.backend_client import BackendClient
from .config import Settings
from .prompts import FORBIDDEN_TERMS, allowed_amount_strings, lookup_limit_reply
from .schemas import Child, Intent, SchoolProfile
from .state import SessionLimitError, SessionStore

AFFIRMATIONS = {"yes", "y", "confirm", "proceed"}
NEGATIVES = {"no", "n", "cancel"}

PHONE_RE = re.compile(r"(?:\+?254|0)\d{9}")
KSH_RE = re.compile(r"(?:kes|ksh|sh)\s*([\d,]+(?:\.\d{1,2})?)", re.IGNORECASE)
_AMOUNT_RE = re.compile(r"(\d[\d,]*(?:\.\d{1,2})?)")
_FORBIDDEN_RES = [re.compile(pattern, re.IGNORECASE) for pattern in FORBIDDEN_TERMS]


def is_affirmation(text: str) -> bool:
    return text.strip().rstrip(".!").lower() in AFFIRMATIONS


def is_negative(text: str) -> bool:
    return text.strip().rstrip(".!").lower() in NEGATIVES


def classify_intent(message: str) -> Intent:
    text = message.lower()
    if re.search(r"\b(pay|paying|mpesa|m-?pesa|stk)\b", text):
        return Intent.PAYMENT_FLOW
    if re.search(r"\b(results?|grades?|marks?|exams?|performance|report|report card|cbc)\b", text):
        return Intent.RESULTS_INQUIRY
    if re.search(r"\b(attendance|present|absent|attend)\b", text):
        return Intent.ATTENDANCE_INQUIRY
    if re.search(r"\b(fees?|balance|owe|billed|paid)\b", text):
        return Intent.FEE_INQUIRY
    return Intent.OUT_OF_SCOPE


def normalize_phone(raw: str) -> str:
    digits = re.sub(r"\D", "", raw)
    if digits.startswith("254"):
        return "+" + digits
    if digits.startswith("0"):
        return "+254" + digits[1:]
    return "+" + digits


def display_phone(phone: str) -> str:
    digits = re.sub(r"\D", "", phone or "")
    if digits.startswith("254"):
        return "0" + digits[3:]
    return phone


def extract_phone(text: str) -> Optional[str]:
    match = PHONE_RE.search(text.replace(" ", ""))
    if not match:
        return None
    return normalize_phone(match.group(0))


def quantize_amount(value: str) -> Optional[str]:
    try:
        amount = Decimal(str(value).replace(",", ""))
    except (InvalidOperation, ValueError):
        return None
    if amount <= 0:
        return None
    return f"{amount.quantize(Decimal('0.01'))}"


def extract_amount(text: str) -> Optional[str]:
    without_phone = PHONE_RE.sub(" ", text.replace(" ", ""))
    match = _AMOUNT_RE.search(without_phone)
    if not match:
        return None
    return quantize_amount(match.group(1))


def extract_stated_amounts(text: str) -> Set[str]:
    stated = set()
    for match in KSH_RE.finditer(text):
        amount = quantize_amount(match.group(1))
        if amount:
            stated.add(amount)
    return stated


def amounts_match(text: str, values: List) -> bool:
    allowed = allowed_amount_strings(values)
    return extract_stated_amounts(text).issubset(allowed)


def counts_present(text: str, counts: List[int]) -> bool:
    digits = text.replace(",", "")
    return all(str(count) in digits for count in counts)


def match_child(children: List[Child], message: str) -> Optional[Child]:
    text = message.lower()
    named = [c for c in children if c.full_name and c.full_name.lower() in text]
    if named:
        return named[0]
    tokens = set(re.findall(r"[a-z]+", text))
    first_name_matches = [
        c for c in children if c.full_name and c.full_name.lower().split()[0] in tokens
    ]
    if len(first_name_matches) == 1:
        return first_name_matches[0]
    return None


def has_forbidden_terms(text: str) -> bool:
    return any(regex.search(text) for regex in _FORBIDDEN_RES)


def sanitize_forbidden_terms(text: str) -> str:
    sentences = re.split(r"(?<=[.!?])\s+", text.strip())
    kept = [s for s in sentences if s and not has_forbidden_terms(s)]
    return " ".join(kept).strip()


async def safe_profile(backend: BackendClient) -> Optional[SchoolProfile]:
    try:
        return await backend.get_school_profile()
    except Exception:  # noqa: BLE001
        return None


async def gate_read(
    store: SessionStore,
    session_id: str,
    backend: BackendClient,
    settings: Settings,
) -> None:
    if store.register_read(session_id):
        return
    profile = await safe_profile(backend)
    raise SessionLimitError("SESSION_LOOKUP_LIMIT_EXCEEDED", lookup_limit_reply(profile))


async def read_school_profile(
    backend: BackendClient,
    store: SessionStore,
    session_id: str,
    settings: Settings,
) -> SchoolProfile:
    await gate_read(store, session_id, backend, settings)
    return await backend.get_school_profile()


async def load_children(
    backend: BackendClient,
    store: SessionStore,
    session_id: str,
    user_id: str,
    settings: Settings,
) -> List[Child]:
    await gate_read(store, session_id, backend, settings)
    return await backend.get_children(user_id)


async def resolve_child(
    student_id_hint: Optional[str],
    message: str,
    user_id: str,
    backend: BackendClient,
    store: SessionStore,
    session_id: str,
    settings: Settings,
) -> Tuple[Optional[Child], List[Child]]:
    children = await load_children(backend, store, session_id, user_id, settings)
    if student_id_hint:
        for child in children:
            if child.student_id == student_id_hint:
                return child, children
        if len(children) == 1:
            return children[0], children
        return None, children
    if len(children) == 1:
        return children[0], children
    return match_child(children, message), children
