# AGENTS.md

AI coding agents (and new contributors) read this file automatically. It gives the
project's commands and conventions so agents don't have to guess.

## Project overview

Ikonex Academy is a multi-school portal for Kenyan secondary schools (CBC,
Grade 7–12). One shared platform hosts many schools; parents, teachers, and
staff (nurse, librarian, bursar, dean of studies, principal/deputy) get
role-specific views. A Go modular monolith backs a Next.js PWA frontend and a
self-hosted Mistral AI service (Python/FastAPI) that assists parents and
teachers. Tenant isolation is enforced with PostgreSQL Row-Level Security
(`school_id`). Release 1 is a synthetic-data proof of concept only.

## Stack

- Backend: Go 1.22 (modular monolith), `net/http` ServeMux
- Database: PostgreSQL 16 (RLS per `school_id`), Redis 7 + Asynq
- Frontend: Next.js 14 (React, TypeScript, Tailwind, PWA)
- AI service: Python 3.11+ / FastAPI, self-hosted Mistral (no direct DB access)
- Integrations: Africa's Talking (SMS), WhatsApp Cloud API, Safaricom Daraja
  (M-Pesa)

## Setup

Prerequisites: Go 1.22+, Node 20+, Python 3.11+, Docker, PostgreSQL 16, Redis 7.

1. `cp .env.example .env` and fill in values.
2. `make up` (starts Postgres + Redis).
3. Apply migrations in order:
   `for f in migrations/*_init_schema.up.sql migrations/*_r1_*.up.sql; do psql -U "$DB_USER" -d "$DB_NAME" -f "$f"; done`
   (or `make migrate-up` once wired).
4. Backend: `cd backend && go mod download && go run cmd/server/main.go`
5. Frontend: `cd frontend && npm install && npm run dev`
6. AI service: `cd ai_service && python3 -m venv .venv && . .venv/bin/activate && pip install -r requirements.txt && uvicorn app.main:app --host 0.0.0.0 --port 8000 --reload`

Or run everything with `make dev` (Docker Compose).

## Build / test / lint commands

- All services: `make dev`
- Backend run: `make backend-run` (`cd backend && go run cmd/server/main.go`)
- Frontend run: `make frontend-run` (`cd frontend && npm run dev`)
- AI service run: `make ai-run` (`cd ai_service && uvicorn app.main:app --host 0.0.0.0 --port 8000 --reload`)
- Tests: `make test`
  - Go: `cd backend && go test -v ./...`
  - Python: `cd ai_service && python3 -m pytest tests/`
- Frontend lint: `cd frontend && npm run lint`
- Frontend build: `cd frontend && npm run build`
- Go format check: `cd backend && gofmt -l .`
- Go vet: `cd backend && go vet ./...`
- Migrations: `make migrate-up` / `make migrate-down` (apply via psql until wired)
- Clean: `make clean`

## Project structure

```
├── Docs/          # Architecture, requirements, decisions, security, agents
├── backend/       # Go modular monolith (cmd/, internal/, Dockerfile)
├── ai_service/    # Python/FastAPI AI service (app/, tests/)
├── frontend/      # Next.js PWA (src/app, components, lib)
├── migrations/    # PostgreSQL migrations (RLS, per-file owners)
├── tasks/         # Master integration plan + per-developer task specs
├── docker-compose.yml
├── Makefile
└── .env.example
```

Backend modules live under `backend/internal/` (academics, auth, billing,
clubs, config, consent, database, internalapi, jobs, library, middleware,
notifications, nurse, onboarding, whatsapp).

## Code style

- Go 1.22, module `github.com/Ashomondi/ikonex-academy/backend`. Format with
  `gofmt`; group imports (stdlib, then third-party). Use the standard
  `net/http` ServeMux. HTTP errors use the `APIError` + standard error envelope.
- Database: `snake_case` tables/columns; Go exported identifiers `PascalCase`;
  JSON tags `snake_case` (see `Docs/API_CONTRACTS.md`).
