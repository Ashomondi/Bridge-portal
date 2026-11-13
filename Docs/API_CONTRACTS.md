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
bound to the user and active membership. For verified chat, Go resolves the
token hash to an unexpired `chat_sessions` row and rechecks guardian access.
For in-app requests, Go validates the user's existing session and membership.
Unsigned user/school headers are never delegation credentials.

Both credentials are required; the service token alone grants no student or
school access. Go verifies each request, derives its principal, and applies the
same permissions as a frontend request. Restrict internal routes to the private
service network, use TLS, rotate service tokens, and redact both headers from
logs. Production must reject the example development secrets. This decision
defines the required protocol; credential validation and delegation middleware
are not implemented in the current scaffolding.

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
- `GET /students/{id}/fee-balance` (parent-scoped, or bursar)
- `POST /payments/stk-push` — AI service calls this after confirming child +
  amount with the parent; never initiates without that confirmation step
- `GET /payments/unmatched` (bursar) — feeds the AI reconciliation feature
- `POST /payments/{id}/match` (bursar confirms an AI-suggested match)
- `POST /payment-plans/{id}/reminder-check` — AI service reads pledge vs.
  actual payments and arrears, triggers a reminder job (does not send
  directly — enqueues via the notification system)

### Notifications
- `POST /notifications/send` (internal use by jobs, not directly callable by
  AI service without going through a job)

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

Pagination uses positive integers: `page` defaults to **1**, maximum
**1,000,000**; `per_page` defaults to **25**, maximum **100**. `total` is the
number of records matching the caller's authorized filters before pagination.
Endpoints must use a stable ordering with an ID tie-breaker. Empty/out-of-range
pages return an empty collection, not an error.

`Decode` limits each JSON request body to **1 MB (1,048,576 bytes)**, rejects
unknown fields, and accepts exactly one JSON value. Oversized bodies currently
return `400 INVALID_JSON`, not `413`. Handlers must call this helper to enforce
the limit. Use canonical, nonzero UUID strings; timestamps use RFC 3339 with an
explicit offset (storage in UTC), dates use `YYYY-MM-DD`, and school-local
presentation uses `Africa/Nairobi`. Financial DTOs should use decimal strings
with two fractional digits in KES to avoid floating-point loss.

## What's not yet decided

- Rate limiting for AI-facing endpoints (a parent-assistant loop must not be
  able to hammer fee-balance lookups).
