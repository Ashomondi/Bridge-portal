# ARCHITECTURE.md

The system design and data model. Update this whenever a structural decision
changes — an AI assistant (and any new teammate) relies on this being accurate,
not just directionally right.

## Shape of the system

```
                       ┌─────────────────────┐
  Parent / Teacher /   │   Next.js frontend   │
  Staff (browser/PWA)  │   (React, PWA)       │
                       └──────────┬───────────┘
                                  │ HTTPS
                       ┌──────────▼───────────┐
                       │   Go backend          │
                       │   (modular monolith)  │      ┌───────────────┐
                       │  - billing            │◄────►│  PostgreSQL   │
                       │  - academics           │      │  (RLS per     │
                       │  - onboarding          │      │   school_id)  │
                       │  - notifications        │      └───────────────┘
                       │  - permission checks    │
                       └───┬───────────────┬────┘
                           │               │
              ┌────────────▼───┐   ┌───────▼────────────┐
              │ Redis + job     │   │  AI service         │
              │ queue (Asynq)   │   │  (Python/FastAPI)   │
              │ - SMS jobs      │   │  - self-hosted       │
              │ - STK polling   │   │    Mistral model      │
              │ - AI reminders  │   │  - calls Go endpoints │
              └─────────────────┘   │    only, no direct DB │
                                    └────────────────────────┘
        External: Africa's Talking (SMS/USSD), WhatsApp Cloud API,
                  Safaricom Daraja (M-Pesa)
```

**Why a modular monolith, not microservices:** at this stage, one deployable Go
service with clear internal package boundaries is far easier to build, test,
and reason about than a distributed system. Split a module out only when it has
a real, separate scaling or ownership need (the AI service is the one exception,
made separate from day one — see below).

**Why the AI service is separate:** it's written in Python to use Mistral's
tooling, it's the one piece most likely to be swapped or re-hosted, and keeping
it out of the Go monolith means it can never accidentally get direct database
access. It talks to the Go backend the same way the frontend does — over an
internal API, with the same permission checks applied.

## Tenancy model

- **One shared PostgreSQL database.** Every tenant-owned table carries
  `school_id`. Chosen over per-school databases for cost and operational
  simplicity — see `DECISIONS.md`.
- **Enforced with row-level security (RLS)**, not just application code. The
  app sets `SET LOCAL app.school_id = '<uuid>'` per request/transaction;
  without it, queries return no rows. Tables are `FORCE ROW LEVEL SECURITY`
  so even the owning role can't bypass it accidentally.
- **People vs. memberships:** a `users` table is global (one login per human).
  A `memberships` table links a user to a school under a role. This is what
  lets a parent with children at two schools, or a teacher who works at two
  schools, log in once.
- **Bind before querying:** `RequireIdentity` receives a verified active
  membership from trusted middleware. `TenantTxFunc.WithTenantTx` accepts that
  `Principal` and derives the school from `Principal.SchoolID`; it does not
  accept a second, request-selected school. The pool adapter must set the
  tenant locally within the transaction and use a non-superuser, non-BYPASSRLS
  application role. Parent/permission checks remain mandatory within a school.

## Roles vs. badges

- **Role** = job title at a school (teacher, nurse, bursar, parent, etc.) —
  stored on `memberships`.
- **Badge** = a permission bundle, defined per school, optionally scoped to a
  class, a club, or a time window. Examples: `class_teacher` (scoped to one
  class), `duty` (scoped to a week), `club_patron` (scoped to one club).
- A membership can hold several badges at once. A teacher can be a class
  teacher, on duty this week, and a club patron simultaneously — three badge
  rows, one membership.
- **Never hardcode "class teacher" as a column or role.** It's a badge scoped
  to a class, so there is one source of truth for who can see what.

## Core entities (see `migrations/` for the executable schema)

- `schools`, `users`, `memberships`
- `permissions`, `badges`, `badge_permissions`, `membership_badges`
- `grade_levels`, `academic_years`, `terms`, `classes`, `dormitories`, `clubs`
- `students`, `student_guardians`, `enrolments`, `club_enrolments`
- `consents`, `biometric_templates`
- `staff_invites`, `staff_applications`
- `fee_structures`, `fee_structure_items`
- `audit_log`

