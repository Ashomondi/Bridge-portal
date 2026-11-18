# Developer 05 checkpoint report — 2026-11-09

This is an incomplete implementation checkpoint, not Release 1 sign-off. Phase Zero, the detailed plan, and NOTIF-01 through NOTIF-03 are complete. The next task, INF-02, requires the missing pgx dependency in module files explicitly protected by the assignment. An exception for the exact dependency proposal in IMPLEMENTATION_PLAN.md has been requested and has not been assumed.

## Git baseline and ownership

- Branch: `feat/ian-backend-engineer-c`.
- Starting commit: `b2b2e3b6f15e60f25947f04c39b8f6b12b5a268f` on `main`, also the verified remote main tip.
- No `release-1` or existing Ian branch was available. Created the requested branch without resetting/stashing/discarding anything.
- Existing unrelated change: `Docs/DECISIONS.md` has a leading blank line. It remains unchanged and unstaged.
- Tested commits preceding this report: `5925aa5` (SMS contract and plan), `515b2f0` (sandbox dispatcher). The commit containing this report adds the verified provider transport. Its exact hash is reported in the user-facing checkpoint response.
- No push or merge performed.
- `git status --short`, `git diff --name-only`, `git diff --cached --name-only`, and `git diff --check` checked before commits. Every staged implementation path belongs to Developer 05.

## Files created

| File | Purpose |
|---|---|
| `backend/internal/billing/IMPLEMENTATION_PLAN.md` | Repository assessment, dependency map, detailed tasks, gates and acceptance tracker |
| `backend/internal/billing/VERIFICATION.md` | This checkpoint's evidence and limitations |
| `backend/internal/notifications/sms.go` | SMS consumer contract, transactional purposes, phone normalization and safe sandbox dispatcher |
| `backend/internal/notifications/sms_test.go` | Normalization, purpose validation, sandbox redaction/determinism/cancellation/concurrency |
| `backend/internal/notifications/sms_provider.go` | Factory and Africa's Talking HTTP transport with safe errors and retry classification |
| `backend/internal/notifications/sms_provider_test.go` | Local HTTP provider contract and failure tests |

No pre-existing tracked source files were modified. The original `notifications.go` remains a placeholder for the later in-app service. No module files, migrations, shared code, frontend, AI service, or external documentation were changed.

## Implemented and tested

- `SMSDispatcher.DispatchSMS(context.Context, SMSMessage) error` is the consumer-facing dependency for auth/onboarding/WhatsApp and, later, the queue worker.
- Supported purposes: app-login OTP, WhatsApp OTP, phone-change OTP, old-number phone-change notice. Other purposes, including absence and sickbay events, are rejected before delivery.
- `NormalizePhone` accepts Kenyan local and country-prefixed numbers and international E.164 forms; normalizes common display separators; rejects malformed input without reflecting private input in errors.
- `NewSMSDispatcher(SMSConfig, *http.Client, *slog.Logger)` selects local-only behavior when `Username == "sandbox"`; no key or network is required in that mode.
- Sandbox logging contains only masked recipient suffix, purpose, and character count. It does not log body, OTP, full number, or credentials. Tests remove the logger timestamp to compare deterministic metadata.
- Provider transport uses a context-bound form POST, API key header, normalized recipient, optional sender ID, and JSON response validation. It verifies a single matching recipient and recognized processed/sent/queued status before success.
- A copied HTTP client receives a finite default timeout and refuses redirects, preventing credentials and SMS bodies from being forwarded to a redirect target.
- `SMSDeliveryError` classifies transport failures, HTTP 408/429/5xx, and selected provider failures for retry without exposing provider text or private data. Invalid/ambiguous response bodies do not trigger automatic retry classification.

## Test evidence

Nine top-level tests, including 18 phone cases and 14 provider-response cases:

