# Notifications integration contract

## Runtime wiring status

This package delivers the reusable in-app notification service, HTTP route
handlers (`RegisterRoutes`), and synchronous transactional SMS transport
(`NewSMSDispatcher`, `NewSandboxSMSDispatcher`). Application-level runtime
wiring remains pending in subsequent PRs:

- `backend/cmd/server/main.go` (route mounting with trusted `contracts.Identity`
  middleware and `contracts.TenantTxFunc`)
- `backend/internal/config/config.go` (`AFRICASTALKING_USERNAME`,
  `AFRICASTALKING_API_KEY`, `AFRICASTALKING_SENDER_ID` loading and validation)
- `docker-compose.yml` (provider environment variables for `backend`)
- `backend/internal/jobs/` and `backend/cmd/worker/main.go` (Redis/Asynq SMS
  job queue and retry worker)
- Consumers still awaiting wiring: `internal/auth`, `internal/onboarding`, and
  `internal/billing` (only `internal/academics` currently integrates
  `contracts.InAppNotifier` directly in code; `cmd/server` wiring of `academics`
  and `notifications` is pending)

## HTTP routes and tenant transactions

The server owner mounts `notifications.RegisterRoutes(mux, service, identity)`.
`identity` is a `contracts.Identity` callback supplied by trusted
authentication middleware; request headers, query strings, or JSON bodies must
never be trusted as identity or school selection.

`Service.DB` implements `contracts.TenantDB` using:

```go
WithTenantTx(ctx context.Context, principal contracts.Principal, fn func(pgx.Tx) error) error
```

In production, wrap the application's tenant transaction runner with
`contracts.TenantTxFunc(pool.WithTenantTx)`. `List` and `MarkRead` pass the
authenticated `contracts.Principal` returned by `contracts.RequireIdentity` to
`s.DB.WithTenantTx(ctx, p, ...)`, deriving the tenant exclusively from
`Principal.SchoolID`.

- `GET /api/v1/notifications?page=1&per_page=25` lists only notices where
  `school_id = Principal.SchoolID` and `recipient_user_id = Principal.UserID`,
  ordered by `created_at DESC, id`. Page size defaults to 25 and is capped at
  100 (`422 INVALID_PAGINATION` for invalid bounds). Success returns
  `{"status":"success","data":[...],"meta":{"page":...,"per_page":...,"total":...}}`.
- `PATCH /api/v1/notifications/{id}/read` requires no request body, updates
  `is_read = true, read_at = coalesce(read_at, now())` scoped to
  `(Principal.SchoolID, Principal.UserID, id)`, and returns
  `{"status":"success","data":{"id":"...","is_read":true}}`. Repeat calls
  preserve the original `read_at`. Unknown, cross-user, or cross-school IDs
  return `404 NOTIFICATION_NOT_FOUND`; malformed IDs return
  `400 INVALID_NOTIFICATION_ID`; missing/invalid identity returns
  `401 UNAUTHENTICATED`. Responses use the `contracts` envelope and
  `Cache-Control: no-store`.

## Transaction-bound in-app notifications (`contracts.InAppNotifier`)

Business modules (`academics`, and future `billing`) inject
`contracts.InAppNotifier`:

```go
CreateInAppNotification(ctx context.Context, tx contracts.SQLTx, input contracts.InAppNotification) error
```

`*notifications.Service` satisfies `contracts.InAppNotifier` (verified at
compile time via `var _ contracts.InAppNotifier = (*Service)(nil)`).

- **Caller owns the transaction:** Start your tenant-scoped transaction first,
  perform domain writes, adapt the active transaction with `contracts.PgxTx(tx)`
  (for `pgx.Tx`) or `contracts.StdSQLTx(tx)` (for `*sql.Tx`, exposed by
  `academics.Store.NotificationTx()`), and call `CreateInAppNotification`
  inside that same transaction.
- **Validation & isolation:** `CreateInAppNotification` validates non-empty
  fields (`SchoolID`, `RecipientUserID`, `Category`, `Title`, `Body`), checks
  that the recipient has an active `memberships` + `users` record in `SchoolID`,
  and inserts into `in_app_notifications` (enforced by PostgreSQL forced RLS).
  It rejects `tx == nil`, never creates its own transaction, and never sends SMS.
- **Error propagation:** Do not discard errors from `CreateInAppNotification`.
  Propagate any error out of the transaction callback so the caller's
  transaction rolls back atomically (leaving zero partial domain writes or
  partial batch notices). Callers should deduplicate recipients per domain event
  and keep secrets, OTPs, and clinical details out of `Title`, `Body`, and
  `Metadata`.

## Transactional SMS transport (`SMSDispatcher`)

Future callers (`auth`, `onboarding`, `jobs`, and Release 3 `whatsapp`) inject
`notifications.SMSDispatcher`:

```go
DispatchSMS(ctx context.Context, message notifications.SMSMessage) error
```

- Permitted `SMSPurpose` values (`C58`): `app_login`, `whatsapp_session`
  (reserved purpose key; WhatsApp channel integration is deferred to Release 3),
  `phone_change`, and `phone_change_notice`. Event alerts (`absence`, `sickbay`,
  etc.) are rejected with `ErrInvalidSMS`.
- `NewSMSDispatcher(config, client, logger)` selects `NewSandboxSMSDispatcher`
  when `Username == "sandbox"` (local redacted metadata logging only, no HTTP
  call), and otherwise configures Africa's Talking with a copied `http.Client`,
  15-second default timeout, `http.ErrUseLastResponse` redirect blocking, and a
  64 KiB response cap.
- A `nil` return means provider/sandbox acceptance, not proven handset delivery.
  `*SMSDeliveryError{Code, Retryable}` omits credentials, bodies, phone numbers,
  and raw provider responses so the future Asynq worker can classify retries
  safely.

## Test setup and verification

PostgreSQL 16 integration tests use the disposable test harness in
`../billing/testenv/database.go` (documented in `../../../README.md`,
`../../../Docs/AGENTS.md`, and `../../../.github/workflows/backend.yml`).
`testenv.New` applies migrations in order `000001` → `000003` → `000004`, runs
tenant queries under a non-superuser, non-`BYPASSRLS` role, and drops the
disposable database/role on cleanup. Missing `DEV05_TEST_DATABASE_URL` skips
DB tests locally when `CI` is unset, and fails the test run when `CI` is set.

Run unit tests (no database required):

```sh
cd backend && go test -count=1 -v ./internal/notifications -run 'Test(AfricaTalking|SMS|Normalize|Sandbox|Transactional)'
```

Run the full PostgreSQL 16 notification integration suite:

```sh
docker compose up -d --wait postgres
export DEV05_TEST_DATABASE_URL='postgres://ikonex_admin:ikonex_secure_pass@localhost:5432/postgres?sslmode=disable'
cd backend && CI=true go test -count=1 -v ./internal/notifications
```

Real PostgreSQL tests verify notification creation, multi-notification
atomicity, caller and `academics` transaction rollback, error propagation,
tenant/recipient RLS isolation, pagination, read receipt ownership, and
idempotent `read_at` timestamps. Local `httptest.Server` fixtures cover SMS
provider behavior without live network calls.
