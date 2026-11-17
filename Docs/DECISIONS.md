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
