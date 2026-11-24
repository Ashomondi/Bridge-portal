import asyncio
import json
import time
from abc import ABC, abstractmethod
from dataclasses import asdict, dataclass
from typing import Any, Dict, Optional

from fastapi import Request
from redis.asyncio import Redis

from .config import get_settings
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
    awaiting_child: bool = False

    def is_complete(self) -> bool:
        return bool(self.student_id and self.amount and self.mpesa_phone)


@dataclass
class SessionState:
    expires_at: float
    read_lookups: int = 0
    payment_inits: int = 0
    payment_state: PaymentState = PaymentState.NONE
    pending: Optional[PendingPayment] = None


class SessionStore(ABC):
    @abstractmethod
    async def register_read(self, session_id: str) -> bool:
        raise NotImplementedError

    @abstractmethod
    async def register_payment_init(self, session_id: str) -> bool:
        raise NotImplementedError

    @abstractmethod
    async def get_payment_state(self, session_id: str) -> PaymentState:
        raise NotImplementedError

    @abstractmethod
    async def get_pending(self, session_id: str) -> Optional[PendingPayment]:
        raise NotImplementedError

    @abstractmethod
    async def set_pending(
        self, session_id: str, pending: Optional[PendingPayment], state: PaymentState
    ) -> None:
        raise NotImplementedError

    @abstractmethod
    async def clear_pending(self, session_id: str) -> None:
        raise NotImplementedError

    @abstractmethod
    async def reset(self) -> None:
        raise NotImplementedError


class InMemorySessionStore(SessionStore):
    def __init__(self, ttl_seconds: int, max_reads: int, max_payments: int):
        self._ttl = ttl_seconds
        self._max_reads = max_reads
        self._max_payments = max_payments
        self._sessions: Dict[str, SessionState] = {}
        self._lock = asyncio.Lock()

    def _get(self, session_id: str) -> SessionState:
        now = time.time()
        state = self._sessions.get(session_id)
        if state is None or state.expires_at < now:
            state = SessionState(expires_at=now + self._ttl)
            self._sessions[session_id] = state
        return state

    async def register_read(self, session_id: str) -> bool:
        async with self._lock:
            state = self._get(session_id)
            if state.read_lookups >= self._max_reads:
                return False
            state.read_lookups += 1
            return True

    async def register_payment_init(self, session_id: str) -> bool:
        async with self._lock:
            state = self._get(session_id)
            if state.payment_inits >= self._max_payments:
                return False
            state.payment_inits += 1
            return True

    async def get_payment_state(self, session_id: str) -> PaymentState:
        async with self._lock:
            return self._get(session_id).payment_state

    async def get_pending(self, session_id: str) -> Optional[PendingPayment]:
        async with self._lock:
            return self._get(session_id).pending

    async def set_pending(
        self, session_id: str, pending: Optional[PendingPayment], state: PaymentState
    ) -> None:
        async with self._lock:
            session = self._get(session_id)
            session.pending = pending
            session.payment_state = state

    async def clear_pending(self, session_id: str) -> None:
        async with self._lock:
            session = self._get(session_id)
            session.pending = None
            session.payment_state = PaymentState.NONE

    async def reset(self) -> None:
        async with self._lock:
            self._sessions.clear()


class RedisSessionStore(SessionStore):
    def __init__(
        self,
        client: Redis,
        ttl_seconds: int,
        max_reads: int,
        max_payments: int,
        prefix: str = "ai:session:",
    ):
        self._client = client
        self._ttl = ttl_seconds
        self._max_reads = max_reads
        self._max_payments = max_payments
        self._prefix = prefix

    def _counter_key(self, session_id: str, kind: str) -> str:
        return f"{self._prefix}{session_id}:{kind}"

    def _state_key(self, session_id: str) -> str:
        return f"{self._prefix}{session_id}:state"

    async def _increment(self, session_id: str, kind: str, cap: int) -> bool:
        key = self._counter_key(session_id, kind)
        count = await self._client.incr(key)
        await self._client.expire(key, self._ttl)
        return count <= cap

    async def register_read(self, session_id: str) -> bool:
        return await self._increment(session_id, "reads", self._max_reads)

    async def register_payment_init(self, session_id: str) -> bool:
        return await self._increment(session_id, "payments", self._max_payments)

    async def _read_state(self, session_id: str) -> Dict[str, Any]:
        raw = await self._client.get(self._state_key(session_id))
        if not raw:
            return {
                "payment_state": PaymentState.NONE.value,
                "pending": None,
            }
        try:
            parsed = json.loads(raw)
        except (TypeError, json.JSONDecodeError):
            return {
                "payment_state": PaymentState.NONE.value,
                "pending": None,
            }
        if not isinstance(parsed, dict):
            return {
                "payment_state": PaymentState.NONE.value,
                "pending": None,
            }
        return parsed

    async def get_payment_state(self, session_id: str) -> PaymentState:
        parsed = await self._read_state(session_id)
        try:
            return PaymentState(parsed.get("payment_state", PaymentState.NONE.value))
        except ValueError:
            return PaymentState.NONE

    async def get_pending(self, session_id: str) -> Optional[PendingPayment]:
        parsed = await self._read_state(session_id)
        pending = parsed.get("pending")
        if not isinstance(pending, dict):
            return None
        return PendingPayment(**pending)

    async def _write_state(
        self, session_id: str, pending: Optional[PendingPayment], state: PaymentState
    ) -> None:
        payload = {
            "payment_state": state.value,
            "pending": asdict(pending) if pending is not None else None,
        }
        await self._client.set(
            self._state_key(session_id),
            json.dumps(payload),
            ex=self._ttl,
        )

    async def set_pending(
        self, session_id: str, pending: Optional[PendingPayment], state: PaymentState
    ) -> None:
        await self._write_state(session_id, pending, state)

    async def clear_pending(self, session_id: str) -> None:
        await self._write_state(session_id, None, PaymentState.NONE)

    async def reset(self) -> None:
        async for key in self._client.scan_iter(match=f"{self._prefix}*"):
            await self._client.delete(key)


def build_session_store(redis_client: Optional[Redis] = None) -> SessionStore:
    settings = get_settings()
    if redis_client is None and settings.ai_redis_url:
        redis_client = Redis.from_url(settings.ai_redis_url, decode_responses=True)
    if redis_client is not None:
        return RedisSessionStore(
            redis_client,
            ttl_seconds=settings.session_ttl_seconds,
            max_reads=settings.max_read_lookups,
            max_payments=settings.max_payment_inits,
            prefix=settings.ai_redis_prefix,
        )
    return InMemorySessionStore(
        ttl_seconds=settings.session_ttl_seconds,
        max_reads=settings.max_read_lookups,
        max_payments=settings.max_payment_inits,
    )


async def get_session_store(request: Request) -> SessionStore:
    store = getattr(request.app.state, "session_store", None)
    if store is None:
        store = build_session_store()
        request.app.state.session_store = store
    return store
