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

## Core entities (see `portal_core_schema.sql` for the full draft)

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

## Curriculum model

CBC only (Grade 7–12), not 8-4-4. `grade_levels.number` is constrained to
7–12, with a `phase` of `junior` or `senior`. If legacy Form 3/4 support is
ever needed, that's a schema change, not a workaround — raise it as a new
decision rather than bolting it on.

## Channels and sessions (added 2026-10-29)

- **WhatsApp Business Cloud API, one bot per school.** The receiving number
  maps to `school_id`, so an admission number is only ever looked up inside
  that school. The Go backend owns the webhook; the AI service is reached only
  through the backend.
- **Parent identity in chat:** sender enters the admission number → masked
  guardian numbers → OTP by SMS to the chosen registered number → a 6-hour
  verified session. Any WhatsApp number may start a chat; the OTP goes only
  to the registered number.
- **Release 1 event notifications are in-app only.** SMS is used for OTP and
  security messages.

## Additional entities (planned)

`school_whatsapp_bots`, `chat_sessions`, `otp_requests`, `parent_qr_codes`,
`staff_qr_codes`, `registration_submissions`, `verification_records`, badge
delegations, `attendance_daily`, `assessments` and `marks`, `report_comments`,
`result_releases`, `grading_scale_versions` (per term), `fee_periods`,
`fee_charges`, `fee_change_notes`, `ledger_entries` (including credit),
`unmatched_payments`, `school_profile` (contact details required), and a
`whatsapp` value for `consents.purpose`. See `ikonex.md` section 9.

Build `fee_periods` on generic billing periods even though Release 1 is terms
only, so custom structures later need no migration (a recommendation).

## Behaviour that shapes the model

- **Results release is a state machine** per class per term: draft → released
  → reopened → re-released. Parents keep seeing the last released version
  while a class is reopened.
- **Grading scale is versioned per term.** A term's levels are computed from
  the scale in force when it was released.
- **Fee due dates follow term start dates.** A term-date change moves due
  dates and creates a fee-change event with a note.
- **Pending registrations are not active records** and grant no access until
  a staff verifier confirms them; uploaded documents are deleted after
  approval.

## Settled vs. open decisions

Settled decisions and their reasoning live in `DECISIONS.md` — don't
re-litigate them here. Open questions that affect this document:

- Hosting location (Kenya only vs regional Africa, and where the AI model
  runs) — to be discussed with the project group.
- Sports/activities performance model (how it's scored, who enters it) — not
  yet designed.
- Store and caterer module data model — not yet designed.
- Fee structure variants (boarder/day, optional items).
- How a shared parent QR links each parent to the right child.
- Payment execution: the system records approved payments; whether it will
  ever initiate transfers itself (beyond the parent-initiated M-Pesa STK push)
  is undecided.
