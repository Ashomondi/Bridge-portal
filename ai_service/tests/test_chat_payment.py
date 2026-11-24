import pytest

from app.clients.backend_client import BackendError
from app.schemas import Child

from tests.helpers import make_headers


def _payload(message, channel="whatsapp"):
    return {"session_id": "session-1", "channel": channel, "message": message}


async def _start_payment(client, settings, message="Pay 5000 for Brian to 0712345678"):
    return await client.post(
        "/api/v1/ai/parent-query",
        json=_payload(message),
        headers=make_headers(settings, channel="whatsapp"),
    )


async def _continue(client, settings, message):
    return await client.post(
        "/api/v1/ai/parent-query",
        json=_payload(message),
        headers=make_headers(settings, channel="whatsapp"),
    )


async def test_payment_does_not_call_stk_before_confirmation(client, settings, fake_backend):
    response = await _start_payment(client, settings)
    assert response.status_code == 200
    data = response.json()["data"]
    assert data["payment_state"] == "AWAITING_CONFIRMATION"
    assert "5,000.00" in data["reply"]
    assert "Brian Otieno" in data["reply"]
    assert "ADM-1024" in data["reply"]
    assert "0712345678" in data["reply"]
    assert fake_backend.stk_calls == []


async def test_payment_cancelled_on_no(client, settings, fake_backend):
    await _start_payment(client, settings)
    response = await _continue(client, settings, "No")
    data = response.json()["data"]
    assert data["payment_state"] == "NONE"
    assert fake_backend.stk_calls == []


async def test_payment_initiated_on_yes(client, settings, fake_backend):
    await _start_payment(client, settings)
    response = await _continue(client, settings, "Yes")
    data = response.json()["data"]
    assert data["payment_state"] == "STK_SENT"
    assert len(fake_backend.stk_calls) == 1
    call = fake_backend.stk_calls[0]
    assert call.amount == "5000.00"
    assert call.phone_number == "+254712345678"
    assert call.parent_confirmed is True
    assert call.student_id == "stu-1"
    assert call.source == "chat"


@pytest.mark.parametrize(
    "message",
    ["Yes, proceed", "confirm payment", "Proceed with payment.", "yes please"],
)
async def test_conversational_affirmations_trigger_payment(client, settings, fake_backend, message):
    await _start_payment(client, settings)
    response = await _continue(client, settings, message)
    assert response.json()["data"]["payment_state"] == "STK_SENT"
    assert len(fake_backend.stk_calls) == 1


@pytest.mark.parametrize("message", ["No", "cancel", "do not proceed"])
async def test_negatives_do_not_trigger_payment(client, settings, fake_backend, message):
    await _start_payment(client, settings)
    response = await _continue(client, settings, message)
    assert response.json()["data"]["payment_state"] != "STK_SENT"
    assert fake_backend.stk_calls == []


async def test_overpayment_is_allowed(client, settings, fake_backend):
    fake_backend.fee.current_balance = "2000.00"
    await _start_payment(client, settings, "Pay 5000 for Brian to 0712345678")
    response = await _continue(client, settings, "Yes")
    assert response.json()["data"]["payment_state"] == "STK_SENT"
    assert len(fake_backend.stk_calls) == 1


async def test_phone_change_freeze_blocks_stk(client, settings, fake_backend):
    fake_backend.stk_error = BackendError(
        status_code=403,
        code="CHAT_PAYMENT_PAUSED_PHONE_CHANGE",
        message="phone change freeze",
    )
    await _start_payment(client, settings)
    response = await _continue(client, settings, "Yes")
    data = response.json()["data"]
    assert data["payment_state"] == "NONE"
    assert "24 hours" in data["reply"]
    assert "+254700111222" in data["reply"]


async def test_ambiguous_child_has_distinct_state_action(client, settings, fake_backend):
    fake_backend.children = [
        Child(student_id="stu-1", full_name="Brian Otieno", admission_number="ADM-1024"),
        Child(student_id="stu-2", full_name="Joy Otieno", admission_number="ADM-1025"),
    ]
    response = await _start_payment(
        client, settings, "Pay 5000 for an unknown child to 0712345678"
    )
    data = response.json()["data"]
    assert data["payment_state"] == "AWAITING_SLOTS"
    assert data["action_taken"] == "PAYMENT_AWAITING_CHILD"
    assert "Which child" in data["reply"]
    assert fake_backend.stk_calls == []

    response = await _continue(client, settings, "Brian")
    data = response.json()["data"]
    assert data["payment_state"] == "AWAITING_CONFIRMATION"
    assert data["action_taken"] == "PAYMENT_AWAITING_CONFIRMATION"
    assert "Brian Otieno" in data["reply"]
    assert fake_backend.stk_calls == []
