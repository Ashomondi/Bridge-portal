import threading
import time
from dataclasses import dataclass
from typing import Dict, Optional

from .schemas import PaymentState


class SessionLimitError(Exception):
    def __init__(self, code: str, reply: str):
        super().__init__(reply)
        self.code = code
        self.reply = reply


@dataclass
class PendingPayment:
    student_id: Optional[str] = None
    amount: Optional[str] = None
    mpesa_phone: Optional[str] = None
    child_name: Optional[str] = None
    admission_number: Optional[str] = None

    def is_complete(self) -> bool:
        return bool(self.student_id and self.amount and self.mpesa_phone)


@dataclass
class SessionState:
    expires_at: float
    read_lookups: int = 0
    payment_inits: int = 0
    payment_state: PaymentState = PaymentState.NONE
    pending: Optional[PendingPayment] = None


class SessionStore:
    def __init__(self, ttl_seconds: int, max_reads: int, max_payments: int):
        self._ttl = ttl_seconds
        self._max_reads = max_reads
        self._max_payments = max_payments
        self._sessions: Dict[str, SessionState] = {}
        self._lock = threading.Lock()

    def _get(self, session_id: str) -> SessionState:
        now = time.time()
        state = self._sessions.get(session_id)
        if state is None or state.expires_at < now:
            state = SessionState(expires_at=now + self._ttl)
            self._sessions[session_id] = state
        return state

    def reset(self) -> None:
        with self._lock:
            self._sessions.clear()

    def register_read(self, session_id: str) -> bool:
        """Count a read lookup. Returns False if the session cap is exceeded."""
        with self._lock:
            state = self._get(session_id)
            if state.read_lookups >= self._max_reads:
                return False
            state.read_lookups += 1
            return True

    def register_payment_init(self, session_id: str) -> bool:
        """Count an STK push initiation. Returns False if the session cap is exceeded."""
        with self._lock:
            state = self._get(session_id)
            if state.payment_inits >= self._max_payments:
                return False
            state.payment_inits += 1
            return True

    def get_payment_state(self, session_id: str) -> PaymentState:
        with self._lock:
            return self._get(session_id).payment_state

    def get_pending(self, session_id: str) -> Optional[PendingPayment]:
        with self._lock:
            return self._get(session_id).pending

    def set_pending(
        self, session_id: str, pending: Optional[PendingPayment], state: PaymentState
    ) -> None:
        with self._lock:
            session = self._get(session_id)
            session.pending = pending
            session.payment_state = state

    def clear_pending(self, session_id: str) -> None:
        with self._lock:
            session = self._get(session_id)
            session.pending = None
            session.payment_state = PaymentState.NONE


_store: Optional[SessionStore] = None
_store_lock = threading.Lock()


def get_session_store() -> SessionStore:
    global _store
    if _store is None:
        with _store_lock:
            if _store is None:
                from .config import get_settings

                settings = get_settings()
                _store = SessionStore(
                    ttl_seconds=settings.session_ttl_seconds,
                    max_reads=settings.max_read_lookups,
                    max_payments=settings.max_payment_inits,
                )
    return _store


def reset_session_store() -> None:
    global _store
    with _store_lock:
        if _store is not None:
            _store.reset()
