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

**2026-11-13 — ADR: Tenant transactions derive the school from the principal**

`TenantTxFunc.WithTenantTx` accepts a `Principal` produced by `RequireIdentity`
using trusted authentication middleware, validates its shape, and passes only
`Principal.SchoolID` to the application's transaction function. That function
must set transaction-local `app.school_id` and use a role subject to RLS.
Request-supplied school IDs cannot select the transaction tenant. Switching
schools requires revalidating the user's active membership before issuing the
principal. Permission and guardian checks still apply within a school.

*Why:* validating a UUID does not authorize access to that school.
*Alternative considered:* accepting a separate school argument and relying on
every handler to compare it with the authenticated school. Removed to avoid a
second tenant selector. Trusted identity verification remains the middleware's
responsibility; the contracts are not an authentication implementation.

**2026-11-13 — ADR: Separate debt and available credit, with a defined downgrade**

Use nonnegative `fee_ledgers.current_balance` for debt and `credit_balance` for
unspent credit. `credit_applied` and `credit_transferred` are cumulative usage
counters, not additional spendable balances. Migration 000004 backfills legacy
negative balances into credit before adding the nonnegative constraint. The
downgrade restores the signed balance as debt minus available credit.

`fee_charges` and `payments` are assessment/settlement sources;
`ledger_entries` records new postings and `fee_ledgers` is the read projection.
Legacy ledger snapshots are opening balances because the migration does not
backfill posting history. STK attempt state is separate from settlement;
unmatched receipts require human resolution before ledger credit. See
`ARCHITECTURE.md` for the ledger, STK, unmatched-payment and chat state machines.

For boarding variants, downgrade retains `all`, otherwise `day`, otherwise
`boarder` per school/year/grade/term, keeping that template and its items intact.
Discard the other variants and their items after exporting them. Day pricing
is the fallback to avoid applying boarding fees to a generic template. Up again
cannot recover discarded variants. Run rollback only with a restorable export.

*Why:* explicit credit avoids displaying overpayment as debt; a deterministic
downgrade avoids a uniqueness failure when both day and boarder variants exist.
*Alternatives considered:* keep signed debt indefinitely; sum variant prices;
or choose an arbitrary template. Separate balances and an explicit precedence
rule make financial meaning and downgrade data loss reviewable.

**2026-11-13 — ADR: Release 1 chat schema, Release 3 WhatsApp delivery**

Release 1 provides in-app parent chat and parent-confirmed payment flows.
Migration 000004 creates the shared chat/verification foundation and reserves
WhatsApp routing support; it does not bring WhatsApp delivery into Release 1.
`school_whatsapp_bots` remains unused until Release 3 integration and activation.
The `is_active` database default is not a deployment feature flag. WhatsApp
webhook routing, provider credentials, templates and outbound delivery must
remain disabled/unwired until that release and the transfer requirements in
`SECURITY.md` are met. SMS and in-app notifications remain Release 1 channels.

*Why:* reconciles the forward-compatible schema with the existing release plan
in `REQUIREMENTS.md`; a table is not a shipped channel.
*Alternative considered:* implicitly move WhatsApp delivery into Release 1
because its table exists. Rejected as an unapproved scope change.

The 12 new tables are `school_mpesa_configs`, `fee_periods`, `fee_charges`,
`fee_change_notes`, `fee_change_note_students`, `ledger_entries`,
`unmatched_payments`, `mpesa_stk_requests`, `school_whatsapp_bots`,
`chat_sessions`, `in_app_notifications`, and `ai_action_logs`.

**2026-11-13 — ADR: AI service credential plus delegated session**

Internal AI calls require a service bearer token and a Go-verifiable delegated
user session, with per-request membership, school, guardian and permission
checks. A service credential alone cannot select a tenant or impersonate a
parent. The header protocol is specified in `API_CONTRACTS.md`; middleware is
pending. Short-lived opaque session tokens allow revocation and are stored only
as hashes. Do not accept unsigned forwarded identity headers.

*Why:* authenticating the service and authorizing the human are separate checks.
*Alternative considered:* a shared service token plus raw user/school headers;
rejected because it permits unverified delegation.
