# Ikonex Academy — AI Service

Standalone Python / FastAPI microservice that powers Release 1 AI features:

- **Parent Assistant** (in-app and WhatsApp) for fee-balance and attendance questions.
- **Chat-initiated M-Pesa payment** (multi-turn, explicit confirmation before STK push).
- **CBC report card comment / re-release note drafter** for teachers.

It talks to a self-hosted Mistral model through an OpenAI-compatible endpoint
(Ollama / vLLM) and to the Go backend's permission-checked internal API.

> Release 1 uses **synthetic data only** and is **English only**.

---

## Non-negotiable data boundary

1. **Zero direct database access.** This service holds no PostgreSQL or Redis credentials
   and never connects to the database. Every fact (fee balance, attendance, marks, school
   contact) is fetched from the Go internal API.
2. **No invented facts.** The LLM is used for language understanding and drafting only.
   Every fee figure and attendance count is validated against the Go payload; if the model
   alters a number the service falls back to a deterministic template.
3. **Humans confirm real-world actions.** No STK push without an explicit `YES`, and report
   comments are drafts (`requires_teacher_approval: true`).
4. **WhatsApp carries minimal facts only.** On `channel == "whatsapp"` the assistant returns
   fees and attendance, but never marks, grades, CBC levels, or comments — it sends a Parent
   Portal link instead.
5. **Every AI action is audited** via `POST /api/v1/internal/ai/audit-log`.

---

## Request flow

```
Go backend ──► POST /api/v1/ai/parent-query
  headers: X-Internal-Service-Token + signed X-User-Context
                 │
                 ▼
        FastAPI (this service)
           ├─ verify token + HMAC context (401 on failure)
           ├─ enforce session rate limits (429)
           ├─ classify intent (heuristic; LLM for phrasing)
           ├─ GET  /api/v1/internal/...   (facts, permission-checked)
           ├─ POST /api/v1/internal/payments/stk-push  (only after "YES")
           ├─ POST /api/v1/internal/ai/audit-log
           └─ reply (LLM text or deterministic fallback)
                 │
                 ▼
        Go backend ──► parent (in-app / WhatsApp)
```

---

## Directory layout

```
ai_service/
├── Dockerfile
├── requirements.txt
├── pytest.ini
├── app/
│   ├── main.py            # FastAPI app, routers, /healthz, error envelope
│   ├── config.py          # Settings (env vars + tunables)
│   ├── security.py        # service token + HMAC X-User-Context verification
│   ├── schemas.py         # Pydantic v2 DTOs and enums
│   ├── state.py           # in-memory TTL session store (rate limits, payment state)
│   ├── prompts.py         # versioned prompts, templates, forbidden-term list, formatting
│   ├── utils.py           # intent/slot extraction, child matching, validators
│   ├── deps.py            # FastAPI dependencies (backend + LLM clients)
│   ├── clients/
│   │   ├── backend_client.py  # async httpx client for Go internal API
│   │   └── llm_client.py      # async httpx client for Mistral
│   └── api/
│       ├── parent_assistant.py   # POST /api/v1/ai/parent-query
│       ├── chat_payment.py       # payment slot state machine
│       ├── report_drafter.py     # POST /api/v1/ai/report-comments/draft
│       ├── syllabus_mapper.py    # Release 2 — returns 501
│       ├── blog_safety.py        # Release 2 — returns 501
│       ├── reconciliation.py     # Release 2 — returns 501
│       └── payment_reminders.py  # Release 2 — returns 501
└── tests/
    ├── conftest.py
    ├── helpers.py
    ├── test_security_and_limits.py
    ├── test_parent_assistant.py
    ├── test_chat_payment.py
    └── test_report_drafter.py
```

---

## Configuration

Loaded by `app/config.py` from environment / `.env` (see `../.env.example`).

| Variable | Default | Purpose |
|---|---|---|
| `AI_SERVICE_HOST` | `0.0.0.0` | Bind host |
| `AI_SERVICE_PORT` | `8000` | Bind port |
| `GO_BACKEND_INTERNAL_URL` | `http://localhost:8080/api/v1/internal` | Go internal API base URL |
| `MISTRAL_API_URL` | `http://localhost:11434/v1` | OpenAI-compatible LLM endpoint |
| `MISTRAL_MODEL_NAME` | `mistral-7b-instruct` | Model name |
| `AI_SERVICE_SECRET_TOKEN` | `dev_internal_secret_token` | Shared secret for service auth + HMAC |
| `request_timeout_seconds` | `10.0` | Go API call timeout |
| `llm_timeout_seconds` | `30.0` | LLM call timeout |
| `session_ttl_seconds` | `21600` (6h) | Session state TTL |
| `max_read_lookups` | `15` | Max read lookups per session |
| `max_payment_inits` | `3` | Max STK push initiations per session |
| `portal_results_url` | `https://portal.ikonex.ac.ke/parent/results` | Deep link for WhatsApp results refusal |

