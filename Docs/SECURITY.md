# SECURITY.md

Consent, encryption, retention, and AI-data-boundary rules. This project
handles children's personal data — including biometric and health data —
under Kenya's Data Protection Act, 2019. This file is not legal advice; have
a Kenyan data protection lawyer or the ODPC review the platform before
processing real student data at scale. Primary sources: the Act itself, the
Data Protection (General) Regulations, 2021, and the ODPC's 2026 Guidance
Notes on Cross-Border Data Transfers — read those directly rather than
relying only on this summary.

## Data categories and how they're treated

| Category | Examples in this system | Rule |
|---|---|---|
| Sensitive personal data | biometric templates, health/nurse records | Extra processing grounds required (Act Part V); strict access control; never sent to the AI service in raw form |
| Personal data (general) | name, DOB, admission number, results | Standard processing rules: lawful, purpose-limited, accurate, not kept longer than needed |
| Financial data | fee balances, payment history | Standard rules, plus signatory approval for anything that moves money |
| Aggregate/derived data | class means, subject means | Not linked to individual students when shown outside the owning class teacher |

## Consent

- **Children's data requires parental/guardian consent.** Record consent per
  student, per purpose (`consents` table: e.g. `biometric_attendance`,
  `photo`, `sms`), with a timestamp and a way to revoke it.
- **Always offer a non-biometric alternative** for attendance and any other
  process that currently defaults to biometrics, for a guardian who declines.
- **Consent is purpose-specific.** Data collected for attendance isn't reused
  for something else (e.g. marketing, a different module) without fresh
  consent.
- **Parents can see what's held and withdraw consent** — this is the planned
  "consent centre" feature (see `REQUIREMENTS.md`, Platform foundation).

## Biometric data

- **Store encrypted templates only. Never store the raw fingerprint image.**
  Discard the raw scan immediately after the template is generated.
- **Per-school encryption keys**, envelope-encrypted via a key management
  service. Templates live in their own table (`biometric_templates`) with
  more restricted access than the rest of the schema.
- **Delete on exit.** When a student leaves the school, their biometric
  templates are deleted, not just marked inactive.
- **Audit every read** of the biometric table.

## Health data (nurse records)

- **Full clinical detail is restricted to the nurse role.**
- **Class teacher notification is a flag, not a copy of the record** — e.g.
  "visited the sickbay, advised rest," never symptoms or diagnosis.
- **Parent-facing SMS follows the same rule**: confirms a visit happened,
  not clinical detail (see `REQUIREMENTS.md`, Nurse module).
- Health data is never sent to the AI service as free text; if the AI ever
  needs to summarise nurse trends for reporting, it works from pre-aggregated,
  de-identified counts, not individual records.

## AI service data boundary

This is the section most specific to this project's AI features — see also
`API_CONTRACTS.md` for how it's enforced technically.

1. **The AI service is self-hosted, in Kenya**, and has no direct database
   access (enforced at the API layer, not just by convention).
2. **It receives only what a specific, permission-checked endpoint returns**
   — e.g. a parent-assistant query about a fee balance gets the balance, not
   the student's full record.
3. **It never receives raw biometric or nurse-record data.**
4. **It never originates a fact.** Fee balances, attendance, and results come
   from the database; the model's job is language (phrasing, translation,
   mapping natural language to CBC strands), never invention.
5. **Every AI-initiated action must be logged**: structured intent, endpoint
   names read, action/outcome, actor, school and feature. Do not log raw prompts,
   replies, access tokens or returned student records. `ai_action_logs` is a
   schema foundation and `AuditLogger` is an interface; durable logging is not
   wired by these contracts and AC-29 remains pending.
6. **A human approves anything sent to a parent or written to a permanent
   record** — report comments, blog posts, plain-language explanations — the
   AI drafts, a teacher/patron/bursar confirms.
7. **Payments: the AI proposes, a human confirms.** The AI service can
   initiate an STK push only after the parent has explicitly confirmed the
   child and amount in the conversation; it never sends money without that
   confirmation, and it never initiates supplier payments at all — those go
   through the bursar signatory workflow.

## Cross-border transfer

- **Default assumption: student and staff personal data stays hosted in
  Kenya.** This is both a compliance position and a product claim (see the
  bootcamp application materials), so don't add a dependency that silently
  routes personal data through a foreign service.
- If a foreign service is ever genuinely needed (e.g. a third-party API with
  no Kenyan equivalent), check Sections 48–50 of the Act first: transfers
  outside Kenya require proof of appropriate safeguards to the Data
  Commissioner, and for sensitive personal data, case law (Federation of
  Kenya Employers v. Ministry of Foreign Affairs, 2023) confirms the
  Commissioner must be informed and satisfied the transfer meets the
  statutory requirements before it happens — this is not just a consent
  question.
