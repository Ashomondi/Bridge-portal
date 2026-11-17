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

## AI service authentication (decided 2026-10-29)

The AI service authenticates to the Go backend with a **service token plus a
short-lived signed user context** (user, school, membership, and, for
WhatsApp, the verified session). Every call carries both; a valid token
without a valid, unexpired context is rejected. A hybrid with mutual TLS may
be added later.

## Rate limits

- **OTP:** one request per minute; more than five requests in thirty minutes
  blocks the *requesting WhatsApp number* for 24 hours (not the guardian's
  account). Applying the same limits to app login is undecided.
- **AI-facing endpoints:** limits are on backend lookups per verified
  session (fee-balance, attendance, payment lookups), with a tighter limit on
  payment starts. Numbers are set by the team and QA. Assistant messages
  themselves are not capped.

## Endpoint groups (fill in as built)

> Each group below should end up as a real section with method, path,
> request/response shape, and required permission. Skeleton only for now.
> No endpoint exists yet; this tracks the contracts agreed so far.

### Auth & onboarding
- `POST /auth/otp/request`, `POST /auth/otp/verify`
- `POST /staff-invites` (permission: `staff.invites.create`)
- `POST /staff-applications/{id}/review` (permission:
  `staff.applications.review`; verifier checks documents in person)
- `POST /parent-qr` (generate, valid 12 hours) and
  `POST /parent-qr/{id}/revoke` (permission: `parents.qr.manage`; dean,
  principal, deputy)
- `POST /registrations` (parent or staff submission via QR; creates a
  *pending* record with no data access)
- `POST /registrations/{id}/verify` (permission: `registrations.verify`;
  held by staff verifiers, delegable for the QR's active window)
- `POST /verifier-grants` (permission: `verifier.grant`; staff only,
  expires with the QR)
- `POST /students/admit` (permission: `students.admit`)
- `POST /guardians/me/phone-change` (2FA; notifies the old number; pauses
  chat payments for 24 hours)

### WhatsApp & parent sessions
- `POST /whatsapp/webhook` (receiving number resolves the school)
- `POST /chat/sessions` (admission number → masked guardian numbers)
- `POST /chat/sessions/{id}/otp` (OTP sent only to the chosen registered
  number); sessions last 6 hours

### Academics
- `GET /classes/{id}/roster` (permission: `class.students.view`)
- `GET /classes/{id}/stats` (permission: `class.stats.view` — means only, no
  names)
- `POST /classes/{id}/attendance` (class teacher; one daily roll-call)
- `PUT /classes/{id}/subjects/{sid}/marks` (assigned subject teacher only)
- `POST /classes/{id}/marks-assignments` (permission:
  `marks.reassign`; dean, principal, deputy)
- `POST /classes/{id}/results/release` (class teacher; rejected unless every
  subject has marks for every student)
- `POST /classes/{id}/results/reopen` and `.../re-release` (class teacher;
  re-release requires a teacher's note per corrected mark; notifies only
  affected parents)
- `POST /report-comments/draft` (AI service drafts; teacher edits and
  publishes)
- `GET/PUT /timetables`, `GET/PUT /calendar` (dean; term-date changes also
  trigger the fee-change notice)
- `POST /syllabus/coverage` — called by AI service after mapping a teacher's
  natural-language input to CBC strands/sub-strands (Release 2)
- `GET /students/{id}/performance-history` (parent/teacher, scoped; returns
  CBC levels per subject per term, never rank)

### Fees & payments
- `GET /students/{id}/fee-balance` (parent-scoped, or Ikonex team in
  Release 1)
- `POST /payments/stk-push` — AI service calls this only after the parent has
  confirmed child, amount, and the typed M-Pesa number; any amount accepted,
  excess held as credit
- `POST /fee-changes` (Ikonex team in Release 1; note required; notifies
  only affected parents)
- `GET /payments/unmatched` (Ikonex team in Release 1; bursar in Release 2)
  — feeds the AI reconciliation feature in Release 2
- `POST /payments/{id}/match` (team in Release 1; bursar confirms an
  AI-suggested match in Release 2)
- `POST /payment-plans/{id}/reminder-check` — Release 2; AI service reads
  pledge vs. actual payments and arrears, triggers a reminder job (does not
  send directly — enqueues via the notification system)

### Notifications
- `POST /notifications/send` (internal use by jobs, not directly callable by
  AI service without going through a job). Release 1 event notifications are
  in-app only; OTPs and the old-number SMS are transactional.

### AI-service-facing (internal only — never exposed to frontend directly)
- `POST /internal/ai/parent-query` — resolves a parent's natural-language
  question by calling the appropriate read endpoints above. On WhatsApp it
  returns only minimal facts and a link for anything else.
- `POST /internal/ai/blog-safety-check` — checks a draft post for children's
  personal details before a patron publishes (Release 2)

## Response conventions

> Fill in once decided: error shape, pagination style, date/time format
> (recommend ISO 8601, always with the school's timezone made explicit),
> versioning approach.

## What's not yet decided

- How OTP limits and blocks apply to app login.
- Response conventions above.
- Exact lookup-limit numbers for AI-facing endpoints.
