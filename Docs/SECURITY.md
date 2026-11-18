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

## SMS and in-app notification security

- **Phone number normalization & masking:** `notifications.NormalizePhone`
  normalizes Kenyan local (`07...`, `01...`, `254...`) and international E.164
  numbers and rejects malformed, newline-containing, or out-of-range inputs
  with `ErrInvalidPhone` (never echoing the raw input). Sandbox logging masks
  recipient numbers to the trailing 3 digits (`***678`) and never logs the full
  number.
- **OTP & message-body redaction:** `SMSDispatcher` never logs message bodies,
  OTP codes, or full recipient numbers (only redacted recipient suffix, purpose,
  and rune length in sandbox mode). Callers must not place OTPs, credentials,
  or clinical details in `in_app_notifications` titles, bodies, or `metadata`.
- **Africa's Talking credential & error sanitization:** `APIKey` is transmitted
  only in the `apiKey` request header to `https://api.africastalking.com/version1/messaging`
  and never logged or surfaced in errors. Provider responses are capped at 64 KiB,
  parsed into a minimal status struct, verified for single-recipient phone
  match (`RECIPIENT_MISMATCH` otherwise), and mapped to safe
  `*SMSDeliveryError{Code, Retryable}` codes without raw provider text.
- **HTTP redirect protection, timeouts & retries:** `NewSMSDispatcher` copies
  the caller's `*http.Client`, enforces a default 15-second timeout when
  `Timeout <= 0`, and sets `CheckRedirect` to `http.ErrUseLastResponse` so
  redirects are never followed and `apiKey`/body cannot leak to a redirect
  target. Retries are not performed inside the transport; `*SMSDeliveryError.Retryable`
  signals safe retry eligibility (`408`, `429`, `5xx`, transport errors, provider
  `500`–`502`) to the future Asynq worker.

### Africa's Talking data boundary

- **What is transmitted & why:** Only the normalized recipient phone number
  (`to`), transactional SMS text (`message`), platform `username`, and optional
  `from` sender ID are sent over HTTPS `POST` to Africa's Talking strictly for
  the four permitted transactional purposes (`app_login`, `whatsapp_session`,
  `phone_change`, `phone_change_notice`). No student records, admission numbers,
  academic marks, health/nurse details, or fee balances are transmitted by the
  Release 1 SMS transport.
- **Recipient & data-location considerations:** Africa's Talking Limited
  processes the outbound SMS payload and passes carrier routing metadata to
  downstream mobile network operators (MNOs). Before processing real Kenyan
  user data in production, legal/compliance must confirm Africa's Talking's
  applicable data-processing agreement, infrastructure/sub-processor locations,
  and whether any cross-border transfer safeguards under Sections 48–50 of the
  Data Protection Act apply — no data-residency guarantee is assumed by code.
- **Control boundary after transmission:** Once Africa's Talking accepts a
  message (`statusCode` `100`/`101`/`102`), the application cannot recall,
  delete, or control retention of the SMS text or recipient number inside the
  provider or telecom carriers, and ambiguous network timeouts cannot prove
  whether the SMS was sent. Keep transactional SMS content minimal and short-lived.

## Retention

<<<<<<< HEAD
Policy is per record type. Periods are to be set with a Kenyan data
protection lawyer; the rules below are the starting point.

| Record | Rule | Status |
|---|---|---|
| Biometric templates | Deleted when the student leaves | Settled |
| Uploaded verification documents | Deleted after approval; only a verification record (who, when, document type) stays | Agreed |
| Rejected or never-verified applications | Deleted automatically after a fixed number of days (number to be agreed) | Agreed |
| Health (nurse) records | Deleted at exit | Proposed |
| Consent records | Deleted at exit | Proposed |
| Results | Archived for a set period | Proposed; period open |
| Financial ledger | Kept as the law requires | Proposed; period open |

What happens to a leftover ledger credit when a student leaves is undecided.
=======
Under Kenya's Data Protection Act, 2019, personal data must not be kept longer
than necessary for its lawful purpose. **Formal retention schedules have not yet
been established** for this platform: every post-session, post-reconciliation,
raw-payload, AI-log, backup, and canonical financial retention period below is a
**policy decision requiring legal and school-controller confirmation** before
production launch, not an existing approved rule. These rules also specify
requirements for future handlers and cleanup workers; expiry, redaction, and
audit-logging workers are **not yet implemented** in the current codebase.