| Test | Evidence |
|---|---|
| TestNormalizePhone | Local/E.164 normalization, incorrect lengths/prefixes, Unicode digits, embedded private/injected text |
| TestTransactionalSMSValidation | All four allowed purposes, rejection of event SMS and blank/invalid content, safe errors |
| TestSandboxSMSRedactsSensitiveData | No full number/body/OTP, safe metadata, deterministic output, cancellation and invalid-purpose silence |
| TestSandboxSMSConcurrentDispatch | 20 concurrent calls, no lost log events and no race findings |
| TestAfricaTalkingRequestAndResponses | Exact method/headers/form, 100/101/102 accepted states, recipient failure/mismatch, malformed/oversized responses, retry classification |
| TestSMSFactorySandboxDoesNotUseHTTP | Sandbox factory uses local logger; no provider transport; missing production config rejected |
| TestAfricaTalkingRejectsRedirects | 307 redirect never reaches target and is not classified as retryable |
| TestAfricaTalkingCancellationAndTimeout | Canceled/invalid requests never reach provider; transport timeout returns retryable safe error |
| TestSMSFactoryDoesNotMutateHTTPClient | Original client remains unchanged; copied client has timeout and redirect policy |

| Command | Result |
|---|---|
| Initial `cd backend && go test -v -race ./...` | Passed; every original package had no tests |
| `gofmt -w` on each affected owned Go file | Applied before each task's verification |
| `go test -v -race ./internal/notifications` after each task | Passed |
| `go vet ./internal/notifications` after each task | Passed |
| Final `cd backend && go test -v -race ./...` | Passed; notifications tests exercised, other packages remain scaffolds |
| `cd backend && go build ./...` | Passed |
| `cd backend && go vet ./...` | Passed |
| `gofmt -l` on the four created Go files | Empty output; formatting clean |
| `make test` | Go tests passed. Python invocation failed: `No module named pytest`. Makefile masks the failure with `|| true`; this is not an all-services pass |
| `cd frontend && npm run lint` | Failed, exit 127: `next: not found` |
| `cd frontend && npm run build` | Failed, exit 127: `next: not found` |
| `git diff --check` | Passed |

All SMS provider tests use local httptest servers. No live Africa's Talking, WhatsApp, Daraja, or AI-service end-to-end test has passed or been claimed. No race findings occurred in the implemented SMS code; financial/OTP-state concurrency remains unimplemented and untested.

## Database, APIs and acceptance status

No database changes or HTTP API endpoints have been implemented at this checkpoint. The ten assigned tables, UP/DOWN migration, RLS tests, and UP/DOWN/UP verification remain pending. Full migration integration also lacks the other developers' 000002 and 000003 files.

| Acceptance criterion | Status |
|---|---|
| AC-BE-C01 — WhatsApp identity/session security | Not implemented; SMS transport alone is not sufficient evidence |
| AC-BE-C02 — cooldown and sender blocking | Not implemented |
| AC-BE-C03 — charging and fee changes | Not implemented |
| AC-BE-C04 — payments, credit, matching, phone lock | Not implemented |
| AC-BE-C05 — internal AI boundary and audit | Not implemented |

## Blockers and integration work

1. The required `WithTenantTx(ctx, schoolID, func(pgx.Tx) error)` contract cannot compile without pgx. Asynq is also absent. `backend/go.mod` currently declares only the module path and Go version. The assignment explicitly forbids editing it or go.sum. The minimal proposed direct requirements are pgx/v5 v5.7.2 and Asynq v0.25.1, plus generated indirect requirements/checksums. No dependency changes or alternate nested module have been made.
2. Evans's real tenant transaction helper, authentication context, audit logger, cleanup integration, and migration 000002 are absent. Consumer interfaces and test doubles are planned, not replacements for those systems.
3. Ema's academics summaries and migration 000003 are absent. Sos's AI client/parent assistant are placeholders. Server route mounting is reserved to Evans. These prevent eventual full integration, independently of unit-test results.
4. The provider has not been exercised with live credentials. A transport timeout can mean delivery was accepted before the response was lost; the API does not provide a client idempotency guarantee here. The queue must handle deduplication where possible and must not claim exactly-once SMS delivery. Queue implementation is pending.
5. HTTP source registration, in-app notifications, all billing/M-Pesa/WhatsApp state-machine/internal API functionality, and background workers remain unimplemented. Nothing in the successful scaffold build changes that status.

The next implementation step is INF-02 after the protected dependency change is supplied or explicitly authorized, followed by transactional in-app notification creation and migration verification. Preserve the task-by-task gates in IMPLEMENTATION_PLAN.md.
