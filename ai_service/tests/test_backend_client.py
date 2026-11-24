import httpx

from app.clients.backend_client import BackendClient
from app.schemas import AuditLogRequest, StkPushRequest


async def test_backend_client_unwraps_envelope_and_forwards_headers():
    seen = {}

    def handler(request: httpx.Request) -> httpx.Response:
        seen["headers"] = request.headers
        seen["path"] = request.url.path
        return httpx.Response(
            200,
            json={
                "status": "success",
                "data": [
                    {
                        "student_id": "stu-1",
                        "admission_number": "ADM-1024",
                        "full_name": "Brian Otieno",
                        "grade_name": "Grade 8",
                        "relationship": "mother",
                    }
                ],
            },
        )

    transport = httpx.MockTransport(handler)
    async with httpx.AsyncClient(transport=transport) as http_client:
        client = BackendClient(
            base_url="https://backend.test/api/v1/internal",
            service_token="token-1",
            user_context="ctx.sig",
            delegated_session="delegated-1",
            client=http_client,
            timeout=5,
        )
        children = await client.get_children("user-1")

    assert children[0].grade_name == "Grade 8"
    assert seen["path"] == "/api/v1/internal/parents/user-1/children"
    assert seen["headers"]["authorization"] == "Bearer token-1"
    assert seen["headers"]["x-internal-service-token"] == "token-1"
    assert seen["headers"]["x-user-context"] == "ctx.sig"
    assert seen["headers"]["x-delegated-session"] == "delegated-1"


async def test_backend_payloads_match_gateway_contract():
    seen = {"bodies": []}

    def handler(request: httpx.Request) -> httpx.Response:
        seen["path"] = request.url.path
        seen["bodies"].append(request.content.decode())
        return httpx.Response(202, json={"status": "success", "data": {"status": "pending"}})

    transport = httpx.MockTransport(handler)
    async with httpx.AsyncClient(transport=transport) as http_client:
        client = BackendClient(
            base_url="https://backend.test/api/v1/internal",
            service_token="token-1",
            user_context="ctx.sig",
            client=http_client,
            timeout=5,
        )
        await client.stk_push(
            StkPushRequest(
                student_id="stu-1",
                amount="5000.00",
                phone_number="+254712345678",
                source="chat",
                parent_confirmed=True,
            )
        )
        await client.audit_log(
            AuditLogRequest(
                session_id="session-1",
                feature="parent_assistant",
                intent="FEE_INQUIRY",
                endpoints_read=["/api/v1/internal/school-profile"],
                action_taken="READ_FEE_BALANCE",
            )
        )

    stk_body, audit_body = seen["bodies"]
    assert '"phone_number":"+254712345678"' in stk_body
    assert '"source":"chat"' in stk_body
    assert "mpesa_phone" not in stk_body
    assert "confirmation_message_text" not in stk_body
    assert '"channel"' not in audit_body
    assert '"feature":"parent_assistant"' in audit_body
