# Bridge Portal AI Service

Standalone FastAPI service for the Release 1 AI features:

- Parent assistant for fee and attendance questions (in-app and WhatsApp).
- Chat-initiated M-Pesa STK push with explicit two-step confirmation.
- CBC report comment and re-release note drafting for teachers.

It calls the Go internal API for every fact and action. It has no tenant database
credentials. Redis is used only for ephemeral session limits and payment state.

---

## Data boundary

1. No PostgreSQL access. Fee balances, attendance, marks, children, and school
   contacts come from permission-checked Go internal endpoints.
2. No invented facts. Currency amounts and attendance counts are validated
   against the Go payload; failures fall back to deterministic templates.
3. No STK push without an explicit parent confirmation in the active session.
4. WhatsApp receives minimal facts only: fees, attendance, and payment prompts.
   Results and report comments are replaced with a Parent Portal link.
5. Every AI request writes an audit entry through the Go internal API.

---

## Configuration

Loaded from environment / `.env` by `app/config.py`.

| Variable | Default | Purpose |
|---|---|---|
| `AI_SERVICE_HOST` | `0.0.0.0` | Bind host |
| `AI_SERVICE_PORT` | `8000` | Bind port |
| `GO_BACKEND_INTERNAL_URL` | `http://localhost:8080/api/v1/internal` | Go internal API |
| `AI_SERVICE_SECRET_TOKEN` | `dev_internal_secret_token` | Service token and default HMAC key |
| `AI_CONTEXT_SIGNING_KEY` | falls back to service token | Optional separate `X-User-Context` signing key |
| `AI_REDIS_URL` | empty | Shared Redis session store; empty uses in-memory (tests/local only) |
| `AI_REDIS_PREFIX` | `ai:session:` | Redis key prefix |
| `MISTRAL_API_URL` | `http://localhost:11434/v1` | OpenAI-compatible LLM endpoint |
| `MISTRAL_MODEL_NAME` | `mistral-7b-instruct` | Model name |
| `SESSION_TTL_SECONDS` | 21600 | Session state TTL |
| `MAX_READ_LOOKUPS` | 15 | Read lookups per session |
| `MAX_PAYMENT_INITS` | 3 | Payment initiations per session |
| `PORTAL_RESULTS_URL` | `https://portal.ikonex.ac.ke/parent/results` | WhatsApp results link |

---

## Authentication

Every route except `GET /healthz` requires:

1. `Authorization: Bearer <token>` or `X-Internal-Service-Token: <token>`, matching
   `AI_SERVICE_SECRET_TOKEN`.
2. `X-User-Context: <base64url_json>.<hmac_sha256_hex>`, signed with
   `AI_CONTEXT_SIGNING_KEY` and containing `user_id`, `school_id`,
   `membership_id`, `role`, `channel`, `session_id`, and `exp`.
3. `X-Delegated-Session` for chat sessions. It is forwarded to the Go API.

The signed context channel is authoritative; the request body cannot downgrade
WhatsApp to in-app.

---

## Endpoints

| Method | Path | Notes |
|---|---|---|
| `GET` | `/healthz` | Liveness, no auth |
| `POST` | `/api/v1/ai/parent-query` | Fee, attendance, results, payment, out-of-scope |
| `POST` | `/api/v1/ai/report-comments/draft` | Teacher-only draft mode |
| `POST` | `/api/v1/ai/syllabus/map` | Release 2 stub (`501`) |
| `POST` | `/api/v1/ai/blog-safety-check` | Release 2 stub (`501`) |
| `POST` | `/api/v1/ai/reconciliation/suggest` | Release 2 stub (`501`) |
| `POST` | `/api/v1/ai/payment-reminders/draft` | Release 2 stub (`501`) |

`payment_state` remains one of `NONE`, `AWAITING_SLOTS`, `AWAITING_CONFIRMATION`,
or `STK_SENT`. When multiple children exist and none matches, the reply uses
`AWAITING_SLOTS` with `action_taken = PAYMENT_AWAITING_CHILD` so the child step is
auditable separately from missing amount/phone slots.

---

## Internal API contract used

- `GET /school-profile`
- `GET /parents/{user_id}/children`
- `GET /students/{id}/fee-balance`
- `GET /students/{id}/attendance-summary`
- `GET /students/{id}/marks-summary`
- `POST /payments/stk-push` with `student_id`, `amount`, `phone_number`,
  `source`, and `parent_confirmed`
- `POST /ai/audit-log`

Responses use the standard `{"status":"success","data":...}` envelope.

---

## Local development

```bash
cd ai_service
python3 -m venv .venv
. .venv/bin/activate
pip install -r requirements.txt
uvicorn app.main:app --host 0.0.0.0 --port 8000 --reload
```

Tests use an in-memory session store and do not require Redis:

```bash
cd ai_service
python -m pytest -q
```
