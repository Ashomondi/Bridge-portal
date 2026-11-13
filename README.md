# Ikonex Academy

A multi-school management and CBC (Competency-Based Curriculum, Grade 7–12) portal for Kenyan secondary schools.

---

## Architecture Overview

```
                      ┌──────────────────────┐
 Parent / Teacher /   │   Next.js Frontend   │
 Staff (browser/PWA)  │   (React, PWA)       │
                      └──────────┬───────────┘
                                 │ HTTPS
                      ┌──────────▼───────────┐
                      │      Go Backend      │
                      │  (Modular Monolith)  │      ┌───────────────┐
                      │  - billing           │◄────►│  PostgreSQL   │
                      │  - academics          │      │  (RLS per     │
                      │  - onboarding         │      │   school_id)  │
                      │  - notifications      │      └───────────────┘
                      │  - internalapi        │
                      └───────┬──────────────┬┘
                              │              │
             ┌────────────────▼─┐   ┌────────▼───────────┐
             │ Redis + Asynq    │   │     AI Service     │
             │ - SMS jobs       │   │  (Python/FastAPI)  │
             │ - STK polling    │   │  - Self-hosted     │
             │ - Reminders      │   │    Mistral model   │
             └──────────────────┘   │  - Calls Go API    │
                                    │    only (no DB)    │
                                    └────────────────────┘
```

---

## Directory Structure

```text
├── Docs/                     # Architecture, requirements, decisions, security & agents
│   ├── AGENTS.md             # Guidelines and conventions for AI coding assistants
│   ├── API_CONTRACTS.md      # API contracts between frontend, backend, and AI service
│   ├── ARCHITECTURE.md       # High-level architecture and tenant isolation design
│   ├── DECISIONS.md          # Architectural Decision Records (ADRs)
│   ├── REQUIREMENTS.md       # Feature scope and release phase breakdown
│   └── SECURITY.md           # Kenyan Data Protection Act 2019 compliance & boundaries
├── backend/                  # Go modular monolith
│   ├── cmd/
│   │   ├── server/           # HTTP API server entry point
│   │   └── worker/           # Asynq background worker entry point
│   ├── internal/             # Domain modules
│   │   ├── academics/        # CBC curriculum, timetables, syllabus tracking
│   │   ├── auth/             # Phone number SMS OTP, sessions, memberships
│   │   ├── billing/          # Fee structures, student ledgers, M-Pesa STK push
│   │   ├── clubs/            # Journalism club, patron reviews, blog posts
│   │   ├── config/           # Environment & configuration loader
│   │   ├── consent/          # Parent consent centre, encrypted biometric templates
│   │   ├── database/         # PostgreSQL connection & RLS session binding
│   │   ├── internalapi/      # Internal endpoints callable strictly by AI service
│   │   ├── jobs/             # Asynq job queues (SMS, STK push polling)
│   │   ├── library/          # Book catalogue, loans, overdue tracking
│   │   ├── middleware/       # Auth, tenant RLS context, badge permissions
│   │   ├── notifications/    # Africa's Talking SMS, WhatsApp, in-app notifications
│   │   ├── nurse/            # Nurse consultations (clinical privacy protected)
│   │   └── onboarding/       # Staff invites via QR, student admission
│   ├── Dockerfile
│   ├── go.mod
│   └── go.sum
├── ai_service/               # Python / FastAPI AI service (zero direct DB access)
│   ├── app/
│   │   ├── api/              # AI feature routes
│   │   │   ├── blog_safety.py        # Content safety scanner (child PII protection)
│   │   │   ├── chat_payment.py       # STK push fee payment flow
│   │   │   ├── parent_assistant.py   # English & Swahili Q&A (fee/attendance/exams)
│   │   │   ├── payment_reminders.py  # 3-stage escalating default reminders
│   │   │   ├── reconciliation.py     # Unmatched M-Pesa match suggestions
│   │   │   ├── report_drafter.py     # CBC report card comments drafter
│   │   │   └── syllabus_mapper.py    # Spoken/typed lesson to CBC strand mapping
│   │   ├── clients/          # External clients
│   │   │   ├── backend_client.py     # Go internal API client (reads facts)
│   │   │   └── llm_client.py         # Self-hosted Mistral LLM client
│   │   ├── config.py         # AI service settings
│   │   └── main.py           # FastAPI application
│   ├── Dockerfile
│   └── requirements.txt
├── frontend/                 # Next.js React frontend (PWA)
│   ├── public/               # Static assets & PWA manifest.json
│   ├── src/
│   │   ├── app/              # Next.js App Router (role-based portals)
│   │   │   ├── (auth)/login/ # Phone SMS OTP login
│   │   │   ├── admin/        # Principal & deputy management, staff QR codes
│   │   │   ├── bursar/       # Fee ledger, fee structures, reconciliation
│   │   │   ├── librarian/    # Book catalogue, loans
│   │   │   ├── nurse/        # Sickbay consultations & teacher alerts
│   │   │   ├── parent/       # Child performance graph, fee ledger, STK push, AI chat
│   │   │   ├── teacher/      # Timetable, class roster, syllabus coverage
│   │   │   ├── layout.tsx    # Root layout
│   │   │   └── page.tsx      # Main landing & portal selector
│   │   ├── components/       # Reusable UI components
│   │   └── lib/              # API clients & utilities
│   ├── package.json
│   ├── tsconfig.json
│   └── tailwind.config.ts
├── migrations/               # PostgreSQL migrations with Row-Level Security (RLS)
│   ├── 000001_init_schema.up.sql
│   ├── 000001_init_schema.down.sql
│   ├── 000003_r1_academics.up.sql
│   ├── 000003_r1_academics.down.sql
│   ├── 000004_r1_billing_whatsapp_notifications.up.sql
│   └── 000004_r1_billing_whatsapp_notifications.down.sql
├── docker-compose.yml        # Multi-service local development orchestration
├── Makefile                  # Developer workflow commands
├── .env.example              # Master configuration template
└── .gitignore
```

