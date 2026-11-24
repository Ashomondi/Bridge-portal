from app.state import RedisSessionStore
from tests.helpers import make_headers


async def _fee_payload():
    return {
        "session_id": "session-1",
        "channel": "whatsapp",
        "message": "How much school fees do I owe for Brian?",
    }


async def test_healthz_is_open(client):
    response = await client.get("/healthz")
    assert response.status_code == 200
    assert response.json()["status"] == "ok"


async def test_missing_service_token_rejected(client, settings):
    headers = make_headers(settings)
    headers.pop("X-Internal-Service-Token")
    response = await client.post(
        "/api/v1/ai/parent-query", json=await _fee_payload(), headers=headers
    )
    assert response.status_code == 401


async def test_wrong_service_token_rejected(client, settings):
    headers = make_headers(settings, secret="wrong-token")
    response = await client.post(
        "/api/v1/ai/parent-query", json=await _fee_payload(), headers=headers
    )
    assert response.status_code == 401


async def test_bearer_token_accepted(client, settings):
    headers = make_headers(settings, use_bearer=True)
    response = await client.post(
        "/api/v1/ai/parent-query", json=await _fee_payload(), headers=headers
    )
    assert response.status_code == 200


async def test_valid_headers_accepted(client, settings):
    headers = make_headers(settings)
    response = await client.post(
        "/api/v1/ai/parent-query", json=await _fee_payload(), headers=headers
    )
    assert response.status_code == 200


async def test_tampered_context_rejected(client, settings):
    headers = make_headers(settings)
    raw = headers["X-User-Context"]
    headers["X-User-Context"] = raw[:-1] + ("0" if raw[-1] != "0" else "1")
    response = await client.post(
        "/api/v1/ai/parent-query", json=await _fee_payload(), headers=headers
    )
    assert response.status_code == 401


async def test_expired_context_rejected(client, settings):
    headers = make_headers(settings, exp=1)
    response = await client.post(
        "/api/v1/ai/parent-query", json=await _fee_payload(), headers=headers
    )
    assert response.status_code == 401


async def test_read_lookup_cap(store):
    for _ in range(15):
        assert await store.register_read("session-x") is True
    assert await store.register_read("session-x") is False


async def test_payment_initiation_cap(store):
    for _ in range(3):
        assert await store.register_payment_init("session-y") is True
    assert await store.register_payment_init("session-y") is False


async def test_endpoint_enforces_read_cap(client, settings):
    headers = make_headers(settings)
    last_status = None
    for _ in range(8):
        response = await client.post(
            "/api/v1/ai/parent-query", json=await _fee_payload(), headers=headers
        )
        last_status = response.status_code
    assert last_status == 429


class FakeRedis:
    def __init__(self):
        self.values = {}
        self.ttls = {}

    async def incr(self, key):
        self.values[key] = int(self.values.get(key, 0)) + 1
        return self.values[key]

    async def expire(self, key, ttl):
        self.ttls[key] = ttl

    async def get(self, key):
        return self.values.get(key)

    async def set(self, key, value, ex=None):
        self.values[key] = value
        if ex is not None:
            self.ttls[key] = ex

    async def scan_iter(self, match=None):
        for key in list(self.values):
            yield key

    async def delete(self, key):
        self.values.pop(key, None)
        self.ttls.pop(key, None)


async def test_redis_store_enforces_caps_and_state():
    client = FakeRedis()
    store = RedisSessionStore(client, ttl_seconds=60, max_reads=2, max_payments=1, prefix="t:")

    assert await store.register_read("s1") is True
    assert await store.register_read("s1") is True
    assert await store.register_read("s1") is False
    assert await store.register_payment_init("s1") is True
    assert await store.register_payment_init("s1") is False

    await store.clear_pending("s1")
    assert (await store.get_payment_state("s1")).value == "NONE"
    assert client.ttls["t:s1:state"] == 60
