# DECISIONS.md

A short, dated log of decisions worth remembering — so nobody (human or AI
assistant) re-litigates a settled question without knowing why it was made.
Add a new entry any time a real decision is made; don't edit old entries,
add a new one that supersedes it if the decision changes.

Format: date, decision, why, alternatives considered.

---

**2026-09-22 — Shared PostgreSQL database with row-level security, not
per-school databases**
Every tenant-owned table carries `school_id`; isolation is enforced with
Postgres RLS (`FORCE ROW LEVEL SECURITY`), not only application-level
filtering.
*Why:* cheapest to run and easiest to back up and migrate across many
schools; RLS gives a database-level safety net against a missed `WHERE
school_id = ?` in application code, which matters given the sensitivity of
the data (children's records).
*Alternative considered:* one database per school — rejected as
operationally heavy past a handful of schools, though the shared-table
design keeps this open as a later migration path for a school that
specifically needs it (e.g. a contractual requirement).

**2026-09-22 — Global `users` table, separate per-school `memberships`**
A person has one login; their role(s) at each school live in `memberships`.
*Why:* supports a parent with children at two schools, or a teacher working
at two schools, without duplicate accounts.

**2026-09-22 — Badges are permissions, not roles or table columns**
"Class teacher," "on duty," and "club patron" are badges — rows in
`membership_badges`, optionally scoped to a class, club, or time window —
not a role type or a column on `classes`.
*Why:* one membership can hold several badges at once (a teacher can be a
class teacher, on duty, and a club patron simultaneously); keeping this as
one badge system avoids one-off columns for every new permission concept.

**2026-09-22 — CBC only (Grade 7–12), not 8-4-4**
`grade_levels.number` is constrained to 7–12 with a `junior`/`senior` phase.
*Why:* stated scope is CBC, both junior and senior school.
*Open:* whether legacy Form 3/4 classes need support during the transition
period is unresolved — see `REQUIREMENTS.md` open decisions.

**2026-09-22 — AI service is separate from the Go backend, with no direct
database access**
The AI service (Python/FastAPI) calls permission-checked Go endpoints only;
it never queries the database directly and never originates a fact.
*Why:* keeps the "AI never invents a fee balance / never handles money"
guarantees enforceable at the API boundary rather than relying on prompt
instructions; also lets the AI service be hosted, scaled, or swapped
independently of the main application.

**2026-09-22 — Payments: system records and approves; it doesn't disburse to
suppliers**
Supplier/bill payments go through a bursar-drafted, signatory-approved
voucher; the actual bank transfer happens outside the system. The one
exception is parent-initiated fee payment, where the AI service can start
an M-Pesa STK push, but only after the parent has confirmed the child and
amount, and the parent still authorises the transaction with their own PIN.
*Why:* automated disbursement carries materially more regulatory and
operational risk than recording an approved payment; parent-initiated STK
push is different because the parent remains the one authorising the actual
transfer.
*Open:* whether automated supplier disbursement is ever added later is an
open question, not a rejected one.

**2026-09-22 — Biometrics: encrypted templates only, per-school keys, manual
fallback always offered**
Raw fingerprint images are never stored; only the encrypted template.
Schools that can't afford biometric hardware use manual attendance.
*Why:* Data Protection Act treatment of biometric data as sensitive
personal data of children; minimising what's stored reduces both legal risk
and breach impact.

**2026-09-22 — Payroll is out of scope**
*Why:* stated explicitly as out of scope; revisit only if a clear
justification and owner emerge.

**2026-09-22 — Hosting: Kenya or regional-Africa cloud, not a default global
region**
*Why:* supports the Data Protection Act cross-border transfer rules and the
"data stays in Kenya" product/compliance claim; see `SECURITY.md`.

---

<<<<<<< HEAD
**2026-10-29 — WhatsApp moves into Release 1; one WhatsApp bot per school**
The parent assistant and chat payment ship on WhatsApp and in-app in Release
1. SMS and USSD come later. Each school has its own bot so admission numbers
never collide.
*Why:* WhatsApp reaches parents where they already are.
*Supersedes:* WhatsApp as a Release 3 could-have.

**2026-10-29 — WhatsApp carries minimal facts only**
Fee balances, attendance status, and payment prompts. Results and report
content stay in the app with a link. Launch gates before real data: a
`whatsapp` consent purpose, an impact assessment, a processing agreement with
Meta, ODPC notification, lawyer review.
*Why:* WhatsApp passes through Meta, a transfer outside Kenya; less child data
in it means less exposure.
*Alternatives considered:* full answers with consent (more legal exposure);
per-school settings (more cases to test).

**2026-10-29 — Parent identity in chat: admission number, masked guardian
numbers, OTP to the registered number, 6-hour session**
Any WhatsApp number may start a chat; OTP requests are limited to one per
minute, and more than five in thirty minutes blocks the requesting number for
24 hours (not the guardian).
*Why:* blocking the guardian would let a stranger lock a real parent out.

**2026-10-29 — AI service authenticates with a service token plus a signed
user context**
*Alternative kept open:* adding mutual TLS later.

**2026-10-29 — Chat payments: parent types the M-Pesa number, any amount,
explicit confirmation**
Excess is held as a ledger credit; a limiter for the top-most grade is planned
later, not agreed.

**2026-10-29 — Release 1 is a test with synthetic data only**
It passes when onboarding works, parents can log in, and badges work well;
every Release 1 feature must work well and stress free. A QA developer owns
verification.
*Why:* nothing real is at stake, so real-data compliance gates apply to the
first release with real data.

**2026-10-29 — Release 1 has no bursar; the Ikonex team loads fee structures**
Fee structure is fixed to terms, set by the school's own structure, loaded as
data by the team, charged automatically once at the start of the year. Mid-year
changes need a note and notify only affected parents. Due dates follow term
start dates.
*Supersedes:* "bursar fee ledger" in Release 1.

**2026-10-29 — Direct Paybill payments**
Auto-matched when the account number is a valid admission number for the
school; the rest go to an unmatched list the team resolves.

**2026-10-29 — Results: marks by subject teachers, hard-gated release, reopen
to correct**
Class teacher releases only when every subject has marks for every student.
The dean reassigns marks entry. Corrections reopen the class; parents keep
seeing the old results until re-release, then only affected parents are
notified with a teacher's note.

**2026-10-29 — Performance graph plots CBC levels; grading scale versioned
per term**

**2026-10-29 — Daily class-teacher roll-call in Release 1; in-app absence
notifications; SMS in Release 2**

**2026-10-29 — Onboarding**
The team creates the school and principal; the principal runs a setup wizard.
Parents register by QR (12 hours, generated by dean/principal/deputy,
revocable) and are verified in person; verification is delegable to staff for
the QR's active window. Staff upload documents and are verified in person.
Documents are deleted after approval; unverified applications are deleted
automatically after a fixed period.

**2026-10-29 — The dean owns teacher assignment, timetables, and the school
calendar (including term dates after setup)**
A term-date change moves fee due dates and triggers the fee-change notice.

**2026-10-29 — Guardian fee visibility: all linked guardians see fees**
May be revisited. Closes the open decision in `REQUIREMENTS.md`.

**2026-10-29 — Legacy 8-4-4 (Form 3/4) is out of scope**
*Why:* 8-4-4 ends in under a year. Closes the open question under the CBC
entry above.

**2026-10-29 — Release 1 is English only; Swahili in Release 2**

**2026-10-29 — Hosting location is unresolved**
This note does not change the earlier hosting entry; it flags that the earlier
entry ("Kenya or regional-Africa cloud") conflicts with the Kenya-only rule in
`AGENTS.md` and `SECURITY.md`, and that regional hosting would count as a
cross-border transfer. To be discussed with the project group. Release 1 is
unaffected (synthetic data).

**2026-10-29 — Retention at student exit is set per record type**
Periods to be set with a data protection lawyer. Biometric templates are
deleted on exit (unchanged).
=======
**2026-11-13 — ADR: Tenant Binding (transactions derive the school from the principal)**

- **Context:** Every tenant-owned table enforces PostgreSQL RLS using `school_id = current_school_id()` via transaction-local `app.school_id`. Accepting a separate school UUID parameter alongside an authenticated caller creates a risk of handlers passing an unchecked URL, query, header, or body school ID into a transaction.
- **Decision:** `TenantTxFunc.WithTenantTx` accepts a `Principal` produced by `RequireIdentity` from trusted authentication middleware, validates its canonical nonzero UUIDs and non-empty role, and passes only `Principal.SchoolID` to the application's transaction function. That function sets transaction-local `app.school_id` (`SET LOCAL`) under a non-superuser, non-`BYPASSRLS` role.
- **Rationale:** Validating UUID syntax does not authorize access to that school. Deriving the tenant solely from the authenticated `Principal` removes any second untrusted tenant selector.
- **Consequences:** Request-supplied school IDs cannot select or override the transaction tenant. Switching schools requires revalidating the user's active `memberships` row before issuing a new `Principal`. Within-school badge permissions and `student_guardians` ownership checks remain mandatory. `RequireIdentity` validates `Principal` structure only; auth/session verification middleware remains pending.
- **Alternatives considered:** Accepting an explicit `schoolID` parameter and relying on every handler to compare it against the authenticated principal — rejected because omitting a comparison in any handler weakens tenant isolation.

**2026-11-13 — ADR: Explicit Credit Balances (separate debt and available credit, with a defined downgrade)**

- **Context:** Legacy `fee_ledgers.current_balance` represented overpayments as negative signed balances, while `fee_structures` enforced a single template per `(school_id, academic_year_id, grade_level_id, term_id)`. Release 1 billing requires nonnegative receivables, explicit prepaid credit tracking, boarding-variant pricing (`all`, `day`, `boarder`), durable STK attempt state (`mpesa_stk_requests`) with per-school unique receipt references (`UNIQUE (school_id, transaction_ref)`), and human-confirmed unmatched-payment resolution.
- **Decision:** Add nonnegative `credit_balance`, `credit_applied`, and `credit_transferred` columns (plus `due_date`) to `fee_ledgers`, backfill negative legacy `current_balance` rows into `credit_balance = -current_balance` and `current_balance = 0` before enforcing `r1_c_ledger_nonnegative`, and track new postings in `ledger_entries` from `fee_charges` and `payments`. On downgrade of `000004`, restore `current_balance = current_balance - credit_balance`, and resolve `fee_structures` boarding variants per `(school_id, academic_year_id, grade_level_id, term_id)` using strict precedence: retain `all` if present, otherwise `day`, otherwise `boarder`, keeping that template's ID, `total_amount`, and `fee_structure_items` while deleting discarded variants and cascading their items.
- **Rationale:** Explicit `credit_balance` prevents overpayments from appearing as negative debt and separates spendable credit from cumulative usage counters (`credit_applied`, `credit_transferred`). Deterministic variant precedence avoids legacy unique-key collisions on rollback and favors `day` pricing over `boarder` pricing so boarding charges are not accidentally applied to a generic fallback template.
- **Consequences:** Financial meaning of negative, zero, and positive legacy balances is preserved across up/down/up round trips. However, the downgrade is intentionally lossy for boarding variants and the 12 new tables: operators must export variant definitions and new tables before running `000004.down`, because reapplying `000004.up` marks surviving templates as `all` and cannot recover deleted variants. `fee_ledgers` remains the materialized read balance; rebuilding from `ledger_entries` must start from the captured legacy opening snapshot.
- **Alternatives considered:** Keeping signed `current_balance` indefinitely (rejected as ambiguous for credit application/transfer); summing variant fees or selecting an arbitrary row on downgrade (rejected as financially corrupt or non-deterministic).

**2026-11-13 — ADR: WhatsApp Release Scope (Release 1 chat schema, Release 3 WhatsApp delivery)**

- **Context:** `REQUIREMENTS.md` scopes in-app parent chat and M-Pesa STK push to Release 1, while deferring WhatsApp Business Cloud API and USSD delivery to Release 3. However, `chat_sessions` and `ai_action_logs` share channel modeling with future WhatsApp sessions.
- **Decision:** Migration `000004` introduces all **12 tables** (`school_mpesa_configs`, `fee_periods`, `fee_charges`, `fee_change_notes`, `fee_change_note_students`, `ledger_entries`, `unmatched_payments`, `mpesa_stk_requests`, `school_whatsapp_bots`, `chat_sessions`, `in_app_notifications`, `ai_action_logs`), creating the shared verification foundation (`chat_sessions`) and reserving `school_whatsapp_bots` for Release 3, while keeping WhatsApp webhook routing, provider credentials, templates, and outbound delivery deferred to Release 3.
- **Rationale:** Reconciles a forward-compatible database schema with the approved release phases in `REQUIREMENTS.md` without pulling Release 3 messaging integration or cross-border compliance work prematurely into Release 1.
- **Consequences:** In-app chat (`channel = 'in_app'`) and SMS/in-app notifications are the Release 1 scope. `school_whatsapp_bots` remains unused in Release 1 (`is_active DEFAULT TRUE` is not an activation flag), and WhatsApp provider integration cannot be enabled until Release 3 and `SECURITY.md` cross-border/ODPC requirements are satisfied.
- **Alternatives considered:** Implicitly bringing WhatsApp delivery into Release 1 because `school_whatsapp_bots` exists (rejected as unapproved scope creep), or splitting `chat_sessions` later with an additional migration (rejected to avoid avoidable schema churn).

**2026-11-13 — ADR: AI Service Credential Plus Delegated Session**

- **Context:** The Python AI service has zero direct database credentials and accesses student/billing data only through internal Go endpoints. Relying on a shared service token with raw forwarded user/school headers would allow an internal bug or header spoof to select arbitrary tenants or guardians.
- **Decision:** Internal AI-to-Go calls require both `Authorization: Bearer <service-token>` (`INTERNAL_SERVICE_TOKEN` / `AI_SERVICE_SECRET_TOKEN`) and `X-Delegated-Session: <opaque-session-token>`. For verified chat, migration `000004` enforces at the database level that `chat_sessions` in `VERIFIED` state carry non-null, non-blank `access_token_hash`, `guardian_user_id`, `verified_at`, and `expires_at > verified_at`; Go hashes `X-Delegated-Session` and resolves it against an unexpired (`expires_at > now()`) `chat_sessions` row before rechecking active membership and `student_guardians` authorization.
- **Rationale:** Authenticating the calling AI service and authorizing the human principal are independent security checks; storing only opaque token hashes prevents token disclosure on DB read and supports immediate revocation.
- **Consequences:** A service credential alone grants no tenant or student access, and unsigned `X-User-ID` / `X-School-ID` headers are never accepted. Delegation middleware and token-issuance handlers remain pending integration in the application runtime.
- **Alternatives considered:** Trusting a shared service token alongside plain `X-User-ID` and `X-School-ID` headers — rejected because it permits unverified delegation.

**2026-11-17 — ADR: SMS Transport and Queue Separation**

- **Context:** Transactional SMS messages (OTPs and phone-change security notices) need asynchronous queuing, backoff retries, and worker execution via Redis/Asynq, while the HTTP provider boundary must enforce phone normalization, redirect blocking, response validation, and error/credential redaction independently of the queue.
- **Decision:** Implement `notifications.SMSDispatcher` (`NewSMSDispatcher` / `NewSandboxSMSDispatcher`) as a synchronous, context-aware transport returning structured `*SMSDeliveryError{Code, Retryable}` errors, and keep Redis/Asynq job enqueueing and retry scheduling separate in `internal/jobs` and `cmd/worker` (pending wiring).
- **Rationale:** Separating transport validation and provider protocol handling from queue scheduling allows deterministic unit testing with `httptest.Server` and local sandbox logging without requiring Redis in transport tests, while giving the future worker explicit `Retryable` classification without leaking provider payloads or secrets.
- **Consequences:** `DispatchSMS` returning `nil` means provider acceptance (or sandbox log acceptance), not proven handset delivery, and ambiguous network timeouts cannot guarantee exactly-once SMS delivery. Durable async queueing and retry execution remain pending until `internal/jobs` and `cmd/worker` are implemented.
- **Alternatives considered:** Embedding Asynq client/retry loops directly inside `notifications.Service` (rejected because it couples provider transport to Redis availability and blurs module boundaries) or returning untyped errors from the provider adapter (rejected because workers need safe, secret-free retry classification).

**2026-11-17 — ADR: Transactional-Only SMS in Release 1 (C58)**

- **Context:** `REQUIREMENTS.md` lists SMS alerts for absence and sickbay (`[S]`), while Release 1 (`C58`) restricts event notifications to in-app delivery and limits SMS to authentication and account-security flows. Allowing arbitrary SMS purposes through the Release 1 transport risks unbudgeted SMS cost (`O4`), unconsented event messaging (`O15`), and leakage of student event/health context over SMS.
- **Decision:** Enforce an explicit allowlist of transactional `SMSPurpose` values in `notifications.SMSMessage` (`app_login`, `whatsapp_session`, `phone_change`, `phone_change_notice`) and reject all other purposes (including `absence`, `sickbay`, and empty purposes) with `ErrInvalidSMS` before any provider call. Event notifications in Release 1 persist exclusively to `in_app_notifications`.
- **Rationale:** Enforcing C58 at the transport validation boundary guarantees that neither callers nor future queue workers can accidentally send Release 2 event SMS alerts through the Release 1 transactional dispatcher.
- **Consequences:** Absence alerts in Release 1 are delivered in-app only (`academics.attendance`); sickbay alerts remain Release 2 (with the nurse module); `whatsapp_session` exists as the transport purpose constant while WhatsApp channel delivery remains deferred to Release 3 per the WhatsApp Release Scope ADR. Expanding SMS to Release 2 event alerts will require an explicit policy/consent decision and an update to the purpose policy.
- **Alternatives considered:** Accepting arbitrary free-form SMS strings in `SMSDispatcher` and relying solely on caller discipline (rejected as easy to bypass accidentally).

**2026-11-17 — ADR: Transaction-Bound In-App Notifications (`contracts.InAppNotifier`)**

- **Context:** Previously, `academics` defined an incompatible `InAppNotifier` interface without a transaction handle, ran notification calls after its store transaction committed, and discarded notification errors; meanwhile `notifications` exposed a `pgx.Tx`-specific signature incompatible with `academics.SQLStore` (`*database/sql` `*sql.Tx`) and `academics.MemoryStore`.
- **Decision:** Define `contracts.InAppNotifier`, `contracts.InAppNotification`, and the minimal `contracts.SQLTx` interface (`PgxTx` and `StdSQLTx` adapters) in `backend/internal/internalapi/contracts`. `notifications.Service` implements `contracts.InAppNotifier` (verified at compile time), and `academics.Service` consumes `contracts.InAppNotifier` inside `Store.WithTx` via `tx.NotificationTx()`, propagating errors so domain writes, audit logs, and mandatory notifications commit or roll back atomically.
- **Rationale:** Placing the contract in `contracts` avoids circular dependencies between `academics` and `notifications`, `contracts.SQLTx` bridges `pgx.Tx` (billing) and `*sql.Tx` (academics) within the caller's real database transaction without forcing a driver rewrite, and executing inside `Store.WithTx` prevents silent partial commits where a domain change succeeds without its required parent/teacher notice.
- **Consequences:** All Release 1 `academics` notification calls (attendance absence, results release/re-release, timetable updates, calendar event updates) are mandatory within their business transaction: if recipient membership validation or notification persistence fails, the entire transaction rolls back and returns an error. Best-effort post-commit fan-out is not used for these paths.
- **Alternatives considered:** Leaving notifications post-commit with best-effort error logging (rejected because it violates atomic business/notice guarantees and silently drops required notices on failure) or rewriting `academics.SQLStore` from `database/sql` to `pgx` (rejected to preserve `academics`'s driver-neutral store boundary).
>>>>>>> 7e300ec074132fac5cb72ca2428a933bfcbc5172
