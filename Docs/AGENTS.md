# AGENTS.md

This file is read by AI coding assistants (Claude, or similar tools) before working
on this repository. Keep it short and current. If a rule here conflicts with what
you're about to do, stop and ask rather than guessing.

## What this project is

A multi-school portal for Kenyan secondary schools (CBC, Grade 7–12), hosted for
many schools on one platform. Parents, teachers, and school staff (nurse,
librarian, bursar, dean of studies, principal/deputy) each get a role-specific
view. An AI layer (self-hosted Mistral models) helps parents ask questions, pay
fees, and receive plain-language updates, and helps teachers and the bursar with
syllabus tracking and payment reconciliation.

See `ARCHITECTURE.md`, `REQUIREMENTS.md`, `API_CONTRACTS.md`, `SECURITY.md`, and
`DECISIONS.md` for the details behind everything summarised here.

## Stack

- **Backend:** Go (modular monolith — one service, packages per domain area:
  billing, academics, onboarding, notifications, etc.)
- **Database:** PostgreSQL, with row-level security (RLS) for tenant isolation
- **Frontend:** Next.js (React), built as a PWA for offline/low-connectivity use
- **AI service:** Python (FastAPI), calling self-hosted Mistral models
  (e.g. via vLLM or Ollama), kept as a separate service — never given direct
  database access
- **Jobs/cache:** Redis, with a Go job queue (e.g. Asynq) for SMS, STK push
  polling, and AI reconciliation/reminder jobs
- **Integrations:** Africa's Talking (SMS/USSD), WhatsApp Business Cloud API,
  Safaricom Daraja API (M-Pesa STK push, per-school Paybill/Till)

## Commands

- **Prerequisites:** Go **1.22.2+** (`backend/go.mod`) and **PostgreSQL 16**
  (`postgres:16-alpine`).
- Backend build & static check:
  `cd backend && go mod verify && go build ./... && go vet ./...`
- Backend unit tests (no database required):
  `cd backend && go test -count=1 -v ./internal/internalapi/contracts ./internal/academics`
  and SMS transport unit tests:
  `cd backend && go test -count=1 -v ./internal/notifications -run 'Test(AfricaTalking|SMS|Normalize|Sandbox|Transactional)'`
- PostgreSQL 16 integration tests: start an isolated PostgreSQL 16 test server
  (e.g., `docker compose up -d --wait postgres` from the repo root), configure
  `DEV05_TEST_DATABASE_URL`, and run all tests with `CI=true` so database tests
  cannot skip silently:
  ```sh
  export DEV05_TEST_DATABASE_URL='postgres://ikonex_admin:ikonex_secure_pass@localhost:5432/postgres?sslmode=disable'
  cd backend && CI=true go test -count=1 -v ./...
  ```
  Run notification integration tests specifically with:
  ```sh
  cd backend && CI=true DEV05_TEST_DATABASE_URL="$DEV05_TEST_DATABASE_URL" go test -count=1 -v ./internal/notifications
  ```
  Run billing integration tests specifically with:
  ```sh
  cd backend && CI=true DEV05_TEST_DATABASE_URL="$DEV05_TEST_DATABASE_URL" go test -count=1 -v ./internal/billing
  ```
  Adjust the URL for your local credentials/port. Use a dedicated local test
  server with a superuser capable of creating/dropping disposable `dev05_*`
  databases and non-superuser, non-`BYPASSRLS` roles; never run against a
  production or shared database. Missing `DEV05_TEST_DATABASE_URL` skips DB
  tests locally without `CI`, but fails when `CI` is nonempty.
  `.github/workflows/backend.yml` provisions PostgreSQL 16 and runs the full
  validation suite.
- Complete validation suite:
  ```sh
  cd backend && go mod verify && go build ./... && go vet ./... && CI=true DEV05_TEST_DATABASE_URL='postgres://ikonex_admin:ikonex_secure_pass@localhost:5432/postgres?sslmode=disable' go test -count=1 -v ./...
  ```
- Backend run (dev): `make backend-run` (entry point is still a placeholder)
- Frontend run (dev): `make frontend-run`
- Migrations: apply paired SQL files using a PostgreSQL migration runner in
  exact sequence **`000001` → `000003` → `000004` → `000005`**
  (`000001_init_schema` → `000003_r1_academics` →
  `000004_r1_billing_whatsapp_notifications` → `000005_r1_mpesa_hardening`).
  **Migration `000002` is absent from this branch.** `make migrate-up` in
  `Makefile` is advertised in help output but not implemented. The test harness
  applies the SQL files directly; see `README.md` for downgrade rules.
- AI service run (dev): `make ai-run`

## Non-negotiable rules

These override convenience or "helpfulness" in any single task. See
`SECURITY.md` for the full reasoning.

1. **Every query on a tenant-owned table must be scoped by `school_id`.**
   Row-level security is the enforcement layer — never write a query or
   migration that bypasses it or reads across schools.
2. **The AI service never touches the database directly.** It calls
   permission-checked Go endpoints and only ever sees data the requesting
   user is allowed to see. See `API_CONTRACTS.md`.
3. **Biometric data:** store encrypted templates only, never raw fingerprint
   images. Encryption keys are per-school.
4. **Health data (nurse records):** class teachers get a notification that a
   visit happened, never the clinical detail.
5. **No feature sends money or approves a payout on its own.** Bursar payments
   go through the signatory approval workflow; the AI proposes matches, a
   human confirms them.
6. **Nothing in scope generates a diagnosis, discipline record, or ranking of
   children.** See `REQUIREMENTS.md` for the full out-of-scope list.
