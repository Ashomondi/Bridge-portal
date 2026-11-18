# REQUIREMENTS.md

The feature register. This is the source of truth for what's in scope, by
role and by release phase. If a task isn't covered here, check with the
product owner before building it — don't infer scope from what would be
technically interesting to add.

Legend: **[M]** must have · **[S]** should have · **[C]** could have ·
**(+)** added during scoping discussion, not in the original brief ·
**[AI]** implemented using the AI service

## A. Platform foundation

- [M] Multi-school hosting with fully isolated data per school
- [M] Roles plus a badge system (badges are permissions, can be tied to a
  class, club, or time period)
- [M] Login by phone number with SMS OTP (+)
- [M] Notifications: in-app for Release 1 events (`C58`), transactional SMS for
  authentication/account security (`app_login`, `phone_change`,
  `phone_change_notice`, plus reserved `whatsapp_session` key for Release 3),
  and WhatsApp [C] deferred to Release 3
- [M] Audit trail of sensitive actions (+)
- [S] English and Swahili (+)
- [S] Excel import for students, staff, and fee data (+)
- [S] School setup wizard (+)
- [M] Parent consent centre for biometrics, photos, and SMS (+)
- [M] Editable standard grading scale per school, aligned to CBC

## B. Onboarding

- [M] Principal and deputy onboarded first
- [M] Staff QR code (expires; valid 8am–6pm by default; principal/DP can
  change the window)
- [M] Applications approved by principal/DP, or a delegate (e.g. dean of
  studies)
- [M] Student admission with class, dormitory, admission number
- [M] Parent self-registration, then in-person confirmation with photo and
  biometrics
- [S] Biometric capture only where the school can afford it, with a manual
  fallback

## C. Parent portal

- [M] Child's results, attendance, fee structure, balance sheet
- [M] Performance graph from the day the child joined, across academics
- [M] Upcoming events, clubs the child is in, books borrowed
- [M] Student blogs and newsletters
- [M] Pay fees by M-Pesa STK push
- [AI] Parent assistant: answers fee/attendance/exam questions in Swahili or
  English via WhatsApp/SMS/app, reading only from the database (never
  invents figures), scoped to the requesting parent's own children
- [AI] Chat-initiated fee payment: parent states an amount and child, AI
  confirms details and starts an STK push; parent authorises with their own
  M-Pesa PIN — AI never handles money directly
- [AI] Plain-language CBC performance-level explanations and report card
  summaries, teacher-approved before sending
- [AI] Payment-default reminders: cross-checks a parent's payment
  pledge/package against actual payments and arrears; escalates tone across
  three stages (reminder → firmer reminder → final warning) based on default
  count
- [S] Sports and activities in the graph (scoring method still to be decided)
- [S] Absence and sickbay SMS alerts — **Release 2** (Release 1 absence
  notifications are in-app only per `C58`; nurse/sickbay module is Release 2,
  and sickbay alerts state only that a visit happened, never clinical detail) (+)
- [C] Notification preferences and weekly digest (+)
- [C] Visiting day and parent-teacher meeting booking (+)
- [C] Digital permission slips, and exeat requests for boarders (+)

## D. Teacher portal

- [M] Personal timetable, events affecting them (e.g. exams), syllabus
  coverage per unit taught
- [M] Class teacher badge: sees every student in own class
- [M] Any teacher sees other classes as class/subject means only, never
  individual student names
- [M] Duty badge: sees the week's expectations/events, can post to the blog
  for parents
- [AI] Syllabus coverage from natural language: teacher types or dictates
  what was taught; AI maps it to CBC strands/sub-strands and updates coverage
- [AI] Report comment drafting from marks and trends, teacher edits before
  publishing
- [S] Behind-schedule syllabus alerts (+)
- [C] Timetable clash checker and duty roster generator (+)

## E. Other staff modules

**Nurse**
- [M] Record consultations; automatic notification to the class teacher
  (notification only, no clinical detail)

**Librarian**
- [M] Book catalogue and who holds each book; borrowed books show on the
  parent side
- [S] Due dates and overdue reminders (+)

**Dean of studies**
- [M] Admissions, class and dormitory assignment, admission numbers

**Principal / deputy**
- [M] Staff onboarding and approvals; leave approvals
- [S] One dashboard: fees, attendance, syllabus coverage (+)

**Leave management**
- [M] Teachers apply for time off in emergencies, with approval
- [S] Cover arrangements for missed lessons (+)