Performance history (the parent-facing graph from the day a child joined) is
built from one `enrolments` row per student per academic year — there is no
separate "history" table to keep in sync.

## Release 1 billing and chat foundation (migration 000004)

Migration 000004 adds **12 tables**, all with `school_id` tenant ownership plus
enabled and forced tenant RLS (`school_id = current_school_id()`). It also adds
composite `(school_id, id)` unique keys on `academic_years`, `terms`, `students`,
and `payments` (plus `(school_id, academic_year_id, id)` on `terms`), adds
`boarding_status` variants (`all`, `boarder`, `day`) to `fee_structures`, and
adds `credit_balance`, `credit_applied`, `credit_transferred`, and `due_date`
to `fee_ledgers`. Schema availability does not mean the related handlers,
workers, permission checks or audit sinks are implemented.

1. **`school_mpesa_configs`**
   - **Purpose:** Stores per-school Safaricom Daraja Paybill/Till shortcode, environment (`sandbox` / `production`), encrypted passkey, and test-payment verification state.
   - **Ownership & tenant relationship:** Owned by `schools` (`school_id` is both `PRIMARY KEY` and `FOREIGN KEY REFERENCES schools(id) ON DELETE CASCADE`); isolated via forced RLS.
   - **Important constraints:** `shortcode` is `UNIQUE NOT NULL`; `environment IN ('sandbox', 'production')` (default `'sandbox'`); `test_payment_verified BOOLEAN NOT NULL DEFAULT FALSE`.
   - **Relationships:** 1:1 with `schools`.
   - **Domain workflow role:** Supplies tenant payment shortcode/credentials for STK push initiation and Daraja callback resolution.

2. **`fee_periods`**
   - **Purpose:** Defines named billing periods and due dates per school, academic year, and optional term.
   - **Ownership & tenant relationship:** Tenant-scoped via `school_id REFERENCES schools(id) ON DELETE CASCADE`; isolated via forced RLS.
   - **Important constraints:** `btrim(period_type) <> ''` (default `'term'`), `btrim(name) <> ''`, `CHECK (period_type <> 'term' OR term_id IS NOT NULL)`, `UNIQUE (school_id, id)`, `UNIQUE (school_id, academic_year_id, term_id)`.
   - **Relationships:** Composite tenant FKs `(school_id, academic_year_id) -> academic_years(school_id, id)` and `(school_id, academic_year_id, term_id) -> terms(school_id, academic_year_id, id)`. Referenced by `fee_charges` and `fee_change_notes`.
   - **Domain workflow role:** Anchors billing due dates and groups student assessments (`fee_charges`) and revision notes (`fee_change_notes`).

3. **`fee_charges`**
   - **Purpose:** Records the assessed fee amount and chosen `fee_structure_id` for a student within a `fee_period`.
   - **Ownership & tenant relationship:** Tenant-scoped via `school_id REFERENCES schools(id) ON DELETE CASCADE`; isolated via forced RLS.
   - **Important constraints:** `amount >= 0`, `UNIQUE (school_id, student_id, fee_period_id)`.
   - **Relationships:** Composite tenant FKs to `students(school_id, id)`, `fee_periods(school_id, id)`, and `fee_structures(school_id, id)`. Indexed on `(school_id, fee_period_id)` and `(school_id, fee_structure_id)`.
   - **Domain workflow role:** Authoritative assessment record driving term charges posted to `ledger_entries` and projected onto `fee_ledgers`.

4. **`fee_change_notes`**
   - **Purpose:** Records authorized fee amount deltas or due-date revisions, reason text, actor, and affected student count.
   - **Ownership & tenant relationship:** Tenant-scoped via `school_id REFERENCES schools(id) ON DELETE CASCADE`; isolated via forced RLS.
   - **Important constraints:** `change_type IN ('amount', 'due_date')`, `btrim(note) <> ''`, `affected_students_count >= 0`, `UNIQUE (school_id, id)`.
   - **Relationships:** `created_by REFERENCES users(id)`; composite tenant FKs `(school_id, term_id) -> terms(school_id, id)` and `(school_id, fee_period_id) -> fee_periods(school_id, id)`. Parent of `fee_change_note_students`.
   - **Domain workflow role:** Captures why and by whom fee charges or due dates changed after initial assessment.

