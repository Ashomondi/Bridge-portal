# Ikonex Academy – Release 1 Task Assignment
## Developer 1: AI Engineer (**Sos** — @sospeter-57)

- **Developer Name:** Sos (@sospeter-57)
- **GitHub Username:** `@sospeter-57`
- **Role:** AI Engineer (Python / FastAPI / Self-Hosted Mistral LLM Integration)
- **Primary Directory Ownership:** `ai_service/`
- **Release Scope:** Release 1 (Synthetic Data Only; English Only; In-App & WhatsApp Parent Assistant, Chat-Initiated M-Pesa STK Push Confirmation Flow, and CBC Report Comment / Re-Release Note Drafter)

---

## 1. Role and Objective

As the **AI Engineer (Sos — @sospeter-57)** for Ikonex Academy Release 1, you own the standalone Python/FastAPI AI microservice (`ai_service/`) and its integration with the self-hosted Mistral model (`mistral-7b-instruct` via Ollama/vLLM OpenAI-compatible API) and the Go backend's internal API (`backend/internal/internalapi`).

Your core architectural mandate is governed by the **AI Service Data Boundary** defined in `Docs/SECURITY.md`, `Docs/ARCHITECTURE.md`, and `Docs/Ikonex.md`:
1. **Zero Direct Database Access:** The AI service holds no PostgreSQL or Redis credentials for the main database and never connects to PostgreSQL directly.
2. **Zero Fact Invention (No Hallucinated Figures):** Every fee balance, attendance record, student name, payment status, or assessment mark must be retrieved synchronously from permission-checked Go internal endpoints (`/api/v1/internal/...`). The LLM is used strictly for natural-language understanding, intent/slot extraction, polite formatting, and drafting teacher comments.
3. **Human Confirmation Before Any Real-World Action:**
   - For **chat-initiated fee payments**, the AI service must extract the child, amount, and typed M-Pesa phone number, repeat all three back to the parent, and wait for an explicit affirmative confirmation (`"yes"`) before calling the Go backend to trigger a Safaricom Daraja STK push.
   - For **report card comments and re-release notes**, the AI service only returns a draft; the class teacher edits and explicitly approves the text in the portal before parents can ever see it.
4. **Strict Channel Content Boundary (WhatsApp vs. In-App):** On WhatsApp (`channel == "whatsapp"`), the assistant may only communicate minimal facts: **fee balances, attendance status, and payment prompts**. It must **never** output student exam marks, CBC performance levels, or teacher report comments over WhatsApp—instead, it must return a deep link to the Parent Portal PWA.
5. **Complete AI Action Logging:** Every AI request, internal fact lookup, prompt execution, and action initiated must be logged and reported back to the Go backend's AI audit endpoint.

---

## 2. Assigned Features

You own **all three Release 1 AI features** and the core AI service infrastructure:

| Feature ID | Feature Name | Target Module | Why Assigned to Sos |
|---|---|---|---|
| **AI-F1** | **AI Service Core Infrastructure, Context Verification & Internal API Client** | `ai_service/app/main.py`, `config.py`, `clients/backend_client.py`, `clients/llm_client.py`, `security.py` | Provides the shared authentication verification (`X-Internal-Service-Token` + HMAC-signed `X-User-Context`), rate-limit tracking per session, LLM retry/fallback wrapper, and Go internal API client required by all AI endpoints. |
| **AI-F2** | **Parent Assistant (In-App & WhatsApp Read-Only Q&A) + Chat-Initiated M-Pesa Payment Flow** | `ai_service/app/api/parent_assistant.py`, `ai_service/app/api/chat_payment.py` | Both features operate on the same parent conversation loop (`in_app` and `whatsapp`), share intent classification, guardian-child scoping, session lookup limits, and multi-turn payment confirmation state. |
| **AI-F3** | **CBC Report Card Comment & Result Re-Release Note Drafter** | `ai_service/app/api/report_drafter.py` | Uses the self-hosted Mistral model to synthesize CBC performance levels (`EE`, `ME`, `AE`, `BE`), subject trends, and teacher correction reasons into constructive, non-ranking draft comments for teachers. |

