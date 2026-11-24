import httpx
import pytest

from app.config import get_settings
from app.deps import get_backend_client, get_llm_client
from app.main import app
from app.state import InMemorySessionStore

from tests.helpers import FakeBackend, FakeLLM


@pytest.fixture
def settings():
    return get_settings()


@pytest.fixture
def fake_backend():
    return FakeBackend()


@pytest.fixture
def fake_llm():
    return FakeLLM()


@pytest.fixture
def store(settings):
    session_store = InMemorySessionStore(
        ttl_seconds=settings.session_ttl_seconds,
        max_reads=settings.max_read_lookups,
        max_payments=settings.max_payment_inits,
    )
    app.state.session_store = session_store
    return session_store


@pytest.fixture
async def client(fake_backend, fake_llm, store):
    app.dependency_overrides[get_backend_client] = lambda: fake_backend
    app.dependency_overrides[get_llm_client] = lambda: fake_llm
    transport = httpx.ASGITransport(app=app)
    async with httpx.AsyncClient(transport=transport, base_url="http://testserver") as http_client:
        yield http_client
    app.dependency_overrides.clear()
