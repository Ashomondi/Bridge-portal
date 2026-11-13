# Ikonex Academy: Project Reference

Draft 1 · 2026-10-29 · Written from the seven repository documents and a requirements discovery session with the project lead. **No code exists yet.**

## 0. How to read this document

Every statement carries one of these statuses. Do not treat a lower status as a higher one.

| Status | Meaning |
|---|---|
| **Docs** | Stated in the original repository documents and not contradicted since. |
| **Agreed (C#)** | Decided with the project lead during discovery. The ID points back to the discovery record. |
| **Assumed (A#)** | My reading of an answer. Not yet confirmed. Listed in Appendix A. |
| **Open (O#)** | Undecided. Listed in Appendix B. Do not build around a guess. |
| **Proposed** | A recommendation, not an agreement. |

When this file and an older document disagree, this file reflects the agreed decisions, and the six older documents were updated to match on 2026-10-29 (see Appendix C). `REQUIREMENTS.md` remains the feature register, `DECISIONS.md` the dated decision log, and this file the place to start.

---

## 1. Vision, objectives, scope

**What it is.** A multi-school management and CBC (Grade 7–12) portal for Kenyan secondary schools, hosted for many schools on one platform. Parents, teachers, and school staff each get a role-specific view. A self-hosted AI layer (Mistral models) helps parents ask questions and pay fees, and helps teachers draft report comments. (Docs)

**Objectives.**
1. Give every school fully isolated data on shared infrastructure. (Docs)
2. Let parents see their child's results, attendance, and fees, and pay fees by M-Pesa, including from a WhatsApp chat. (Docs, C2)
3. Keep teachers' marks entry, report release, attendance, and timetables inside one system with clear human approval steps. (C30–C35, C59)
4. Protect children's data under the Kenya Data Protection Act, 2019. (Docs)

**In scope.** CBC Grade 7–12 only. (Docs, C11)

**Out of scope.** See section 8.

---

## 2. Repository state

All seven documents were reviewed in full. Nothing else was available. (C1)

| Reviewed | Not available |
|---|---|
| README, AGENTS, API_CONTRACTS, ARCHITECTURE, DECISIONS, REQUIREMENTS, SECURITY | Any Go, Python, or Next.js code, migrations, `docker-compose.yml`, `Makefile`, `.env.example`, `portal_core_schema.sql` (referenced in ARCHITECTURE but absent), tests, issues, commit history, any file naming people or teams |

**Consequence.** Every feature below is **Not started**. Status in section 18 describes the plan, not progress.

---

## 3. Release plan

| Release | Contents | Status of plan |
|---|---|---|
| **1** | Release 1 as listed in `REQUIREMENTS.md`, **plus WhatsApp** for the parent assistant and chat payment (C2, C3). Built and tested with **synthetic data only** (C52). | Agreed |
| **2** | The Release 2 list as written in `REQUIREMENTS.md`, plus Swahili (A18), SMS alerts (C36), the consent centre with real parents (O15), and the real-data launch gates (section 12). | Agreed (C3) |
| **3** | The Release 3 list as written, minus WhatsApp (moved to Release 1). | Agreed (C3) |

### 3.1 What Release 1 is

- Release 1 is **a test of the idea**. It most likely will not interact with real parents. (C51)
- It uses **synthetic data only**. Internal testers and demo accounts play every role. No real child's record enters the system. (C52)
- Because nothing real is at stake, Release 1 runs against the **Daraja sandbox** and a **test WhatsApp number**, and its hosting is not constrained by the Kenya-only rule. (A17, O10)
- **It passes when** (C53):
  1. onboarding works (school and principal creation, staff and parent QR registration, in-person verification);
  2. parents can log in (SMS OTP and the WhatsApp OTP session);
  3. badges work well and stress free.
- **All Release 1 features must also work well and stress free**, not only the three above. (C54)
- The QA developer owns how this is verified. (C55)

### 3.2 Release 1 feature list

| Area | Feature | Section |
|---|---|---|
| Foundation | Multi-school isolation, roles and badges, audit trail, phone + SMS OTP login, in-app notifications, setup wizard, English | 4, 5, 6.1 |
| Onboarding | School and principal created by team; staff QR and applications; parent QR registration; student records (two routes); in-person verification and delegation | 6.1–6.4 |
| Academics | Classes and teacher assignment, timetables, school calendar, daily attendance, marks entry, results release/reopen, performance graph, AI report comment drafting | 6.5 |
| Fees | Fee structure (team-loaded), automatic charging, STK push, direct Paybill matching, ledger | 6.6 |
| Parent portal | Results, attendance, fees, events, placeholders for clubs/books/blogs | 6.7 |
| Parent assistant | WhatsApp and in-app chat, chat-initiated payment | 6.8 |

---

## 4. Users, roles, and badges

**Role** = job title on a `membership`. **Badge** = permission bundle, optionally scoped to a class, subject, club, or time window. A membership can hold several badges. Never hard-code "class teacher" as a column. (Docs)

| Actor | Release | What they do | Notes |
|---|---|---|---|
| **Ikonex team** | 1 | Creates school and principal; sets up the school's WhatsApp bot and Daraja credentials; loads fee structures; applies fee changes on written request; resolves unmatched payments | C37, C16, C44, C49, A15, C50 |
| **Principal / deputy** | 1 | Wizard setup (calendar, classes, dormitories, grading scale); staff QR and application approval; parent QR; in-person verification; reassigns marks entry | C38, C28, C32 |
| **Dean of studies** | 1 | Parent QR; verification and delegation; assigns teachers to classes and subjects; owns all timetables and the school calendar, **including term dates after setup**; reassigns marks entry | C59, C60, C28, C32 |
| **Class teacher** (badge, one class) | 1 | Sees every student in the class; daily roll-call; edits report comments; approves, reopens, and re-releases results; enters student records on the first load | C35, C30, C33, C39 |
| **Subject teacher** (assignment: class + subject) | 1 | Enters marks for the assigned class and subject; corrects marks after a reopen | C30, C33 |
| **Any teacher** | 1 | Sees other classes as class/subject means only, never names | Docs |
| **Delegated verifier** (staff badge, 12 h) | 1 | Verifies parents and students in person while a QR is active | C28, C29 |
| **Parent / guardian** | 1 | Sees own children's data; chats; pays fees; changes registered number | C10, C27 |
| **QA developer** | 1 | Verification of all features | C55 |
| Bursar, nurse, librarian, duty teacher, club patron | 2 | As listed in `REQUIREMENTS.md`; **no bursar in Release 1** | C48 |

**Dean absence.** Whether the principal/deputy can override or stand in for the dean on assignments, timetables, and calendar is **Open (O16)**. They already can on QR, verification, and reassignment (C29, C32).

**Guardians.** All guardians linked to a child can see fees (C10, may be revisited). Any linked guardian can start a payment (A2).

---

## 5. Cross-cutting business rules

1. **Tenant isolation.** Every tenant-owned table has `school_id`; RLS enforces it with `SET LOCAL app.school_id`. (Docs)
2. **AI never touches the database and never invents a fact.** It calls permission-checked Go endpoints. (Docs, C21)
3. **No feature moves money on its own.** AI starts an STK push only after the parent confirms child, amount, and number; the parent authorises with their own M-Pesa PIN. (Docs, A4 confirmed)
4. **Human approval before anything reaches a permanent record or a parent.** The AI drafts; a teacher confirms. (Docs)
5. **No ranking, leaderboards, discipline records, diagnoses, or training on student data.** (Docs)
6. **Every sensitive action is audit-logged**, including QR creation and revocation, verification, delegation, reassignment, result release/reopen/re-release, fee changes, term-date changes, and number changes. (Docs, plus C28–C34, C60, C61, C27)
7. **School contact details are required profile fields** because the assistant refers parents to them. (C57)

---

## 6. Feature specifications (Release 1)

### 6.1 School setup and the setup wizard

1. The **Ikonex team** creates the school and the principal's membership. The principal logs in with SMS OTP. This is the first-stage route; the later route is **Open (O11)**. (C37)
2. The team also sets up the school's WhatsApp bot, connects the school's Paybill/Till (Daraja credentials), runs a test payment before fee payments go live, loads the fee structure, and enters the school's contact details. (C16, C44, C49, C57)
3. The principal completes the **setup wizard**: academic years and terms, grade levels and classes, dormitories, and the first CBC grading scale. (C38, A10) What "all the help they need" means in the wizard is **Open (O12)**.
4. Term dates are set in the wizard by the principal; afterwards the dean can edit them. (C60)
5. The grading scale is **versioned per term**; a term uses the scale in force when its results were released; edits apply from the next term. (C43)

### 6.2 Staff onboarding

1. Principal or deputy issues a staff QR (default valid 8am–6pm, window adjustable by principal/deputy). (Docs)
2. The applicant scans it and submits details plus uploaded documents (for example a national ID and, where relevant, a professional registration number). (C62)
3. The principal/deputy, or a delegate such as the dean, **verifies the documents against the person in person** and approves. (C62, Docs)
4. Documents are deleted after approval; a verification record remains. (C63)

### 6.3 Parent onboarding and student records

**Student records (per-school choice, C41).** A school uses either or both:
- **Class-teacher route.** On first load every student is entered manually by their class teacher, with the guardian's current number; parents are notified on that number. (C39, C40)
- **Opening-day route.** Every student is registered through the parent using the QR. (C40)
- On re-enrolment (existing students in a new year) the system carries records forward automatically. (C39)
- When both routes are used, a parent registering an existing student is matched by admission number and flagged to the verifier. (A12) How the shared QR links each parent to the right child is **Open (O9)**.

**Parent QR registration (C26, C28).**
1. The dean, principal, or deputy generates a parent QR, **valid 12 hours**, revocable at any time. One QR serves many parents. (A7 confirmed)
2. The parent scans it, fills in forms and questions, and uploads documents.
3. The submission is **pending**. It is not an active record and grants no access to any child's data. (A8 confirmed)
4. A staff verifier **physically confirms the parent exists and the student is real**. On approval the parent is added to the database. In-person verification applies on both routes. (A11)
5. On long days the dean/principal/deputy can **delegate verification to other staff**. The privilege goes only to staff and lasts only while the QR is active (12 hours). Grants and verifications are audit-logged. (C28, C29)
6. Uploaded documents are deleted after approval, keeping who verified whom, when, and which document type. Rejected or never-verified submissions are **deleted automatically after a fixed number of days** (number **Open, O17**). (C63, C64, A20)

**Changing the registered number (C25, C27).**
- During onboarding a guardian may edit the number in the app.
- Afterwards a verified parent can change it in the app using **2FA** (an OTP to the new number, A5), and the **old number gets an SMS notice**.
- Chat payments pause for **24 hours** after a change.
- If the parent has lost access entirely, staff assist in person (A6).

### 6.4 Login and parent identity

**App login.** Phone number + SMS OTP (Africa's Talking). (Docs)

**WhatsApp session.** Each school has its **own WhatsApp bot**; the receiving number tells the backend which school. (C16)
1. A chat can start from **any** WhatsApp number. (C13)
2. The sender enters the **student's admission number**. The system finds the guardian numbers registered for that student. (C14)
3. If several guardians exist, the chat shows **masked numbers** (for example "07•• ••• 482"); the sender picks their own. (C15)
4. An OTP is sent by SMS **only** to the chosen registered number; the sender types it into the chat. (C12, C13)
5. A verified session lasts **6 hours**, then the parent must verify again. (C13)

**Limits.** One OTP request per minute. After **more than five requests in thirty minutes**, the **requesting WhatsApp number** is blocked for 24 hours; the guardian's account is not blocked. (C18, C19, C20) A3 (awaiting validation): the one-per-minute limit also applies per admission number regardless of sender. How the same limits apply to **app login** is **Open (O20)**.

### 6.5 Academics

**Classes and teaching assignments.** The dean assigns teachers to classes and subjects. The class teacher is a badge scoped to one class. (C59)

**Timetables and calendar.** The dean owns all timetables (teachers and students) and the school calendar. Changes notify parents and teachers in-app. A term-date change moves fee due dates and triggers the fee-change notice (6.6). (C59, C60, C61)

**Attendance.** One **daily roll-call per student**, taken by the class teacher. Per-lesson and boarding roll-calls are not in Release 1. Biometric attendance (Release 2) must write into the same daily record. (C35) Parents see absences as in-app notifications; SMS alerts arrive in Release 2. (C36)

**Marks and results release.**
1. Each subject teacher enters marks in the portal for their assigned class and subject. (C30)
2. If a teacher cannot enter marks, the dean (or principal/deputy) **reassigns** that subject and class to another teacher. (C32)
3. The class teacher cannot release until **every subject has marks for every student** (hard gate). There is no override in Release 1. (C31)
4. The class teacher enters report comments (AI may draft from marks and trends; the teacher edits before publishing) and **releases** the class's results to parents. (C30, C5)
5. **Reopen.** The class teacher can reopen a released class; the subject teacher corrects the mark; the class teacher re-releases. (C33)
6. While reopened, parents **keep seeing the previously released results**. After re-release, **only parents of affected students** are notified, the affected subjects are marked, and a **teacher's note explaining why** is shown. The note is required and written by the teacher; the AI may help draft it from the teacher's stated reason. (C34, A9 resolved)

**Performance graph.** Plots **CBC performance levels per subject per term**, from the day the child joined. It never shows class position. Marks are stored; levels are derived using the term's grading-scale version. (C42, C43, Docs)

### 6.6 Fees and payments

**Fee structure.** Each school defines its own structure entirely. In Release 1 it is **fixed to terms** and **hard coded by the Ikonex team** (loaded as database records, not source constants, **Proposed**). Schools have no fee screens and there is **no bursar role** in Release 1. (C46, C47, C48, C49)

**Charging.** Charges are applied to enrolled students **automatically, once, at the start of the year**, divided into terms. A term's fee due date **follows the term's start date**. (C45, C46, C61) Whether the structure has variants such as boarder versus day scholar is **Open (O13)**.

**Mid-year changes.** In Release 1 the team applies a change on the school's written request (A15). Every change requires a **note**, notifies **only parents of the affected students**, and is audit-logged. (C45, A13 confirmed) A term-date change that moves a due date counts as a fee change and also requires the dean's note. (C61)

**Paying.**
- **STK push** from the app or WhatsApp. The payment carries the student reference so it matches automatically. (Docs)
- In chat the parent states child and amount, and **types the M-Pesa number to charge**; the chat repeats child, amount, and number and waits for an explicit "yes" before the push. (C23, A4 confirmed)
- **Any amount** is accepted. Any excess is held as a **credit on the student's ledger**. (C24) Applying credit to the next charge is assumed (A22). A limiter for the top-most grade is planned later, **not agreed**. What happens to a credit when a student leaves is **Open (O6)**.
- **Direct Paybill payments** are auto-matched when the account number is a valid admission number for that school. Anything else goes to an **unmatched list** that the Ikonex team resolves. The team's process is **Open (O14)**. (C50)

**Who pays for WhatsApp and SMS messages** is **Open (O4)**.

### 6.7 Parent portal (Release 1)

Child's results (as released), attendance, fee structure and ledger, performance graph, upcoming events and timetable, STK push, and the assistant. **Clubs, books borrowed, and student blogs/newsletters show placeholders** until Release 2. (C4)

### 6.8 Parent assistant (WhatsApp and in-app)

- Reads facts only through permission-checked endpoints, scoped to the requesting parent's own children. English only in Release 1. (Docs, C56)
- **On WhatsApp it carries only minimal facts**: fee balances, attendance status, and payment prompts. Results, report comments, and anything about a child's performance stay in the app, and the chat sends a link. (C22)
- When it cannot or should not answer, it **refuses politely and gives the school's contact details**. There is no hand-off inbox in Release 1. (C57)
- **Limits** are on backend lookups per session (fee-balance, attendance, payment), with a tighter limit on payment starts. Exact numbers are left to the team and QA. Model compute is not capped (O19). (C66)
- It authenticates to the Go backend with a **service token plus a short-lived signed user context** (user, school, membership, WhatsApp session). A hybrid with mutual TLS may come later. (C21)
- Every AI-initiated action is logged: what was asked, what was read, what action followed. (Docs)

### 6.9 Notifications

- **Event notifications in Release 1 are in-app only.** (C58)
- **Transactional messages are separate:** login and WhatsApp OTPs by SMS, and the SMS to the old number when a registered number changes. (C12, C27)
- SMS alerts (absence, sickbay) and WhatsApp event messages are not Release 1.

### 6.10 Consent and privacy in Release 1

The consent centre is designed now, with purposes **biometrics, photos, SMS, and WhatsApp**. Collection from real parents is not tested in Release 1 and moves to Release 2 (**Open, O15**, parked after C51).

---

## 7. Release 2 and Release 3 scope

Specified in `REQUIREMENTS.md` (sections C to E, release phases). Details not yet agreed.

**Release 2.** Nurse, librarian, leave management, blogs and clubs, syllabus coverage with AI mapping, biometric attendance, duty badge, bursar payables and signatories, AI M-Pesa reconciliation, AI default reminders, parent-facing plain-language explanations (C5), SMS alerts, Swahili (A18), consent collection with real parents.

**Default reminders and payment plans (C6–C8).**
- Plans come from the **bursar** or from a **parent's pledge**.
- A parent's **custom** pledge stays pending until the bursar accepts it. A pledge that picks a **predefined scheme** is accepted automatically.
- Each school's bursar creates their own schemes; the platform ships **examples** to teach the bursar (A1: examples are not live schemes).
- The scheme designs and the reminder logic are not yet specified. Release 1 has no plan or pledge entity.

**Release 3.** Store, caterer, exeat and permission slips, dashboards, USSD.

---

## 8. Out of scope

Payroll; automated payment disbursement; discipline records; any ranking, leaderboard, or profiling beyond class/subject means; AI diagnoses or health assessments; training AI on student data; **legacy 8-4-4 (Form 3/4)** (C11); a Release 1 bursar role (C48); a human hand-off inbox in Release 1 (C57); Excel import for the first student load (C39, **Proposed** to confirm).

---

## 9. Data model notes

Existing entities are in `ARCHITECTURE.md`. Additions implied by agreed decisions:

| Entity | Purpose | Source |
|---|---|---|
| `school_whatsapp_bots` | Maps a receiving WhatsApp number to `school_id` | C16 |
| `chat_sessions` | WhatsApp/app sessions, 6-hour expiry, OTP state | C12, C13 |
| `otp_requests` | Rate limiting per sender and per admission number | C18–C20 |
| `parent_qr_codes`, `staff_qr_codes` | Expiry, revocation, issuer | C28 |
| `registration_submissions` | Pending parent/staff applications; auto-delete | C28, C64 |
| `verification_records` | Who verified whom, when, document type | C63 |
| `badge delegations` | Time-bound verifier and marks-reassignment grants | C29, C32 |
| `attendance_daily` | One row per student per day | C35 |
| `assessments`/`marks`, `report_comments` | Marks and drafts | C30 |
| `result_releases` | Release state per class per term; reopen history | C31, C33, C34 |
| `grading_scale_versions` | Per-term scale | C43 |
| `fee_periods`, `fee_charges` | Per-term charges, due dates tied to term start | C45, C46, C61 |
| `fee_change_notes` | Mid-year change notes | C45 |
| `ledger_entries` (incl. credit) | Charges, payments, credit | C24 |
| `unmatched_payments` | Direct payments with no valid reference | C50 |
| `school_profile` | Contact details (required) | C57 |
| `consents` | Add purpose `whatsapp` | C22 |

**Proposed.** Build `fee_periods` on generic billing periods even though Release 1 is terms only, so custom structures need no migration (C47).

---

## 10. API behaviour and integrations

**Integrations.** Africa's Talking (SMS/OTP); WhatsApp Business Cloud API (one number per school); Safaricom Daraja (STK push and C2B, per-school Paybill/Till). Release 1 uses the Daraja sandbox and a test WhatsApp number (A17).

**Rules.**
- Every request resolves identity and `school_id`; permissions are checked before handler logic. (Docs)
- The AI service only calls endpoints and never reads storage. (Docs)
- AI-to-Go authentication: service token + signed user context. (C21)
- New endpoint groups (WhatsApp webhook, OTP, QR, verification, results release, fee notes) are listed in `API_CONTRACTS.md`.
- Still undecided: response shape, pagination, date format (recommend ISO 8601 with the school's timezone), versioning.

---

## 11. Architectural decisions

The dated log is `DECISIONS.md`. Headlines: shared PostgreSQL with RLS; global `users` with per-school `memberships`; badges as permissions; CBC only; separate AI service with no database access; system records and approves payments but does not disburse; biometrics as encrypted templates only. Added 2026-10-29: per-school WhatsApp bot, WhatsApp content limits, OTP rules, AI service authentication, hard-gated results release, versioned grading scale, team-loaded fees in Release 1, synthetic-data Release 1, hosting marked unresolved.

---

## 12. Security and compliance

**Hosting is unresolved (O10).** `DECISIONS.md` said "Kenya or regional-Africa cloud"; `AGENTS.md` and `SECURITY.md` say personal data stays in Kenya, and hosting abroad counts as a transfer. The project lead will discuss this with the group. Release 1 is unaffected because it uses synthetic data.

**WhatsApp is a deliberate, documented exception to the Kenya-hosting rule.** WhatsApp messages pass through Meta. (C22)

**Launch gates for the first release with real data** (not Release 1):
1. `whatsapp` consent purpose in the consent centre.
2. Data protection impact assessment for the WhatsApp channel.
3. Processing agreement with Meta.
4. ODPC notification of the transfer.
5. Minimal message content.
6. Review by a Kenyan data protection lawyer.
7. ODPC registration as a controller/processor.
8. The hosting decision (O10).
9. Retention periods (O18) and the breach process, which is still `TODO` in `SECURITY.md`.

I checked on 2026-10-29 that the ODPC published its cross-border transfer guidance in September 2026; read the final text directly rather than relying on summaries. Have the lawyer confirm whether a child's fee data falls under a sensitive category.

**Document handling.** Uploaded documents are deleted after approval; rejected or never-verified submissions are auto-deleted after a fixed time. (C63, C64)

**Retention at exit (C65).** Each record type has its own rule; periods are to be set with the lawyer.

| Record | Rule | Status |
|---|---|---|
| Biometric templates | Deleted when the student leaves | Docs |
| Health (nurse) records | Deleted at exit | Proposed |
| Consent records | Deleted at exit | Proposed |
| Results | Archived for a set period | Proposed, period open |
| Financial ledger | Kept as the law requires | Proposed, period open |

---

## 13. Edge cases and error handling

| Situation | Expected behaviour | Source |
|---|---|---|
| WhatsApp sender enters an unknown admission number | No data, no OTP sent; generic reply | C14 |
| Guardian's registered number differs from the WhatsApp number | Allowed; the OTP goes only to the registered number | C13 |
| More than 5 OTP requests in 30 minutes | Requesting number blocked 24 h; guardian unaffected | C20 |
| OTP requested twice within a minute | Second request refused | C18 |
| Verified session older than 6 hours | New OTP required | C13 |
| Parent asks for results on WhatsApp | Link to the app; no results in chat | C22 |
| Question outside the assistant's scope | Polite refusal with school contact details | C57 |
| Parent mistypes the M-Pesa number | Repeat-back step catches it before the push | A4 |
| Parent pays more than owed | Excess held as credit | C24 |
| Payment arrives with wrong/missing reference | Unmatched list, resolved by the team | C50 |
| Subject teacher cannot enter marks | Dean reassigns; release stays blocked meanwhile | C31, C32 |
| Mark wrong after release | Reopen, correct, re-release with note and notice | C33, C34 |
| QR scanned after expiry or revocation | Registration refused | C28 |
| Submission never verified | Auto-deleted after the fixed period | C64 |
| Term date moved | Due dates move; fee-change notice with dean's note | C61 |
| Grading scale edited | Applies from the next term only | C43 |
| Registered number changed | Old number notified; chat payments paused 24 h | C27 |

---

## 14. Acceptance criteria (Release 1)

Written for the QA developer. Each is testable on synthetic data.

**Onboarding**
- AC-01: The team can create a school and principal; the principal logs in by SMS OTP and reaches the wizard.
- AC-02: A staff QR is refused outside its valid window and after revocation.
- AC-03: A parent QR is accepted for 12 hours from creation, refused after that or after revocation, and serves more than one parent.
- AC-04: A pending submission cannot see any child's data until verified.
- AC-05: A delegated verifier can verify while the QR is active and cannot after it expires; a non-staff user cannot receive the grant.
- AC-06: Documents are removed after approval; the verification record remains.
- AC-07: A rejected or never-verified submission and its files are removed after the configured number of days.

**Login and WhatsApp**
- AC-08: A chat from an unregistered number can start; the OTP goes only to the registered number.
- AC-09: With several guardians, the sender sees masked numbers and the OTP goes only to the one chosen.
- AC-10: The second OTP request within one minute is refused; the sixth within thirty minutes blocks the sender for 24 hours and does not block the guardian.
- AC-11: A session expires after 6 hours.
- AC-12: One school's bot cannot read another school's students, even with a matching admission number.
- AC-13: On WhatsApp, results and report comments are never returned; a link is sent instead.

**Academics**
- AC-14: A class teacher can mark a daily roll-call; a teacher outside the class cannot.
- AC-15: A subject teacher can enter marks only for their assigned class and subject.
- AC-16: Release is blocked until every subject has marks for every student.
- AC-17: After reassignment by the dean, the new teacher can enter marks and the old teacher cannot.
- AC-18: Reopening keeps the previous results visible to parents; re-release notifies only affected parents, marks the subjects, and requires a teacher's note.
- AC-19: The graph shows levels per subject per term and never a class position; a scale edit does not change past terms.

**Fees and payments**
- AC-20: Charges are created once at the start of the year, divided into terms.
- AC-21: A mid-year change without a note is rejected; with a note, only affected parents are notified.
- AC-22: Moving a term date moves due dates and sends the fee-change notice.
- AC-23: A chat payment proceeds only after an explicit "yes" to child, amount, and number, and any amount is accepted.
- AC-24: An overpayment appears as a ledger credit.
- AC-25: A direct payment with a valid admission number is credited; others appear in the unmatched list.
- AC-26: After a number change, chat payments are blocked for 24 hours and the old number gets an SMS.

**Cross-cutting**
- AC-27: With the tenant context unset, queries return no rows; RLS is forced on every tenant table.
- AC-28: The AI service holds no database credentials and every action it takes appears in the AI action log.
- AC-29: Every item listed in section 5, rule 6 appears in the audit log.

---

## 15. Testing requirements

- The QA developer decides the method (automated tests, usability testing, load testing). (C55)
- Minimum: every acceptance criterion above has an automated or documented manual test before its feature counts as done.
- Release 1 passes only when the three core criteria hold. (C53)
- Cross-tenant isolation is tested explicitly, not assumed. (Docs)
- Commands are all `TODO` in `AGENTS.md`; fill them in as the scaffolding is created.

---

## 16. Team and ownership

| Role | Person | Owns |
|---|---|---|
| QA / testing developer | Not named yet | Verifying Release 1 features (C55) |
| Product owner | Mentioned in `REQUIREMENTS.md`, not identified | Scope decisions |
| Backend (Go), frontend (Next.js), AI service (Python), infrastructure, compliance contact | **Unassigned** | Not yet provided |

The project lead will send the full team and each member's role (C9). Until then no feature has a named owner, and nothing here should be read as assigning one.

**Collaboration points to assign once the team is known:** the Go to AI service boundary (C21), WhatsApp and Daraja per-school setup (C16, C44), the first-stage school onboarding checklist (C37), and the legal review before real data.

---

## 17. Development workflow

- Commands, Go package layout, error-handling style, naming, and commit style are all `TODO` in `AGENTS.md`. Do not guess them.
- The Ikonex team's per-school onboarding checklist (**Proposed** as an operations document): create school and principal; set up WhatsApp bot; connect Paybill/Till and run a sandbox or live test payment; load fee structure; enter school contact details.

---

## 18. Status and priorities

| Area | Status |
|---|---|
| All Release 1, 2, and 3 features | **Not started** (no code) |
| Documentation | Seven documents updated 2026-10-29 |
| Verified/unverified | Nothing is verified |

**Proposed build order.** Scope is unchanged (C54); order only reduces risk.
1. Core: tenancy and RLS, badges, school/principal creation, staff and parent onboarding, login (SMS and WhatsApp). This is the Release 1 pass condition (C53).
2. Academics: classes, assignments, timetable and calendar, attendance, marks, release and reopen, graph.
3. Fees: structure loading, charging, STK push, matching.
4. Assistant: WhatsApp and in-app chat, chat payment, report comment drafting.

**Honest risk.** The Release 1 scope is large against a no-rough-edges standard, and the team is not yet named.

---

## Appendix A: Assumptions awaiting validation

| ID | Assumption |
|---|---|
| A1 | Example payment schemes are teaching aids; none is live until the bursar creates or enables one. |
| A2 | Any linked guardian can start a payment (C10 covered visibility only). |
| A3 | The one-OTP-per-minute limit also applies per admission number regardless of sender. |
| A5 | The second factor for a number change is an OTP to the new number. |
| A6 | Staff-assisted change is a fallback for a parent with no access, not a normal path. |
| A10 | The wizard covers calendar, classes, dormitories, and the grading scale. |
| A11 | In-person verification applies on both student-record routes. |
| A12 | Both routes can run together; duplicates are matched by admission number and flagged. |
| A15 | Mid-year fee changes in Release 1 are applied by the team on written request. |
| A17 | Release 1 uses the Daraja sandbox, a test WhatsApp number, and unconstrained hosting. |
| A18 | Swahili moves to Release 2. |
| A20 | Document deletion applies equally to staff and parent uploads. |
| A22 | A ledger credit is applied to the next charge. |

Confirmed since first raised: A4, A7, A8, A9 (resolved), A13.

## Appendix B: Open decisions

| ID | Question | Owner |
|---|---|---|
| O4 | Who pays for WhatsApp and SMS messages | Project lead / group |
| O6 | What happens to a leftover credit when a student leaves | Project lead |
| O9 | How a shared QR links each parent to the right child | Project lead |
| O10 | Hosting: Kenya only, or regional Africa, and where the AI model runs | Group discussion |
| O11 | How later schools' first principals are onboarded | Project lead |
| O12 | What help the wizard gives a first-time principal | Project lead |
| O13 | Fee variants (boarder/day, optional items) | Project lead |
| O14 | How the team learns of and resolves unmatched payments | Project lead |
| O15 | Consent collection with real parents | Release 2 |
| O16 | Whether principal/deputy can override the dean | Project lead |
| O17 | Days before rejected/never-verified submissions are deleted | Team |
| O18 | Retention periods | Lawyer |
| O19 | Whether to cap assistant messages for compute cost | Later |
| O20 | How OTP limits and blocks apply to app login | Project lead |
| — | Sports/activities scoring; store and caterer design; breach process; API response conventions; commands and code conventions; team roster; guardian fee visibility (agreed, may revisit) | Various |

## Appendix C: Discrepancies in the original documents and how they were handled

| Discrepancy | Handling |
|---|---|
| `AGENTS.md` titled `CLAUDE.md`; `API_CONTRACTS.md` refers to it by that name | Title and references changed to `AGENTS.md`. |
| README says docs live in `Docs/`; the files provided were flat | Not resolved; location does not matter (project lead). |
| `portal_core_schema.sql` referenced but absent | Noted; the file does not exist yet. |
| WhatsApp [C] in Release 3 but required for the Release 1 assistant | WhatsApp moved to Release 1. |
| Syllabus coverage [M] for teachers but placed in Release 2 | Left in Release 2 as written in the release phases. Not discussed. |
| Parent books/blogs [M] but modules in Release 2 | Placeholders in Release 1 (C4). |
| AI report comments and explanations in no release | Comments in Release 1, parent explanations in Release 2 (C5). |
| Bursar fee ledger in Release 1 but no bursar | Fees hard coded by the team (C48, C49). |
| Payment plans [C] "covered" by AI reminders without a data model | Plans and pledges specified (C6–C8), Release 2. |
| Hosting: "Kenya or regional Africa" vs "stays in Kenya" | Unresolved (O10). |
| Swahili [S] and Swahili assistant | English only in Release 1; Swahili in Release 2. |
| `SECURITY.md` cross-border rule vs WhatsApp | Documented exception with launch gates (C22). |
| Dean owns admissions only vs also timetables/calendar/teacher assignment | Dean owns all of these (C59). |
