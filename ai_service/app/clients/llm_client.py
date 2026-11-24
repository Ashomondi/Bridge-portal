from typing import Optional

import httpx


class LLMError(Exception):
    pass


class LLMClient:
    def __init__(
        self,
        base_url: str,
        model: str,
        client: httpx.AsyncClient,
        timeout: float = 30.0,
    ):
        self._base_url = base_url.rstrip("/")
        self._model = model
        self._client = client
        self._timeout = timeout

    async def complete_text(
        self,
        system_prompt: str,
        user_prompt: str,
        temperature: float = 0.2,
    ) -> str:
        body = {
            "model": self._model,
            "messages": [
                {"role": "system", "content": system_prompt},
                {"role": "user", "content": user_prompt},
            ],
            "temperature": temperature,
        }
        try:
            response = await self._client.post(
                f"{self._base_url}/chat/completions",
                json=body,
                timeout=self._timeout,
            )
        except httpx.HTTPError as exc:
            raise LLMError("LLM request failed") from exc
        if response.status_code >= 400:
            raise LLMError(f"LLM returned {response.status_code}")
        try:
            payload = response.json()
            content: Optional[str] = payload["choices"][0]["message"]["content"]
        except (KeyError, IndexError, TypeError, ValueError) as exc:
            raise LLMError("Invalid LLM response") from exc
        if not content:
            raise LLMError("Empty LLM response")
        return content.strip()