### Explicitly Out of Scope for Release 1 (Do Not Implement Yet)
The following files exist as placeholders in `ai_service/app/api/` for **Release 2** and must remain stubbed (`501 Not Implemented`) in Release 1:
- `ai_service/app/api/syllabus_mapper.py` (Release 2: Natural-language lesson to CBC strand mapping)
- `ai_service/app/api/blog_safety.py` (Release 2: Journalism club PII/child safety scanner)
- `ai_service/app/api/reconciliation.py` (Release 2: Unmatched M-Pesa fuzzy matching for bursars)
- `ai_service/app/api/payment_reminders.py` (Release 2: 3-stage escalating payment plan default reminders)
- Swahili language generation (`Docs/Ikonex.md` C56, A18: Release 1 is strictly **English only**).

---

## 3. Detailed Feature Specifications

### Feature AI-F1: AI Service Core Infrastructure, Context Verification & Internal API Client
- **What the feature does:**
  - Bootstraps the FastAPI server on port `8000`, validates incoming service-to-service headers from the Go backend, verifies the HMAC-SHA256 signature and expiry of `X-User-Context`, manages outbound calls to the Go backend (`GO_BACKEND_INTERNAL_URL`) forwarding both headers, and wraps the local Mistral OpenAI-compatible endpoint (`MISTRAL_API_URL`) with structured JSON output parsing, timeouts, and deterministic fallbacks when the model is slow or unreachable.
- **Business Logic & Security Rules:**
  - Every request to `/api/v1/ai/*` (except `/healthz`) must carry:
    1. `X-Internal-Service-Token`: Must match `AI_SERVICE_SECRET_TOKEN` using constant-time comparison (`hmac.compare_digest`).
    2. `X-User-Context`: Base64URL-encoded JSON payload signed with HMAC-SHA256 (`<base64url_payload>.<hex_signature>`) using `AI_SERVICE_SECRET_TOKEN`.
  - Decoded `X-User-Context` schema:
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
  - If the token is missing, invalid, or `exp < current_unix_timestamp`, immediately return `401 Unauthorized` without calling the LLM or Go backend.
  - **Session Lookup Rate Limiting (`Docs/Ikonex.md` C66):**
    - Track backend lookups per `session_id` in memory (with TTL matching the 6-hour session window):
      - Maximum **15 read lookups** (fee balance, attendance, school contact profile) per `session_id`.
      - Maximum **3 STK push payment initiations** per `session_id`.
    - When a session exceeds its lookup cap, do not call the Go backend; return a polite message instructing the parent to wait or contact the school office using cached/provided school contact details.

### Feature AI-F2: Parent Assistant (In-App & WhatsApp) & Chat-Initiated M-Pesa Payment Flow
- **What the feature does:**
  - Handles natural-language messages from verified parents via both the Next.js Parent Portal (`channel: "in_app"`) and the school's WhatsApp bot (`channel: "whatsapp"`).
  - Supports four core intents:
    1. `FEE_INQUIRY`: Answers questions about a child's fee balance, billed amount, paid amount, credit balance, and term due date.
    2. `ATTENDANCE_INQUIRY`: Answers questions about a child's daily attendance record (present/absent days in the current term).
    3. `RESULTS_INQUIRY`:
       - If `channel == "whatsapp"`: **Refuses to show marks/results in chat** (`Docs/Ikonex.md` C22, AC-13) and replies with a polite message and a direct link to the Parent Portal (`https://portal.ikonex.ac.ke/parent/results`).
       - If `channel == "in_app"`: Summarizes released term performance levels (`EE`, `ME`, `AE`, `BE`) fetched from the Go internal endpoint (only for released terms) or directs the parent to the Performance Graph tab.
    4. `PAYMENT_FLOW`: Multi-turn conversational M-Pesa STK push initiation (`chat_payment.py`).
    5. `OUT_OF_SCOPE` / `UNSUPPORTED`: Politely refuses and always provides the school's official contact details (`phone`, `email`, `office_hours`) fetched from `GET /api/v1/internal/school-profile` (`Docs/Ikonex.md` C57).
