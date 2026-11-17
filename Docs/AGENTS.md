# AGENTS.md

This file is read by AI coding assistants (Claude, or similar tools) before working
on this repository. Keep it short and current. If a rule here conflicts with what
you're about to do, stop and ask rather than guessing.

## What this project is

A multi-school portal for Kenyan secondary schools (CBC, Grade 7–12), hosted for
many schools on one platform. Parents, teachers, and school staff (nurse,
librarian, bursar, dean of studies, principal/deputy) each get a role-specific
view. An AI layer (self-hosted Mistral models) helps parents ask questions, pay
fees, and receive plain-language updates, and helps teachers and the bursar with
syllabus tracking and payment reconciliation.

See `ARCHITECTURE.md`, `REQUIREMENTS.md`, `API_CONTRACTS.md`, `SECURITY.md`, and
`DECISIONS.md` for the details behind everything summarised here.

## Stack

- **Backend:** Go (modular monolith — one service, packages per domain area:
  billing, academics, onboarding, notifications, etc.)
- **Database:** PostgreSQL, with row-level security (RLS) for tenant isolation
- **Frontend:** Next.js (React), built as a PWA for offline/low-connectivity use
- **AI service:** Python (FastAPI), calling self-hosted Mistral models
  (e.g. via vLLM or Ollama), kept as a separate service — never given direct
  database access
- **Jobs/cache:** Redis, with a Go job queue (e.g. Asynq) for SMS, STK push
  polling, and AI reconciliation/reminder jobs
- **Integrations:** Africa's Talking (SMS/USSD), WhatsApp Business Cloud API,
  Safaricom Daraja API (M-Pesa STK push, per-school Paybill/Till)

## Commands

> Fill these in as the project scaffolding is created — an AI assistant should
> never guess at commands.

- Backend tests: `TODO`
- Backend run (dev): `TODO`
- Frontend run (dev): `TODO`
- Migrations: `TODO`
- AI service run (dev): `TODO`

## Non-negotiable rules

These override convenience or "helpfulness" in any single task. See
`SECURITY.md` for the full reasoning.

1. **Every query on a tenant-owned table must be scoped by `school_id`.**
   Row-level security is the enforcement layer — never write a query or
   migration that bypasses it or reads across schools.
2. **The AI service never touches the database directly.** It calls
   permission-checked Go endpoints and only ever sees data the requesting
   user is allowed to see. See `API_CONTRACTS.md`.
3. **Biometric data:** store encrypted templates only, never raw fingerprint
   images. Encryption keys are per-school.
4. **Health data (nurse records):** class teachers get a notification that a
   visit happened, never the clinical detail.
5. **No feature sends money or approves a payout on its own.** Bursar payments
   go through the signatory approval workflow; the AI proposes matches, a
   human confirms them.
6. **Nothing in scope generates a diagnosis, discipline record, or ranking of
   children.** See `REQUIREMENTS.md` for the full out-of-scope list.
7. **Student and staff personal data stays hosted in Kenya**, per the Data
   Protection Act. Don't add a dependency that routes personal data through a
   foreign service without checking `SECURITY.md` first.

8. **Release 1 uses synthetic data only.** Never load real student, parent,
   or staff data into any environment until the launch gates in `SECURITY.md`
   are cleared.
9. **WhatsApp carries minimal facts only** (fee balances, attendance status,
   payment prompts). Never put results, report comments, health, or
   discipline information in a WhatsApp message; send a link to the app.
10. **Don't guess open decisions.** `ikonex.md` Appendix B lists them
    (hosting, who pays for messages, fee variants, and others). If a task
    depends on one, stop and ask.

## Conventions

> Expand this section as real conventions are established in the codebase —
> don't invent conventions here that the code doesn't actually follow.

- Go package layout: `TODO`
- Error handling style: `TODO`
- Naming conventions (DB, API, frontend): `TODO`
- Commit message style: `TODO`

## Where to look before asking the user

- Product behaviour, agreed decisions, open questions, status → `ikonex.md`
- Data model and service boundaries → `ARCHITECTURE.md`
- What's in scope, by release phase → `REQUIREMENTS.md`
- Endpoint shapes and what the AI service may call → `API_CONTRACTS.md`
- Consent, encryption, retention rules → `SECURITY.md`
- Why a past decision was made → `DECISIONS.md`