5. **`fee_change_note_students`**
   - **Purpose:** Preserves the exact subset of students affected by each `fee_change_notes` entry.
   - **Ownership & tenant relationship:** Tenant-scoped via `school_id REFERENCES schools(id) ON DELETE CASCADE`; isolated via forced RLS.
   - **Important constraints:** `PRIMARY KEY (school_id, note_id, student_id)`.
   - **Relationships:** Composite tenant FKs `(school_id, note_id) -> fee_change_notes(school_id, id) ON DELETE CASCADE` and `(school_id, student_id) -> students(school_id, id)`. Indexed on `(school_id, student_id)`.
   - **Domain workflow role:** Prevents losing student-level scope when a fee change note only applies to specific students in a period.

6. **`ledger_entries`**
   - **Purpose:** Append-style idempotent posting log for post-migration ledger operations (`charge`, `payment`, `credit_applied`, `adjustment`).
   - **Ownership & tenant relationship:** Tenant-scoped via `school_id REFERENCES schools(id) ON DELETE CASCADE`; isolated via forced RLS.
   - **Important constraints:** `entry_type IN ('charge', 'payment', 'credit_applied', 'adjustment')`, `UNIQUE (school_id, student_id, term_id, entry_type, reference)`.
   - **Relationships:** Composite tenant FKs `(school_id, student_id) -> students(school_id, id)` and `(school_id, term_id) -> terms(school_id, id)`. Indexed on `(school_id, student_id, created_at, id)`.
   - **Domain workflow role:** Makes billing postings idempotent via `(school_id, student_id, term_id, entry_type, reference)` alongside materialized `fee_ledgers` updates.

7. **`unmatched_payments`**
   - **Purpose:** Queue of settled payments whose student could not be automatically matched, together with human resolution tracking.
   - **Ownership & tenant relationship:** Tenant-scoped via `school_id REFERENCES schools(id) ON DELETE CASCADE`; isolated via forced RLS.
   - **Important constraints:** `amount > 0`, `status IN ('unmatched', 'resolved')`, `UNIQUE (school_id, payment_id)`, `UNIQUE (school_id, transaction_ref)`, and a `CHECK` enforcing that `unmatched` rows have NULL `resolved_at`, `resolved_student_id`, and `resolved_by` while `resolved` rows require all three non-NULL. Partial queue index on `(school_id, created_at, id) WHERE status = 'unmatched'`.
   - **Relationships:** Composite tenant FKs `(school_id, payment_id) -> payments(school_id, id) ON DELETE CASCADE` and `(school_id, resolved_student_id) -> students(school_id, id)`; `resolved_by REFERENCES users(id)`.
   - **Domain workflow role:** Feeds AI reconciliation suggestions and requires bursar confirmation before ledger posting.

8. **`mpesa_stk_requests`**
   - **Purpose:** Tracks durable M-Pesa STK Push payment attempts and settlement/timeout/uncertain states separately from settled `payments`.
   - **Ownership & tenant relationship:** Tenant-scoped via `school_id REFERENCES schools(id) ON DELETE CASCADE`; isolated via forced RLS.
   - **Important constraints:** `amount > 0`, `checkout_request_id VARCHAR(100) UNIQUE`, `UNIQUE (school_id, transaction_ref)` (per-school unique receipt reference when non-null while allowing multiple `NULL` attempts), `status IN ('initiating', 'pending', 'succeeded', 'failed', 'timeout', 'uncertain')`.
   - **Relationships:** Composite tenant FK `(school_id, student_id) -> students(school_id, id)`; `guardian_user_id REFERENCES users(id)`. Indexed on `(school_id, status, created_at)`.
   - **Domain workflow role:** Supports idempotent STK initiation, callback/polling reconciliation, and duplicate receipt prevention per school.

