# Developer 05 implementation plan

## Repository assessment — 2026-11-09

- Branch: `feat/ian-backend-engineer-c`, created from `main`.
- Starting commit: `b2b2e3b6f15e60f25947f04c39b8f6b12b5a268f`; remote main agrees. No local or remote `release-1` or existing Ian branch.
- Pre-existing working-tree change: one leading blank line in `Docs/DECISIONS.md`. Preserve it; exclude it from every commit. Index initially empty.
- Read both assignment documents in full, baseline UP/DOWN migrations, repository instructions, API/architecture/security/requirements documents, and relevant collaborator contracts.
- Backend: Go 1.22.2, standard-library placeholder packages, empty server and worker entry points. No third-party dependencies or go.sum. No existing tests. Frontend and Python services are scaffolds.
- Schema: only migration 000001 exists. Tenant policies use `current_school_id()` and transaction-local `app.school_id`. Users and schools are global. Fee structures are unique per school/year/grade/term; ledgers are per school/student/year/term. Payments have globally unique transaction references.
- Schema gaps: fee items have neither school_id nor RLS; baseline foreign keys do not enforce matching school IDs; audit_log is not RLS protected. All owned queries must explicitly scope and validate related records. Never bypass RLS to route webhooks: inject trusted identifier-to-school routing, then verify configuration within that tenant.
- Missing migration 000002 supplies school_profile, otp_requests, sender_rate_blocks, and users.chat_payment_paused_until. Missing 000003 supplies academics data. Do not create replacements for these migrations.
- Documentation differences: old API_CONTRACTS.md leaves envelopes unspecified; the master plan defines them. Chat state is called both state/uppercase and status/lowercase: use the specified `state` column and uppercase state machine. StudentChargeHook/StudentBillingHook name the same admission capability. Fee variants are unresolved product policy; preserve the plan's default `all`, without inventing boarder classification. Term ledgers require explicit credit transfer, not repeated summing of rolled-forward balances.
- Tooling: Go 1.22.2 and Docker 27.3.1 available; PostgreSQL/Redis executables absent, images not cached. Live provider credentials and deployed AI service have not been verified. No real SMS/payment requests are authorized by test execution.

## Dependency map

| Dependency | Owner | Available | Required contract / approach |
|---|---|---|---|
| Tenant transactions | Evans | No | `WithTenantTx(ctx, schoolID, func(pgx.Tx) error) error`; injected consumer contract, test double; no edits to database package. |
| Audit logger | Evans | No | Transaction-bound audit writer; injected contract; failures roll back financial changes. |
| Authentication | Evans | No | Trusted user, membership, school, role and channel accessor; deny absent identity; no caller-controlled identity headers. |
| SMSDispatcher | Ian | No | Context-aware transactional SMS message with purpose; local sandbox and Africa's Talking implementations. |
| InAppNotifier | Ian | No | `CreateInAppNotification` participating in the caller's tenant transaction. |
| Admission billing | Ian / Evans consumer | No | StudentChargeHook (StudentBillingHook alias if needed), caller transaction, tenant/student/year and actor. |
| Term-date sync | Ian / Ema consumer | No | `OnTermStartDateChanged(ctx, tx, schoolID, termID, newStartDate, actorUserID, note) error`. |
| Academic summaries | Ema | No | GetStudentAttendanceSummary / GetStudentMarksSummary consumer interface; mock only, no substitute academics system. |
| Cleanup | Evans | No | `PurgeExpiredUnverifiedSubmissions(ctx)`; configuration retained by adapter; default seven days, scheduled daily. |
| AI | Sos | Placeholder | HTTP parent-query; token plus raw base64url JSON + dot + hex HMAC-SHA256 of encoded payload, keyed by service token. |
| Frontend | Ashy | Placeholder | Standard success/error envelopes, fixed two-decimal money strings, paginated lists, explicit offsets. |
| pgx / Asynq | Evans module files | Absent | Native implementations cannot compile in the root module until dependencies are added by owner or explicitly authorized. |
| Redis / PostgreSQL | Infrastructure | Docker available | Disposable synthetic services for migration, concurrency and worker verification. |

## Protected-file coordination proposal

The minimal dependency change needed for the prescribed pgx transaction contract and real Asynq worker is the following direct requirements in `backend/go.mod`, plus their generated indirect requirements and `backend/go.sum`:

