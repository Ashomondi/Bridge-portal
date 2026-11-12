
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