- Remote access to Kenyan personal data from outside Kenya, and hosting it on
  a server abroad, both count as a "transfer" under the ODPC's 2026 guidance
  — not only a literal file copy.

## Retention

The following are engineering retention defaults for billing/chat; legal and
school-controller review must confirm the applicable financial retention period
before production. These are requirements for handlers and cleanup jobs, not a
claim that expiry/redaction workers are already implemented. Other categories
(results, attendance, consent records) still need controller-approved schedules.

| Data | Retention and deletion rule | Redaction/access rule |
|---|---|---|
| `chat_sessions.sender_wa_number` | Retain only while the session is active; clear within 24 hours of expiry/revocation. Remove expired session rows after 30 days, after detaching retained AI logs as described below. | Normalize for routing; restrict raw values to verification/delivery code. Mask all but the final 4 digits in staff UI; exclude from URLs, app logs, analytics and AI prompts. A WhatsApp number does not prove guardian identity. |
| `otp_hash`, `otp_expires_at`, `otp_attempts` | OTP lifetime at most 5 minutes and at most 5 failed attempts. Clear the hash on success, expiry, replacement or exhaustion; purge expired hashes within 24 hours even if no request returns. | Never persist/log plaintext OTPs. Use a keyed hash with a server-held secret because short numeric OTPs are vulnerable to offline guessing. Do not expose hashes to clients or the AI service. |
| `access_token_hash` and chat verification timestamps | Verification/access lifetime at most 24 hours. Clear hashes on revocation or expiry; cleanup within 24 hours. Validate expiry on every request; cleanup delay must never extend access. | Tokens must be cryptographically random and stored only as a hash. Redact bearer/delegation headers, cookies and token-bearing request fields at proxy and application layers. |
| `mpesa_stk_requests.phone_number`, `payments` / `unmatched_payments` payer phone, payer name and raw account reference | Retain through settlement/reconciliation, then redact within 90 days of terminal resolution. Review unresolved/uncertain records after 90 days and monthly thereafter; record an owner and time-bounded reason for extension. | Raw payer details only for authorized bursar reconciliation and delivery. Mask phone numbers in displays and notifications; never include them in model prompts, telemetry or error details. AI suggestions use minimal, scoped facts. |
| Provider `payments.raw_payload` | Allowlist fields at ingestion. Delete raw payload within 30 days of settlement, or resolution for an unmatched receipt. Keep canonical receipt fields separately. | Never retain PINs, credentials, authorization headers, or unrelated provider fields. Do not log raw webhooks. |
| Canonical financial records | Retain receipt reference, amount, posting dates and required ledger/audit evidence for 7 years after financial-year close by default; a documented legal hold overrides deletion. Payer contact details follow the shorter rule above unless specifically required by the hold. | Tenant- and permission-scoped access only; encrypt storage and backups. Avoid copying financial records into chat or AI logs. |
| `ai_action_logs` | Retain for 90 days, then delete; keep any legally required financial audit evidence separately in the financial audit trail for its retention period. | Allowlist feature, intent, endpoint templates, action/outcome, opaque resource IDs and timing. No prompts, message transcripts, response bodies, numbers, names, payer details, secrets or hashes in `metadata`. Endpoint templates must omit query strings and personal identifiers. |

Cleanup must execute per school with the same tenant protections as request
traffic. `ai_action_logs.session_id` currently has a non-cascading foreign key:
set it to NULL in the same tenant transaction before deleting a referenced
expired chat session (the log retains its own actor/school/action). Purge or
redact copies in queue payloads, dead-letter jobs, exports and notification
metadata too. Keep backups for at most 30 days; apply recorded deletions and
redactions after restore and before serving traffic. A legal hold must specify
scope, owner, reason and review/expiry date, not suspend deletion indefinitely.

Tests and examples use synthetic identities and payer data. Audit deletion
counts and policy outcomes without recording the removed personal data. Review
new metadata keys and provider fields against this allowlist before storage.

## Breach response

- Data controllers must report a breach to the ODPC within 72 hours, and to
  affected data subjects without undue delay.
- `TODO`: define the internal breach process (who's notified internally,
  how a breach is confirmed, who contacts the ODPC) before handling real
  student data.

## Registration

- Acting as a data controller or processor without registering with the ODPC
  is an offence under the Act. Confirm registration status before
  processing real student data — this is a business/legal task, not a
  coding task, but it blocks going live.