```go
require (
    github.com/hibiken/asynq v0.25.1
    github.com/jackc/pgx/v5 v5.7.2
)
```

Both versions support this Go toolchain. This is a proposal, not authorization to edit protected files. Do not add nested modules, vendor libraries, replace Asynq, or hide unresolved imports behind build tags. Standard-library tasks can proceed independently. Migration 000002's global sender-block semantics and OTP school/session binding also need coordination; use boundary tests until the real schema exists.

## Execution and verification gates

Paths below are relative to the repository root. `B`, `N`, `W`, `I`, `J` mean `backend/internal/billing`, `notifications`, `whatsapp`, `internalapi`, `jobs` respectively. `M` means `migrations/000004_r1_billing_whatsapp_notifications`. Tests live beside their implementation. Each row is one task: inspect, implement its ordered steps, gofmt only affected files, run verification, review the diff and ownership, then update status. Never advance a dependent task past a failed gate.

`G(package)` means `cd backend && go test -v -race ./internal/<package>` and `go vet ./internal/<package>`; both must pass. `DB` means disposable PostgreSQL 16: baseline then owned UP, inspect schema/constraints/indexes, use a non-superuser role to test RLS, owned DOWN, verify baseline unchanged, UP again. Missing external integration is reported separately from a successful unit test. Commit tested coherent tasks only; no push or merge.