9. **`school_whatsapp_bots`**
   - **Purpose:** Stores school-to-WhatsApp Business Cloud API phone number mappings (`phone_number_id`, `display_phone_number`) for future Release 3 activation.
   - **Ownership & tenant relationship:** Tenant-scoped via `school_id REFERENCES schools(id) ON DELETE CASCADE`; isolated via forced RLS.
   - **Important constraints:** `phone_number_id VARCHAR(100) UNIQUE NOT NULL`, `display_phone_number VARCHAR(30) UNIQUE NOT NULL`, `is_active BOOLEAN NOT NULL DEFAULT TRUE`.
   - **Relationships:** References `schools(id)`. Indexed on `(school_id)`.
   - **Domain workflow role:** Reserved schema for Release 3 WhatsApp bot routing; unused in Release 1 (`is_active` default is not a Release 1 activation flag).

10. **`chat_sessions`**
    - **Purpose:** Tracks parent chat verification state, OTP challenge state/attempts, delegated access-token hashes, expiry, and lookup/payment counters.
    - **Ownership & tenant relationship:** Tenant-scoped via `school_id REFERENCES schools(id) ON DELETE CASCADE`; isolated via forced RLS.
    - **Important constraints:** `channel IN ('whatsapp', 'in_app')`, `state IN ('AWAITING_ADMISSION_NUMBER', 'AWAITING_GUARDIAN_SELECTION', 'AWAITING_OTP', 'VERIFIED')`, `otp_attempts >= 0`, `lookup_count >= 0`, `payment_init_count >= 0`, `UNIQUE (school_id, id)`, partial unique index `chat_sessions_sender_idx ON (school_id, sender_wa_number) WHERE channel = 'whatsapp'`, and `CHECK (state <> 'VERIFIED' OR (guardian_user_id IS NOT NULL AND access_token_hash IS NOT NULL AND btrim(access_token_hash) <> '' AND verified_at IS NOT NULL AND expires_at IS NOT NULL AND expires_at > verified_at))`.
    - **Relationships:** Composite tenant FK `(school_id, student_id) -> students(school_id, id)`; `guardian_user_id REFERENCES users(id)`. Referenced by `ai_action_logs(school_id, session_id)`.
    - **Domain workflow role:** Enforces step-up verification before delegated parent assistant queries or STK payment initiation.

11. **`in_app_notifications`**
    - **Purpose:** Stores per-recipient in-app inbox notifications and read status/timestamps.
    - **Ownership & tenant relationship:** Tenant-scoped via `school_id REFERENCES schools(id) ON DELETE CASCADE`; isolated via forced RLS.
    - **Important constraints:** `btrim(category) <> ''`, `btrim(title) <> ''`, `btrim(body) <> ''`, `jsonb_typeof(metadata) = 'object'`, `CHECK (is_read = (read_at IS NOT NULL))`.
    - **Relationships:** `recipient_user_id REFERENCES users(id) ON DELETE CASCADE`. Indexed on `(school_id, recipient_user_id, created_at DESC, id)`.
    - **Domain workflow role:** Release 1 in-app notification inbox delivery and read-state tracking.

12. **`ai_action_logs`**
    - **Purpose:** Records redacted AI assistant feature invocations, intent, endpoint templates read, and action outcomes.
    - **Ownership & tenant relationship:** Tenant-scoped via `school_id REFERENCES schools(id) ON DELETE CASCADE`; isolated via forced RLS.
    - **Important constraints:** `channel IN ('whatsapp', 'in_app')`, `jsonb_typeof(endpoints_read) = 'array'`, `jsonb_typeof(metadata) = 'object'`.
    - **Relationships:** `user_id REFERENCES users(id)`; composite tenant FK `(school_id, session_id) -> chat_sessions(school_id, id)` (non-cascading). Indexed on `(school_id, session_id, created_at)`.
    - **Domain workflow role:** Provides redacted operational AI action logging (`SECURITY.md`), separate from `audit_log` / `AuditLogger` financial audit requirements.

### Fee ledgers: sources and balance transitions

`fee_structures` and their items are pricing templates. `fee_charges` records
the amount assessed to a student; subsequent changes need a `fee_change_notes`
record and exact membership in `fee_change_note_students`. `payments` is the
settled receipt source. An STK attempt is never itself a receipt.

