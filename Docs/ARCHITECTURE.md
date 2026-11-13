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

Migration 000004 adds **12 tables**, all with enabled and forced tenant RLS:

| Table | Responsibility |
|---|---|
| `school_mpesa_configs` | School shortcode, environment, encrypted passkey and test-payment verification |
| `fee_periods` | Named billing period and due date within a school/year/term |
| `fee_charges` | Student's assessed amount and selected structure for a period |
| `fee_change_notes` | Actor, reason, amount/date changes and affected count |
| `fee_change_note_students` | Exact students affected by each change note |
| `ledger_entries` | Idempotent financial postings and references |
| `unmatched_payments` | Reconciliation queue and human resolution record |
| `mpesa_stk_requests` | Durable payment attempts, including uncertain outcomes |
| `school_whatsapp_bots` | School-to-provider phone identity mapping for future WhatsApp activation |
| `chat_sessions` | Channel, guardian verification, OTP/token hashes and expiry |
| `in_app_notifications` | Recipient inbox and read state |
| `ai_action_logs` | Redacted AI intent, endpoint reads and action metadata |

It also adds boarding variants to `fee_structures` and credit/due-date fields
to `fee_ledgers`. Schema availability does not mean the related handlers,
workers, permission checks or audit sinks are implemented.

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
posting in the same idempotent transaction. Enforce unique checkout/receipt
identifiers; retries create a new attempt only after checking the earlier one.
The schema checks state names; transition and callback authentication rules
must be implemented in the billing service.

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

`VERIFIED` requires a guardian, verification timestamp, and a non-null expiry
strictly after verification. At request time, also require `expires_at > now()`
and a current guardian relationship. Expiry/revocation clears OTP/token hashes
and returns to the admission state; expiry is enforced even before cleanup
runs. No `EXPIRED` enum exists. `sender_wa_number` is routing data, never proof
of guardianship. The constraint verifies field shape, not OTP possession or
current authorization. Release 1 uses in-app chat; WhatsApp routing tables
prepare for Release 3 (see the ADR in `DECISIONS.md`).

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

## Settled vs. open decisions

Settled decisions and their reasoning live in `DECISIONS.md` — don't
re-litigate them here. Open questions that affect this document:

- Sports/activities performance model (how it's scored, who enters it) — not
  yet designed.
- Store and caterer module data model — not yet designed.
- Payment execution: the system records approved payments; whether it will
  ever initiate transfers itself (beyond the parent-initiated M-Pesa STK push)
  is undecided.
