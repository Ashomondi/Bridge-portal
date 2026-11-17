from fastapi import Depends, Request

from .clients.backend_client import BackendClient
from .clients.llm_client import LLMClient
from .config import Settings, get_settings
from .security import AuthContext, require_auth


def get_backend_client(
    request: Request,
    auth: AuthContext = Depends(require_auth),
    settings: Settings = Depends(get_settings),
) -> BackendClient:
    return BackendClient(
        base_url=settings.go_backend_internal_url,
        service_token=settings.ai_service_secret_token,
        user_context=auth.raw_user_context,
        client=request.app.state.http_client,
        timeout=settings.request_timeout_seconds,
    )


def get_llm_client(
    request: Request,
    settings: Settings = Depends(get_settings),
) -> LLMClient:
    return LLMClient(
        base_url=settings.mistral_api_url,
        model=settings.mistral_model_name,
        client=request.app.state.http_client,
        timeout=settings.llm_timeout_seconds,
    )
