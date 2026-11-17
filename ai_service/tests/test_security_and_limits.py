from app.state import SessionLimitError


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
    from tests.helpers import make_headers

    headers = make_headers(settings)
    headers.pop("X-Internal-Service-Token")
    response = await client.post(
        "/api/v1/ai/parent-query", json=await _fee_payload(), headers=headers
    )
    assert response.status_code == 401


async def test_wrong_service_token_rejected(client, settings):
    from tests.helpers import make_headers

    headers = make_headers(settings, secret="wrong-token")
    response = await client.post(
        "/api/v1/ai/parent-query", json=await _fee_payload(), headers=headers
    )
    assert response.status_code == 401


async def test_valid_headers_accepted(client, settings):
    from tests.helpers import make_headers

    headers = make_headers(settings)
    response = await client.post(
        "/api/v1/ai/parent-query", json=await _fee_payload(), headers=headers
    )
    assert response.status_code == 200


async def test_tampered_context_rejected(client, settings):
    from tests.helpers import make_headers

    headers = make_headers(settings)
    raw = headers["X-User-Context"]
    headers["X-User-Context"] = raw[:-1] + ("0" if raw[-1] != "0" else "1")
    response = await client.post(
        "/api/v1/ai/parent-query", json=await _fee_payload(), headers=headers
    )
    assert response.status_code == 401


async def test_expired_context_rejected(client, settings):
    from tests.helpers import make_headers

    headers = make_headers(settings, exp=1)
    response = await client.post(
        "/api/v1/ai/parent-query", json=await _fee_payload(), headers=headers
    )
    assert response.status_code == 401


def test_read_lookup_cap(store):
    for _ in range(15):
        assert store.register_read("session-x") is True
    assert store.register_read("session-x") is False


def test_payment_initiation_cap(store):
    for _ in range(3):
        assert store.register_payment_init("session-y") is True
    assert store.register_payment_init("session-y") is False


async def test_endpoint_enforces_read_cap(client, settings):
    from tests.helpers import make_headers

    headers = make_headers(settings)
    last_status = None
    for _ in range(8):
        response = await client.post(
            "/api/v1/ai/parent-query", json=await _fee_payload(), headers=headers
        )
        last_status = response.status_code
    assert last_status == 429