| ID | Feature / objective | Dependencies | Exact files | Ordered implementation / expected behavior | Tests and verification | Acceptance / status |
|---|---|---|---|---|---|---|
| INF-01 | C1–C3 baseline and architecture | None | B/IMPLEMENTATION_PLAN.md | Read documents; inventory actual code/schema; preserve edits; inspect remote; create branch; map gaps | Git status/branches/HEAD, baseline race suite | Completed; baseline race suite passed (no pre-existing tests) |
| NOTIF-01 | C3 SMS boundary and normalization | INF-01 | N/sms.go, N/sms_test.go | Define SMSDispatcher and allowed transactional purposes; normalize Kenyan/E.164 phones; validate without leaking content | Table tests for accepted/rejected numbers, purposes and missing content; G(notifications) | Completed; race tests and vet passed |
| NOTIF-02 | C3 deterministic sandbox SMS | NOTIF-01 | N/sms.go, N/sms_test.go | Log masked recipient and safe metadata, never message/OTP; propagate cancellation; username selection follows in NOTIF-03 factory | Capture slog output, assert no OTP/full phone/API key; concurrent dispatch; G(notifications) | Completed; race tests and vet passed |
| NOTIF-03 | C3 Africa's Talking transport | NOTIF-01 | N/sms.go, N/sms_test.go | Context-bound HTTP client; encoded form and apiKey; validate per-recipient response; classify retryable/permanent failures | httptest request contract, provider rejection, timeout, malformed response, cancellation; G(notifications) | No false delivery success or secret-bearing errors; Pending |
| INF-02 | C1–C3 transaction/auth/error contracts | INF-01, pgx | I/contracts/contracts.go, I/contracts/contracts_test.go | Small consumer contracts for pgx tenant transactions/audit/identity; APIError and standard envelopes; deny missing identity | Contract doubles, malformed identity, error envelope and safe error mapping; G(internalapi/...) | Compatible injection, no substitute auth; Blocked on module dependency |
| NOTIF-04 | C3 transaction-bound notification creation | INF-02 | N/notifications.go, N/notifications_test.go | Validate content/recipient; scoped insertion within supplied transaction; define InAppNotifier | Rollback, tenant/user checks; PostgreSQL insertion and RLS; G(notifications) | Usable by Ema and billing without separate commit; Pending |
| NOTIF-05 | C3 notification HTTP routes | NOTIF-04, DB-03 | N/http.go, N/notifications_test.go | Register GET list / PATCH read; bounded pagination; recipient+school scope; preserve read_at on repeat | HTTP authorization, paging, ownership, read idempotency; G(notifications) | Both endpoints and envelopes verified; Pending |
| DB-01 | C1–C3 migration design and UP | INF-01 | M.up.sql, B/migration_test.go | Add ten assigned tables; ledger credit/date; necessary financial constraints, indexes and tenant policies; enforce reference ownership | Review against baseline; DB schema and cross-tenant FK tests | No duplicated baseline/core tables; Pending |
| DB-02 | C1–C3 reversible DOWN | DB-01 | M.down.sql, B/migration_test.go | Reverse owned changes in dependency order; no CASCADE into baseline; preserve unrelated records | DB UP/DOWN/UP and baseline schema/data comparison | Rollback verified; Pending |
| DB-03 | C1–C3 RLS verification | DB-02 | B/migration_test.go, B/testdata/schema_checks.sql | Two synthetic schools; non-superuser sessions; no-tenant, wrong-tenant read/write and transaction-local context | DB policies, indexes, constraints, rollback | Actual PostgreSQL isolation evidence; Pending |
| BILL-01 | C1 exact money and ledger arithmetic | INF-01 | B/money.go, B/money_test.go | Integer minor units; strict decimal parser/JSON strings/NUMERIC limits; compute nonnegative balance and credit | Precision, overflow, negative/zero, overpayment, serialization, conservation; G(billing) | No float arithmetic; Pending |
| BILL-02 | C1 fee-structure validation | BILL-01, INF-02 | B/structures.go, B/structures_test.go | Validate team role, references, term/year agreement and items; check sums and duplicates | Invalid/cross-school/year/grade/items and duplicate request tests; G(billing) | Admin-only valid consistent structure; Pending |
| BILL-03 | C1 fee periods and structure loading | BILL-02, DB-03 | B/billing.go, B/structures.go, B/structures_test.go | Transactionally persist itemized structure and term period with due_date=start_date | PostgreSQL reference/rollback/duplicate tests; G(billing) | No partial structure; Pending |
| BILL-04 | C1 enrolled-student term charging | BILL-03 | B/charges.go, B/charges_test.go | Select applicable enrolled students; serialize student financial operations; insert unique charge and ledger entry | Concurrent repeat charge, inactive/other-grade exclusions, dates; G(billing) | One charge per student/period; Pending |
| BILL-05 | C1 admission hook | BILL-04 | B/billing.go, B/charges_test.go | Export caller-transaction StudentChargeHook; load applicable structures; reuse charging logic | New admission, repeat hook, absent structures, rollback; G(billing) | Onboarding can inject without importing its code; Pending |
| BILL-06 | C1 credit carry-forward | BILL-04 | B/ledgers.go, B/ledgers_test.go | Serialize student ledger changes; consume existing credit once; preserve term history; record credit_applied | Multi-term/year, partial/exact/excess credit and concurrency; G(billing) | Financial conservation and no double credit; Pending |
| BILL-07 | C1 balance and ledger endpoints | BILL-06 | B/http.go, B/ledgers.go, B/ledgers_test.go | Guardian/team authorization; retrieve balances/items/history/notes; fixed money and date formats | Own/unrelated guardian, cross-tenant, item/entry accuracy; G(billing) | Two GET routes conform to envelope; Pending |
| BILL-08 | C1 explained fee adjustments | BILL-07, NOTIF-04 | B/fee_changes.go, B/fee_changes_test.go | Require note; select affected charges; apply deltas, preserve entries; create note and audit in transaction | 422 code, positive/negative changes, rollback and history; G(billing) | Atomic adjustment with audit; Pending |
| BILL-09 | C1 targeted fee notices | BILL-08 | B/fee_changes.go, B/fee_changes_test.go | Deduplicate affected guardians; create notifications inside same transaction | Shared guardian, unaffected guardian, notifier failure rollback; G(billing) | Only affected guardians receive one notice; Pending |
| BILL-10 | C1 term-date synchronization | BILL-09 | B/fee_changes.go, B/fee_changes_test.go | Export exact date hook; shift periods/charges/ledger due dates; note/audit/notify | Same/date-shift, rollback, targeted notices; G(billing) | AC-22 hook verified; Pending |
| MPESA-01 | C2 Daraja client and encrypted config | BILL-01 | B/daraja.go, B/daraja_test.go | School credential resolver; OAuth; safe HTTP timeout/errors; exact request encoding | httptest OAuth/STK/query payloads, secrets, provider failures; G(billing) | No hardcoded credentials or live calls; Pending |
| MPESA-02 | C2 STK authorization and validation | MPESA-01, BILL-07 | B/mpesa_stk.go, B/mpesa_stk_test.go | Linked guardian/registered phone; positive amount; chat-only freeze; internal parent confirmation | Under/equal/over balance, 0.01, unauthorized, freeze boundary/web, missing confirmation; G(billing) | AC-23/26 enforced before provider call; Pending |
| MPESA-03 | C2 durable STK initiation | MPESA-02 | B/mpesa_stk.go, B/mpesa_stk_test.go, M.up.sql, M.down.sql | Persist correlation before provider side effect; admission AccountReference; schedule 45s polling; recover queue failures | Provider failure, delayed callback, correlation and queue tests; G(billing), DB if schema changes | Pending STK survives process restart; Pending |
| MPESA-04 | C2 idempotent callback credit | MPESA-03, BILL-06, NOTIF-04 | B/mpesa_stk.go, B/payments.go, B/payments_test.go | Authenticate callback route context; validate response against pending request; persist receipt/payment/entry/credit/notice atomically | Failed/malformed/duplicate/concurrent callback, wrong amount/receipt/school; G(billing) | Exactly one financial credit; Pending |
| MPESA-05 | C2 direct Paybill and unmatched storage | MPESA-04 | B/mpesa_c2b.go, B/mpesa_c2b_test.go | Resolve receiving shortcode through trusted boundary; school-scoped normalized admission; persist match or full unmatched facts | Two schools/same admission, invalid/blank/ambiguous reference, duplicate receipt; G(billing) | No lost incoming payment; Pending |
| MPESA-06 | C2 unmatched list and manual resolution | MPESA-05 | B/unmatched.go, B/unmatched_test.go | Team-only list; lock unresolved payment; validate target; credit once; resolve with actor/audit | Forbidden/cross-school/duplicate/concurrent resolution and overpayment; G(billing) | No fuzzy matching; Pending |
| WA-01 | C3 guardian phone masking | NOTIF-01 | W/masking.go, W/masking_test.go | Normalize and mask; reject invalid values without echoing private input | Exact +254712345678 -> 07•• ••• 678; malformed cases; G(whatsapp) | No full guardian number returned; Pending |
| WA-02 | C3 session persistence and bot routing | INF-02, DB-03 | W/whatsapp.go, W/repository.go, W/whatsapp_test.go | Trusted bot mapping then tenant verification; create/reuse sender-bound sessions; scoped admission/guardian lookup | Unknown admission zero SMS; two schools; masked multi/single selection; G(whatsapp) | No cross-school lookup or identity assumption; Pending |
| WA-03 | C3 secure OTP and sender limits | WA-02, NOTIF-03, migration 000002 contract | W/otp.go, W/otp_test.go | Cryptographic 6-digit OTP/hash/expiry; bind guardian/session; serialize global sender and school/admission limits; persist sixth-request 24h block | 59/60s, 5/6 requests, 30min/24h boundaries, concurrent senders, selected recipient only; G(whatsapp) | Guardian account never blocked; Pending |
| WA-04 | C3 verification and six-hour sessions | WA-03 | W/whatsapp.go, W/whatsapp_test.go | Atomic consume; bounded wrong-code attempts; verified_at/expires_at; deny at exact expiry; restart identity flow | Replay, wrong/expired code, exact six-hour boundary, ownership; G(whatsapp) | Fresh OTP after expiry; Pending |
| WA-05 | C3 webhook and REST adapters | WA-04 | W/http.go, W/cloud.go, W/http_test.go | Verify Meta challenge and request signature; route events to same state machine; bind REST sender; bounded payloads and delivery retries | Invalid signature/session, duplicate events, REST 429, Cloud API HTTP tests; G(whatsapp) | One state machine across transports; Pending |
| API-01 | C3 signed user context | INF-01 | I/context_signer.go, I/context_signer_test.go | Implement prescribed base64url/hex HMAC; constant-time comparisons; strict fields/channels/expiry; injectable clock | Fixed cross-language vector, tampering, malformed/expired/missing context/token; G(internalapi) | Verify signature before decoding/trusting payload; Pending |
| API-02 | C3 internal authentication and authorization | API-01, INF-02, WA-04 | I/internalapi.go, I/internalapi_test.go | Authenticate both headers; recheck active membership/session; guardian-child tenant guard; inject academics interface | Cross-parent/school, revoked/expired session, bad membership; G(internalapi) | No AI trust for authorization; Pending |
| API-03 | C3 five internal read endpoints | API-02, BILL-07 | I/reads.go, I/reads_test.go | School profile, own children, fees, attendance, marks; minimal WhatsApp facts and app link for marks | Route/data/envelope tests; academics mocks labeled; WhatsApp results denial; G(internalapi) | Backend enforces own-child and channel restrictions; Pending |
| API-04 | C3 internal confirmed STK and audit | API-03, MPESA-03 | I/actions.go, I/actions_test.go | Force chat source/confirmation; call billing; persist safe action/read metadata; configurable lookup/start limiter | Missing confirmation/freeze, audit failures, secret metadata, rate limits; G(internalapi) | AC-28 and payment boundaries; Pending |
| API-05 | C3 chat proxy and WhatsApp forwarding | API-04, WA-05 | I/client.go, I/client_test.go, W/whatsapp.go | Sign outbound identity; call AI parent-query with both headers; forward validated response; safe failure/fallback | httptest contract, context TTL<=session expiry, canceled/failed AI, channel limits; G(internalapi), G(whatsapp) | No AI end-to-end claim from mocks; Pending |
| JOB-01 | C3 Asynq client and task payloads | Asynq dependency, NOTIF-03, MPESA-03 | J/jobs.go, J/jobs_test.go | Typed sms:transactional, mpesa:stk_poll, onboarding:purge_unverified tasks; validate/serialize; stable task IDs | Serialization, invalid payload, deduplication and 45s delay; G(jobs) | Actual Asynq, no substitute queue; Blocked on module dependency |
| JOB-02 | C3 transactional SMS worker | JOB-01 | J/sms.go, J/sms_test.go | Context-bound dispatch; retry temporary errors; skip permanent/expired OTP; delivery deduplication | Redis-backed retry, permanent failure, cancellation, duplicate task; G(jobs) | No event SMS or unbounded OTP retries; Pending |
| JOB-03 | C2 STK polling worker | JOB-01, MPESA-04 | J/mpesa.go, J/mpesa_test.go, B/mpesa_stk.go | Check pending; query provider; terminal failure/timeout; success requires verified receipt facts; same idempotent settlement | Callback/poll race, retry, pending, success without receipt; G(jobs), G(billing) | Query response never fabricates a receipt; Pending |
| JOB-04 | C3 cleanup schedule and lifecycle | JOB-02, JOB-03 | J/worker.go, J/worker_test.go, backend/cmd/worker/main.go | Redis configuration; inject Evans cleanup; retention default 7; daily scheduler; signals/graceful shutdown | Test cleanup double, retention/env, handler registration/start/shutdown; G(jobs), go build ./cmd/worker | Missing cleanup explicitly reported, never silently skipped; Pending |
| QA-01 | C1–C3 final verification and report | All above | B/VERIFICATION.md | Format owned Go; test/build/vet/race; DB cycle; inspect security/concurrency/contracts; map ACs; ownership audit; conventional commits | go test -v -race ./...; go build ./...; go vet ./...; make test; frontend lint/build; git status/diff/cached paths | Report evidence and blockers separately; Pending |