| Data category | Stored & why | Who can access | Retention status (confirmation required) | Redaction / deletion & protection rules |
|---|---|---|---|---|
| **Chat phone numbers** (`chat_sessions.sender_wa_number`) | Stored in PostgreSQL (`chat_sessions`, RLS-isolated) solely to route inbound/outbound Release 3 WhatsApp chat turns to the active session. Not used in Release 1 in-app chat and never proof of guardian identity. | Restricted to backend messaging/verification code for that `school_id`; masked to final 4 digits if ever surfaced in staff UI; excluded from URLs, app logs, analytics, and AI prompts. | **Policy decision required:** confirm maximum inactive/post-expiry retention for routing phone numbers and expired `chat_sessions` rows (proposed engineering target: clear `sender_wa_number` on expiry/revocation and purge expired rows on a short schedule once detached from retained logs). | Clear `sender_wa_number` upon session revocation/expiry; normalize for routing only; protect at rest under forced tenant RLS (`school_id`) and encrypted storage. |
| **OTP hashes & challenge fields** (`chat_sessions.otp_hash`, `otp_expires_at`, `otp_attempts`) | **Plaintext OTPs are never stored or logged.** Only a server-keyed cryptographic hash and expiry/attempt counters are stored transiently in `chat_sessions` to enforce step-up guardian verification and rate/attempt limits. | Backend authentication/verification service only (`school_id` RLS-isolated). Never exposed to frontend clients, staff UI, logs, or the AI service. | **Policy decision required:** confirm exact OTP validity window, max failed attempts, and dead-challenge purge window (proposed engineering target: short-lived validity and immediate clearing on verify/replace/exhaustion, plus periodic purge of expired challenges). | Clear `otp_hash` immediately on successful verification, replacement, exhaustion, or expiry. Use a keyed hash with a server-held secret (never plain unsalted hashes for short numeric OTPs). |
| **Delegated access-token hashes & verification timestamps** (`chat_sessions.access_token_hash`, `verified_at`, `expires_at`) | **Plaintext delegated tokens are never stored or logged.** Only the cryptographic hash (`access_token_hash`, required non-null/non-blank in `VERIFIED` state by migration `000004`) and `verified_at`/`expires_at` are stored so Go can resolve `X-Delegated-Session` and enforce expiry. | Backend auth/delegation validation middleware only (`school_id` RLS-isolated). Never returned to clients after issuance or exposed to AI prompts or logs. | **Policy decision required:** confirm maximum delegated session TTL and post-expiry cleanup window (proposed engineering target: short-lived TTL with prompt cleanup after revocation/expiry). Cleanup delay must **never** extend access: handlers must enforce `expires_at > now()` on every request. | Generate cryptographically random opaque tokens; store only hashes; redact `Authorization` and `X-Delegated-Session` headers, cookies, and token fields at proxy and application layers. Transitioning out of `VERIFIED` or revoking clears `access_token_hash`, `verified_at`, and `expires_at`. |
| **M-Pesa payer information** (`mpesa_stk_requests.phone_number`, `payments.payer_phone`, `payments.payer_name`, `unmatched_payments.payer_phone`, `payer_name`, `raw_account_ref`) | Stored in PostgreSQL (RLS-isolated per `school_id`) to initiate STK pushes, display payment confirmation, and allow bursar/AI-assisted matching of unmatched receipts. | Tenant-scoped authorized bursar/admin staff and the paying guardian (for their own STK attempt); never exposed to teachers or unrelated parents. AI reconciliation receives only minimal, scoped facts—never raw phone dumps in prompts. | **Policy decision required:** confirm how long payer phone/name/raw account references may be retained after terminal STK outcome or payment reconciliation, and review cadence for `uncertain`/`unmatched` items. | Mask phone numbers in staff displays and notifications; exclude from telemetry, error payloads, and LLM prompts; redact/anonymize payer contact fields once the controller-confirmed post-settlement/reconciliation window closes (unless covered by a documented legal hold). |
| **Payment-related personal information & raw payloads** (`payments.raw_payload`, `payments`, `fee_ledgers`, `ledger_entries`, `fee_charges`, `fee_change_notes`) | Canonical billing records (`payments`, `fee_ledgers`, `ledger_entries`, `fee_charges`, `fee_change_notes`) are stored for school accounting and auditability. `payments.raw_payload` stores raw provider callback JSON if ingested. | Tenant- and permission-scoped access only (bursar/admin roles, and guardians scoped to their own children via `student_guardians`). | **Policy decision required:** confirm statutory/financial retention period for canonical accounting records and retention limit for `payments.raw_payload` under Kenyan tax/education/accounting law and school controller policy. | **Never** store or log M-Pesa PINs, API secrets, or authorization headers in `raw_payload` or logs; allowlist callback fields before storage; avoid copying financial/payer details into chat logs; protect via forced RLS and encrypted database backups. |
| **AI action logs** (`ai_action_logs`) | Stored in PostgreSQL (RLS-isolated per `school_id`) to record structured `feature`, `intent`, `endpoints_read` (JSON array), `action_taken`, and redacted `metadata` for accountability under the AI data boundary. | Tenant-scoped authorized administrators/auditors only; `ai_action_logs` is separate from `AuditLogger` / financial `audit_log` (AC-29 remains pending). | **Policy decision required:** confirm retention period for operational AI action logs and backup retention/deletion propagation rules (other categories like results, attendance, and consent records also require controller-approved schedules). | Allowlist `endpoints_read` (route templates without query strings/IDs) and `metadata` keys; **never** store raw prompts, message transcripts, response bodies, phone numbers, names, payer details, secrets, or token hashes. `ai_action_logs.session_id` is a non-cascading FK to `chat_sessions(school_id, id)`—cleanup must set `session_id = NULL` in the same tenant transaction before deleting expired `chat_sessions`. |
| **Transactional SMS payloads & provider responses** (in-memory transport; future `internal/jobs` queue payloads & logs) | Normalized recipient phone numbers, transactional message bodies (including OTPs), and provider responses are handled transiently by `SMSDispatcher`. No SMS body or provider-response persistence table exists in Release 1. | Restricted to backend SMS transport/worker execution; never exposed to frontend clients, staff UI, or AI service. | **Policy decision required:** confirm whether future worker job/dead-letter retention is strictly bounded (proposed target: never persist plaintext OTPs in queue dead-letter logs; purge delivery attempt metadata on a short operational window) and confirm provider-side retention terms with Africa's Talking. | **Never** log SMS message bodies, OTP values, `APIKey`, raw provider responses, or unmasked phone numbers; mask recipients to trailing digits (`***678`) in sandbox logs; store only safe error codes in delivery diagnostics. |
| **In-app notifications & read timestamps** (`in_app_notifications`, `read_at`) | Stored in PostgreSQL (RLS-isolated per `school_id`) for `recipient_user_id` (`category`, `title`, `body`, `metadata`, `is_read`, `read_at`, `created_at`) to deliver Release 1 event notices and track read receipts. | Authenticated recipient (`recipient_user_id = Principal.UserID`) in the active school (`school_id = Principal.SchoolID`) only; enforced by service queries and PostgreSQL forced RLS. | **Policy decision required:** confirm how long read and unread in-app notifications (and `read_at` timestamps) are retained after term close or membership/student exit; no automated purge worker is implemented yet. | Avoid storing secrets, OTPs, clinical details, or unnecessary personal data in `body`/`metadata`; preserve first `read_at` (`coalesce(read_at, now())`) for audit consistency until purged under the approved retention schedule. |

