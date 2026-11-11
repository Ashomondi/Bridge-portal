from typing import Any, Dict, List, Optional

import httpx

from ..schemas import (
    AttendanceSummary,
    AuditLogRequest,
    Child,
    FeeBalance,
    MarksSummary,
    SchoolProfile,
    StkPushRequest,
)


class BackendError(Exception):
    def __init__(
        self,
        status_code: int,
        code: Optional[str] = None,
        message: str = "",
        payload: Optional[Dict[str, Any]] = None,
    ):
        super().__init__(message or code or "backend error")
        self.status_code = status_code
        self.code = code
        self.message = message
        self.payload = payload or {}

    @classmethod
    def from_response(cls, response: httpx.Response) -> "BackendError":
        code = None
        message = ""
        payload: Dict[str, Any] = {}
        try:
            payload = response.json()
            error = payload.get("error")
            if isinstance(error, dict):
                code = error.get("code")
                message = error.get("message", "")
            elif isinstance(error, str):
                code = error
        except Exception:  # noqa: BLE001
            message = response.text
        return cls(status_code=response.status_code, code=code, message=message, payload=payload)


class BackendClient:
    def __init__(
        self,
        base_url: str,
        service_token: str,
        user_context: str,
        client: httpx.AsyncClient,
        timeout: float = 10.0,
    ):
        self._base_url = base_url.rstrip("/")
        self._client = client
        self._timeout = timeout
        self._headers = {
            "X-Internal-Service-Token": service_token,
            "X-User-Context": user_context,
        }

    async def _get(self, path: str, params: Optional[Dict[str, Any]] = None) -> Any:
        response = await self._client.get(
            f"{self._base_url}{path}", headers=self._headers, params=params, timeout=self._timeout
        )
        if response.status_code >= 400:
            raise BackendError.from_response(response)
        return response.json()

    async def _post(self, path: str, json_body: Dict[str, Any]) -> Any:
        response = await self._client.post(
            f"{self._base_url}{path}", headers=self._headers, json=json_body, timeout=self._timeout
        )
        if response.status_code >= 400:
            raise BackendError.from_response(response)
        if not response.content:
            return {}
        return response.json()

    async def get_school_profile(self) -> SchoolProfile:
        data = await self._get("/school-profile")
        return SchoolProfile(**data)

    async def get_children(self, user_id: str) -> List[Child]:
        data = await self._get(f"/parents/{user_id}/children")
        if isinstance(data, dict):
            data = data.get("children", data.get("data", []))
        return [Child(**item) for item in data or []]

    async def get_fee_balance(self, student_id: str) -> FeeBalance:
        data = await self._get(f"/students/{student_id}/fee-balance")
        return FeeBalance(**data)

    async def get_attendance_summary(self, student_id: str) -> AttendanceSummary:
        data = await self._get(f"/students/{student_id}/attendance-summary")
        return AttendanceSummary(**data)

    async def get_marks_summary(
        self, student_id: str, term_id: Optional[str] = None
    ) -> MarksSummary:
        params = {"term_id": term_id} if term_id else None
        data = await self._get(f"/students/{student_id}/marks-summary", params=params)
        return MarksSummary(**data)

    async def stk_push(self, payload: StkPushRequest) -> Dict[str, Any]:
        return await self._post("/payments/stk-push", payload.model_dump())

    async def audit_log(self, payload: AuditLogRequest) -> None:
        try:
            await self._post("/ai/audit-log", payload.model_dump())
        except Exception:  # noqa: BLE001
            return None
