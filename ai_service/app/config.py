from functools import lru_cache

from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    model_config = SettingsConfigDict(env_file=".env", extra="ignore")

    ai_service_host: str = "0.0.0.0"
    ai_service_port: int = 8000
    go_backend_internal_url: str = "http://localhost:8080/api/v1/internal"
    mistral_api_url: str = "http://localhost:11434/v1"
    mistral_model_name: str = "mistral-7b-instruct"
    ai_service_secret_token: str = "dev_internal_secret_token"

    request_timeout_seconds: float = 10.0
    llm_timeout_seconds: float = 30.0

    session_ttl_seconds: int = 6 * 60 * 60
    max_read_lookups: int = 15
    max_payment_inits: int = 3

    portal_results_url: str = "https://portal.ikonex.ac.ke/parent/results"


@lru_cache
def get_settings() -> Settings:
    return Settings()
