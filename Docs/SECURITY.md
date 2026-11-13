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
5. **Every AI-initiated action is logged**: what was asked, what data was
   read, what action (if any) was taken, and by which AI feature.
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
