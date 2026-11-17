import json
from typing import Optional

import httpx


class LLMError(Exception):
    pass


def _strip_code_fences(text: str) -> str:
    text = text.strip()
    if text.startswith("```"):
        lines = text.splitlines()
        if lines and lines[0].startswith("```"):
            lines = lines[1:]
        if lines and lines[-1].startswith("```"):
            lines = lines[:-1]
        text = "\n".join(lines)
    return text.strip()


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

    async def _chat(self, system_prompt: str, user_prompt: str) -> str:
        payload = {
            "model": self._model,
            "messages": [
                {"role": "system", "content": system_prompt},
                {"role": "user", "content": user_prompt},
            ],
            "temperature": 0.2,
        }
        try:
            response = await self._client.post(
                f"{self._base_url}/chat/completions",
                json=payload,
                timeout=self._timeout,
            )
            response.raise_for_status()
            data = response.json()
            content = data["choices"][0]["message"]["content"]
        except Exception as exc:  # noqa: BLE001
            raise LLMError(str(exc)) from exc
        if not isinstance(content, str) or not content.strip():
            raise LLMError("empty LLM response")
        return _strip_code_fences(content)

    async def complete_text(self, system_prompt: str, user_prompt: str) -> str:
        return await self._chat(system_prompt, user_prompt)

    async def complete_json(
        self, system_prompt: str, user_prompt: str
    ) -> Optional[dict]:
        text = await self._chat(system_prompt, user_prompt)
        try:
            parsed = json.loads(text)
        except json.JSONDecodeError as exc:
            raise LLMError("LLM did not return valid JSON") from exc
        if not isinstance(parsed, dict):
            raise LLMError("LLM JSON was not an object")
        return parsed