All names are case-insensitive (pydantic-settings).

---

## Security model

Every route except `GET /healthz` requires two headers, both issued by the Go backend:

1. **`X-Internal-Service-Token`** — must equal `AI_SERVICE_SECRET_TOKEN`
   (constant-time comparison).
2. **`X-User-Context`** — `<base64url_payload>.<hex_hmac_sha256>`, where the signature is
   `HMAC-SHA256(secret, base64url_payload)`.

Decoded payload:

```json
{
  "user_id": "uuid",
  "school_id": "uuid",
  "membership_id": "uuid",
  "role": "parent | teacher | dean_of_studies | principal | deputy",
  "channel": "in_app | whatsapp",
  "session_id": "uuid",
  "exp": 1762300000
}
```

Missing/invalid signature or `exp < now` → `401 Unauthorized` (no LLM or Go call is made).
Channel and session are taken from the **signed context**, not the request body, so callers
cannot downgrade `whatsapp` to `in_app` to bypass the results restriction.

### Generating a valid header (for local testing)

```python
import base64, hashlib, hmac, json, time

secret = "dev_internal_secret_token"
payload = {
    "user_id": "user-1",
    "school_id": "school-1",
    "membership_id": "membership-1",
    "role": "parent",
    "channel": "in_app",
    "session_id": "session-1",
    "exp": int(time.time()) + 3600,
}
b64 = base64.urlsafe_b64encode(json.dumps(payload).encode()).rstrip(b"=").decode()
sig = hmac.new(secret.encode(), b64.encode(), hashlib.sha256).hexdigest()
print(f"{b64}.{sig}")
```

---

## HTTP API

### `GET /healthz`
No auth. Returns `{"status": "ok"}`.

### `POST /api/v1/ai/parent-query`

Headers: `X-Internal-Service-Token`, `X-User-Context`.

Request:
```json
{
  "session_id": "550e8400-e29b-41d4-a716-446655440000",
  "channel": "whatsapp",
  "message": "How much school fees do I owe for Brian?",
  "student_id": "7c9e6679-7425-40de-944b-e07fc1f90ae7"
}
```

Success (`200`):
```json
{
  "status": "success",
  "data": {
    "reply": "The current fee balance for Brian Otieno (ADM: ADM-1024) for Term 1 2026 is KES 12,500.00 (...). Due date: 2026-01-15. Would you like to make a payment via M-Pesa?",
    "intent": "FEE_INQUIRY",
    "channel": "whatsapp",
    "payment_state": "NONE",
    "action_taken": "READ_FEE_BALANCE",
    "portal_link": null
  }
}
```

`payment_state` is one of `NONE`, `AWAITING_SLOTS`, `AWAITING_CONFIRMATION`, `STK_SENT`.

Intents: `FEE_INQUIRY`, `ATTENDANCE_INQUIRY`, `RESULTS_INQUIRY`, `PAYMENT_FLOW`,
`OUT_OF_SCOPE`, `UNSUPPORTED`.

### `POST /api/v1/ai/report-comments/draft`

Role must be `teacher`, `dean_of_studies`, `principal`, or `deputy`.

Request:
```json
{
  "mode": "term_comment",
  "student_id": "7c9e6679-7425-40de-944b-e07fc1f90ae7",
  "term_id": "9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d",
  "teacher_reason": null,
  "subject_name": null
}
```

`mode` is `term_comment` or `rerelease_note` (the latter requires `teacher_reason`).

Success (`200`):
```json
{
  "status": "success",
  "data": {
    "draft_comment": "Brian demonstrated EE in Mathematics and ME in English this term. Attendance was 62 of 63 school days. Continued focus will support further progress next term.",
    "requires_teacher_approval": true,
    "model_used": "mistral-7b-instruct",
    "fallback_used": false
  }
}
```

### Status codes

| Code | When |
|---|---|
| `200` | Success |
| `400` | Invalid draft mode / missing `teacher_reason` |
| `401` | Missing/invalid service token or user context (missing, tampered, expired) |
| `403` | Role not allowed to draft report comments |
| `429` | Session lookup limit (15) or payment limit (3) exceeded |
| `501` | Release 2 endpoints (stubs) |

---

## Internal API dependencies (called by this service)

All calls forward both inbound headers and are scoped by `school_id` on the Go side.