- **User Workflows Supported:**
  - **Workflow A (Read-Only Fee/Attendance Inquiry):**
    1. Parent asks: *"How much school fees do I owe for Brian?"*
    2. AI service fetches the parent's linked children from `GET /api/v1/internal/parents/{user_id}/children`.
    3. Matches *"Brian"* to `student_id`. (If the parent has multiple children and didn't specify which child, asks the parent to clarify which child they mean).
    4. Calls `GET /api/v1/internal/students/{student_id}/fee-balance`.
    5. Uses the exact figures returned by Go (`current_balance`, `credit_balance`, `term_name`, `due_date`) to formulate a clear English response. Validates that every number in the LLM output matches the Go payload; if the LLM alters a number, falls back to a deterministic template.
  - **Workflow B (Chat-Initiated Fee Payment – Multi-Turn Confirmation):**
    1. Parent says: *"I want to pay 5,000 for Brian using 0712345678"*.
    2. If any of the three required slots (`student_id`, `amount`, `mpesa_phone`) is missing, the assistant prompts for the missing field(s). Note: The parent **must explicitly type the M-Pesa phone number** to charge (`Docs/Ikonex.md` C23, A4).
    3. Once all three slots are present, the assistant transitions the session payment state to `AWAITING_CONFIRMATION` and replies with the mandatory repeat-back prompt:
       > *"Please confirm: You are about to initiate an M-Pesa STK push of **KES 5,000.00** for **Brian Otieno (Adm: ADM-1024)** to M-Pesa number **0712345678**. Reply **YES** to confirm or **NO** to cancel."*
    4. Only when the next message from the parent in that session is an unambiguous confirmation (`"yes"`, `"y"`, `"confirm"`, `"proceed"`, case-insensitive) does the AI service call `POST /api/v1/internal/payments/stk-push`.
    5. If the parent replies anything else (or changes the amount/number), the pending state is cancelled or updated and re-confirmed.
    6. **Any positive amount (`> 0`) is accepted** (`Docs/Ikonex.md` C24)—even if it exceeds `current_balance` (the backend holds excess as a credit on the student's ledger).
    7. **24-Hour Phone Change Freeze (`Docs/Ikonex.md` C27, AC-26):** If the Go backend returns `403 Forbidden` with error code `CHAT_PAYMENT_PAUSED_PHONE_CHANGE`, the assistant explains clearly that chat payments are paused for 24 hours following a registered phone number change for security, and provides the school's contact details.

### Feature AI-F3: CBC Report Card Comment & Result Re-Release Note Drafter
- **What the feature does:**
  - Provides two teacher-facing drafting modes via `POST /api/v1/ai/report-comments/draft`:
    1. `mode = "term_comment"`: Drafts a personalized, constructive 2–3 sentence CBC report card comment for a student based on their current term subject performance levels (`EE` = Exceeding Expectations, `ME` = Meeting Expectations, `AE` = Approaching Expectations, `BE` = Below Expectations), historical trend compared to prior terms, and attendance summary.
    2. `mode = "rerelease_note"`: Assists a class teacher in drafting a clear, professional explanation note when reopening and re-releasing a class's results after a subject mark correction (`Docs/Ikonex.md` C34, A9).
- **Business Logic & Guardrails:**
  - **Never ranks or compares a student to classmates** (`Docs/Ikonex.md` Section 5 Rule 5). Words/phrases like *"ranked"*, *"top of the class"*, *"bottom"*, *"position"*, *"beat"*, or medical/diagnostic labels (*"ADHD"*, *"dyslexic"*, *"diagnosed"*) must be stripped or rejected by a post-generation validator.
  - Returns `draft_comment` with `requires_teacher_approval: true`. Does not save to the database; the teacher edits and submits the final text to the Go backend.

---

## 4. Technical Implementation

### 4.1 Directory & File Structure (`ai_service/`)
You own and will implement the following files:
```text
ai_service/
├── Dockerfile
├── requirements.txt
├── app/
│   ├── __init__.py
│   ├── main.py                      # FastAPI app, router registration, /healthz, exception handlers
│   ├── config.py                    # Pydantic BaseSettings loading env vars (.env.example)
│   ├── security.py                  # X-Internal-Service-Token & HMAC X-User-Context verifier
│   ├── schemas.py                   # Pydantic v2 request/response models & internal DTOs
│   ├── prompts.py                   # Versioned system prompts & deterministic fallback templates
│   ├── state.py                     # In-memory/TTL session state for payment confirmation & rate limits
│   ├── clients/
│   │   ├── __init__.py
│   │   ├── backend_client.py        # Async httpx client for Go /api/v1/internal/* endpoints
│   │   └── llm_client.py            # Async httpx client for Mistral (/v1/chat/completions)
│   └── api/
│       ├── __init__.py
│       ├── parent_assistant.py      # POST /api/v1/ai/parent-query
│       ├── chat_payment.py          # Payment slot state machine & confirmation handler
│       ├── report_drafter.py        # POST /api/v1/ai/report-comments/draft
│       ├── syllabus_mapper.py       # R2 placeholder (returns 501 Not Implemented)
│       ├── blog_safety.py           # R2 placeholder (returns 501 Not Implemented)
│       ├── reconciliation.py        # R2 placeholder (returns 501 Not Implemented)
│       └── payment_reminders.py     # R2 placeholder (returns 501 Not Implemented)
└── tests/
    ├── __init__.py
    ├── conftest.py
    ├── test_security_and_limits.py
    ├── test_parent_assistant.py
    ├── test_chat_payment.py
    └── test_report_drafter.py
```

### 4.2 Inbound API Contracts (Called by Go Backend -> AI Service)

#### Endpoint 1: `POST /api/v1/ai/parent-query`
- **Caller:** Go Backend (`backend/internal/internalapi` or `backend/internal/whatsapp`)
- **Headers Required:**
  - `X-Internal-Service-Token: <secret>`
  - `X-User-Context: <base64url_json>.<hmac_sha256_hex>`
- **Request Body:**
  ```json
  {
    "session_id": "550e8400-e29b-41d4-a716-446655440000",
    "channel": "whatsapp",
    "message": "How much fees do I owe for Joy?",
    "student_id": "7c9e6679-7425-40de-944b-e07fc1f90ae7"
  }
  ```
  *(Note: `student_id` is optional if the parent has not selected a child yet or is on in-app chat with multiple children; for WhatsApp sessions, the session was verified against a specific student's admission number so `student_id` is pre-populated by the Go backend).*
- **Response Body (`200 OK`):**
  ```json
  {
    "status": "success",
    "data": {
      "reply": "The current fee balance for Joy Wanjiku (ADM-1042) for Term 1 2026 is KES 12,500.00 (Billed: KES 35,000.00, Paid: KES 22,500.00, Credit: KES 0.00). Due date: 2026-01-15. Would you like to make a payment via M-Pesa?",
      "intent": "FEE_INQUIRY",
      "channel": "whatsapp",
      "payment_state": "NONE",
      "action_taken": "READ_FEE_BALANCE",
      "portal_link": null
    }
  }
  ```
- **Error Responses:**
  - `401 Unauthorized`: Missing/invalid service token or expired/tampered `X-User-Context`.
  - `429 Too Many Requests`: Session lookup limit or payment initiation limit exceeded (returns a polite `reply` in `data` alongside error code `SESSION_LOOKUP_LIMIT_EXCEEDED` so the Go backend can still deliver the message to the parent).

#### Endpoint 2: `POST /api/v1/ai/report-comments/draft`
- **Caller:** Go Backend (`POST /api/v1/report-comments/draft` in `academics`)
- **Headers Required:** `X-Internal-Service-Token`, `X-User-Context` (role must be `teacher`, `dean_of_studies`, `principal`, or `deputy`).
- **Request Body:**
  ```json
  {
    "mode": "term_comment",
    "student_id": "7c9e6679-7425-40de-944b-e07fc1f90ae7",
    "term_id": "9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d",
    "teacher_reason": null
  }
  ```
  *(When `mode == "rerelease_note"`, `teacher_reason` is required, e.g. `"Mathematics continuous assessment paper 2 mark was transposed from 68 to 86"`, and `subject_name` is included).*
- **Response Body (`200 OK`):**
  ```json
  {
    "status": "success",
    "data": {
      "draft_comment": "Joy has demonstrated strong mastery this term, achieving Exceeding Expectations (EE) in Mathematics and Integrated Science and Meeting Expectations (ME) in English. Her consistent attendance (62/63 days) reflects commendable dedication, and continued focus on Pre-Technical Studies will support her further growth next term.",
      "requires_teacher_approval": true,
      "model_used": "mistral-7b-instruct",
      "fallback_used": false
    }
  }
  ```

### 4.3 Outbound Internal API Contracts (Called by AI Service -> Go Backend `internalapi`)
Your `backend_client.py` must call these exact endpoints on `GO_BACKEND_INTERNAL_URL` (`http://backend:8080/api/v1/internal`), always forwarding `X-Internal-Service-Token` and `X-User-Context`:
1. `GET /api/v1/internal/school-profile` -> Returns `{ "school_id", "name", "contact_phone", "contact_email", "office_hours", "portal_url" }`.
2. `GET /api/v1/internal/parents/{user_id}/children` -> Returns list of linked students (`student_id`, `admission_number`, `full_name`, `class_name`, `grade_level`).
3. `GET /api/v1/internal/students/{student_id}/fee-balance` -> Returns `{ "student_id", "full_name", "admission_number", "term_name", "due_date", "opening_balance", "billed_amount", "paid_amount", "current_balance", "credit_balance", "chat_payment_blocked_until" }`.
4. `GET /api/v1/internal/students/{student_id}/attendance-summary` -> Returns `{ "student_id", "term_name", "total_school_days", "days_present", "days_absent", "recent_absences": ["2026-02-10"] }`.
5. `GET /api/v1/internal/students/{student_id}/marks-summary?term_id=<uuid>` -> Returns student's subject performance levels (`EE`, `ME`, `AE`, `BE`), previous term levels for trend comparison, and release status (`is_released`).
6. `POST /api/v1/internal/payments/stk-push` -> Initiates M-Pesa STK push **only after explicit parent `"YES"` confirmation**:
   ```json
   {
     "student_id": "7c9e6679-7425-40de-944b-e07fc1f90ae7",
     "amount": "5000.00",
     "mpesa_phone": "+254712345678",
     "session_id": "550e8400-e29b-41d4-a716-446655440000",
     "parent_confirmed": true,
     "confirmation_message_text": "YES"
   }
   ```
7. `POST /api/v1/internal/ai/audit-log` -> Logs every AI interaction:
   ```json
   {
     "feature": "parent_assistant | chat_payment | report_drafter",
     "session_id": "uuid",
     "channel": "whatsapp | in_app",
     "intent": "FEE_INQUIRY",
     "endpoints_read": ["/api/v1/internal/students/7c9e.../fee-balance"],
     "action_taken": "READ_FEE_BALANCE",
     "metadata": { "fallback_used": false }
   }
   ```

---

## 5. Exact Deliverables

When you declare your assignment complete, the following exact functionality must work end-to-end:
1. **Authenticated Service Boundary:** Requests to `ai_service` without a valid `X-Internal-Service-Token` or with an expired/invalid `X-User-Context` signature are rejected with `401 Unauthorized`.
2. **Accurate Fee & Attendance Q&A (Zero Hallucination):** When a parent asks for their child's fee balance or attendance on either WhatsApp or in-app chat, the service fetches the live numbers from the Go internal API and returns a response whose numeric figures match the Go response 100%, even if the Mistral LLM is offline (via deterministic template fallback).
3. **WhatsApp Content Restriction Enforcement (`AC-13`):** When `channel == "whatsapp"` and a parent asks for exam marks, grades, report card comments, or performance levels, the AI service **never** calls the marks endpoint or outputs grades; it returns a polite refusal and the Parent Portal deep link.
4. **Polite Out-of-Scope Refusal with School Contact Details (`C57`):** When a parent asks an out-of-scope question (e.g., discipline, medical/sickbay details, transport routes, fee waiver negotiations), the assistant politely declines and includes the school's phone number and email from `GET /api/v1/internal/school-profile`.
5. **Two-Step Explicit Payment Confirmation (`AC-23`):**
   - A single message like *"Pay 5000 for Joy to 0712345678"* **never** triggers `POST /api/v1/internal/payments/stk-push` immediately.
   - Instead, it transitions the session to `AWAITING_CONFIRMATION` and repeats back Child Name, Admission Number, Amount (`KES 5,000.00`), and Phone Number (`0712345678`).
   - Only a subsequent explicit `"YES"` message triggers `POST /api/v1/internal/payments/stk-push` with `parent_confirmed: true`.
   - Overpayments or arbitrary positive amounts are accepted without artificial caps.
6. **CBC Report Comment & Re-Release Note Drafting:** Calling `POST /api/v1/ai/report-comments/draft` returns a well-formed, encouraging CBC comment (or re-release explanation note) free of any class rankings, positions, or medical diagnoses.
7. **Complete Audit Trail (`AC-28`):** Every call to `parent-query` or `report-comments/draft` writes an audit entry via `POST /api/v1/internal/ai/audit-log`.

---

## 6. Acceptance Criteria

- [ ] **AC-AI-01 (aligns with AC-28):** `ai_service` contains zero database driver packages (`psycopg2`, `asyncpg`, `sqlalchemy` are absent from `requirements.txt`) and reads/writes exclusively via `GO_BACKEND_INTERNAL_URL`.
- [ ] **AC-AI-02 (aligns with AC-13):** Any query with `channel == "whatsapp"` asking for results, grades, CBC levels, or teacher comments returns zero academic data and includes the app link.
- [ ] **AC-AI-03 (aligns with AC-23):** `POST /api/v1/internal/payments/stk-push` is never called unless the session state is `AWAITING_CONFIRMATION` for the exact `(student_id, amount, mpesa_phone)` tuple and the user's current message matches the affirmative confirmation pattern.
- [ ] **AC-AI-04 (aligns with AC-26):** When the Go backend indicates a 24-hour chat payment pause due to a phone number change, `chat_payment.py` blocks the STK push and informs the parent with the school's contact details.
- [ ] **AC-AI-05 (Fact Validation):** Post-generation regex/number verification ensures that if the LLM output contains any currency amount that does not appear in the Go `fee-balance` JSON response, the service discards the LLM text and uses the deterministic template.
- [ ] **AC-AI-06 (No Ranking / No Diagnosis Guardrail):** `report_drafter.py` rejects/sanitizes any draft containing forbidden terms (`rank`, `position`, `1st`, `last`, `top`, `bottom`, clinical diagnoses) and guarantees `requires_teacher_approval: true`.
- [ ] **AC-AI-07 (Session Lookup Cap - C66):** Exceeding 15 read lookups or 3 payment starts in a single `session_id` returns a rate-limit response with school contact details without calling the Go backend.

---

## 7. Testing Requirements

All tests must be written using `pytest` and `httpx.AsyncClient` (`respx` or `unittest.mock` for mocking Go internal API and Mistral LLM responses) under `ai_service/tests/` and pass via `make test`:
1. **Unit Tests (`test_security_and_limits.py`):**
   - Valid `X-Internal-Service-Token` + valid HMAC `X-User-Context` passes.
   - Expired `exp` timestamp, tampered `school_id` in context, or wrong token returns `401`.
   - Session rate limiter blocks the 16th read lookup and the 4th payment initiation for the same `session_id`.
2. **Unit & Integration Tests (`test_parent_assistant.py`):**
   - Fee balance query in-app and on WhatsApp returns exact figures from mocked Go endpoint.
   - Attendance query returns exact present/absent counts.
   - Results query on WhatsApp (`channel="whatsapp"`) is blocked and returns the portal URL (`AC-13`).
   - Out-of-scope query returns polite refusal + school contact phone/email (`C57`).
   - Simulated LLM hallucination (LLM invents `KES 9,999` when Go returned `KES 12,500`) triggers the deterministic fallback template so `KES 12,500.00` is returned.
   - Simulated LLM timeout (`httpx.ConnectTimeout`) still returns accurate facts via the deterministic fallback template.
3. **Multi-Turn Payment Flow Tests (`test_chat_payment.py`):**
   - Turn 1: `"Pay 5000 for Joy to 0712345678"` -> Returns confirmation prompt; asserts `stk-push` endpoint was **not** called (`call_count == 0`).
   - Turn 2a: `"No"` -> Cancels pending payment; asserts `stk-push` was **not** called.
   - Turn 2b: `"Yes"` -> Calls `POST /api/v1/internal/payments/stk-push` with `parent_confirmed: true`, `amount: "5000.00"`, `mpesa_phone: "+254712345678"`.
   - Overpayment test: Balance is `KES 2,000.00`, parent pays `KES 5,000.00` -> Allowed and confirmed cleanly.
   - 24h phone-change lock test: Mock Go returning `403 CHAT_PAYMENT_PAUSED_PHONE_CHANGE` -> Assistant explains the 24-hour security pause.
4. **Report Drafter Tests (`test_report_drafter.py`):**
   - Generates a valid CBC comment from `EE/ME/AE/BE` payload.
   - Generates a valid re-release teacher note from `teacher_reason`.
   - Verifies forbidden ranking/diagnostic words are stripped or replaced.

---

## 8. Integration Responsibilities

- **With Developer 5 (Ian — @frihk – Backend Engineer C):**
  - Ian (@frihk) implements `backend/internal/internalapi/internalapi.go` and `backend/internal/whatsapp/whatsapp.go`.
  - Coordinate with Ian (@frihk) on the exact HMAC signing algorithm (`HMAC-SHA256` over the base64url-encoded JSON string of `X-User-Context` using `INTERNAL_SERVICE_TOKEN`).
  - Both services already share `INTERNAL_SERVICE_TOKEN` / `AI_SERVICE_SECRET_TOKEN` in `docker-compose.yml` and `.env.example`.
- **With Developer 4 (Ema — @ejao-000 – Backend Engineer B):**
  - Ema (@ejao-000) implements `POST /api/v1/report-comments/draft` in `backend/internal/academics`, which forwards teacher drafting requests to your `POST /api/v1/ai/report-comments/draft` endpoint and reads internal student mark summaries via `GET /api/v1/internal/students/{id}/marks-summary`.
- **With Developer 2 (Ashy — @Ashomondi – Frontend Engineer):**
  - Ashy's (@Ashomondi) Parent Portal chat component and Teacher Portal report comment modal consume the JSON structures returned by your endpoints (proxied through the Go backend). Ensure `payment_state` (`NONE`, `AWAITING_SLOTS`, `AWAITING_CONFIRMATION`, `STK_SENT`) is always included in `parent-query` responses so Ashy's UI can render clear confirmation badges/buttons.

---

## 9. File Ownership and Conflict Prevention

| Category | Paths | Rule |
|---|---|---|
| **Exclusively Owned by Sos (@sospeter-57, Dev 1)** | `ai_service/**` (all files in `ai_service/app/`, `ai_service/tests/`, `ai_service/requirements.txt`, `ai_service/Dockerfile`) | Full read/write ownership. No other developer may modify `ai_service/`. |
| **Shared Read-Only / Coordinate Before Edit** | `Docs/API_CONTRACTS.md`, `.env.example`, `docker-compose.yml` | Propose changes in PR and tag affected developers before merging. |
| **Strictly Do Not Modify** | `backend/**`, `frontend/**`, `migrations/**` | Owned by Evans (@eojuma), Ema (@ejao-000), Ian (@frihk), and Ashy (@Ashomondi). Never modify Go, SQL, or Next.js files. |

---

## 10. Definition of Done

- [ ] All endpoints (`POST /api/v1/ai/parent-query`, `POST /api/v1/ai/report-comments/draft`, `GET /healthz`) are implemented and match the JSON contracts in Section 4.
- [ ] `X-Internal-Service-Token` and HMAC-signed `X-User-Context` verification is enforced on every AI route.
- [ ] Deterministic fallback templates work when the local Mistral server (`Ollama`/`vLLM`) is unavailable, ensuring local/CI tests pass with or without a GPU.
- [ ] Post-generation numeric validator prevents any altered fee balance or attendance count from being sent to a parent.
- [ ] WhatsApp minimal-fact rule (`AC-13`) and two-step STK push confirmation (`AC-23`) are verified by automated tests.
- [ ] Every AI request dispatches an audit record to `POST /api/v1/internal/ai/audit-log` (`AC-28`).
- [ ] `pytest` suite in `ai_service/tests/` passes with 100% pass rate and `docker compose up --build ai_service` starts cleanly on port `8000`.