`ledger_entries` is the posting history for new billing operations;
`fee_ledgers` is the materialized per-student/year/term balance used for reads.
The existing ledger is the opening source for legacy history: migration 000004
does not synthesize entries for old transactions. Rebuilding balances must
start from a captured legacy opening snapshot plus subsequent postings, not
from the new entries alone. Use the unique entry reference to make retries
idempotent; write source records, postings, balance updates and audit events in
one tenant transaction while locking the affected ledger rows.

`current_balance` is outstanding debt and `credit_balance` is available prepaid
credit, both nonnegative. Charging consumes available credit first, then raises
debt; payment reduces debt first, then adds excess to available credit.
`credit_applied` accumulates credit consumed in that ledger and
`credit_transferred` accumulates credit moved out to another ledger. A transfer
locks both ledgers and decreases source credit/increases destination credit
atomically; it creates no new payment. Charge/adjustment postings use signed
debt deltas, payment and credit-applied postings record positive applied amounts;
transfer adjustments must have paired references and descriptions.

There is no persisted ledger status enum. Derived states are owing
(`current_balance > 0`), settled (both balances zero), and in-credit
(`credit_balance > 0`); service operations should net debt against credit before
commit. The database enforces nonnegativity, not the accounting workflow.
Up migration converts negative legacy debt to credit. Down migration restores
`current_balance - credit_balance` and removes the added credit counters.

### STK requests

`mpesa_stk_requests` is authoritative for attempt state:
`initiating` → `pending` when Daraja accepts; `initiating` → `failed` for an
explicit rejection, or `uncertain` if acceptance cannot be determined.
`pending` → `succeeded` / `failed` on an authenticated provider result, or
`timeout` after the polling deadline. `uncertain` / `timeout` may resolve to
`pending`, `succeeded` or `failed` after provider reconciliation. A timeout is
not proof that no money moved. `succeeded` and provider-confirmed `failed` are
terminal; duplicate callbacks are no-ops. Contradictory evidence requires
operator reconciliation, never blind replay.

Only verified settlement creates a `payments` receipt and applies the ledger
posting in the same idempotent transaction. Checkout identifiers are globally
unique (`checkout_request_id UNIQUE`), and M-Pesa receipt references are unique
per school (`UNIQUE (school_id, transaction_ref)` on `mpesa_stk_requests` while
allowing multiple `NULL` attempts prior to settlement); retries create a new
attempt only after checking the earlier one. The schema checks state names and
identifier uniqueness; transition and callback authentication rules must be
implemented in the billing service.

### Unmatched payments

`payments` owns receipt amount, reference, provider and settlement time.
`unmatched_payments` stores a queue snapshot plus resolution state; it is not a
second receipt and must never independently credit a ledger. State is
`unmatched` → `resolved` only when a permitted bursar confirms the student.
An AI suggestion leaves the queue unmatched (`payments.matched_status` may be
`ai_suggested`). Resolution sets student, actor and timestamp, updates the
payment to `matched`, and posts exactly once in the same locked transaction.
`resolved` is terminal; corrections require an audited compensating operation.
The database checks resolution fields but does not implement this workflow.

### Chat sessions

`chat_sessions` owns verification state, not the sender's claimed identity:
`AWAITING_ADMISSION_NUMBER` → `AWAITING_GUARDIAN_SELECTION` when multiple
guardians require selection, or directly to `AWAITING_OTP` for a single
eligible guardian. Selection → `AWAITING_OTP`; a valid OTP → `VERIFIED`.
Admission lookup alone never grants access. OTP failure increments attempts;
expiry or attempt exhaustion clears the challenge and restarts verification.

`VERIFIED` requires a guardian (`guardian_user_id`), non-null/non-blank
`access_token_hash`, verification timestamp (`verified_at`), and a non-null
expiry (`expires_at`) strictly after verification (`expires_at > verified_at`).
At request time, also require `expires_at > now()` and a current guardian
relationship. Expiry/revocation clears OTP/token hashes and returns to the
admission state (`AWAITING_ADMISSION_NUMBER`, where `access_token_hash` may be
NULL); expiry is enforced even before cleanup runs. No `EXPIRED` enum exists.
`sender_wa_number` is routing data, never proof of guardianship. The constraint
verifies field shape, not OTP possession or current authorization. Release 1
uses in-app chat; WhatsApp routing tables prepare for Release 3 (see the ADR in
`DECISIONS.md`).

