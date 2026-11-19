# API_CONTRACTS.md

How the frontend, Go backend, and AI service agree on interfaces. Update this
as real endpoints are built — this is meant to track actual contracts, not
aspirational ones. Once the project has a real number of endpoints, consider
generating this from an OpenAPI spec instead of maintaining it by hand.

## Auth model

- Every authenticated request carries the caller's identity (user + active
  membership) and resolves to a `school_id`.
- The Go backend sets `app.school_id` for the request's database transaction
  so RLS does the enforcement — see `ARCHITECTURE.md`.
- Every endpoint declares which permission(s) it requires. A request without
  the right permission (via the caller's badges) is rejected before it
  reaches any handler logic — not filtered afterwards.
- Parent-facing endpoints additionally scope to "children this parent is a
  guardian of" — never just "this school."

### Tenant and permission boundary

`RequireIdentity` accepts an application-supplied `Identity` callback that must
verify the session, active membership, and current school against trusted auth
state. It validates the returned nonzero UUIDs and role; it does not itself
verify credentials or grant permissions. `TenantTxFunc.WithTenantTx(ctx,
principal, fn)` derives the transaction school from `Principal.SchoolID`.
Pass the principal returned by `RequireIdentity`, never one assembled from
request JSON, query parameters, or `X-School-ID` / `X-User-ID` headers. The
underlying application transaction function must use that school in
transaction-local `app.school_id`, execute the callback, and commit or roll back.

Roles describe membership; they are not a blanket permission grant. Endpoint
authorization must evaluate the required badge permissions, their class/club
scope and validity window, plus resource ownership. Parent access requires a
current `student_guardians` relationship in the same school. Fee management and
manual matching require bursar permissions; payment initiation also requires
explicit confirmation of child and amount. RLS restricts schools, not these
within-school permissions. Missing or invalid identity is `401 UNAUTHENTICATED`;
insufficient permission is `403 FORBIDDEN`. Return `404 NOT_FOUND` for a resource
outside the caller's authorized scope without disclosing whether it exists.

These are integration requirements. The shared contracts implement identity
shape validation and principal-derived tenant selection; the server, auth,
permission middleware, and database pool are still scaffolding. `AuditLogger`
is only an interface: AC-29 is **not addressed** by this foundation. Production
handlers must wire durable, redacted audit events into the same transaction as
the protected change and propagate logging failure before claiming AC-29.

### AI-service authentication decision

Internal calls require both a service credential and a delegated user session.
Use `Authorization: Bearer <service-token>` for the shared secret configured as
`INTERNAL_SERVICE_TOKEN` in Go and `AI_SERVICE_SECRET_TOKEN` in Python. Use
`X-Delegated-Session: <opaque-session-token>` for a Go-issued, short-lived session
bound to the user and active membership. For verified chat, migration `000004`
enforces that `chat_sessions` rows in `VERIFIED` state carry a non-null,
non-blank `access_token_hash` alongside `guardian_user_id`, `verified_at`, and
`expires_at > verified_at`; Go resolves the hash of `X-Delegated-Session` to an
unexpired (`expires_at > now()`) `chat_sessions` row and rechecks guardian
access. For in-app requests, Go validates the user's existing session and
membership. Unsigned user/school headers are never delegation credentials.

Both credentials are required; the service token alone grants no student or
school access. Go verifies each request, derives its principal, and applies the
same tenant, badge-permission, and guardian-relationship checks as a frontend
request. The AI service is restricted to permission-checked Go endpoints and
never accesses PostgreSQL, raw biometric templates, or nurse clinical records.
Restrict internal routes to the private service network, use TLS, rotate
service tokens, and redact both headers from logs. Production must reject the
example development secrets. This decision defines the required protocol;
credential validation and delegation middleware are not implemented in the
current scaffolding.

## Rule: the AI service never touches the database directly

The AI service (Python/FastAPI) has no direct database credentials. It:

1. Receives a request already carrying the caller's identity and permission
   context (forwarded by the Go backend, not re-derived by the AI service).
2. Calls specific, permission-checked Go endpoints to fetch facts (fee
   balance, attendance record, syllabus status, etc.).
3. Uses the model only to phrase, translate, summarise, or map — never to
   originate a fact that should come from the database.
4. Any action with a real-world effect (starting an STK push, updating
   syllabus coverage, posting a reminder) calls a Go endpoint to perform it;
   the AI service does not write to storage on its own.

This is the enforcement point for "AI never invents a fee balance" and "AI
never handles money directly" from `AGENTS.md` and `SECURITY.md`.

## Endpoint groups (fill in as built)

> Each group below should end up as a real section with method, path,
> request/response shape, and required permission. Skeleton only for now.

### Auth & onboarding
- `POST /auth/otp/request`, `POST /auth/otp/verify`
- `POST /staff-invites` (permission: `staff.invites.create`)
- `POST /staff-applications/{id}/review` (permission:
  `staff.applications.review`)
- `POST /students/admit` (permission: `students.admit`)

### Academics
- `GET /classes/{id}/roster` (permission: `class.students.view`)
- `GET /classes/{id}/stats` (permission: `class.stats.view` — means only, no
  names)
- `POST /syllabus/coverage` — called by AI service after mapping a teacher's
  natural-language input to CBC strands/sub-strands
- `GET /students/{id}/performance-history` (parent/teacher, scoped)

### Fees & payments

Handler implementation status: `billing.RegisterRoutes` implements the four
endpoints below (`POST /api/v1/billing/fee-structures`,
`POST /api/v1/billing/fee-structures/charge-current`,
`GET /api/v1/students/{id}/fee-balance`, and `GET /api/v1/students/{id}/ledger`)
in `backend/internal/billing`, but route mounting in `backend/cmd/server/main.go`
remains pending application composition. All monetary fields serialize as
fixed two-decimal KES strings (e.g. `"10000.00"`).

- `POST /api/v1/billing/fee-structures`
  - **Auth & tenant scope:** Requires a valid `contracts.Principal` via
    `contracts.RequireIdentity` (`401 UNAUTHENTICATED`). `Service.LoadFeeStructure`
    binds the transaction solely to `Principal.SchoolID` via `contracts.TenantDB.WithTenantTx`.
  - **Permission:** Requires an active non-parent `memberships` row (`status = 'active'`)
    with an active, unexpired school-wide badge granting `billing.fee_structures.manage`
    (`403 FORBIDDEN` otherwise). Requires a configured `AuditLogger` (`500 AUDIT_LOGGER_REQUIRED` if nil).
  - **Request body (`LoadFeeStructureInput`, `Decode` unknown fields rejected):**
    `{academic_year_id, term_id, grade_level_id, boarding_status?, items: [{title|name, amount, is_compulsory?}]}`.
  - **Validation & variants:** Validates that `academic_year_id`, `term_id`
    (belonging to `academic_year_id`), and `grade_level_id` exist in `Principal.SchoolID`;
    rejects empty items, duplicate case-insensitive item titles, non-positive amounts,
    >2 decimal places, and overflow beyond `NUMERIC(12,2)` (`422`).
    In Release 1, `boarding_status` defaults to `"all"` and rejects `"boarder"` / `"day"`
    (`422 UNSUPPORTED_BOARDING_VARIANT`) or unknown values (`422 INVALID_BOARDING_STATUS`).
  - **Idempotency, conflict & charging:** Upserts the term `fee_periods` row.
    If an identical structure already exists for `(school_id, academic_year_id, grade_level_id, term_id, 'all')`,
    returns it idempotently; if items/amounts differ, returns `409 FEE_STRUCTURE_CONFLICT`.
    Whenever `academic_years.is_current` or `terms.is_current` is true (on initial insert
    or identical replay), triggers idempotent `ChargeGradeForTerm` for active enrolled students.
    Persists a `billing.fee_structure.load` audit event in the same transaction.
  - **Success (`201`):** `{"status":"success","data":{id, school_id, academic_year_id, term_id, grade_level_id, fee_period_id, boarding_status, due_date, total_amount, items, created_at}}`.
- `POST /api/v1/billing/fee-structures/charge-current`
  - **Auth & permission:** Same `RequireIdentity` + `billing.fee_structures.manage` badge check + mandatory `AuditLogger`.
  - **Behavior:** Idempotently charges all preloaded `'all'`-variant fee structures in
    `Principal.SchoolID` whose academic year or term is now `is_current = true`.
    Emits `billing.fee_structure.charge_current` audit event.
  - **Success (`200`):** `{"status":"success","data":{"structures_processed":<int>,"charges_created":<int>}}`.
- `GET /api/v1/students/{id}/fee-balance`
  - **Auth & tenant scope:** Requires `contracts.RequireIdentity` (`401 UNAUTHENTICATED`).
    Validates UUID `{id}` (`400 INVALID_STUDENT_ID`). Looks up student strictly in
    `Principal.SchoolID` (`404 STUDENT_NOT_FOUND` when missing or in another school).
  - **Visibility:** Allowed for either (a) an active non-parent staff/bursar membership
    holding an active school-wide `billing.ledgers.view` badge, or (b) an active
    guardian linked in `student_guardians` with `is_fee_payer = TRUE`. Non-fee-paying
    guardians (`is_fee_payer = FALSE`), unlinked guardians, and staff without
    `billing.ledgers.view` receive `403 FORBIDDEN`.
  - **Success (`200`):** `{"status":"success","data":{student_id, admission_number, student_name, opening_balance, billed_amount, paid_amount, current_balance, credit_balance, due_date, terms: [{academic_year_id, year, term_id, term_number, due_date, opening_balance, billed_amount, paid_amount, credit_applied, credit_transferred, current_balance, credit_balance}], items, fee_change_notes}}`.
    Aggregate balances equal the sum across preserved per-term `fee_ledgers` rows (`terms`).
- `GET /api/v1/students/{id}/ledger`
  - **Auth & visibility:** Identical authentication, tenant isolation, `billing.ledgers.view` / `is_fee_payer = TRUE` rules as `fee-balance`.
  - **Success (`200`):** Returns the `StudentFeeBalance` fields plus `entries: [{id, term_id, entry_type, amount, reference, description, created_at}]`
    ordered deterministically by `created_at ASC, id ASC` (`entry_type` in `charge | payment | credit_applied | adjustment`).
  - **Note on pagination:** Currently returns the student's full entry list without query pagination parameters.
- `POST /api/v1/payments/stk-push`
  - **Auth & tenant scope:** Requires `contracts.RequireIdentity` (`401 UNAUTHENTICATED`). Student lookup and M-Pesa config resolution are strictly scoped to `Principal.SchoolID` (`404 STUDENT_NOT_FOUND` across schools).
  - **Guardian eligibility:** Requires an active `student_guardians` link with `is_fee_payer = TRUE`, active `memberships.status = 'active'`, and `users.is_active = TRUE` (`403 FORBIDDEN` otherwise).
  - **Request schema:** `{"student_id":"<uuid>","amount":"<whole KES string/number >= 1.00>","phone_number":"<optional E.164/KE phone>","source":"web|chat|whatsapp|in_app","parent_confirmed":true|false,"idempotency_key":"<optional <=128 chars>"}` (also accepts `Idempotency-Key` HTTP header).
  - **Amount precision:** Release 1 STK Push accepts **whole KES shillings >= 1.00 KES** only (`amount >= 100 && amount % 100 == 0` in minor units); fractional KES, zero, or negative values return `422 INVALID_PAYMENT_AMOUNT`. Overpayments above the student's current balance are permitted and carried forward as `credit_balance`.
  - **Chat-initiated payment controls (`source` in `chat | whatsapp | in_app`):**
    - Mandatory explicit `"parent_confirmed": true`; omitted or `false` returns `422 PARENT_CONFIRMATION_REQUIRED`.
    - Enforces the 24-hour phone-change freeze (`users.chat_payment_paused_until > NOW()` → `403 CHAT_PAYMENT_PAUSED_PHONE_CHANGE`). Fails closed (`503 CHAT_PAYMENT_FREEZE_UNAVAILABLE`) if the freeze state cannot be read.
    - Requires a valid registered phone number (`422 INVALID_REGISTERED_PHONE`) and rejects any destination phone override (`422 CHAT_PHONE_OVERRIDE_FORBIDDEN`).
  - **Idempotency:** Scoped to `(school_id, guardian_user_id, idempotency_key)`. Identical retries return the existing STK attempt without re-invoking Daraja; reusing a key with different student, amount, phone, or source returns `409 IDEMPOTENCY_KEY_CONFLICT`; concurrent in-flight attempts return `409 STK_REQUEST_IN_PROGRESS`.
  - **Mandatory dependencies:** Requires non-nil `AuditLogger`, `InAppNotifier`, `DarajaProvider`, and `STKPollQueue` (fails closed if unconfigured). Sets `AccountReference = student.admission_number` and enqueues a 45-second status poll (`poll_scheduled` reflects queue status; queue failures are durably audited and recoverable via `RetryUnscheduledSTKPolls`).
  - **Success (`202`):** `{"status":"success","data":{stk_request_id, student_id, admission_number, account_reference, amount, phone_number, checkout_request_id, merchant_request_id, status, customer_message, poll_scheduled}}`.
- `POST /api/v1/billing/mpesa/callback` and `POST /api/v1/billing/mpesa/callback/{callback_token}` *(Safaricom Daraja STK Webhook — flat Daraja JSON, NOT the standard app envelope)*
  - **Authentication (fail-closed):**
    - Per-request path route (`/{callback_token}`): Safaricom invokes the server-generated 64-hex-char token embedded in the `CallBackURL` path segment during `InitiateSTKPush`. The backend hashes the token (SHA-256), validates it in constant time against `mpesa_stk_requests.callback_token_hash`, and verifies `CheckoutRequestID` binding.
    - Static route (`/api/v1/billing/mpesa/callback`): Requires a configured `VerifyCallback` hook OR a trusted reverse proxy / API gateway whose `RemoteAddr` is in `TrustedCallbackProxies`, strips client auth headers, and injects `X-Callback-Token` (or `Authorization: Bearer`) matching `MPESA_CALLBACK_TOKEN`. Query-string tokens are always rejected (`401`).
  - **Tenant resolution:** Resolved strictly from trusted database state (`mpesa_stk_requests` via `callback_token_hash` or globally unique `CheckoutRequestID`), never from caller-supplied `school_id` query parameters or `BusinessShortCode`. If `BusinessShortCode` is present, it is verified against `school_mpesa_configs.shortcode`; if `MerchantRequestID` is stored and present, it is verified to match.
  - **Amount & state integrity:** Successful callbacks (`ResultCode == 0`) require both `MpesaReceiptNumber` and a valid positive `Amount` matching the initiated `mpesa_stk_requests.amount` exactly (`422 MPESA_AMOUNT_MISMATCH` otherwise, with zero ledger mutation and a rejection audit entry). Provider-confirmed success may settle `initiating`, `pending`, `timeout`, `uncertain`, or earlier `failed` attempts; duplicate successes are idempotent no-ops; late failure callbacks never overwrite `succeeded`.
  - **Response schema (Daraja format — intentionally distinct from standard app envelope):**
    - Accepted / duplicate (`200`): `{"ResultCode":0,"ResultDesc":"Accepted"}`
    - Rejected / unauthenticated / invalid (`400`/`401`/`404`/`409`/`422`/`500`): `{"ResultCode":1,"ResultDesc":"<safe error description>"}`
- `POST /api/v1/billing/mpesa/c2b-confirmation` *(Safaricom Daraja C2B Paybill Webhook — flat Daraja JSON)*
  - **Authentication:** Same fail-closed trusted reverse proxy (`TrustedCallbackProxies` + `X-Callback-Token` / `Bearer`) or `VerifyCallback` hook as the static STK callback route. Query-string tokens are rejected (`401`).
  - **Scope & tenant resolution:** Release 1 supports **Paybill** (`Pay Bill` / `CustomerPayBillOnline`) only; Till (`Buy Goods` / `CustomerBuyGoodsOnline`) is rejected (`422 UNSUPPORTED_MPESA_SHORTCODE_TYPE`). Resolves `school_id` strictly from `BusinessShortCode` via `ResolveShortcode` and verifies against `school_mpesa_configs`.
  - **Matching & idempotency:** Matches `UPPER(TRIM(students.admission_number)) = UPPER(TRIM(BillRefNumber))` strictly within the resolved `school_id`. Matched payments credit the ledger and notify `is_fee_payer = TRUE` guardians; unknown admission numbers persist into `payments` (`unmatched`) + `unmatched_payments` without discarding money. Duplicate `TransID` receipts are idempotent no-ops.
  - **Response schema (Daraja format):** `{"ResultCode":0,"ResultDesc":"Accepted"}` on `200`, or `{"ResultCode":1,"ResultDesc":"<safe error description>"}` on rejection.
- `GET /api/v1/payments/unmatched`
  - **Auth & permissions:** Requires `contracts.RequireIdentity` (`401 UNAUTHENTICATED`) and an active non-parent school membership holding an active school-wide badge granting `billing.unmatched_payments.manage`, `billing.ledgers.view`, or `billing.fee_structures.manage` (`403 FORBIDDEN` otherwise). Strictly tenant-scoped to `Principal.SchoolID`.
  - **Pagination:** Standard `page` / `per_page` (`contracts.Pagination`).
  - **Success (`200`):** Standard paginated envelope with `data: [{id, payment_id, transaction_ref, amount, payer_phone, payer_name, raw_account_ref, status, created_at}]`.
- `POST /api/v1/payments/{id}/match`
  - **Auth & permissions:** Requires `contracts.RequireIdentity` (`401 UNAUTHENTICATED`), valid UUID `{id}` (`400 INVALID_PAYMENT_ID`), and an active non-parent school membership holding an active school-wide badge granting `billing.unmatched_payments.manage` or `billing.fee_structures.manage` (`403 FORBIDDEN` for teachers and read-only `billing.ledgers.view` staff).
  - **Request schema:** `{"student_id":"<uuid>","resolution_note":"<optional string>"}`.
  - **Behavior:** Locks the `unmatched_payments` row `FOR UPDATE` in `Principal.SchoolID` (`404 UNMATCHED_PAYMENT_NOT_FOUND` across schools; `409 PAYMENT_ALREADY_RESOLVED` if already resolved), validates target student belongs to `Principal.SchoolID` (`422 STUDENT_NOT_FOUND` otherwise), atomically updates `unmatched_payments` + `payments`, credits the student ledger, notifies `is_fee_payer = TRUE` guardians, and writes a `billing.payment.manual_match` audit event.
  - **Success (`200`):** `{"status":"success","data":{payment_id, unmatched_id, student_id, transaction_ref, amount, status:"resolved", resolved_at}}`.
- `POST /payment-plans/{id}/reminder-check` *(planned/unimplemented on this branch)* — AI service reads pledge vs. actual payments and arrears, triggers a reminder job (does not send directly — enqueues via the notification system)

### Notifications

Handler implementation status: `notifications.RegisterRoutes` implements the
two HTTP handlers below in `backend/internal/notifications/http.go`, but they
are **not yet mounted** in `backend/cmd/server/main.go` (still a placeholder
pending auth/session middleware and server composition).

- `GET /api/v1/notifications`
  - **Auth & tenant scope:** Requires a valid `contracts.Principal` via
    `contracts.RequireIdentity` (`401 UNAUTHENTICATED` when missing or invalid).
    `Service.List` passes the `Principal` to `contracts.TenantDB.WithTenantTx`,
    deriving `school_id` solely from `Principal.SchoolID` (query/header school
    overrides are ignored).
  - **Recipient access:** Returns only rows in `in_app_notifications` where
    `school_id = Principal.SchoolID` and `recipient_user_id = Principal.UserID`,
    ordered deterministically by `created_at DESC, id`.
  - **Pagination:** Offset-based via `contracts.Pagination` (`page` default `1`,
    max `1,000,000`; `per_page` default `25`, max `100`). Invalid values return
    `422 INVALID_PAGINATION`. Out-of-range pages return `200` with `data: []`
    and `meta.total` set to the caller's total notification count.
  - **Success (`200`):** Uses the shared paginated envelope (`status: "success"`,
    `data` array of `{id, category, title, body, metadata, is_read, read_at, created_at}`
    with UTC RFC 3339 timestamps, and `meta: {page, per_page, total}`), with
    `Content-Type: application/json` and `Cache-Control: no-store`.
  - **Errors:** `401 UNAUTHENTICATED`, `422 INVALID_PAGINATION`,
    `500 INTERNAL_ERROR`.
- `PATCH /api/v1/notifications/{id}/read`
  - **Auth & tenant scope:** Requires a valid `contracts.Principal` via
    `contracts.RequireIdentity` (`401 UNAUTHENTICATED` when missing or invalid)
    and derives `school_id` solely from `Principal.SchoolID` via
    `contracts.TenantDB.WithTenantTx`.
  - **Ownership & tenant isolation:** Updates only where
    `school_id = Principal.SchoolID AND recipient_user_id = Principal.UserID AND id = {id}`,
    additionally enforced by PostgreSQL forced RLS. Notices belonging to another
    user or another school return `404 NOTIFICATION_NOT_FOUND` without disclosing
    existence.
  - **Request body:** None required.
  - **Idempotent read timestamp:** Sets `is_read = true` and
    `read_at = coalesce(read_at, now())`, so repeat calls succeed with `200`
    and preserve the original `read_at` timestamp.
  - **Success (`200`):** `{"status":"success","data":{"id":"<uuid>","is_read":true}}`
    with `Cache-Control: no-store`.
  - **Errors:** `400 INVALID_NOTIFICATION_ID` (non-UUID/zero `{id}`),
    `401 UNAUTHENTICATED`, `404 NOTIFICATION_NOT_FOUND`, `500 INTERNAL_ERROR`.
- `POST /notifications/send` *(planned/unimplemented — not built in
  `internal/notifications`; internal job-queue scheduling remains pending in
  `internal/jobs` and `cmd/worker`)*.

### AI-service-facing (internal only — never exposed to frontend directly)
- `POST /internal/ai/parent-query` — resolves a parent's natural-language
  question by calling the appropriate read endpoints above
- `POST /internal/ai/blog-safety-check` — checks a draft post for children's
  personal details before a patron publishes

## Response conventions

The helpers in `backend/internal/internalapi/contracts` define these envelopes
for consumers that adopt them. Existing handlers must be checked individually;
this is not a claim that every endpoint already uses these helpers. New routes
use the `/api/v1` prefix; the endpoint sketches above omit that prefix.

Success (`200`, or `201` for creation):

```json
{"status":"success","data":{"id":"10000000-0000-0000-0000-000000000001"}}
```

Paginated success (empty collections use `[]`):

```json
{"status":"success","data":[],"meta":{"page":1,"per_page":25,"total":0}}
```

Error (the HTTP status is not repeated inside `error`):

```json
{"status":"error","error":{"code":"FORBIDDEN","message":"Access denied.","details":{}}}
```

`data` is the endpoint payload; `meta` is omitted for non-paginated responses.
`details` is an object for safe, structured validation information. Never expose
SQL errors, secrets, provider payloads, or another tenant's identifiers. Unknown
errors become the generic `500 INTERNAL_ERROR` response. Helpers set
`Content-Type: application/json` and `Cache-Control: no-store`.

| HTTP | Code | Meaning / implementation |
|---|---|---|
| 400 | `INVALID_JSON` | `Decode`: malformed JSON, unknown fields, multiple JSON values, or body over the limit |
| 401 | `UNAUTHENTICATED` | `RequireIdentity` / tenant adapter: absent or invalid principal; auth middleware must use this for invalid/expired credentials |
| 403 | `FORBIDDEN` | Handler permission denial; emitted using `Fail` |
| 404 | `NOT_FOUND` | Handler resource lookup outside authorized scope; emitted using `Fail` |
| 409 | `CONFLICT` | Handler rejects a duplicate or invalid state transition; emitted using `Fail` |
| 422 | `INVALID_PAGINATION` | `Pagination`: out-of-range or non-integer page values |
| 422 | `VALIDATION_ERROR` | Handler field/domain validation; emitted using `Fail` |
| 429 | `RATE_LIMITED` | Reserved for rate-limit middleware, not implemented yet |
| 500 | `INTERNAL_ERROR` | Unexpected error or response serialization failure |

Pagination is **offset-based** (not cursor-based) using positive integer query
parameters: `page` defaults to **1** (maximum **1,000,000**) and `per_page`
defaults to **25** (maximum **100**), where SQL queries compute
`OFFSET = (page - 1) * per_page` and `LIMIT = per_page`. Non-integer,
non-positive, or out-of-range values return `422 INVALID_PAGINATION`. `total` in
`meta` is the number of records matching the caller's authorized filters before
pagination. Endpoints must use a stable ordering with an ID tie-breaker.
Empty/out-of-range pages return an empty `[]` collection with `200`, not an
error.

`Decode` limits each JSON request body to **1 MB (1,048,576 bytes)** via
`http.MaxBytesReader`, rejects unknown fields (`DisallowUnknownFields`), and
accepts exactly one JSON value. Oversized bodies currently return
`400 INVALID_JSON`, not `413`. Because `Decode` is an explicit per-handler
helper rather than global server middleware in the current scaffolding, handlers
must call `Decode` on every JSON endpoint to enforce the limit consistently.
Use canonical, nonzero UUID strings (`ValidID`); timestamps use RFC 3339 with an
explicit offset (storage in UTC), dates use `YYYY-MM-DD`, and school-local
presentation uses `Africa/Nairobi`. Financial DTOs should use decimal strings
with two fractional digits in KES to avoid floating-point loss.

## What's not yet decided

- Rate limiting for AI-facing endpoints (a parent-assistant loop must not be
  able to hammer fee-balance lookups).