Operational cleanup requirements (pending worker implementation):

- Execute all retention purges and redactions per `school_id` inside tenant-scoped
  transactions (`SET LOCAL app.school_id`), honoring forced RLS.
- Purge or redact personal copies across queue payloads, dead-letter jobs,
  exports, and notification metadata alongside primary table rows.
- Any legal hold suspending deletion must be documented with scope, owner,
  reason, and review/expiry date; never suspend deletion indefinitely without
  review.
- Tests and examples must use synthetic identities and payer data only.
- Audit deletion counts and policy outcomes without recording the deleted
  personal data. Review new metadata keys and provider fields against the
  redaction rules above before storage.
>>>>>>> 7e300ec074132fac5cb72ca2428a933bfcbc5172

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

## Release 1 and launch gates (added 2026-10-29)

- **Release 1 uses synthetic data only.** No real student, parent, or staff
  data enters any environment. It runs against the Daraja sandbox and a test
  WhatsApp number.
- **Gates before the first release with real data:**
  1. `whatsapp` consent purpose in the consent centre (purposes are now
     biometrics, photos, SMS, and WhatsApp)
  2. Data protection impact assessment for the WhatsApp channel
  3. Processing agreement with Meta
  4. ODPC notification of the transfer
  5. Minimal WhatsApp message content (see below)
  6. Review by a Kenyan data protection lawyer, including whether a child's
     fee data falls under a sensitive category
  7. ODPC registration
  8. The hosting decision (Kenya only vs regional Africa; where the AI model
     runs) — unresolved
  9. Retention periods and the breach process

## WhatsApp exception to Kenya-only hosting

WhatsApp messages pass through Meta, which counts as a transfer under the
rules above. This is a deliberate, documented exception, limited by content:
WhatsApp carries only fee balances, attendance status, and payment prompts.
Results, report comments, health, and discipline information are never sent on
WhatsApp; the chat sends a link to the app instead.

## Parent identity, OTP, and sessions

- Chats identify a student by admission number inside the school whose bot
  received the message; the OTP goes only to a number registered for that
  student, chosen by the sender from masked numbers.
- OTPs: one request per minute; more than five in thirty minutes blocks the
  requesting number for 24 hours (not the guardian's account). How this applies
  to app login is undecided.
- A verified WhatsApp session lasts 6 hours.
- Changing a registered number needs 2FA (an OTP to the new number), notifies
  the old number by SMS, and pauses chat payments for 24 hours.
- Pending registrations are not active records and grant no access until a
  staff verifier confirms them in person. Verification can be delegated only
  to staff, only while the QR is active (12 hours).