### Downgrade rule

Before downgrading 000004, export its new tables and boarding variant
definitions. The old uniqueness key cannot represent alternatives: for each
school/year/grade/term retain `all`, otherwise `day`, otherwise `boarder`, with
the retained row's ID, total and items. Other variants and their items are
deleted; alternative fees are never summed. The fallback favors day pricing
over boarding pricing to avoid adding boarding charges to the generic template.
This is a lossy compatibility downgrade, not a historical restore. Reapplying
000004 marks the retained template as `all`; restore exports to recover other
variants. Rollback tests cover these choices and credit round trips.

## Curriculum model

CBC only (Grade 7–12), not 8-4-4. `grade_levels.number` is constrained to
7–12, with a `phase` of `junior` or `senior`. If legacy Form 3/4 support is
ever needed, that's a schema change, not a workaround — raise it as a new
decision rather than bolting it on.

## Channels and sessions (added 2026-10-29)

- **WhatsApp Business Cloud API, one bot per school.** The receiving number
  maps to `school_id`, so an admission number is only ever looked up inside
  that school. The Go backend owns the webhook; the AI service is reached only
  through the backend.
- **Parent identity in chat:** sender enters the admission number → masked
  guardian numbers → OTP by SMS to the chosen registered number → a 6-hour
  verified session. Any WhatsApp number may start a chat; the OTP goes only
  to the registered number.
- **Release 1 event notifications are in-app only.** SMS is used for OTP and
  security messages.

## Additional entities (planned)

`school_whatsapp_bots`, `chat_sessions`, `otp_requests`, `parent_qr_codes`,
`staff_qr_codes`, `registration_submissions`, `verification_records`, badge
delegations, `attendance_daily`, `assessments` and `marks`, `report_comments`,
`result_releases`, `grading_scale_versions` (per term), `fee_periods`,
`fee_charges`, `fee_change_notes`, `ledger_entries` (including credit),
`unmatched_payments`, `school_profile` (contact details required), and a
`whatsapp` value for `consents.purpose`. See `ikonex.md` section 9.

Build `fee_periods` on generic billing periods even though Release 1 is terms
only, so custom structures later need no migration (a recommendation).

## Behaviour that shapes the model

- **Results release is a state machine** per class per term: draft → released
  → reopened → re-released. Parents keep seeing the last released version
  while a class is reopened.
- **Grading scale is versioned per term.** A term's levels are computed from
  the scale in force when it was released.
- **Fee due dates follow term start dates.** A term-date change moves due
  dates and creates a fee-change event with a note.
- **Pending registrations are not active records** and grant no access until
  a staff verifier confirms them; uploaded documents are deleted after
  approval.

## Settled vs. open decisions

Settled decisions and their reasoning live in `DECISIONS.md` — don't
re-litigate them here. Open questions that affect this document:

- Hosting location (Kenya only vs regional Africa, and where the AI model
  runs) — to be discussed with the project group.