## Acceptance evidence tracker

| Criterion | Required coverage | Evidence | Status |
|---|---|---|---|
| AC-BE-C01 | School routing, masked selected guardian, OTP, exact 6h expiry | WA-01–05, PostgreSQL/HTTP tests | Not implemented |
| AC-BE-C02 | Sender/admission cooldown, sixth-request sender-only 24h block | WA-03 concurrent and boundary tests | Not implemented |
| AC-BE-C03 | Idempotent term charge, credit, explained targeted fee/date changes | BILL-02–10, PostgreSQL tests | Not implemented |
| AC-BE-C04 | STK rules, receipt idempotency, overpayment, school-scoped C2B/manual matching | MPESA-01–06, JOB-03 | Not implemented |
| AC-BE-C05 | Token/HMAC/expiry, own-child permissions, AI audit | API-01–05 | Not implemented |

## Outstanding integration, not substitute implementations

Evans must mount exported routes, supply authenticated identities/tenant transactions/audit/cleanup, and land core migration/dependencies. Ema supplies academic reads and invokes term-date synchronization. Sos supplies AI service and compatible signed-context verification; Ashy consumes documented owned endpoint shapes. Provider credentials, a reachable authenticated callback ingress, and provider sandbox end-to-end tests remain separate verification gates. Full integration of migrations 000001–000004 cannot be claimed before 000002/000003 exist.
