import pytest

from app.clients.llm_client import LLMError
from tests.helpers import make_headers


def _payload(message, channel="whatsapp"):
    return {"session_id": "session-1", "channel": channel, "message": message}


@pytest.mark.parametrize("channel", ["in_app", "whatsapp"])
async def test_fee_inquiry_returns_exact_figures(client, settings, fake_backend, channel):
    response = await client.post(
        "/api/v1/ai/parent-query",
        json=_payload("How much school fees do I owe for Brian?", channel=channel),
        headers=make_headers(settings, channel=channel),
    )
    assert response.status_code == 200
    data = response.json()["data"]
    assert "12,500.00" in data["reply"]
    assert "35,000.00" in data["reply"]
    assert "22,500.00" in data["reply"]
    assert data["action_taken"] == "READ_FEE_BALANCE"


async def test_attendance_returns_exact_counts(client, settings):
    response = await client.post(
        "/api/v1/ai/parent-query",
        json=_payload("What is Brian's attendance?"),
        headers=make_headers(settings),
    )
    assert response.status_code == 200
    assert "62 of 63" in response.json()["data"]["reply"]


async def test_whatsapp_results_are_blocked(client, settings, fake_backend):
    response = await client.post(
        "/api/v1/ai/parent-query",
        json=_payload("Can you share Brian's exam results?"),
        headers=make_headers(settings, channel="whatsapp"),
    )
    assert response.status_code == 200
    data = response.json()["data"]
    assert data["action_taken"] == "RESULTS_BLOCKED_WHATSAPP"
    assert "portal.ikonex.ac.ke/parent/results" in data["reply"]
    assert fake_backend.marks_calls == 0


async def test_out_of_scope_refusal_includes_contacts(client, settings):
    response = await client.post(
        "/api/v1/ai/parent-query",
        json=_payload("What bus route does Brian use?"),
        headers=make_headers(settings),
    )
    assert response.status_code == 200
    reply = response.json()["data"]["reply"]
    assert "can only help with school fee balances and attendance" in reply
    assert "+254700111222" in reply
    assert "office@ikonex.ac.ke" in reply


async def test_llm_hallucination_falls_back_to_template(client, settings, fake_llm):
    fake_llm.text = "Your fee balance is KES 9,999.00."
    response = await client.post(
        "/api/v1/ai/parent-query",
        json=_payload("How much do I owe for Brian?"),
        headers=make_headers(settings),
    )
    assert response.status_code == 200
    reply = response.json()["data"]["reply"]
    assert "12,500.00" in reply
    assert "9,999.00" not in reply


async def test_llm_timeout_falls_back_to_template(client, settings, fake_llm):
    fake_llm.error = LLMError("timeout")
    response = await client.post(
        "/api/v1/ai/parent-query",
        json=_payload("How much do I owe for Brian?"),
        headers=make_headers(settings),
    )
    assert response.status_code == 200
    assert "12,500.00" in response.json()["data"]["reply"]


async def test_audit_log_written(client, settings, fake_backend):
    await client.post(
        "/api/v1/ai/parent-query",
        json=_payload("How much do I owe for Brian?"),
        headers=make_headers(settings),
    )
    assert fake_backend.audit_calls
    audit = fake_backend.audit_calls[0]
    assert audit.feature == "parent_assistant"
    assert audit.intent == "FEE_INQUIRY"
    assert not hasattr(audit, "channel")