- Frontend: Next.js 14 App Router, TypeScript, Tailwind. Routes in
  `src/app/`, shared UI in `src/components/`, API clients in `src/lib/`.
- Python: FastAPI, pydantic v2, httpx. Keep zero direct DB access.
- Migrations: four numbered files with single owners (see master plan). Every
  tenant-owned table carries `school_id` plus `ENABLE/FORCE ROW LEVEL
  SECURITY` and a tenant-isolation policy.
- Commits: conventional style observed in history — `feat(scope): ...`,
  `fix(scope): ...`, `docs: ...`, `chore: ...`.

## Boundaries (do not touch)

- Generated/vendored dirs: `venv/`, `node_modules/`, `.next/`, `bin/`,
  `dist/`, `__pycache__/`, `.pytest_cache/`.
- Secrets: `.env`, `*.pem`, `*.key` — never commit.
- Baseline migration `migrations/000001_init_schema.*.sql` must not be edited;
  migration files `000002`–`000004` are owned by other developers.
- Never bypass RLS or read across schools; never give the AI service DB
  credentials.
- Release 1 uses synthetic data only — no real student/parent/staff records.
- Don't guess open decisions (hosting, fee variants, etc.); see
  `Docs/Ikonex.md` Appendix B and `Docs/REQUIREMENTS.md`.

## Change rules

- Keep changes small and focused; no unrelated refactors.
- Every query on a tenant-owned table must be scoped by `school_id`.
- Run tests before finishing (`make test`, `npm run lint`).
- Update `Docs/API_CONTRACTS.md` when endpoint shapes change.
- Add a migration for any schema change (follow the numbered-file ownership).
- AI proposes; humans confirm for payments and permanent records.

## Change size rules

- Make the smallest change that fully solves the request. Fewer lines is better, as long as it works.
- Touch only the files and functions the task requires. No refactors, renames, reformatting, or "cleanup" unless I ask.
- Do not add features, options, abstractions, config, or error handling that I did not ask for.
- Reuse existing code and project dependencies before writing new helpers or adding libraries.
- Do not write new tests, docs, or comments unless I ask.
- After editing, run the project's build/tests and confirm they pass. Do not report done on code you have not run.
- If the task is bigger than expected, split it into small steps and do one step at a time.

## PR conventions

- Commit messages: concise imperative, conventional style
  (`feat(module): ...`).
- A PR must include tests for new behaviour, docs updates for contract
  changes, and migrations for schema changes.
- Reference acceptance criteria (`AC-xx`) from `Docs/Ikonex.md` §14 where
  applicable.

## Non-negotiable rules

1. Every tenant query scoped by `school_id` (RLS is the enforcement layer).
2. AI service never touches the database directly.
3. Biometrics: encrypted templates only, never raw images; per-school keys.
4. Nurse records: clinical detail stays with the nurse; teachers get a flag.
5. No feature moves money or approves a payout on its own.
6. No diagnoses, discipline records, or child rankings.
7. Student/staff personal data stays hosted in Kenya (DPA 2019).
8. Release 1 uses synthetic data only.
9. WhatsApp carries minimal facts only (no results/report/health/discipline).
10. Don't guess open decisions — ask.

## Where to look before asking

- Product behaviour, decisions, open questions, status → `Docs/Ikonex.md`
- Data model and service boundaries → `Docs/ARCHITECTURE.md`
- Scope by release phase → `Docs/REQUIREMENTS.md`
- Endpoint shapes and AI-callable endpoints → `Docs/API_CONTRACTS.md`
- Consent, encryption, retention → `Docs/SECURITY.md`
- Why a past decision was made → `Docs/DECISIONS.md`
- Coding conventions → `Docs/AGENTS.md`
- Task breakdown and ownership → `tasks/MASTER-INTEGRATION-PLAN.md`
filePath: /home/banker/code/open/Ikonex-Academy/AGENTS.md