7. **Student and staff personal data stays hosted in Kenya**, per the Data
   Protection Act. Don't add a dependency that routes personal data through a
   foreign service without checking `SECURITY.md` first.

## Conventions

> Expand this section as real conventions are established in the codebase —
> don't invent conventions here that the code doesn't actually follow.

- Go package layout: `TODO`
- Shared consumer contracts: `backend/internal/internalapi/contracts`;
  `TenantTxFunc.WithTenantTx` takes the principal from trusted `RequireIdentity`,
  never a school chosen from request input. See `API_CONTRACTS.md` for envelopes,
  limits and the distinction between contracts and pending runtime wiring.
- Error handling style: shared `Fail` / `WriteError` helpers redact unexpected
  failures. `AuditLogger` is only an interface; do not claim AC-29 is addressed
  without a wired durable implementation.
- Naming conventions (DB, API, frontend): `TODO`
- Commit message style: `TODO`
- Notifications package (`backend/internal/notifications`):
  - **SMS purpose restrictions (`C58`):** `SMSDispatcher` only accepts
    `app_login`, `whatsapp_session` (Release 3 reserved purpose key),
    `phone_change`, and `phone_change_notice`. Never add event notifications
    (`absence`, `sickbay`, results, timetable, calendar, fee reminders) to the
    Release 1 SMS transport; those go to `in_app_notifications`.
  - **Provider security:** Normalize phones via `NormalizePhone`; never log
    message bodies, OTPs, full phone numbers, `APIKey`, or raw provider
    responses; preserve `http.ErrUseLastResponse` redirect blocking, copied
    HTTP clients with explicit timeouts, 64 KiB response bounds, and safe
    `*SMSDeliveryError{Code, Retryable}` error classification.
  - **Tenant transactions:** `notifications.Service.List` and `MarkRead` must
    receive a `contracts.Principal` from `contracts.RequireIdentity` and pass it
    to `contracts.TenantDB.WithTenantTx(ctx, principal, fn)` — never pass a raw
    school ID string or accept school overrides from request inputs.
  - **Transaction-bound in-app notifications:** Implement and consume
    `contracts.InAppNotifier` (`CreateInAppNotification(context.Context,
    contracts.SQLTx, contracts.InAppNotification) error`) inside the caller's
    active tenant transaction (`contracts.PgxTx(tx)` or `contracts.StdSQLTx(tx)`),
    verifying `var _ contracts.InAppNotifier = (*notifications.Service)(nil)` at
    compile time. Never start a separate transaction inside `CreateInAppNotification`,
    never send SMS from it, and always propagate mandatory notification errors
    so the caller's transaction rolls back atomically.
- Billing package (`backend/internal/billing`):
  - **Exact integer minor-unit money (`billing.Money`):** Represent all KES
    amounts as `int64` cents matching `NUMERIC(12,2)`; never use `float64`.
  - **Release 1 boarding variants:** Only `boarding_status = 'all'` is supported;
    reject `'boarder'` / `'day'` until O13 is decided.
  - **Tenant & authorization checks:** Pass `contracts.Principal` to
    `contracts.TenantDB.WithTenantTx`; evaluate `billing.fee_structures.manage` /
    `billing.ledgers.view` badge permissions in PostgreSQL (never hardcode
    invented role strings); enforce `student_guardians.is_fee_payer = TRUE` for
    guardian balance/ledger visibility and fee-change notices.
  - **Per-term ledgers & payment allocation:** Preserve per-term `fee_ledgers`
    rows (never zero out earlier terms); allocate unallocated payments
    oldest-due-first or apply explicit `targetTermID`; record paired
    `adjustment` (`XFER:...`) + `credit_applied` (`CRD:...`) entries for
    cross-term credit transfers; check payment reference idempotency before
    allocating.
  - **Mandatory audit & term sync:** Every financial mutation requires a non-nil
    `contracts.AuditLogger` (`PostgresAuditLogger`) and fails closed if nil.
    `Service.OnTermStartDateChanged` implements `academics.TermDateFeeSyncer`
    inside `academics`'s active SQL transaction via `academics.SQLTxFromContext`.
  - **Safaricom Daraja (M-Pesa) Paybill integration:** Release 1 supports
    Paybill (`CustomerPayBillOnline`) only; reject Till (`Buy Goods`). STK Push
    enforces whole KES (`>= 1.00 KES`), `Africa/Nairobi` (EAT) timestamps,
    scoped `Idempotency-Key` (`school_id, guardian_user_id, idempotency_key`),
    mandatory `AuditLogger`/`InAppNotifier`/`STKPollQueue`, and per-request
    hashed callback path tokens (`/api/v1/billing/mpesa/callback/{callback_token}`).
    Static callback routes (`/api/v1/billing/mpesa/callback` and
    `/api/v1/billing/mpesa/c2b-confirmation`) require `VerifyCallback` or a
    trusted reverse proxy (`MPESA_TRUSTED_CALLBACK_PROXIES` +
    `MPESA_CALLBACK_TOKEN`); never accept query-string tokens and always return
    flat Daraja JSON (`{"ResultCode":0,"ResultDesc":"Accepted"}`), never the
    standard API envelope. Unmatched payment manual matching requires
    `billing.unmatched_payments.manage` (or `billing.fee_structures.manage`).

## Where to look before asking the user

- Data model and service boundaries → `ARCHITECTURE.md`
- What's in scope, by release phase → `REQUIREMENTS.md`
- Endpoint shapes and what the AI service may call → `API_CONTRACTS.md`
- Consent, encryption, retention rules → `SECURITY.md`
- Why a past decision was made → `DECISIONS.md`