---

## Core Non-Negotiable Rules

1. **PostgreSQL Row-Level Security (RLS)**: Every tenant-owned table carries `school_id`. Queries must be scoped to the active tenant via `SET LOCAL app.school_id = '<uuid>'`.
2. **AI Service Boundary**: The AI service never accesses PostgreSQL directly. It calls internal Go endpoints with verified permissions.
3. **Biometric Privacy**: Raw fingerprint images are never stored. Only encrypted templates with per-school keys are kept, with mandatory manual fallbacks.
4. **Nurse Record Isolation**: Clinical notes remain restricted to the nurse role. Class teachers receive only a notification flag that a visit occurred.
5. **Human Confirmation on Financial Actions**: System records and approves payments. Bursar payments follow a signatory workflow. AI can initiate STK push only after explicit parent confirmation.

## Backend build and tests

Run from the repository root with Go 1.22.2 or later and Docker Compose:

```sh
docker compose up -d --wait postgres
export DEV05_TEST_DATABASE_URL='postgres://ikonex_admin:ikonex_secure_pass@localhost:5432/postgres?sslmode=disable'
cd backend
go mod download
go build ./...
go test -count=1 -v ./...
```

Adjust the URL to your local PostgreSQL credentials/port. Use a dedicated
development/test server, never production: the database user needs superuser
capabilities to create disposable `dev05_*` databases and roles. Fixtures use
synthetic data, apply migrations 000001 and 000003 before 000004, and delete
their databases/roles at cleanup. Tenant behavior runs under a separate
non-superuser, non-BYPASSRLS role, so the RLS checks are meaningful.

Without `DEV05_TEST_DATABASE_URL`, database tests explicitly skip locally;
that is not full migration verification. If `CI` is nonempty, a missing URL
fails instead. `.github/workflows/backend.yml` supplies PostgreSQL 16 and the
URL and runs both build and uncached tests. Regression tests cover legacy
overpayment backfill, boarding-variant rollback, required verified-session
expiry, tenant isolation and schema up/down/up restoration.

## Migration and implementation status

Apply migrations in numeric order: **000001** (core), **000003** (academics),
**000004** (billing/chat/notifications). There is no 000002 file. Migration 000004
adds **12 tables**; their responsibilities and state machines are documented in
[ARCHITECTURE.md](Docs/ARCHITECTURE.md). Use a PostgreSQL migration runner;
the Makefile's advertised migration targets do not yet implement one.

Before downgrading 000004, export its new tables and boarding variant templates.
The old schema supports only one structure per school/year/grade/term: rollback
keeps `all`, otherwise `day`, otherwise `boarder`, with its original total/items,
and deletes the alternatives. Available credit returns to a signed balance
(`current_balance - credit_balance`). Reapplying the migration cannot recover
deleted variant definitions. See [the ADRs](Docs/DECISIONS.md).

The consumer contracts define identity/tenant boundaries and response helpers;
auth middleware, permission enforcement, billing handlers and durable audit
logging still require application integration. `AuditLogger` alone does not
address AC-29. In-app chat is Release 1 scope; WhatsApp routing schema prepares
for Release 3 delivery, as specified in the release ADR.