| Method | Path | Returns |
|---|---|---|
| GET | `/school-profile` | name, phone, email, office hours, portal URL |
| GET | `/parents/{user_id}/children` | linked students |
| GET | `/students/{student_id}/fee-balance` | balances, term, due date, payment-blocked-until |
| GET | `/students/{student_id}/attendance-summary` | present/absent counts, recent absences |
| GET | `/students/{student_id}/marks-summary?term_id=...` | released CBC levels + trend |
| POST | `/payments/stk-push` | M-Pesa STK push (only after explicit `YES`) |
| POST | `/ai/audit-log` | audit record for every AI interaction |

Base URL: `GO_BACKEND_INTERNAL_URL`.

---

## Behaviour rules

- **Fee / attendance:** figures come from the Go payload. Post-generation validation rejects
  any LLM currency amount not in the payload (fallback to template); attendance requires the
  exact counts to appear in the text.
- **WhatsApp results:** the marks endpoint is **never called**; the reply contains the portal
  deep link (`AC-13`).
- **Out of scope** (discipline, medical, transport, fee waivers, etc.): polite refusal plus
  school phone/email from `school-profile` (`C57`).
- **Payments:** the assistant extracts `student`, `amount`, `mpesa_phone`, transitions to
  `AWAITING_CONFIRMATION`, and repeats them back. Only `yes` / `y` / `confirm` / `proceed`
  triggers the STK push; anything else cancels or re-confirms. Any amount `> 0` is accepted
  (overpayment becomes credit). A `403 CHAT_PAYMENT_PAUSED_PHONE_CHANGE` response is surfaced
  as a 24-hour security pause with school contacts (`AC-23`, `AC-26`).
- **Report comments:** drafts are post-filtered to remove ranking and clinical/diagnostic
  terms (`rank`, `position`, `top`, `bottom`, `ADHD`, `dyslexic`, …); the response always has
  `requires_teacher_approval: true` (`AC-AI-06`).
- **Rate limits:** 15 read lookups and 3 payment initiations per `session_id` (6h TTL). When
  exceeded, the polite limit message includes school contacts and no Go call is made (`C66`).

---

## Local development

Prerequisites: Python 3.11+.

```bash
cd ai_service
python3 -m venv .venv
. .venv/bin/activate
pip install -r requirements.txt

# run (reload)
uvicorn app.main:app --host 0.0.0.0 --port 8000 --reload
```

Or from the repo root: `make ai-run`.

Interactive docs: http://localhost:8000/docs

---

## Docker

The service is part of `docker-compose.yml`:

```bash
# from repo root
docker compose up --build ai_service
# or the whole stack
make dev
```

The compose file supplies `GO_BACKEND_INTERNAL_URL`, `AI_SERVICE_SECRET_TOKEN`,
`MISTRAL_API_URL`, and `MISTRAL_MODEL_NAME`. Port `8000` is exposed.

---

## Tests

```bash
cd ai_service
. .venv/bin/activate
python -m pytest -q
```

Tests use FastAPI dependency overrides with fake backend/LLM clients (see
`tests/helpers.py`) — no real Go backend or Mistral needed. Coverage:

- `test_security_and_limits.py` — token/HMAC/expiry rejection, read + payment caps.
- `test_parent_assistant.py` — exact fee/attendance figures, WhatsApp results block,
  out-of-scope contacts, hallucination + timeout fallbacks, audit logging.
- `test_chat_payment.py` — no STK before confirmation, cancel on `NO`, push on `YES`,
  overpayment, 24h phone-change freeze.
- `test_report_drafter.py` — forbidden-term stripping, re-release note, role gate, audit.

---

## Release scope

**Release 1 (implemented):** parent assistant, chat payments, report drafter/comments.

**Release 2 (stubs returning `501`):** syllabus mapper, blog safety check, M-Pesa
reconciliation, payment reminders.

**Out of scope for R1:** Swahili generation (English only), direct DB access.

---

## Troubleshooting

- **`401`** — missing/mismatched `X-Internal-Service-Token`, or the `X-User-Context` signature
  doesn't match the base64url payload, or `exp` is in the past.
- **`429`** — session hit 15 reads or 3 payment initiations; wait or contact the school.
- **Template replies instead of LLM text** — the LLM was unreachable, timed out, returned
  empty output, or produced a number that failed validation. This is expected and safe.
- **Config not picked up** — ensure `.env` exists (copy `.env.example`) and the variable names
  match the table above.

See also `../tasks/developer-01-ai-engineer.md`, `../Docs/API_CONTRACTS.md`, and
`../Docs/SECURITY.md`.
