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