**Journalism club**
- [M] Patron reviews and publishes blog posts and newsletters
- [AI] Blog/newsletter safety check: flags children's personal details or
  inappropriate content before the patron publishes
- [S] Other clubs get their own patron posting

**Bursar**
- [M] Fee structure per school and grade, reviewed yearly
- [M] Per-student fee ledger with arrears carried forward
- [AI] M-Pesa reconciliation: proposes the likely student for payments with
  wrong/missing admission numbers; bursar confirms with one click
- [M] Payments to suppliers/bills with signatory approval workflow
- [M] Receipts and fee statements
- [S] Bursaries and sponsors tracked separately (+)
- [S] Budget vs. actual and board-ready reports (+)
- [C] Payment plans and reminders — see parent-portal AI default-reminder
  feature above, which covers this (+)

**Store** — [S] stock records, receipts, issues, linked to bursar payables;
details still to be defined

**Caterer** — [S] menu planning and confirming items received from store;
details still to be defined

## Out of scope

- Payroll
- Automated payment disbursement (the system records and approves; it does
  not itself send money to suppliers)
- Discipline records
- Any student ranking, leaderboard, or profiling beyond class/subject means
- AI-generated diagnoses or health assessments
- Training AI models on student data

## Release phases

1. **Release 1:** foundation, onboarding, parent portal core (fees, results,
   attendance, events), teacher timetable and class performance, bursar/team-loaded
   fee ledger with STK push, in-app event notifications (`C58`), transactional
   SMS (`app_login`, `phone_change`, `phone_change_notice`; `whatsapp_session`
   purpose key reserved), parent AI assistant (in-app read-only) and
   chat-initiated payment.
2. **Release 2:** nurse (including in-app class-teacher flag and `[S]` sickbay
   SMS alert), librarian, leave management, blogs and clubs, syllabus coverage
   (including AI mapping), biometric attendance, duty badge, bursar payables and
   signatories, AI M-Pesa reconciliation, AI default reminders, absence SMS
   alerts (`[S]`), Swahili, consent collection with real parents.
3. **Release 3:** store, caterer, exeat and permission slips, dashboards,
   WhatsApp, USSD.

## Notification channel scope (C58 reconciliation)

**C58 (Release 1 event-notification rule):** All Release 1 event notifications
(attendance absence, results release, results re-release, timetable updates,
calendar updates, and fee-change notices) are delivered **in-app only**
(`in_app_notifications`).

1. **Permitted Release 1 SMS purposes:** The Release 1 SMS transport
   (`internal/notifications`) permits only four transactional purposes:
   - `app_login` — phone-number login OTP
   - `whatsapp_session` — Release 3 WhatsApp session OTP purpose key (reserved in
     the transport enum; WhatsApp delivery itself is excluded from Release 1)
   - `phone_change` — guardian 2FA OTP sent to the new phone number
   - `phone_change_notice` — security alert sent to the old phone number on
     registered-number change
2. **Prohibited event-related SMS in Release 1:** Event notifications —
   specifically `absence`, `sickbay`, results release/amendment, timetable,
   calendar, and fee/payment reminders — are prohibited and rejected by the
   Release 1 SMS transport (`ErrInvalidSMS`).
3. **Absence and sickbay notification status:**
   - **Absence notifications:** In Release 1, daily roll-call absences notify
     guardians **in-app only** (`academics.attendance`). Absence **SMS** alerts
     (`[S]`) are deferred to **Release 2**.
   - **Sickbay notifications:** The **Nurse** module is **Release 2**; when
     implemented, class-teacher notifications are in-app flags without clinical
     detail (`[M]`), and parent sickbay SMS alerts (`[S]`) are **Release 2**
     (subject to Release 2 consent and SMS-cost policy decisions).
4. **Why WhatsApp provider integration is excluded from Release 1:** Per the
   Release phases above and the WhatsApp Release Scope ADR in `DECISIONS.md`,
   WhatsApp Business Cloud API webhook routing, bot credentials, and outbound
   delivery belong to **Release 3** (alongside USSD and ODPC cross-border
   review under `SECURITY.md`). Release 1 uses in-app chat while keeping
   `school_whatsapp_bots` and `chat_sessions` schema-compatible for Release 3.

## Open decisions

- Sports/activities scoring: patron-entered scores vs. participation records
  only
- Guardian fee visibility: both guardians, or only the one flagged as fee
  payer
- Legacy 8-4-4 (Form 3/4) support: in or out of scope
- Store/caterer module: not yet designed
