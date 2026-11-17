import base64
import hashlib
import hmac
import json
import time
from dataclasses import dataclass

from fastapi import Depends, Header, HTTPException

from .config import Settings, get_settings
from .schemas import UserContext


@dataclass
class AuthContext:
    user: UserContext
    raw_user_context: str


def _unauthorized(code: str, message: str = "Unauthorized") -> HTTPException:
    return HTTPException(status_code=401, detail={"code": code, "message": message})


def _b64url_decode(data: str) -> bytes:
    padding = "=" * (-len(data) % 4)
    return base64.urlsafe_b64decode(data + padding)


def verify_service_token(token: str | None, settings: Settings) -> None:
    expected = settings.ai_service_secret_token
    if not token or not hmac.compare_digest(token, expected):
        raise _unauthorized("INVALID_SERVICE_TOKEN", "Invalid internal service token")


def verify_user_context(raw: str | None, settings: Settings) -> UserContext:
    if not raw or "." not in raw:
        raise _unauthorized("INVALID_USER_CONTEXT", "Missing or malformed user context")

    payload_b64, signature = raw.rsplit(".", 1)
    expected_sig = hmac.new(
        settings.ai_service_secret_token.encode("utf-8"),
        payload_b64.encode("utf-8"),
        hashlib.sha256,
    ).hexdigest()
    if not hmac.compare_digest(signature, expected_sig):
        raise _unauthorized("INVALID_USER_CONTEXT_SIGNATURE", "Bad user context signature")

    try:
        payload = json.loads(_b64url_decode(payload_b64))
    except Exception as exc:  # noqa: BLE001
        raise _unauthorized("INVALID_USER_CONTEXT", "Undecodable user context") from exc

    if not isinstance(payload, dict):
        raise _unauthorized("INVALID_USER_CONTEXT", "Undecodable user context")

    if int(payload.get("exp", 0)) < int(time.time()):
        raise _unauthorized("USER_CONTEXT_EXPIRED", "User context has expired")

    try:
        return UserContext(**payload)
    except Exception as exc:  # noqa: BLE001
        raise _unauthorized("INVALID_USER_CONTEXT", "Invalid user context payload") from exc


async def require_auth(
    x_internal_service_token: str | None = Header(default=None),
    x_user_context: str | None = Header(default=None),
    settings: Settings = Depends(get_settings),
) -> AuthContext:
    verify_service_token(x_internal_service_token, settings)
    user = verify_user_context(x_user_context, settings)
    return AuthContext(user=user, raw_user_context=x_user_context or "")
