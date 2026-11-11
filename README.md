# Ikonex Academy

A multi-school management and CBC (Competency-Based Curriculum, Grade 7–12) portal for Kenyan secondary schools.

**Start with `ikonex.md`.** It is the central project reference: scope, roles, feature behaviour, decisions, open questions, and status. The other documents in `Docs/` hold detail on architecture, API contracts, security, and the decision log.

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
│   └── go.mod
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
│   └── 000001_init_schema.down.sql
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

6. **WhatsApp is minimal.** Each school has its own WhatsApp bot. It carries only minimal facts (fee balances, attendance status, payment prompts). Results and report content stay in the app. WhatsApp is a documented exception to the Kenya-hosting rule; see `SECURITY.md`.
7. **Human approval before parents see results.** Marks are entered by subject teachers; the class teacher releases results only when every subject has marks for every student.

---

## Release status (2026-10-29)

- **No code exists yet.** The repository holds documentation only. Directory entries above describe the *planned* layout, not what is built.
- **Release 1** is a test of the idea, built and tested with synthetic data only, and includes WhatsApp. It passes when onboarding works, parents can log in, and badges work well. See `ikonex.md` section 3.
- Modules for nurse, librarian, clubs, bursar, and leave arrive in Release 2. In Release 1 the Ikonex team loads each school's fee structure, and there is no bursar portal.
- `backend/internal/` will also need a `whatsapp/` module (webhook, per-school bot routing, OTP chat sessions); it is not in the tree above yet.