- Sports/activities performance model (how it's scored, who enters it) — not
  yet designed.
- Store and caterer module data model — not yet designed.
- Fee structure variants (boarder/day, optional items).
- How a shared parent QR links each parent to the right child.
- Payment execution: the system records approved payments; whether it will
  ever initiate transfers itself (beyond the parent-initiated M-Pesa STK push)
  is undecided.

## Notifications engine (`backend/internal/notifications`)

The notifications package provides the Release 1 synchronous transactional SMS
transport and transaction-bound in-app notifications service. Application-level
wiring (`cmd/server/main.go`, `internal/config`, `internal/jobs`,
`cmd/worker/main.go`, auth/onboarding callers) remains pending.

### SMS transport versus SMS queue

- **Synchronous SMS transport (implemented):** `notifications.SMSDispatcher`
  (`DispatchSMS(context.Context, SMSMessage) error`) validates phone numbers
  and allowed purposes, then dispatches synchronously via either
  `NewSandboxSMSDispatcher` (local redacted logging for username `sandbox`) or
  the Africa's Talking HTTP adapter (`NewSMSDispatcher`).
- **Redis/Asynq SMS queue (pending future integration):** Asynchronous job
  scheduling, Redis persistence, backoff retries, and dead-letter handling
  belong to `internal/jobs` and `cmd/worker` (currently placeholders). No
  durable queue exists yet in this codebase.
- **Retry & error classification boundary:** The transport surfaces structured
  `*SMSDeliveryError{Code, Retryable}` values (e.g., HTTP `408`/`429`/`5xx`,
  transport timeouts, and provider `500`–`502` recipient codes marked
  `Retryable: true`; redirects, `4xx` rejections, recipient mismatch, and
  malformed/oversized responses marked `Retryable: false`) so the future Asynq
  worker can decide whether to retry without inspecting raw provider bodies.
  Ambiguous network failures after a request is sent cannot guarantee
  exactly-once SMS delivery.

### Transactional-only SMS purposes (Release 1)

Release 1 SMS transport is restricted to four explicit `SMSPurpose` values
(`C58`):

- `app_login` — app login OTP
- `whatsapp_session` — Release 3 WhatsApp session OTP purpose key (reserved in
  the transport enum; WhatsApp integration itself is deferred to Release 3)
- `phone_change` — guardian 2FA phone-change OTP to the new number
- `phone_change_notice` — security alert to the old number on phone change

Event-based messages (`absence`, `sickbay`, results, timetable, calendar, fee
alerts) are explicitly rejected by `SMSMessage` validation (`ErrInvalidSMS`).
Release 1 event notifications use `in_app_notifications` instead.

### Africa's Talking provider boundary

- **HTTP client & timeouts:** `NewSMSDispatcher` copies any caller-supplied
  `*http.Client` (never mutating a shared client) and enforces a default
  15-second timeout when `Timeout <= 0`.
- **Redirect protection:** Installs `CheckRedirect` returning
  `http.ErrUseLastResponse` so HTTP redirects are rejected (`Retryable: false`)
  and `apiKey` headers / form bodies are never forwarded to redirect targets.
- **Credential & error redaction:** `SMSConfig` carries `Username`, `APIKey`,
  and optional `SenderID` (`AFRICASTALKING_*` env vars; config wiring is still
  pending). Returned errors (`ErrInvalidPhone`, `ErrInvalidSMS`,
  `*SMSDeliveryError`) never include the phone number, message body, OTP,
  `APIKey`, or raw provider response body. Response reads are capped at 64 KiB
  and verify single-recipient normalized phone matching.

### In-app notifications

- **Ownership & tenant isolation:** Every row in `in_app_notifications` belongs
  to `(school_id, recipient_user_id)` with PostgreSQL `ENABLE` and `FORCE ROW
  LEVEL SECURITY` (`school_id = current_school_id()`). `List` and `MarkRead`
  derive `school_id` exclusively from `contracts.Principal` via
  `contracts.TenantDB.WithTenantTx` and filter by `recipient_user_id =
  Principal.UserID`.
- **Transaction-bound creation (`contracts.InAppNotifier`):** Business services
  begin their tenant-scoped transaction, perform domain writes, and call
  `CreateInAppNotification(ctx, tx, notification)` using `contracts.SQLTx`
  (`contracts.PgxTx` for `pgx.Tx`, `contracts.StdSQLTx` for `*sql.Tx`) inside
  that same transaction. `notifications.Service` verifies an active
  `memberships` + `users` row for `(school_id, recipient_user_id)` and inserts
  the notification without starting its own transaction or sending SMS. Any
  failure propagates to the caller so the entire business transaction rolls back
  atomically.
- **Read receipt semantics & pagination:** `MarkRead` sets `is_read = true` and
  `read_at = coalesce(read_at, now())` so repeat marks preserve the first read
  timestamp; cross-user or cross-school IDs return `404 NOTIFICATION_NOT_FOUND`.
  `List` uses `contracts.Pagination` (default `page=1`, `per_page=25`, max
  `per_page=100`) ordered by `created_at DESC, id`.
- **Consumer integration status:** `internal/academics` adopts
  `contracts.InAppNotifier` directly and executes all in-app notifications
  (attendance absence, results release, results re-release, timetable update,
  calendar event update) inside `Store.WithTx` with error propagation.
  `internal/billing`, `internal/auth`, `internal/onboarding`, `internal/jobs`,
  and `cmd/server` route mounting remain pending wiring.
