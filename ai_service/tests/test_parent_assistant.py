from app.clients.llm_client import LLMError

from tests.helpers import make_headers


def _payload(message, channel="in_app"):
    return {"session_id": "session-1", "channel": channel, "message": message}


async def test_fee_inquiry_in_app_returns_exact_figures(client, settings):
    response = await client.post(
        "/api/v1/ai/parent-query",
        json=_payload("How much school fees do I owe for Brian?"),
        headers=make_headers(settings, channel="in_app"),
    )
    assert response.status_code == 200
    data = response.json()["data"]
    assert data["intent"] == "FEE_INQUIRY"
    assert "12,500.00" in data["reply"]
    assert "35,000.00" in data["reply"]


async def test_fee_inquiry_whatsapp_returns_exact_figures(client, settings):
    response = await client.post(
        "/api/v1/ai/parent-query",
        json=_payload("How much school fees do I owe for Brian?", channel="whatsapp"),
        headers=make_headers(settings, channel="whatsapp"),
    )
    assert response.status_code == 200
    assert "12,500.00" in response.json()["data"]["reply"]


async def test_attendance_returns_exact_counts(client, settings):
    response = await client.post(
        "/api/v1/ai/parent-query",
        json=_payload("What is Brian's attendance this term?"),
        headers=make_headers(settings, channel="in_app"),
    )
    assert response.status_code == 200
    data = response.json()["data"]
    assert data["intent"] == "ATTENDANCE_INQUIRY"
    assert "62" in data["reply"]
    assert "63" in data["reply"]


async def test_whatsapp_results_are_blocked(client, settings, fake_backend):
    response = await client.post(
        "/api/v1/ai/parent-query",
        json=_payload("Show me Brian's exam results and grades", channel="whatsapp"),
        headers=make_headers(settings, channel="whatsapp"),
    )
    assert response.status_code == 200
    data = response.json()["data"]
    assert data["intent"] == "RESULTS_INQUIRY"
    assert data["portal_link"]
    assert "portal" in data["reply"].lower()
    assert fake_backend.marks_calls == 0


async def test_out_of_scope_refusal_includes_contacts(client, settings):
    response = await client.post(
        "/api/v1/ai/parent-query",
        json=_payload("What is the school transport route?"),
        headers=make_headers(settings, channel="in_app"),
    )
    assert response.status_code == 200
    reply = response.json()["data"]["reply"]
    assert "+254700111222" in reply
    assert "office@ikonex.ac.ke" in reply


async def test_llm_hallucination_falls_back_to_template(client, settings, fake_llm):
    fake_llm.text = "The fee balance is KES 9,999.00 only."
    response = await client.post(
        "/api/v1/ai/parent-query",
        json=_payload("How much fees do I owe for Brian?"),
        headers=make_headers(settings, channel="in_app"),
    )
    assert response.status_code == 200
    reply = response.json()["data"]["reply"]
    assert "12,500.00" in reply
    assert "9,999" not in reply


async def test_llm_timeout_falls_back_to_template(client, settings, fake_llm):
    fake_llm.error = LLMError("connect timeout")
    response = await client.post(
        "/api/v1/ai/parent-query",
        json=_payload("How much fees do I owe for Brian?"),
        headers=make_headers(settings, channel="in_app"),
    )
    assert response.status_code == 200
    assert "12,500.00" in response.json()["data"]["reply"]


async def test_audit_log_written(client, settings, fake_backend):
    await client.post(
        "/api/v1/ai/parent-query",
        json=_payload("How much fees do I owe for Brian?"),
        headers=make_headers(settings, channel="in_app"),
    )
    assert fake_backend.audit_calls
    assert fake_backend.audit_calls[0].feature == "parent_assistant"
