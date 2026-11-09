# Codex conversation

## User

.

## Assistant

What would you like me to work on in Ikonex Academy?

## User

# MASTER DEVELOPMENT PROMPT
## Ikonex Academy – Release 1
### Developer 05: Backend Engineer C (Go Backend)

---

# 1. YOUR ROLE AND RESPONSIBILITY

You are acting as a Senior Go Backend Engineer, Software Architect, Database Engineer, Security Engineer, and Test Engineer.

You have been assigned Developer 05's responsibilities in the Ikonex Academy Release 1 project.

Your task is to independently design, implement, test, and prepare your assigned backend features for integration with the other developers' work.

You are not being asked to produce a quick prototype, demonstrate a few endpoints, or create placeholder implementations.

You must deliver production-quality, maintainable, secure, testable Go backend functionality that follows the architecture and requirements of the existing project.

You must take ownership of the entire implementation process, from repository inspection and architectural analysis through implementation, testing, and final verification.

You must work methodically.

**Do not attempt to implement everything in one large operation.**

First understand the project, create a comprehensive implementation plan, break that plan into small independently verifiable tasks, and then execute those tasks one at a time.

Every task must be tested before you proceed to the next dependent task.

---

# 2. AUTHORITATIVE PROJECT DOCUMENTS

Two important documents are provided:

1. `developer-05-backend-engineer-c.md`
2. `MASTER-INTEGRATION-PLAN.md`

You MUST read both documents in their entirety before implementing anything.

## 2.1 Developer 05 Assignment

Treat `developer-05-backend-engineer-c.md` as the authoritative description of your individual responsibilities.

It defines:

- Your role and objectives.
- Your assigned features.
- Functional requirements.
- Business rules.
- Security requirements.
- Database migration responsibilities.
- Expected Go package structure.
- API endpoints.
- Cross-developer integration responsibilities.
- File ownership restrictions.
- Testing requirements.
- Acceptance criteria.
- Definition of Done.

Do not casually reinterpret, omit, or simplify any of these requirements.

## 2.2 Master Integration Plan

Treat `MASTER-INTEGRATION-PLAN.md` as the authoritative source for understanding how your implementation fits into the wider Ikonex Academy architecture.

Study it carefully, particularly:

- Developer responsibilities and ownership.
- Feature ownership matrix.
- API conventions.
- Database migration ownership.
- PostgreSQL Row-Level Security requirements.
- Cross-module interfaces.
- Dependencies between developers.
- Shared resource ownership.
- Recommended implementation order.
- Integration milestones.
- Git branching conventions.
- Testing and quality assurance.
- Open decisions and configurable assumptions.
- Final Release 1 integration checklist.

You must understand the complete architecture before deciding how to implement your assigned features.

Do not treat Developer 05's assignment as an isolated project.

Your implementation must be designed to integrate with the other modules without violating their ownership boundaries.

---

# 3. IMPORTANT DEVELOPMENT CONTEXT

The other developers have not yet implemented their assigned features.

This means:

- Developer 03's authentication and onboarding implementation is not yet available.
- Developer 04's academics implementation is not yet available.
- Developer 01's Python AI service implementation is not yet available.
- Developer 02's frontend implementation is not yet available.

There is currently no completed implementation from these developers that you can rely on.

You must therefore approach this assignment as an independent development effort within a larger planned architecture.

## 3.1 Do not wait for the other developers

You must not postpone your implementation because another developer has not completed their work.

Instead:

1. Study the interfaces and integration contracts specified in the documents.
2. Identify what your modules need from other developers.
3. Define compatible interfaces at your module boundaries.
4. Use small test doubles, mocks, and controlled adapters where appropriate.
5. Implement your own functionality independently.
6. Verify that your modules can compile and be tested without the unfinished implementations.
7. Preserve the integration contracts so the real implementations can be connected later.

Do not create fake implementations of other developers' entire systems.

Only create the minimum interfaces and test doubles necessary to independently implement and test your responsibilities.

## 3.2 Do not assume missing functionality exists

For example:

- Do not assume `WithTenantTx` already exists.
- Do not assume the audit logger has been implemented.
- Do not assume authentication middleware is available.
- Do not assume the academics service exists.
- Do not assume the AI service is running.
- Do not assume the onboarding cleanup function exists.

Inspect the actual repository first.

If a dependency is missing, identify the required contract and create a compatible consumer-facing interface within your owned code where appropriate.

Never modify another developer's files to make your implementation easier.

## 3.3 Distinguish unavailable dependencies from implementation failures

Your modules must remain independently testable.

If a full integration test cannot run because another developer's implementation is unavailable:

- Run the appropriate unit tests.
- Use mocks or interface implementations.
- Test the expected integration contract.
- Clearly record the unavailable dependency.
- Do not falsely claim that full integration has passed.

---

# 4. STRICT FILE OWNERSHIP AND MODIFICATION RULES

This is one of the most important requirements.

You are ONLY permitted to create, edit, rename, or delete files within the paths explicitly assigned to Developer 05 in `developer-05-backend-engineer-c.md`.

The exclusive ownership paths are:

```text
backend/
├── cmd/
│   └── worker/
│       └── main.go
│
└── internal/
    ├── billing/
    │   └── **
    │
    ├── whatsapp/
    │   └── **
    │
    ├── notifications/
    │   └── **
    │
    ├── jobs/
    │   └── **
    │
    └── internalapi/
        └── **

migrations/
├── 000004_r1_billing_whatsapp_notifications.up.sql
└── 000004_r1_billing_whatsapp_notifications.down.sql
```

These are your exclusive development boundaries.

You may create additional files and subdirectories inside these locations when genuinely necessary.

For example, you may introduce:

- Additional service files.
- Repository files.
- Validation helpers.
- Provider clients.
- Test fixtures.
- Unit tests.
- Integration tests.
- Mock implementations.
- Internal documentation, if needed.

However, all such files must remain inside your assigned directories.

## 4.1 Files and directories you MUST NOT modify

You must not modify any of the following:

```text
backend/internal/auth/**
backend/internal/onboarding/**
backend/internal/middleware/**
backend/internal/database/**
backend/internal/academics/**

backend/cmd/server/main.go

backend/go.mod
backend/go.sum

ai_service/**
frontend/**

Docs/API_CONTRACTS.md

docker-compose.yml
Makefile
.env.example

migrations/000001_*
migrations/000002_*
migrations/000003_*
```

Also do not modify any other files outside your explicitly assigned ownership paths.

The above restrictions are intentional.

Even if you discover a problem in another developer's implementation or a shared file, do not fix it directly.

Instead:

1. Identify the problem.
2. Explain why it affects your implementation.
3. Determine whether an adapter or interface inside your owned directories can accommodate it.
4. Record the required coordination point.
5. Continue with the work that can be completed independently.

If a change outside your ownership is absolutely necessary, stop before making it and explain the issue. Do not silently violate the ownership boundary.

## 4.2 Mandatory file-change verification

Before beginning development:

- Inspect Git status.
- Record the existing working-tree state.
- Identify existing modifications.
- Identify the current branch.
- Record the current commit.

Before every commit and at the end of development, inspect:

```bash
git status --short
git diff --name-only
git diff --cached --name-only
```

Verify that every changed file belongs to your assigned paths.

Do not overwrite existing work belonging to another developer.

---

# 5. GIT BRANCH REQUIREMENTS

All development must happen in a separate feature branch.

Do not develop directly on `main`, `release-1`, or another shared integration branch.

Use the Developer 05 naming convention from the Master Integration Plan.

For this combined assignment, create:

```text
feat/ian-backend-engineer-c
```

If the repository already contains an appropriate Developer 05 branch, inspect it and determine whether it is suitable before creating another one.

## 5.1 Branch creation procedure

Before creating the branch:

1. Inspect the current branch.
2. Inspect available local branches.
3. Inspect remote branches if the remote is available.
4. Identify whether `main` or `release-1` is the correct starting point.
5. Inspect the current working tree.
6. Preserve any existing uncommitted work.

Create your feature branch from the appropriate existing base.

Do not automatically assume that `main` is the correct base if the project is using `release-1`.

If the working tree contains unrelated changes, do not discard, reset, clean, or overwrite them.

Preserve them and work safely.

## 5.2 Commit strategy

Use meaningful conventional commits.

Examples:

```text
feat(notifications): implement transactional SMS dispatcher

feat(billing): implement fee structures and term charging

feat(billing): implement student ledger and credit carry-forward

feat(whatsapp): implement school-scoped OTP sessions

feat(billing): integrate Daraja STK push

feat(billing): implement C2B payment reconciliation

feat(jobs): implement Asynq background workers

feat(internalapi): implement authenticated AI gateway

test(billing): add ledger and payment integration tests

test(whatsapp): add OTP security and tenant isolation tests
```

Make commits only after the associated task has been implemented and tested.

Do not create meaningless commits just to increase commit count.

Do not push or merge into shared branches unless explicitly authorized.

---

# 6. DEVELOPMENT PHILOSOPHY

Follow these principles throughout the assignment.

### 6.1 Understand before implementing

Never edit a file simply because its name appears in the assignment.

First understand:

- What the existing code does.
- How the existing packages are structured.
- Which architectural conventions are already established.
- Which dependencies are available.
- Which database tables already exist.
- Which integration contracts must be respected.

### 6.2 Small, incremental tasks

Every feature must be broken into small tasks.

Each task should have:

- A clear objective.
- A defined scope.
- Expected files to create or modify.
- Dependencies.
- Implementation steps.
- Expected behaviour.
- Tests.
- A completion condition.

Avoid tasks such as:

"Implement the entire billing system."

Prefer tasks such as:

"Implement validation for fee structure creation, including academic year, grade, term, item amounts, and tenant ownership."

### 6.3 Test continuously

Do not postpone testing until the end.

After implementing a small unit of functionality:

1. Format the code.
2. Compile the relevant package.
3. Run the relevant tests.
4. Examine failures.
5. Fix problems.
6. Rerun the tests.
7. Verify that existing tests still pass.

Only proceed when the current task passes its verification gate.

### 6.4 Avoid premature complexity

Use clean Go architecture without introducing unnecessary abstractions.

Prefer:

- Small interfaces.
- Explicit dependencies.
- Clear service boundaries.
- Repository abstractions where useful.
- Strong input validation.
- Database transactions.
- Context-aware operations.
- Deterministic tests.

Do not introduce unnecessary frameworks or dependencies.

---

# 7. PHASE ZERO: COMPLETE REPOSITORY AND ARCHITECTURAL INVESTIGATION

This phase is mandatory.

Do not begin feature implementation until it is complete.

## Task 0.1: Read the assignment documents

Read the entire contents of:

```text
developer-05-backend-engineer-c.md
MASTER-INTEGRATION-PLAN.md
```

Extract all requirements relevant to Developer 05.

Build an internal requirements checklist covering:

- BE-C1.
- BE-C2.
- BE-C3.
- Database migration requirements.
- API endpoints.
- Security invariants.
- Integration interfaces.
- Testing requirements.
- Acceptance criteria.
- Definition of Done.

## Task 0.2: Inspect the repository

Inspect the complete project structure.

Understand:

```text
backend/
migrations/
Docs/
ai_service/
frontend/
```

You are allowed to READ files outside your ownership.

You are NOT allowed to MODIFY them.

Inspect the existing Go backend architecture.

Pay particular attention to:

- `backend/go.mod`
- Existing internal packages.
- Database connection management.
- Transaction handling.
- Tenant isolation.
- Authentication conventions.
- Middleware conventions.
- Error handling.
- API response structures.
- Logging.
- Configuration access.
- Existing migrations.
- Existing database models.

Do not assume that the repository matches the architecture documents perfectly.

Identify differences between the documented architecture and the actual implementation.

## Task 0.3: Analyse the Master Integration Plan

Perform a focused architectural analysis of Developer 05's relationships with the other developers.

Create an internal dependency map.

At minimum, analyse:

**Developer 03 – Evans**

Required dependencies:

- `database.WithTenantTx`
- `database.AuditLogger`
- Authentication and authorization context.
- Middleware conventions.
- `notifications.SMSDispatcher`
- `billing.StudentChargeHook`
- Unverified registration cleanup integration.

**Developer 04 – Ema**

Required dependencies:

- `notifications.Service.CreateInAppNotification`
- `billing.Service.OnTermStartDateChanged`
- `academics.Service.GetStudentAttendanceSummary`
- `academics.Service.GetStudentMarksSummary`

**Developer 01 – Sos**

Required integration:

- `X-Internal-Service-Token`
- Signed `X-User-Context`
- AI parent-query endpoint.
- Internal school and student information endpoints.
- Confirmed STK Push endpoint.
- AI action audit logging.

**Developer 02 – Ashy**

Required API compatibility:

- Fee balance.
- Student ledger.
- STK Push.
- Unmatched payments.
- Notifications.
- Chat messaging.

Document the interfaces your implementation must expose.

Do not wait for these developers to implement their modules.

## Task 0.4: Inspect the existing database schema

Read the existing migrations.

Identify the actual structure of:

- `schools`
- `users`
- `memberships`
- `academic_years`
- `terms`
- `grade_levels`
- `students`
- `student_guardians`
- `enrolments`
- `fee_structures`
- `fee_structure_items`
- `fee_ledgers`
- `payments`
- `audit_log`
- `sender_rate_blocks`
- Other relevant existing tables.

Understand existing:

- Primary keys.
- Foreign keys.
- Unique constraints.
- Tenant ownership.
- RLS policies.
- Existing indexes.
- Existing enum or CHECK constraints.
- Relevant existing columns.

Your migration must extend the existing schema without duplicating existing tables or introducing conflicting definitions.

## Task 0.5: Identify missing dependencies

Prepare a dependency matrix:

| Dependency | Expected owner | Available? | Required contract | Approach |
|---|---|---|---|---|
| Tenant transaction helper | Evans | Inspect | WithTenantTx | Adapter/interface |
| Audit logger | Evans | Inspect | AuditLogger | Consumer contract |
| SMS integration | Developer 05 | Your implementation | SMSDispatcher | Implement |
| Academics summary service | Ema | Inspect | Read methods | Interface/mock |
| AI service | Sos | Not yet implemented | HTTP contract | HTTP client |
| Redis | Shared infrastructure | Inspect | Connection | Existing configuration |

Do not invent implementations for unavailable external dependencies.

## Task 0.6: Produce a repository assessment

Before implementation, report:

1. Current repository structure.
2. Current Git branch.
3. Current working-tree state.
4. Existing Go architecture.
5. Existing migration state.
6. Available dependencies.
7. Missing dependencies.
8. Conflicts between documentation and actual repository.
9. Your proposed integration strategy.
10. Any genuine blockers.

If a documented assumption cannot be verified, identify it explicitly.

**Completion gate:** Repository and architecture analysis is complete, and the development branch has been created safely.

---

# 8. PHASE ONE: CREATE THE IMPLEMENTATION PLAN

Before writing feature code, create a detailed implementation plan.

The plan must cover all three feature groups.

Do not begin implementation until the plan has been presented and its internal dependencies are understood.

## 8.1 Required implementation plan structure

For every task, include:

| Field | Description |
|---|---|
| Task ID | Unique identifier |
| Feature | BE-C1, BE-C2, or BE-C3 |
| Task name | Specific implementation objective |
| Description | What must be implemented |
| Dependencies | What must exist first |
| Files | Exact owned files involved |
| Implementation steps | Small ordered actions |
| Tests | Tests required |
| Verification | Commands and expected results |
| Acceptance criteria | Conditions for completion |
| Status | Pending, In Progress, Blocked, Completed |

Use identifiers such as:

```text
INF-01
DB-01
NOTIF-01
BILL-01
MPESA-01
WA-01
JOB-01
API-01
```

## 8.2 Required implementation order

Use the following as the initial execution order, adjusting it only when repository inspection demonstrates a technical reason.

### Phase A – Foundation and Integration Contracts

1. Git branch and repository baseline.
2. Dependency and architecture analysis.
3. Shared response and error handling conventions within owned packages.
4. Database integration adapter boundaries.
5. Notifications package skeleton.
6. SMS dispatcher interface and implementation.
7. In-app notification service.
8. Notification tests.

### Phase B – Database Migration

1. Inspect existing schema.
2. Design Release 1 migration.
3. Create migration UP file.
4. Create migration DOWN file.
5. Add constraints and indexes.
6. Add tenant isolation policies.
7. Test migration application.
8. Test migration rollback.
9. Verify schema integrity.

### Phase C – Core Billing Engine

1. Billing service structure.
2. Fee structure validation.
3. Fee structure loading.
4. Fee period creation.
5. Automated term charging.
6. Idempotency and duplicate-charge prevention.
7. Student admission billing hook.
8. Ledger entry creation.
9. Balance calculation.
10. Credit balance calculation.
11. Credit carry-forward.
12. Student ledger retrieval.
13. Fee balance endpoint.
14. Billing tests.

### Phase D – Mid-Year Fee Changes

1. Fee change request validation.
2. Mandatory note enforcement.
3. Fee adjustment calculations.
4. Ledger updates.
5. Fee change history.
6. Audit logging.
7. Targeted guardian notifications.
8. Term start-date change hook.
9. Tests.

### Phase E – M-Pesa Integration

1. Daraja configuration.
2. Credentials and secret handling.
3. OAuth token retrieval.
4. STK Push request construction.
5. STK Push validation.
6. Guardian authorization.
7. Phone-change payment freeze.
8. STK Push HTTP integration.
9. Callback processing.
10. Payment persistence.
11. Ledger crediting.
12. Callback idempotency.
13. Direct Paybill C2B confirmation.
14. Admission-number matching.
15. Unmatched payment creation.
16. Unmatched payment listing.
17. Manual payment resolution.
18. Audit logging.
19. M-Pesa tests.

### Phase F – WhatsApp Integration

1. WhatsApp configuration.
2. Webhook verification and request validation.
3. Per-school bot identification.
4. Tenant resolution.
5. Admission number lookup.
6. Generic unknown-admission response.
7. Guardian lookup.
8. Phone masking.
9. Guardian selection.
10. OTP generation.
11. OTP storage and verification.
12. SMS dispatch.
13. OTP cooldown.
14. Sender rate limiting.
15. Twenty-four-hour sender blocking.
16. Six-hour verified sessions.
17. Expired-session handling.
18. AI service forwarding.
19. WhatsApp response delivery.
20. Security tests.

### Phase G – Internal AI API

1. Service-token authentication.
2. HMAC context verification.
3. Expiration validation.
4. Context parsing.
5. Tenant context enforcement.
6. Guardian-child authorization.
7. School profile endpoint.
8. Parent children endpoint.
9. Fee balance endpoint.
10. Attendance summary endpoint.
11. Marks summary endpoint.
12. Confirmed STK Push endpoint.
13. AI audit logging.
14. Chat proxy.
15. Security tests.

### Phase H – Asynq Background Jobs

1. Redis configuration integration.
2. Asynq client.
3. Task definitions.
4. Transactional SMS job.
5. SMS retries.
6. STK Push polling job.
7. STK timeout handling.
8. Unverified registration purge integration.
9. Periodic scheduling.
10. Worker entry point.
11. Graceful shutdown.
12. Worker tests.

### Phase I – Final Verification

1. Full Go formatting.
2. Full Go compilation.
3. Unit tests.
4. Race detection.
5. Migration tests.
6. Security tests.
7. Tenant isolation tests.
8. Cross-module contract tests.
9. Acceptance criteria verification.
10. File ownership verification.
11. Git diff review.
12. Final implementation report.

---

# 9. PHASE TWO: DATABASE MIGRATION

Implement:

```text
migrations/000004_r1_billing_whatsapp_notifications.up.sql

migrations/000004_r1_billing_whatsapp_notifications.down.sql
```

These are your assigned migration files.

Follow the schema requirements in Developer 05's assignment.

The migration must support:

- M-Pesa configuration.
- Generic billing periods.
- Fee charges.
- Fee change notes.
- Ledger entries.
- Credit balances.
- Unmatched payments.
- WhatsApp bot configuration.
- Chat sessions.
- In-app notifications.
- AI action logs.

Required tables:

```text
school_mpesa_configs
fee_periods
fee_charges
fee_change_notes
ledger_entries
unmatched_payments
school_whatsapp_bots
chat_sessions
in_app_notifications
ai_action_logs
```

Required alteration:

```text
fee_ledgers:
    credit_balance
    due_date
```

## 9.1 Database safety requirements

For every new tenant-owned table:

- Include `school_id`.
- Reference `schools(id)`.
- Enable Row-Level Security.
- Force Row-Level Security.
- Create tenant isolation policies.
- Add appropriate indexes.
- Add necessary foreign keys.
- Add appropriate unique constraints.

Follow the integration plan's RLS convention.

Do not assume that creating an RLS policy alone guarantees tenant isolation.

Consider how transactions establish the tenant context.

Your implementation must work with the project's intended `WithTenantTx` mechanism.

## 9.2 Migration quality

The migration must:

- Apply cleanly after the existing migrations.
- Avoid redefining existing tables.
- Avoid modifying other developers' migration files.
- Be reversible.
- Preserve referential integrity.
- Avoid destructive rollback behaviour where existing data could be affected.

Test both migration directions.

Do not proceed to dependent billing implementation until the migration has been reviewed and verified.

---

# 10. PHASE THREE: BE-C1 – BILLING AND STUDENT LEDGERS

Implement the complete billing engine.

Owned package:

```text
backend/internal/billing/
```

Expected files include:

```text
billing.go
structures.go
ledgers.go
fee_changes.go
billing_test.go
```

You may add more files within this directory when necessary.

## 10.1 Team-loaded fee structures

Implement:

```http
POST /api/v1/billing/fee-structures
```

Only authorized Ikonex administrators may load fee structures.

A fee structure belongs to:

- A school.
- An academic year.
- A grade level.
- A term.

Support itemized fee structure lines.

Validate:

- School ownership.
- Academic year.
- Term.
- Grade level.
- Fee items.
- Amounts.
- Duplicate structures.
- Invalid or inconsistent references.

Respect the project's existing fee structure schema.

The Release 1 system does not introduce a bursar workflow.

Fee structures are loaded by the Ikonex team.

## 10.2 Generic billing periods

Implement `fee_periods` using a generic period model.

For Release 1:

```text
period_type = term
```

The due date must correspond to:

```text
terms.start_date
```

Preserve the ability to introduce additional period types in future releases without redesigning the entire schema.

## 10.3 Automated term charging

When a fee structure is loaded for an active academic year and term:

- Identify enrolled students in the applicable grade.
- Generate the applicable student charges.
- Generate corresponding ledger entries.
- Assign the correct term due date.
- Apply available student credit.
- Prevent duplicate charges.

Term charging must be idempotent.

Repeatedly submitting the same fee structure must not charge students twice.

Use database constraints and transactional safeguards.

The system must also expose the required student billing hook for new student admission.

Implement:

```text
StudentChargeHook
```

with a stable, documented contract that can be injected into Evans's onboarding module later.

Do not modify onboarding files.

## 10.4 Student financial ledgers

Implement reliable student ledger calculations.

The system must track:

- Opening balance.
- Total billed amount.
- Total paid amount.
- Current balance.
- Credit balance.
- Due date.
- Fee item breakdown.
- Historical ledger entries.
- Fee change notes.

Maintain:

```text
current_balance >= 0
credit_balance >= 0
```

Use precise monetary arithmetic.

Do not use `float32` or `float64` for financial calculations.

Use PostgreSQL `NUMERIC(12,2)` and an appropriate Go representation.

## 10.5 Overpayment and credit carry-forward

If a student's payments exceed their total outstanding billed amount:

- Set current balance to zero.
- Store the excess as credit.
- Preserve the payment history.
- Automatically apply available credit to the next term charge.
- Record the credit application as a ledger entry.

Do not lose or duplicate credits.

Ensure that concurrent payment processing cannot apply the same credit twice.

## 10.6 Ledger endpoints

Implement:

```http
GET /api/v1/students/{id}/fee-balance

GET /api/v1/students/{id}/ledger
```

Authorization:

- Any guardian linked to the student may access the student's fees.
- Authorized Ikonex team members may access the appropriate records.

One linked guardian must not be able to access unrelated students.

Return the required financial breakdown.

Use the project's standard JSON response envelope.

---

# 11. PHASE FOUR: MID-YEAR FEE CHANGES

Implement:

```text
backend/internal/billing/fee_changes.go
```

Endpoint:

```http
POST /api/v1/fee-changes
```

This functionality is restricted to authorized Ikonex team members.

## 11.1 Mandatory explanation note

Every fee adjustment must include a non-empty explanation note.

If missing or empty:

```http
422 Unprocessable Entity
```

Error code:

```text
FEE_CHANGE_NOTE_REQUIRED
```

Do not silently accept missing notes.

## 11.2 Adjustment processing

A valid fee change must:

1. Identify the affected students.
2. Calculate the financial adjustment.
3. Update the relevant fee charges.
4. Update the corresponding ledger balances.
5. Preserve previous ledger entries.
6. Create a `fee_change_notes` record.
7. Write an audit log entry.
8. Notify affected guardians.

The adjustment must be transactional.

A failed transaction must not leave partially updated student ledgers.

## 11.3 Targeted notifications

Notify only guardians linked to affected students.

Do not notify every guardian in the school.

Avoid duplicate notifications.

## 11.4 Term date changes

Expose:

```go
OnTermStartDateChanged(
    ctx,
    tx,
    schoolID,
    termID,
    newStartDate,
    actorUserID,
    note,
)
```

Use the exact compatible method contract defined by the integration requirements, adapting Go types as needed to match the repository.

It must:

- Update the relevant `fee_periods.due_date`.
- Update corresponding `fee_charges.due_date`.
- Record the change.
- Preserve auditability.
- Notify affected guardians.

This method will eventually be called by Ema's calendar implementation.

Do not modify Ema's files.

---

# 12. PHASE FIVE: BE-C2 – SAFARICOM DARAJA M-PESA

Owned package:

```text
backend/internal/billing/
```

Expected files:

```text
mpesa_stk.go
mpesa_c2b.go
unmatched.go
```

Implement the complete M-Pesa integration against Safaricom Daraja sandbox.

## 12.1 Configuration

Support school-specific M-Pesa configuration.

The system must resolve the appropriate Paybill configuration using the receiving business shortcode.

Never hardcode school-specific credentials.

Protect secrets.

Do not expose passkeys, access tokens, or internal credentials through API responses or logs.

## 12.2 STK Push

Implement:

```http
POST /api/v1/payments/stk-push
```

Support linked guardians initiating payments.

Validate:

- Student existence.
- School ownership.
- Guardian-child relationship.
- Positive payment amount.
- Registered phone number.
- Applicable M-Pesa configuration.

### Important: Any positive amount

Accept any:

```text
amount > 0
```

Do not reject a payment because it is:

- Less than the outstanding balance.
- Equal to the outstanding balance.
- Greater than the outstanding balance.

Overpayment must be handled through ledger credit.

### Phone-change security freeze

If:

```text
source = chat
```

and:

```text
users.chat_payment_paused_until > NOW()
```

Reject the request.

Return:

```http
403 Forbidden
```

Code:

```text
CHAT_PAYMENT_PAUSED_PHONE_CHANGE
```

Do not accidentally block ordinary web-initiated payments under this specific chat restriction.

### Explicit AI payment confirmation

For internal AI-initiated STK Push:

Require:

```text
parent_confirmed = true
```

Otherwise return:

```http
422 Unprocessable Entity
```

The AI service must never initiate a payment without explicit parental confirmation.

### Account reference

Use:

```text
student.admission_number
```

as the Daraja AccountReference.

### STK polling

If the callback does not arrive within approximately 45 seconds, enqueue the appropriate polling task.

## 12.3 Daraja callbacks

Implement:

```http
POST /api/v1/billing/mpesa/callback
```

Process callback results safely.

Requirements:

- Validate callback structure.
- Handle unsuccessful transactions.
- Persist successful transactions.
- Prevent duplicate processing.
- Update payment records.
- Update student ledgers.
- Record payment ledger entries.
- Notify the relevant guardians.

Callback processing must be idempotent.

A callback received twice must not credit the student twice.

## 12.4 Direct Paybill C2B

Implement:

```http
POST /api/v1/billing/mpesa/c2b-confirmation
```

Resolve the school from the receiving Paybill shortcode.

Match the payment using:

```sql
UPPER(TRIM(admission_number))
=
UPPER(TRIM(BillRefNumber))
```

The match must be scoped to the resolved school.

Never match an admission number globally across all schools.

### Valid match

If a valid student exists:

- Create the payment.
- Associate the student.
- Record the payment ledger entry.
- Update the financial ledger.
- Apply overpayment credit if applicable.
- Send an in-app receipt notification.

### Invalid or missing reference

If no valid student matches:

- Create the payment record.
- Mark it unmatched.
- Create an `unmatched_payments` record.
- Preserve the raw account reference.
- Preserve the payer phone.
- Preserve the payer name.
- Preserve the transaction reference.
- Preserve the amount.

Never silently discard an incoming payment.

## 12.5 Unmatched payment resolution

Implement:

```http
GET /api/v1/payments/unmatched

POST /api/v1/payments/{id}/match
```

Only authorized Ikonex team members may resolve unmatched payments.

The resolution operation must:

- Validate the payment.
- Validate the target student.
- Confirm school ownership.
- Credit the correct ledger.
- Apply overpayment rules.
- Mark the unmatched record resolved.
- Record the resolving user.
- Write an audit log entry.
- Prevent duplicate resolution.

Do not implement Release 2 AI fuzzy reconciliation.

Manual resolution is the Release 1 requirement.

---

# 13. PHASE SIX: BE-C3 – NOTIFICATIONS

Owned package:

```text
backend/internal/notifications/
```

Implement:

```text
notifications.go
notifications_test.go
```

## 13.1 In-app notifications

Implement:

```http
GET /api/v1/notifications

PATCH /api/v1/notifications/{id}/read
```

Requirements:

- List notifications for the authenticated user.
- Support pagination where appropriate.
- Return the standard response envelope.
- Mark notifications as read.
- Record `read_at`.
- Prevent access to another user's notifications.
- Enforce tenant isolation.

Expose:

```go
Service.CreateInAppNotification(...)
```

The method must be usable by:

- Billing.
- WhatsApp.
- Other backend modules.
- Ema's academics module when integrated later.

Keep the interface stable and documented.

## 13.2 Africa's Talking SMS

Implement an SMS dispatcher.

Expose a compatible:

```go
SMSDispatcher
```

interface.

It must support:

1. App login OTP.
2. WhatsApp session OTP.
3. Old-number security notifications after phone changes.

Do not implement Release 2 SMS event alerts.

## 13.3 Sandbox behaviour

When:

```text
AFRICASTALKING_USERNAME=sandbox
```

use a deterministic local sandbox logger.

The sandbox implementation must:

- Avoid real SMS delivery.
- Make tests deterministic.
- Record the intended recipient and message metadata safely.
- Avoid logging sensitive OTP values in production logs.

For production configuration, use the Africa's Talking integration.

Normalize phone numbers to E.164 at the API boundary.

---

# 14. PHASE SEVEN: WHATSAPP BOT AND OTP STATE MACHINE

Owned package:

```text
backend/internal/whatsapp/
```

Expected files:

```text
whatsapp.go
masking.go
whatsapp_test.go
```

This feature has important security requirements.

Do not treat it as a simple messaging webhook.

Implement it as a carefully controlled identity-verification state machine.

## 14.1 Per-school bot routing

Each school has its own WhatsApp Business number.

Use:

```text
school_whatsapp_bots
```

Resolve the school using:

- `metadata.display_phone_number`
- `phone_number_id`

Every subsequent student lookup must be scoped to that school.

Use tenant-aware transactions.

School A's bot must never resolve School B's students.

## 14.2 State 1 – Admission number

Initial state:

```text
AWAITING_ADMISSION_NUMBER
```

Any WhatsApp number may initiate a chat.

Ask for the student's admission number.

Do not assume that the sender's WhatsApp number is registered to the guardian.

## 14.3 State 2 – Guardian selection

State:

```text
AWAITING_GUARDIAN_SELECTION
```

Look up the student using:

- The resolved school ID.
- The submitted admission number.

If the student does not exist:

Return a generic response:

"We could not verify that admission number. Please check and try again or contact the school office."

Do not reveal student information.

Do not send an OTP.

If the student exists:

- Fetch linked guardians.
- Retrieve their registered phone numbers.
- Mask the phone numbers.
- Ask the sender to select the intended guardian.

Example:

```text
1. 07•• ••• 482
2. 07•• ••• 910
```

If only one guardian exists, show the masked number and ask for confirmation before sending OTP.

## 14.4 Phone masking

Implement a dedicated helper.

Example:

```text
+254712345678
```

Expected display:

```text
07•• ••• 678
```

Do not expose the full registered phone number through WhatsApp.

## 14.5 State 3 – OTP verification

State:

```text
AWAITING_OTP
```

Generate a secure six-digit OTP.

Send the OTP ONLY to the selected guardian's registered phone number.

Never send the OTP to the requesting WhatsApp number.

### Rate limits

Enforce:

**Limit 1: One-minute cooldown**

Do not allow another OTP request within 60 seconds from:

- The same requesting WhatsApp number.
- The same `(school_id, admission_number)` combination.

Return HTTP 429 for the REST testing endpoint when applicable.

**Limit 2: Six requests within 30 minutes**

Count OTP requests made by the requesting WhatsApp number.

When the sixth request occurs:

- Block that requesting WhatsApp number.
- Set the block duration to 24 hours.
- Persist the block in `sender_rate_blocks`.
- Refuse the request.

Crucially:

Do not block:

- The guardian's user account.
- The guardian's registered phone number.
- The guardian's ability to log into the web application.

Only the requesting WhatsApp number is blocked.

## 14.6 State 4 – Verified session

State:

```text
VERIFIED
```

After successful OTP verification:

Set:

```text
verified_at = NOW()
expires_at = NOW() + INTERVAL '6 hours'
```

The verified session must remain valid for six hours.

For each valid incoming message:

1. Validate the session.
2. Confirm that it has not expired.
3. Build the signed user context.
4. Forward the message to the AI service.
5. Receive the AI response.
6. Send the response back through WhatsApp.

If the session has expired:

- Do not process the query.
- Require fresh OTP verification.

The exact expiry boundary must be tested.

## 14.7 REST testing endpoints

Expose:

```http
POST /api/v1/chat/sessions

POST /api/v1/chat/sessions/{id}/otp
```

These endpoints must exercise the same security and state-transition logic as the WhatsApp webhook.

Do not implement two separate versions of the OTP state machine.

Share the underlying service logic.

---

# 15. PHASE EIGHT: INTERNAL AI API

Owned package:

```text
backend/internal/internalapi/
```

Expected files:

```text
internalapi.go
context_signer.go
internalapi_test.go
```

This module is a critical security boundary.

The Python AI service must not access the database directly.

All access to school and student information must pass through your Internal AI API.

## 15.1 Authentication

Every internal request must validate:

```text
X-Internal-Service-Token
```

and:

```text
X-User-Context
```

The user context contains:

- User ID.
- School ID.
- Membership ID.
- Role.
- Channel.
- Session ID.
- Expiration.

Follow the exact signing format defined in the Master Integration Plan.

Use HMAC-SHA256.

Verify:

- Token presence.
- Token validity.
- HMAC signature.
- Context structure.
- Expiration.
- Required fields.
- Allowed channel values.

Use constant-time comparison for signatures.

Reject invalid requests.

Do not trust a decoded payload until its signature has been verified.

## 15.2 Context signing

Implement:

```text
context_signer.go
```

This will sign user context for outbound requests to the AI service.

The format must remain compatible with Sos's planned Python implementation.

Do not invent a different encoding or signature convention.

## 15.3 Tenant isolation

For every internal endpoint:

- Establish the correct tenant context.
- Use tenant-aware transactions.
- Scope student queries to the context school.
- Validate guardian-child relationships.
- Prevent cross-school data exposure.

## 15.4 Required internal endpoints

Implement:

```http
GET /api/v1/internal/school-profile

GET /api/v1/internal/parents/{user_id}/children

GET /api/v1/internal/students/{id}/fee-balance

GET /api/v1/internal/students/{id}/attendance-summary

GET /api/v1/internal/students/{id}/marks-summary

POST /api/v1/internal/payments/stk-push

POST /api/v1/internal/ai/audit-log
```

Also implement:

```http
POST /api/v1/chat/message
```

The chat endpoint is the frontend's chat proxy.

## 15.5 Guardian-child authorization

For parent-facing endpoints:

The student must belong to the authenticated guardian.

If Parent A attempts to access Parent B's child:

Return:

```http
403 Forbidden
```

Do not rely on the AI service to enforce this.

The Go backend must enforce authorization independently.

## 15.6 Academics integration

The attendance and marks endpoints will call Ema's academics service.

Define an interface for the expected operations.

Use a mock in your tests.

Do not create an alternative implementation of the academics system.

## 15.7 AI action logging

Every relevant AI action must be recorded in:

```text
ai_action_logs
```

Record appropriate metadata, including:

- School.
- User.
- Session.
- Channel.
- Feature.
- Intent.
- Endpoints read.
- Action taken.
- Additional safe metadata.

Never log secrets, OTPs, or sensitive authentication headers.

---

# 16. PHASE NINE: ASYNQ BACKGROUND WORKERS

Owned paths:

```text
backend/internal/jobs/
backend/cmd/worker/main.go
```

Expected:

```text
jobs.go
```

and:

```text
main.go
```

Implement Redis-backed Asynq processing.

## 16.1 Required queues

Implement:

```text
sms:transactional
mpesa:stk_poll
onboarding:purge_unverified
```

## 16.2 Transactional SMS

Implement asynchronous SMS delivery.

Requirements:

- Enqueue messages.
- Process them in the worker.
- Handle provider errors.
- Retry temporary failures.
- Avoid duplicate delivery where possible.
- Handle permanent failures appropriately.

## 16.3 STK polling

Implement the STK polling handler.

When the callback does not arrive within approximately 45 seconds:

- Query the Daraja STK status.
- Interpret the response.
- Update the transaction appropriately.
- Avoid duplicate ledger crediting.

## 16.4 Unverified registration cleanup

Register a periodic task that calls the cleanup integration supplied by Evans.

Expected behaviour:

- Use the configured retention period.
- Respect `UNVERIFIED_SUBMISSION_RETENTION_DAYS`.
- Default to seven days as specified in the integration plan.
- Do not implement an independent replacement for Evans's cleanup logic.

Define a compatible interface and use a test double until the actual implementation is available.

## 16.5 Worker lifecycle

Implement:

- Redis connection.
- Asynq server setup.
- Queue configuration.
- Handler registration.
- Periodic task registration.
- Graceful shutdown.
- Structured error handling.

Do not modify Docker Compose, Makefile, or shared environment files.

---

# 17. API AND CODE CONVENTIONS

Follow the Master Integration Plan.

## 17.1 JSON response format

Success:

```json
{
  "status": "success",
  "data": {}
}
```

Paginated responses:

```json
{
  "status": "success",
  "data": [],
  "meta": {
    "page": 1,
    "per_page": 25,
    "total": 100
  }
}
```

Error:

```json
{
  "status": "error",
  "error": {
    "code": "MACHINE_READABLE_ERROR_CODE",
    "message": "Human-readable explanation",
    "details": {}
  }
}
```

## 17.2 Monetary values

Use:

```text
NUMERIC(12,2)
```

Never use floating-point arithmetic for monetary values.

## 17.3 Dates and timestamps

Calendar dates:

```text
YYYY-MM-DD
```

Timestamps:

- UTC with an explicit offset, or
- Africa/Nairobi with `+03:00`.

Do not assume the server's local timezone.

## 17.4 Phone numbers

Normalize phone numbers to E.164 at API boundaries.

## 17.5 Go quality

Follow idiomatic Go.

Use:

- `context.Context`.
- Explicit dependency injection.
- Appropriate error wrapping.
- Small interfaces.
- Structured logging.
- Clear package boundaries.
- Transactional database operations.
- Deterministic tests.

Do not create unnecessarily large service methods.

---

# 18. TESTING STRATEGY

Testing is part of implementation, not a final optional activity.

Every task must have tests appropriate to its functionality.

## 18.1 Billing tests

Cover:

- Fee structure creation.
- Invalid fee structure rejection.
- Term charging.
- Duplicate charge prevention.
- New student billing hook.
- Correct term due dates.
- Ledger calculations.
- Overpayment.
- Credit carry-forward.
- Credit application.
- Mid-year changes.
- Missing note rejection.
- Targeted notifications.
- Term start-date changes.

## 18.2 M-Pesa tests

Cover:

- STK request validation.
- Guardian authorization.
- Any-positive-amount acceptance.
- Chat phone-change freeze.
- Internal AI confirmation requirement.
- Daraja callback processing.
- Duplicate callback protection.
- C2B matching.
- Cross-school admission-number isolation.
- Unmatched payment creation.
- Manual payment resolution.
- Overpayment credit.

## 18.3 WhatsApp tests

Cover:

- Unknown admission number.
- Zero OTP dispatch for unknown students.
- Masked guardian numbers.
- Multiple guardian selection.
- OTP delivered only to selected guardian.
- OTP verification.
- One-minute cooldown.
- Sixth-request blocking.
- Twenty-four-hour sender block.
- Guardian account remaining unblocked.
- Six-hour session expiry.
- Cross-school bot isolation.
- Invalid session handling.

## 18.4 Notification tests

Cover:

- Notification creation.
- Notification listing.
- Read status.
- User ownership.
- Tenant isolation.
- SMS sandbox.
- SMS provider errors.
- Retry behaviour.

## 18.5 Internal API tests

Cover:

- Missing service token.
- Invalid service token.
- Missing user context.
- Invalid HMAC signature.
- Expired context.
- Malformed payload.
- Invalid user context.
- Cross-school student access.
- Guardian-child ownership.
- Unauthorized STK Push.
- Missing parent confirmation.
- AI audit logging.

## 18.6 Worker tests

Cover:

- Task creation.
- Task serialization.
- SMS processing.
- Retry behaviour.
- STK polling.
- Cleanup scheduling.
- Invalid payload handling.
- Graceful shutdown where testable.

## 18.7 Race detection

Run:

```bash
cd backend && go test -race ./...
```

Pay particular attention to:

- Concurrent payment callbacks.
- Duplicate payment processing.
- Concurrent credit application.
- OTP request rate limiting.
- Session transitions.

---

# 19. IMPLEMENTATION EXECUTION RULES

This section governs how you must actually work.

## 19.1 Work one task at a time

Once the implementation plan is complete:

Select the first incomplete task.

Implement only that task and its immediate dependencies.

Do not jump to unrelated features.

## 19.2 Required workflow for every task

For every task, follow this exact cycle:

**Step 1 – Explain the task**

Briefly state:

- What you are implementing.
- Why it is necessary.
- Which files will be affected.
- What must already exist.

**Step 2 – Inspect**

Read the relevant existing code and understand its conventions.

**Step 3 – Implement**

Make the smallest coherent implementation that satisfies the task.

**Step 4 – Format**

Run Go formatting on the affected files.

**Step 5 – Test**

Run the relevant tests.

**Step 6 – Analyse**

If tests fail:

- Read the actual error.
- Identify the underlying cause.
- Correct the implementation.
- Rerun the tests.

Do not hide or suppress test failures.

**Step 7 – Verify**

Confirm:

- The task works.
- No unrelated files were modified.
- Existing functionality remains intact.
- The task's acceptance conditions are met.

**Step 8 – Update progress**

Mark the task completed in your implementation checklist.

**Step 9 – Continue**

Move to the next task only after the current task passes its verification gate.

## 19.3 No fake progress

Never mark a task complete simply because:

- The code was generated.
- A file was created.
- A function was written.
- The package compiled.
- A mock test passed.

Completion requires the verification conditions specified for that task.

## 19.4 Do not write all tests at the end

Tests must be introduced alongside implementation.

For example:

- Implement phone masking and test it immediately.
- Implement fee validation and test it immediately.
- Implement OTP cooldown and test it immediately.
- Implement HMAC verification and test it immediately.

## 19.5 Do not skip difficult requirements

If a requirement is technically difficult:

- Investigate it.
- Break it into smaller tasks.
- Implement it carefully.
- Test it.
- Document any genuine external blocker.

Do not silently omit functionality.

---

# 20. SECURITY AND FINANCIAL INVARIANTS

The following requirements are non-negotiable.

### Tenant security

- Every tenant-owned query must be properly scoped.
- RLS must be respected.
- School A must never access School B's data.
- WhatsApp admission lookup must be school-scoped.
- Internal AI queries must use verified tenant context.

### Guardian privacy

- Only linked guardians may access student information.
- WhatsApp must show masked guardian numbers.
- OTPs must only go to registered guardian numbers.
- OTPs must never be sent to the requesting WhatsApp number.

### Payment security

- Only authorized guardians may initiate payments.
- AI payments require explicit confirmation.
- Chat payments are blocked during the phone-change freeze.
- Duplicate callbacks must not duplicate financial credits.
- Payment references must be matched within the correct school.
- Unmatched payments must never be discarded.

### Financial integrity

- Never use floating-point arithmetic for money.
- Never create duplicate term charges.
- Never apply the same credit twice.
- Never allow negative current balances.
- Never lose excess payment credit.
- Preserve ledger history.
- Use database transactions for multi-record financial changes.

### AI security

- The AI service has no direct database access.
- Internal requests require service-token authentication.
- User context requires valid HMAC verification.
- Context expiry must be enforced.
- Guardian-child ownership must be checked by Go.
- AI actions must be logged.

---

# 21. RELEASE 1 SCOPE RESTRICTIONS

Do not accidentally implement Release 2 features.

Explicitly out of scope:

- Bursar payables.
- Supplier vouchers.
- Signatory workflows.
- Payment plans.
- Parent pledges.
- AI escalating default reminders.
- AI fuzzy payment reconciliation.
- SMS absence alerts.
- SMS sickbay alerts.

Keep the architecture extensible without implementing these features.

---

# 22. FINAL QUALITY ASSURANCE

After completing all planned tasks, perform a complete verification pass.

## 22.1 Go formatting

```bash
gofmt -w
```

Apply formatting only to your owned Go files.

## 22.2 Package compilation

Verify your assigned packages compile.

## 22.3 Full backend tests

Run:

```bash
cd backend
go test -v -race ./...
```

If another developer's unfinished module prevents the complete suite from running, distinguish that failure from failures in your own packages.

Do not modify their files to force the suite to pass.

## 22.4 Migration verification

Verify:

- Migration UP succeeds.
- Migration DOWN succeeds in a disposable test database.
- Migration UP succeeds again.
- RLS policies exist.
- Constraints exist.
- Indexes exist.
- No existing schema was unintentionally modified.

## 22.5 Acceptance criteria mapping

Create a verification matrix covering:

| Acceptance criterion | Required behaviour | Test | Result |
|---|---|---|---|
| AC-BE-C01 | WhatsApp OTP and session security | WhatsApp tests | |
| AC-BE-C02 | OTP cooldown and sender blocking | WhatsApp tests | |
| AC-BE-C03 | Fee charging and fee changes | Billing tests | |
| AC-BE-C04 | Payments, credits, matching and phone lock | Billing/M-Pesa tests | |
| AC-BE-C05 | Internal AI API security | Internal API tests | |

Do not mark criteria as passed without supporting test evidence.

## 22.6 Ownership verification

Run:

```bash
git status --short
git diff --name-only
git diff --cached --name-only
```

Check every changed path.

There must be no unauthorized changes outside your assigned directories.

## 22.7 Code review

Review your implementation as if you were reviewing another senior engineer's pull request.

Look for:

- Security vulnerabilities.
- Tenant isolation mistakes.
- Race conditions.
- Duplicate payment processing.
- Incorrect credit calculations.
- Missing transactions.
- Weak validation.
- Leaked secrets.
- Poor error handling.
- Inconsistent API responses.
- Missing tests.
- Unnecessary abstractions.
- Unimplemented requirements.

Fix identified issues before reporting completion.

---

# 23. FINAL DELIVERABLES

Your assignment is complete only when the following deliverables are ready.

### Database

- [ ] Migration UP.
- [ ] Migration DOWN.
- [ ] All required tables.
- [ ] Required ledger alterations.
- [ ] Tenant RLS policies.
- [ ] Appropriate constraints and indexes.

### Billing

- [ ] Fee structure loader.
- [ ] Generic billing periods.
- [ ] Automated term charging.
- [ ] Idempotent charging.
- [ ] Student billing hook.
- [ ] Student ledgers.
- [ ] Credit balance.
- [ ] Credit carry-forward.
- [ ] Fee balance endpoint.
- [ ] Ledger endpoint.
- [ ] Mid-year fee changes.
- [ ] Mandatory explanation notes.
- [ ] Term-date fee synchronization.

### M-Pesa

- [ ] Daraja configuration.
- [ ] STK Push.
- [ ] Any-positive-amount acceptance.
- [ ] Phone-change payment freeze.
- [ ] Explicit AI confirmation.
- [ ] Callback processing.
- [ ] Duplicate callback protection.
- [ ] Direct Paybill matching.
- [ ] Unmatched payments.
- [ ] Manual resolution.
- [ ] Payment receipts.

### WhatsApp

- [ ] Per-school bot routing.
- [ ] School-scoped admission lookup.
- [ ] Masked guardian phone numbers.
- [ ] Guardian selection.
- [ ] Secure OTP generation.
- [ ] OTP delivery.
- [ ] OTP cooldown.
- [ ] Sender-only 24-hour block.
- [ ] Six-hour verified sessions.
- [ ] Session expiry.
- [ ] AI forwarding.
- [ ] REST testing endpoints.

### Notifications

- [ ] In-app notification service.
- [ ] Notification listing.
- [ ] Mark-as-read.
- [ ] Africa's Talking dispatcher.
- [ ] Deterministic sandbox SMS.
- [ ] Integration interfaces.

### Background Jobs

- [ ] Asynq setup.
- [ ] Transactional SMS.
- [ ] STK polling.
- [ ] Unverified submission cleanup integration.
- [ ] Worker binary.
- [ ] Graceful shutdown.

### Internal AI API

- [ ] Service-token authentication.
- [ ] HMAC verification.
- [ ] Context signing.
- [ ] Context expiration.
- [ ] Tenant isolation.
- [ ] Guardian-child authorization.
- [ ] Seven internal endpoints.
- [ ] Frontend chat proxy.
- [ ] AI audit logging.

### Quality

- [ ] Automated tests.
- [ ] Race detection.
- [ ] Migration verification.
- [ ] Acceptance criteria verification.
- [ ] Full code review.
- [ ] Git ownership verification.
- [ ] Conventional commits.

---

# 24. REQUIRED FINAL REPORT

At the end of the assignment, provide a detailed report.

Include:

1. Git branch used.
2. Starting commit.
3. Final commit.
4. Features implemented.
5. Files created.
6. Files modified.
7. Database changes.
8. APIs implemented.
9. Integration interfaces exposed.
10. Security controls implemented.
11. Tests created.
12. Commands executed.
13. Test results.
14. Race detection results.
15. Migration results.
16. Acceptance criteria status.
17. Dependencies still awaiting other developers.
18. Outstanding integration work.
19. Known limitations.
20. Final Git status.

Distinguish clearly between:

- Fully implemented and tested.
- Implemented but awaiting external integration.
- Blocked by an unavailable dependency.
- Not implemented.

Never claim that the entire feature works end-to-end if you have only tested it using mocks.

---

# 25. MOST IMPORTANT EXECUTION INSTRUCTION

You must follow this sequence:

```text
READ DOCUMENTS
      |
      v
ANALYSE MASTER INTEGRATION
      |
      v
INSPECT REPOSITORY
      |
      v
INSPECT DATABASE AND EXISTING CODE
      |
      v
IDENTIFY DEPENDENCIES
      |
      v
CREATE FEATURE BRANCH
      |
      v
CREATE DETAILED IMPLEMENTATION PLAN
      |
      v
BREAK PLAN INTO SMALL TASKS
      |
      v
IMPLEMENT FIRST TASK
      |
      v
TEST FIRST TASK
      |
      v
VERIFY FIRST TASK
      |
      v
IMPLEMENT NEXT TASK
      |
      v
TEST AND VERIFY
      |
      v
REPEAT UNTIL ALL TASKS ARE COMPLETE
      |
      v
FULL QUALITY ASSURANCE
      |
      v
VERIFY FILE OWNERSHIP
      |
      v
FINAL REPORT
```

**Do not skip steps.**

**Do not implement everything at once.**

**Do not modify files outside Developer 05's ownership.**

**Do not depend on unfinished work from other developers.**

**Do not compromise security or financial integrity to make tests pass.**

**Do not claim functionality is complete without testing it.**

Your objective is to deliver a complete, independently testable Developer 05 backend implementation that can be integrated into Ikonex Academy Release 1 with minimal friction.

Begin with Phase Zero.

Your first action must be repository inspection and architectural analysis, NOT feature coding.

After completing Phase Zero, present the implementation plan and begin executing its first small task.

## Assistant

I’ll begin with Phase Zero: record the Git baseline, read both assignment documents in full, and inspect the backend and schema. Before writing feature code, I’ll report the findings, create the Developer 05 branch safely, and present the incremental implementation plan.

## Activity

    $ pwd && git status --short && git branch --show-current && git rev-parse HEAD && git branch -a && rg --files -g AGENTS.md -g '*developer-05*' -g '*MASTER-INTEGRATION*' -g 'go.mod' -g 'go.sum' -g '*.sql' -g 'Makefile' -g 'package.json'
    /home/fian/frihk/Ikonex-Academy
     M Docs/DECISIONS.md
    main
    b2b2e3b6f15e60f25947f04c39b8f6b12b5a268f
      Docs
      dev
    * main
      remotes/origin/Docs
      remotes/origin/dev
      remotes/origin/docs/add-agents-md
      remotes/origin/feat/ema-academics-b1
      remotes/origin/feat/ema-academics-b2
      remotes/origin/feat/ema-academics-b3
      remotes/origin/feat/finances
      remotes/origin/main
    migrations/000001_init_schema.down.sql
    migrations/000001_init_schema.up.sql
    Makefile
    backend/go.mod
    frontend/package.json
    tasks/MASTER-INTEGRATION-PLAN.md
    tasks/developer-05-backend-engineer-c.md
    AGENTS.md
    Docs/AGENTS.md
    ✓ • 32ms

## Activity

    $ wc -l tasks/developer-05-backend-engineer-c.md tasks/MASTER-INTEGRATION-PLAN.md && cat tasks/developer-05-backend-engineer-c.md && cat tasks/MASTER-INTEGRATION-PLAN.md
      252 tasks/developer-05-backend-engineer-c.md
      307 tasks/MASTER-INTEGRATION-PLAN.md
      559 total
    # Ikonex Academy – Release 1 Task Assignment
    ## Developer 5: Backend Engineer C (**Ian** — @frihk)
    
    - **Developer Name:** Ian (@frihk)
    - **GitHub Username:** `@frihk`
    - **Role:** Backend Engineer C (Go — Fees & Student Ledgers, Safaricom Daraja M-Pesa STK/C2B, Per-School WhatsApp Bot & 6h OTP Sessions, Notifications, Asynq Workers & Internal AI Bridge)
    - **Primary Directory Ownership:** `backend/internal/billing/`, `backend/internal/whatsapp/`, `backend/internal/notifications/`, `backend/internal/jobs/`, `backend/internal/internalapi/`, `backend/cmd/worker/`, and `migrations/000004_r1_billing_whatsapp_notifications.*`
    - **Release Scope:** Release 1 (Team-Loaded Fee Structures, Automated Term Charging, Ledger Credit Handling, Mid-Year Fee Changes, Daraja Sandbox STK Push & Direct Paybill Matching, Per-School WhatsApp Webhook & 6-Hour OTP Sessions, In-App & Transactional SMS Notifications, Asynq Background Jobs, and Go-to-AI Internal API)
    
    ---
    
    ## 1. Role and Objective
    
    As **Backend Engineer C (Ian — @frihk)**, you own the **Financial Ledger & M-Pesa Billing Engine**, the **External Communications & Per-School WhatsApp Channel**, the **Asynq Background Worker (`cmd/worker`)**, and the **Internal AI API Bridge (`internal/internalapi`)**.
    
    Your responsibilities enforce several of Ikonex Academy's most critical security and financial invariants (`Docs/Ikonex.md` Section 5 & Section 6.6–6.9):
    1. **No Bursar in Release 1 — Team-Loaded Fee Structures (`C46–C49`):** Fee structures are loaded as database records by the Ikonex team (`fee_periods`, `fee_structures`, `fee_structure_items`), charged automatically once at the start of the year divided into terms (`AC-20`), with due dates tied directly to each term's `start_date` (`C61, AC-22`).
    2. **Mid-Year Fee Changes Require a Note (`C45, AC-21`):** Every mid-year fee adjustment requires a mandatory explanation note (`fee_change_notes`), is audit-logged, and sends an in-app notification **only** to guardians of the affected students.
    3. **M-Pesa STK Push, Any-Amount Credit & Direct Paybill Matching (`C23, C24, C50, AC-23–AC-26`):**
       - Any positive payment amount is accepted; any overpayment beyond `current_balance` is held as a **credit** on the student's ledger (`AC-24`) and automatically applied to the next term charge (`A22`).
       - Direct Paybill (C2B) payments whose account number matches a valid `admission_number` within that school are auto-matched (`AC-25`); invalid/missing references are stored in `unmatched_payments` for the Ikonex team to resolve.
       - Chat-initiated payments must be blocked for **24 hours** after a guardian changes their registered phone number (`C27, AC-26`).
    4. **Per-School WhatsApp Bot & 6-Hour OTP Chat Sessions (`C12–C20, AC-08–AC-12`):**
       - Each school has its own WhatsApp Business number mapped to `school_id` (`school_whatsapp_bots`) so admission numbers never collide across schools (`AC-12`).
       - Any WhatsApp number can start a chat by entering a student's `admission_number` -> the bot displays **masked guardian numbers** (`07•• ••• 482`) -> sends the SMS OTP **only** to the chosen registered number -> establishes a **6-hour verified session** (`AC-08, AC-09, AC-11`).
       - Rate limits: 1 OTP/minute (`AC-10`, plus `A3` per admission number); `> 5 requests in 30 minutes` blocks the **requesting WhatsApp number** for 24 hours without blocking the guardian's account (`AC-10`).
    5. **AI Service Boundary Enforcement (`C21, AC-28`):** Your `internalapi` package is the **only** gateway through which Sos's Python AI service can read school/student facts or trigger a confirmed STK push, authenticated via `X-Internal-Service-Token` + short-lived HMAC-signed `X-User-Context`.
    
    ---
    
    ## 2. Assigned Features
    
    You own **3 cohesive, end-to-end backend features**:
    
    | Feature ID | Feature Name | Owned Go Packages | Why These Belong Together |
    |---|---|---|---|
    | **BE-C1** | **Team-Loaded Fee Structures, Automated Term Charging, Student Ledgers & Mid-Year Fee Changes** | `internal/billing` (`structures.go`, `ledgers.go`, `fee_changes.go`) | Manages the core accounting ledger, generic billing periods (`fee_periods`), term due-date synchronization, credit carry-forward, and mid-year fee adjustment notes. |
    | **BE-C2** | **Safaricom Daraja M-Pesa Integration (STK Push, Direct Paybill C2B Matching & Unmatched Resolution)** | `internal/billing` (`mpesa_stk.go`, `mpesa_c2b.go`, `unmatched.go`) | Connects external M-Pesa money flows directly to student ledgers, enforcing the 24h phone-change payment lock, overpayment credit rules, and unmatched C2B queue. |
    | **BE-C3** | **Per-School WhatsApp Bot, 6h OTP Sessions, Notifications (In-App + SMS), Asynq Workers & Internal AI API** | `internal/whatsapp`, `internal/notifications`, `internal/jobs`, `internal/internalapi`, `cmd/worker` | Unifies all external messaging channels (WhatsApp Cloud API + Africa's Talking SMS), background job execution, and the internal bridge to Sos's AI service. |
    
    ### Explicitly Out of Scope for Release 1 (Do Not Implement Yet)
    - **Bursar Payables, Supplier Vouchers & Signatory Workflows:** Release 2 (`C48`).
    - **Payment Plans, Parent Pledges & AI Escalating Default Reminders (`POST /api/v1/payment-plans/{id}/reminder-check`):** Release 2 (`C6–C8`).
    - **AI Fuzzy Reconciliation for Unmatched Payments:** Release 2 (in Release 1, the Ikonex team resolves unmatched payments manually via `POST /api/v1/payments/{id}/match`).
    - **SMS Event Alerts (Absences/Sickbay):** Release 2 (`C36, C58` — Release 1 event notifications are strictly **in-app only**; SMS is used only for login/WhatsApp/phone-change OTPs and the old-number security notice).
    
    ---
    
    ## 3. Detailed Feature Specifications
    
    ### Feature BE-C1: Fee Structures, Automated Term Charging, Student Ledgers & Mid-Year Fee Changes
    1. **Generic Billing Periods & Team-Loaded Fee Structures (`C46–C49`, Section 9 Proposed):**
       - Build `fee_periods` on generic billing periods (`period_type = 'term'` in Release 1, linked to `term_id` and `due_date = terms.start_date`) so Release 2 custom structures need no migration.
       - `POST /api/v1/billing/fee-structures`: Callable by the Ikonex team (`ikonex_admin`). Loads the fee structure and itemized lines (`fee_structure_items`) for a `(school_id, academic_year_id, grade_level_id, term_id)`.
    2. **Automated Start-of-Year / Term Charging (`C45, C46, C61, AC-20`):**
       - When a fee structure is loaded for an active year/term (or when a new student is admitted via Evans's `StudentChargeHook`), automatically generate per-term `fee_charges` and double-entry `ledger_entries` (`entry_type = 'charge'`) for every enrolled student in that grade level, divided by term (`AC-20`).
       - Each term charge's `due_date` equals `terms.start_date` (`C61`).
       - If a student has an existing `credit_balance > 0` on their `fee_ledgers` from a prior overpayment, automatically apply that credit against the new term's charge (`A22`).
    3. **Mid-Year Fee Changes & Term Start-Date Shifts (`C45, C61, AC-21, AC-22`):**
       - `POST /api/v1/fee-changes`: Callable by the Ikonex team on the school's written request (`A15`).
       - **Mandatory Note Rule (`AC-21`):** Request must include a non-empty `note` explaining the reason for the fee change. If `note` is empty, reject with `422 Unprocessable Entity` (`FEE_CHANGE_NOTE_REQUIRED`).
       - Updates the affected students' `fee_charges` / `fee_ledgers`, writes a row to `fee_change_notes` and `audit_log` (`AC-29`), and sends an in-app notification **only** to guardians of the affected students (`AC-21`).
       - Expose `OnTermStartDateChanged(ctx, tx, schoolID, termID, newStartDate, actorUserID, note)` for Ema's `PUT /api/v1/calendar` handler: updates `fee_periods.due_date` and `fee_charges.due_date` to `newStartDate`, records a `fee_change_notes` entry, and notifies affected parents (`AC-22`).
    4. **Parent & Team Fee Balance / Ledger Read (`C10, AC-24`):**
       - `GET /api/v1/students/{id}/fee-balance` and `GET /api/v1/students/{id}/ledger`:
       - Accessible to **any guardian linked to the student** in `student_guardians` (`C10`: all linked guardians see fees) or the Ikonex team.
       - Returns `opening_balance`, `billed_amount`, `paid_amount`, `current_balance` (`>= 0`), `credit_balance` (`>= 0`, representing excess paid over total billed, `AC-24`), `due_date`, item breakdown, and any `fee_change_notes`.
    
    ### Feature BE-C2: Safaricom Daraja M-Pesa Integration (STK Push, C2B Matching & Unmatched Queue)
    1. **STK Push Initiation (`POST /api/v1/payments/stk-push` — `C23, C24, C27, AC-23, AC-24, AC-26`):**
       - Callable by any linked guardian of `student_id` (`A2`) from the Parent Portal, or by Sos's AI service via `POST /api/v1/internal/payments/stk-push` after explicit parent confirmation.
       - **Validation & Business Rules:**
         1. **24-Hour Phone-Change Security Freeze (`C27, AC-26`):** If the request comes from a chat session (`source == "chat"`) and the guardian's `users.chat_payment_paused_until > NOW()`, reject immediately with `403 Forbidden` (`code: "CHAT_PAYMENT_PAUSED_PHONE_CHANGE"`).
         2. **Explicit Confirmation Check (`AC-23`):** If called via the internal AI endpoint, `parent_confirmed` must be `true`; otherwise reject with `422 Unprocessable Entity`.
         3. **Any Positive Amount Accepted (`C24, AC-23`):** Accept any `amount > 0`. Never reject an amount for being less than or greater than `current_balance`.
         4. **Account Reference:** Set Daraja `AccountReference` to the student's `admission_number` so callback matching is deterministic.
         5. Enqueue an Asynq background job (`jobs:mpesa:stk_poll`) to poll Daraja if the webhook callback does not arrive within 45 seconds.
    2. **Daraja STK Callback & Direct Paybill C2B Confirmation (`C24, C50, AC-24, AC-25`):**
       - `POST /api/v1/billing/mpesa/callback` (STK Push result) and `POST /api/v1/billing/mpesa/c2b-confirmation` (Direct Paybill payment):
       - Resolve `school_id` from the receiving Paybill/BusinessShortCode configured in `school_mpesa_configs`.
       - Look up `students` in that `school_id` where `UPPER(TRIM(admission_number)) = UPPER(TRIM(BillRefNumber))`.
       - **If Valid Student Match Found (`AC-25`):**
         - Insert row in `payments` (`matched_status = 'matched'`, `student_id = student.id`).
         - Insert row in `ledger_entries` (`entry_type = 'payment'`).
         - Update `fee_ledgers`: increase `paid_amount`. If `paid_amount > opening_balance + billed_amount`, set `current_balance = 0.00` and store the excess in `credit_balance = paid_amount - (opening_balance + billed_amount)` (`C24, AC-24`).
         - Create an in-app payment receipt notification for the student's guardians.
       - **If Admission Number Missing or Not Found in School (`C50, AC-25`):**
         - Insert row in `payments` (`matched_status = 'unmatched'`, `student_id = NULL`) and insert row in `unmatched_payments` with payer phone, payer name, raw `BillRefNumber`, amount, and M-Pesa receipt number.
    3. **Unmatched Payment Resolution by Ikonex Team (`C50, AC-25`):**
       - `GET /api/v1/payments/unmatched`: Lists all unresolved rows in `unmatched_payments` for the school.
       - `POST /api/v1/payments/{id}/match`: Callable by Ikonex team in Release 1. Assigns the payment to a valid `student_id`, credits the student's `fee_ledgers` (including `credit_balance` if overpaid), marks `unmatched_payments.resolved_at = NOW()`, and writes to `audit_log`.
    
    ### Feature BE-C3: Per-School WhatsApp Bot, 6h OTP Sessions, Notifications, Asynq Workers & Internal AI API
    1. **Per-School WhatsApp Bot & 6-Hour OTP Verification State Machine (`C12–C20, AC-08–AC-12`):**
       - Create package `backend/internal/whatsapp/`.
       - **Tenant Routing (`C16, AC-12`):** `POST /api/v1/whatsapp/webhook` reads the receiving WhatsApp Business phone number (`metadata.display_phone_number` / `phone_number_id`) and resolves `school_id` from `school_whatsapp_bots`. Every subsequent student lookup is strictly scoped to that `school_id` via `WithTenantTx`—a bot for School A can never match an admission number belonging to School B (`AC-12`).
       - **Conversational Identity & OTP Flow (also exposed via REST `POST /api/v1/chat/sessions` & `POST /api/v1/chat/sessions/{id}/otp` for testing):**
         - **State 1 (`AWAITING_ADMISSION_NUMBER`):** Any WhatsApp number (`sender_wa_number`) can start a chat (`AC-08`). Bot prompts for the student's admission number.
         - **State 2 (`AWAITING_GUARDIAN_SELECTION`):** Sender sends an admission number.
           - If no student in `school_id` matches that admission number (`C14`), return a generic message (*"We could not verify that admission number. Please check and try again or contact the school office."*)—send **no data and no OTP**.
           - If matched, fetch all linked guardians in `student_guardians`. If multiple guardians exist (`C15, AC-09`), reply with numbered **masked phone numbers** (e.g., `1. 07•• ••• 482`, `2. 07•• ••• 910`) and ask the sender to pick their registered number. (If only 1 guardian exists, display the masked number and prompt to confirm sending OTP).
         - **State 3 (`AWAITING_OTP`):**
           - Enforce OTP Rate Limits (`C18–C20, A3, AC-10`):
             1. If `sender_wa_number` is currently blocked in `sender_rate_blocks` (`blocked_until > NOW()`), refuse immediately.
             2. If an OTP was requested within the last **60 seconds** by this `sender_wa_number` OR for this `(school_id, admission_number)` (`A3`), refuse the second request (`AC-10`).
             3. Count OTP requests from `sender_wa_number` in the last **30 minutes**. If this is the **6th request** (`> 5 in 30 minutes`), insert a 24-hour block in `sender_rate_blocks` for `sender_wa_number` (`blocked_until = NOW() + INTERVAL '24 hours'`) and refuse. **Crucial (`C20, AC-10`):** Never block the guardian's user account or registered phone number—only block the requesting `sender_wa_number`.
           - Send the 6-digit SMS OTP **strictly to the chosen guardian's registered phone number** in `users.phone_number` (`C12, C13, AC-08, AC-09`), never to `sender_wa_number`.
         - **State 4 (`VERIFIED` — 6-Hour Session `C13, AC-11`):**
           - When the sender replies with the valid OTP, mark `chat_sessions.status = 'verified'` and `expires_at = NOW() + INTERVAL '6 hours'`.
           - For any message arriving when `NOW() < expires_at`, sign `X-User-Context` (`channel = "whatsapp"`) and forward the message to Sos's AI service (`POST /api/v1/ai/parent-query`), then send the reply back via WhatsApp.
           - Once `NOW() >= expires_at` (`AC-11`), require a fresh OTP before processing any query.
    2. **In-App Notifications & Transactional SMS (`internal/notifications` — `C58`):**
       - `GET /api/v1/notifications` & `PATCH /api/v1/notifications/{id}/read`: Serves in-app notifications (`in_app_notifications`) to the logged-in user.
       - `SMSDispatcher`: Sends transactional SMS via Africa's Talking API (or deterministic local sandbox logger when `AFRICASTALKING_USERNAME=sandbox`) strictly for: (a) App login OTPs, (b) WhatsApp session OTPs, and (c) Old-number security alerts on phone change (`C12, C27, C58`).
    3. **Asynq Background Worker (`internal/jobs`, `cmd/worker/main.go`):**
       - Implements Redis/Asynq queues for:
         - `sms:transactional`: Async SMS delivery with retries.
         - `mpesa:stk_poll`: Polls Daraja STK push query API if callback is delayed.
         - `onboarding:purge_unverified`: Periodic task calling Evans's cleanup logic to delete rejected/unverified `registration_submissions` after the configured retention days (`AC-07`).
    4. **Internal AI API (`internal/internalapi` — `C21, AC-28`):**
       - Validates `X-Internal-Service-Token` and verifies the HMAC-SHA256 signature and `exp` of `X-User-Context`.
       - Sets tenant RLS (`WithTenantTx`) from `X-User-Context.school_id` and enforces guardian-child ownership (`student_guardians`) on every student lookup (`Docs/API_CONTRACTS.md`: *"Parent-facing endpoints additionally scope to children this parent is a guardian of"*).
       - Implements all 7 internal endpoints called by Sos (`GET /api/v1/internal/school-profile`, `GET /api/v1/internal/parents/{user_id}/children`, `GET /api/v1/internal/students/{id}/fee-balance`, `GET /api/v1/internal/students/{id}/attendance-summary`, `GET /api/v1/internal/students/{id}/marks-summary`, `POST /api/v1/internal/payments/stk-push`, `POST /api/v1/internal/ai/audit-log`) and the frontend chat proxy `POST /api/v1/chat/message`.
    
    ---
    
    ## 4. Technical Implementation
    
    ### 4.1 Database Migration Owned by Ian (`migrations/000004_r1_billing_whatsapp_notifications.up.sql`)
    Create `000004_r1_billing_whatsapp_notifications.up.sql` and `.down.sql` with all Release 1 Billing, WhatsApp, Notification, and AI Audit tables and RLS policies:
    - `school_mpesa_configs`: `school_id (PK/FK)`, `shortcode VARCHAR(30) UNIQUE NOT NULL`, `environment VARCHAR(20) DEFAULT 'sandbox'`, `passkey_encrypted TEXT`, `test_payment_verified BOOLEAN DEFAULT FALSE`, `updated_at`.
    - `fee_periods`: `id`, `school_id`, `academic_year_id`, `term_id`, `period_type VARCHAR(30) NOT NULL DEFAULT 'term'`, `name VARCHAR(100) NOT NULL`, `due_date DATE NOT NULL`, `created_at`, `UNIQUE(school_id, academic_year_id, term_id)`.
    - `fee_charges`: `id`, `school_id`, `student_id`, `fee_period_id`, `fee_structure_id`, `amount NUMERIC(12,2) NOT NULL`, `due_date DATE NOT NULL`, `created_at`, `UNIQUE(school_id, student_id, fee_period_id)`.
    - `fee_change_notes`: `id`, `school_id`, `term_id`, `fee_period_id`, `change_type VARCHAR(50) NOT NULL`, `amount_delta NUMERIC(12,2) DEFAULT 0`, `old_due_date DATE`, `new_due_date DATE`, `note TEXT NOT NULL`, `created_by UUID NOT NULL`, `affected_students_count INT NOT NULL`, `created_at`.
    - `ledger_entries`: `id`, `school_id`, `student_id`, `term_id`, `entry_type VARCHAR(30) NOT NULL CHECK (entry_type IN ('charge', 'payment', 'credit_applied', 'adjustment'))`, `amount NUMERIC(12,2) NOT NULL`, `reference VARCHAR(100)`, `description TEXT`, `created_at`.
    - Alter `fee_ledgers`: Add `credit_balance NUMERIC(12,2) NOT NULL DEFAULT 0.00`, `due_date DATE`.
    - `unmatched_payments`: `id`, `school_id`, `payment_id UUID NOT NULL REFERENCES payments(id) ON DELETE CASCADE`, `transaction_ref VARCHAR(100) NOT NULL`, `amount NUMERIC(12,2) NOT NULL`, `payer_phone VARCHAR(20)`, `payer_name VARCHAR(255)`, `raw_account_ref VARCHAR(100)`, `status VARCHAR(30) NOT NULL DEFAULT 'unmatched'`, `resolved_student_id UUID REFERENCES students(id)`, `resolved_by UUID REFERENCES users(id)`, `resolution_note TEXT`, `resolved_at TIMESTAMPTZ`, `created_at`.
    - `school_whatsapp_bots`: `id`, `school_id`, `phone_number_id VARCHAR(100) UNIQUE NOT NULL`, `display_phone_number VARCHAR(30) UNIQUE NOT NULL`, `is_active BOOLEAN DEFAULT TRUE`, `created_at`.
    - `chat_sessions`: `id`, `school_id`, `channel VARCHAR(20) NOT NULL CHECK (channel IN ('whatsapp', 'in_app'))`, `sender_wa_number VARCHAR(30)`, `student_id UUID REFERENCES students(id)`, `guardian_user_id UUID REFERENCES users(id)`, `state VARCHAR(40) NOT NULL DEFAULT 'AWAITING_ADMISSION_NUMBER'`, `verified_at TIMESTAMPTZ`, `expires_at TIMESTAMPTZ`, `lookup_count INT NOT NULL DEFAULT 0`, `payment_init_count INT NOT NULL DEFAULT 0`, `created_at`, `updated_at`.
    - `in_app_notifications`: `id`, `school_id`, `recipient_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE`, `category VARCHAR(50) NOT NULL`, `title VARCHAR(255) NOT NULL`, `body TEXT NOT NULL`, `metadata JSONB`, `is_read BOOLEAN NOT NULL DEFAULT FALSE`, `read_at TIMESTAMPTZ`, `created_at`.
    - `ai_action_logs`: `id`, `school_id`, `user_id UUID`, `session_id UUID`, `channel VARCHAR(20) NOT NULL`, `feature VARCHAR(50) NOT NULL`, `intent VARCHAR(50)`, `endpoints_read JSONB`, `action_taken VARCHAR(100)`, `metadata JSONB`, `created_at`.
    - Apply `ENABLE ROW LEVEL SECURITY`, `FORCE ROW LEVEL SECURITY`, and `tenant_isolation` policies on all tenant tables above.
    
    ### 4.2 Owned Go Files & Package Structure
    ```text
    backend/
    ├── cmd/
    │   └── worker/
    │       └── main.go                      # Asynq worker binary entry point
    └── internal/
        ├── billing/
        │   ├── billing.go                   # Service struct, RegisterRoutes(mux, deps), cross-module hooks
        │   ├── structures.go                # Team fee structure loader & auto start-of-year/term charging
        │   ├── ledgers.go                   # Student balance, credit_balance carry-forward, ledger history
        │   ├── fee_changes.go               # Mid-year fee changes (note required) & term start-date hook
        │   ├── mpesa_stk.go                 # Daraja STK push, 24h phone-change lock check, callback handler
        │   ├── mpesa_c2b.go                 # Direct Paybill C2B auto-matching by school + admission number
        │   ├── unmatched.go                 # Unmatched payments list & team manual match resolution
        │   └── billing_test.go
        ├── whatsapp/
        │   ├── whatsapp.go                  # Webhook handler, school_whatsapp_bots lookup, session state machine
        │   ├── masking.go                   # Phone masking helper ("+254712345678" -> "07•• ••• 678")
        │   └── whatsapp_test.go
        ├── notifications/
        │   ├── notifications.go             # In-app notifications CRUD + Africa's Talking SMS client
        │   └── notifications_test.go
        ├── jobs/
        │   └── jobs.go                      # Asynq task definitions & handlers (SMS, STK poll, purge)
        └── internalapi/
            ├── internalapi.go               # Service token + HMAC X-User-Context middleware & internal routes
            ├── context_signer.go            # Signs X-User-Context for outbound calls to Sos's AI service
            └── internalapi_test.go
    ```
    
    ---
    
    ## 5. Exact Deliverables
    
    1. **Automated Term Charging & Credit Carry-Forward (`AC-20`, `AC-24`):** Loading a school's fee structure charges enrolled students once at the start of the year divided into terms (`AC-20`), sets due dates to `terms.start_date`, and holds any overpayment as `credit_balance` on `fee_ledgers` (`AC-24`).
    2. **Mid-Year Fee Changes & Term-Date Due-Date Shifts (`AC-21`, `AC-22`):**
       - `POST /api/v1/fee-changes` without a `note` fails with `422`; with a `note`, it updates affected ledgers, writes to `fee_change_notes` and `audit_log`, and notifies **only** affected parents (`AC-21`).
       - Invoking `OnTermStartDateChanged` shifts term fee due dates and sends the fee-change notice (`AC-22`).
    3. **STK Push, 24h Phone-Change Lock & Direct Paybill Matching (`AC-23`, `AC-25`, `AC-26`):**
       - STK push accepts any positive amount (`AC-23`) and blocks chat-initiated payments for 24 hours after a guardian phone number change (`AC-26`).
       - Direct Paybill C2B payments with a valid school admission number auto-credit the student's ledger (`AC-25`); invalid/missing references land in `unmatched_payments` and can be resolved via `POST /api/v1/payments/{id}/match`.
    4. **Per-School WhatsApp Bot & 6-Hour OTP Sessions (`AC-08` to `AC-12`):**
       - Unregistered WhatsApp numbers can start a chat; multiple guardians are shown as masked numbers (`07•• ••• 482`), and the SMS OTP is sent **only** to the selected registered number (`AC-08`, `AC-09`).
       - Rate limiter refuses a 2nd OTP within 1 minute and blocks the requesting WhatsApp number for 24 hours on the 6th request in 30 minutes without locking the guardian's account (`AC-10`).
       - Verified sessions expire after 6 hours (`AC-11`), and School A's bot cannot resolve School B's students (`AC-12`).
    5. **Internal AI API & Action Logging (`AC-28`):** Every call from Sos's AI service is authenticated with `X-Internal-Service-Token` + HMAC-signed `X-User-Context`, scoped to the parent's own children, and logged in `ai_action_logs`.
    
    ---
    
    ## 6. Acceptance Criteria
    
    - [ ] **AC-BE-C01 (aligns with AC-08, AC-09, AC-11, AC-12):** WhatsApp session state machine masks guardian phones, sends OTP only to the chosen registered number, expires after 6 hours, and isolates admission numbers strictly by `school_whatsapp_bots.school_id`.
    - [ ] **AC-BE-C02 (aligns with AC-10):** Requesting a 2nd WhatsApp OTP within 60s returns `429`; the 6th request within 30 minutes blocks `sender_wa_number` for 24h while the guardian can still log in from the web app or their own number.
    - [ ] **AC-BE-C03 (aligns with AC-20, AC-21, AC-22):** Term charges are created once per year divided into terms; `POST /api/v1/fee-changes` rejects missing notes (`422`) and notifies only affected guardians when valid.
    - [ ] **AC-BE-C04 (aligns with AC-23, AC-24, AC-25, AC-26):** Overpaying a fee balance records `current_balance = 0` and `credit_balance = excess`; valid C2B Paybill references auto-match while invalid references create `unmatched_payments` rows; chat payments within 24h of a phone change return `403 CHAT_PAYMENT_PAUSED_PHONE_CHANGE`.
    - [ ] **AC-BE-C05 (aligns with AC-28):** `internalapi` rejects requests with missing/invalid `X-Internal-Service-Token` or expired `X-User-Context`, rejects parent queries for a `student_id` not linked to `X-User-Context.user_id`, and writes to `ai_action_logs`.
    
    ---
    
    ## 7. Testing Requirements
    
    Write automated Go tests across your packages:
    1. **Billing & M-Pesa Tests (`internal/billing/billing_test.go`):**
       - Start-of-year term charging (`AC-20`), overpayment credit calculation (`AC-24`), and credit application to next term (`A22`).
       - Mid-year fee change missing note rejection (`422`) and targeted parent notification (`AC-21`).
       - Term start-date change shifting fee due dates (`AC-22`).
       - Direct Paybill C2B valid admission number auto-match vs. invalid reference landing in `unmatched_payments` (`AC-25`).
       - 24h phone-change chat payment block (`AC-26`).
    2. **WhatsApp Session & OTP Rate Limit Tests (`internal/whatsapp/whatsapp_test.go`):**
       - Unknown admission number returns generic reply and sends zero OTPs (`C14`).
       - Multiple guardians return masked numbers (`07•• ••• 482`) and OTP dispatches only to chosen registered number (`AC-08`, `AC-09`).
       - 1-minute cooldown and 6-in-30-minutes 24h sender block leaving guardian account unblocked (`AC-10`).
       - 6-hour session expiry (`AC-11`) and cross-school bot isolation (`AC-12`).
    3. **Internal AI API Security Tests (`internal/internalapi/internalapi_test.go`):**
       - HMAC signature verification on `X-User-Context`, guardian-child ownership check (`403` when Parent A queries Parent B's child), and AI audit logging (`AC-28`).
    
    ---
    
    ## 8. Integration Responsibilities
    
    - **Depends on Developer 3 (Evans — @eojuma – Backend Engineer A):**
      - Uses `database.WithTenantTx`, `database.AuditLogger`, and `middleware` helpers.
      - Exposes `notifications.SMSDispatcher` and `billing.StudentChargeHook` interfaces for Evans's (@eojuma) `auth` and `onboarding` packages.
    - **Integrates with Developer 4 (Ema — @ejao-000 – Backend Engineer B):**
      - Exposes `notifications.Service.CreateInAppNotification` and `billing.Service.OnTermStartDateChanged` for Ema's (@ejao-000) `academics` package.
      - Calls Ema's (@ejao-000) `academics.Service.GetStudentAttendanceSummary` and `GetStudentMarksSummary` inside `internalapi` to serve Sos's (@sospeter-57) AI service.
    - **Integrates with Developer 1 (Sos — @sospeter-57 – AI Engineer):**
      - Signs `X-User-Context` (`context_signer.go`) and forwards verified WhatsApp and in-app chat messages to Sos's (@sospeter-57) `POST /api/v1/ai/parent-query`, and serves `/api/v1/internal/*` read/STK endpoints to Sos (@sospeter-57).
    
    ---
    
    ## 9. File Ownership and Conflict Prevention
    
    | Category | Paths | Rule |
    |---|---|---|
    | **Exclusively Owned by Ian (@frihk, Dev 5)** | `backend/cmd/worker/main.go`, `backend/internal/billing/**`, `backend/internal/whatsapp/**`, `backend/internal/notifications/**`, `backend/internal/jobs/**`, `backend/internal/internalapi/**`, `migrations/000004_r1_billing_whatsapp_notifications.*` | Full read/write ownership. No other developer edits these packages or migration files. |
    | **Shared Files (Coordinate via PR)** | `backend/cmd/server/main.go` (owned by Evans — @eojuma; Ian exports `RegisterRoutes` in `billing`, `whatsapp`, `notifications`, `internalapi`), `Docs/API_CONTRACTS.md` | Coordinate route mounting with Evans (@eojuma). |
    | **Strictly Do Not Modify** | `backend/internal/auth/**`, `backend/internal/onboarding/**`, `backend/internal/middleware/**` (Evans — @eojuma), `backend/internal/academics/**` (Ema — @ejao-000), `ai_service/**` (Sos — @sospeter-57), `frontend/**` (Ashy — @Ashomondi) | Owned by other developers. |
    
    ---
    
    ## 10. Definition of Done
    
    - [ ] `migrations/000004_r1_billing_whatsapp_notifications.up.sql` and `.down.sql` apply and roll back cleanly with `FORCE ROW LEVEL SECURITY` on all tenant tables.
    - [ ] `billing`, `whatsapp`, `notifications`, `jobs`, `internalapi`, and `cmd/worker` compile and pass `go test -v ./...`.
    - [ ] Per-school WhatsApp webhook, masked guardian OTP verification, 6-hour session expiry, and sender-only 24h rate blocking work end-to-end.
    - [ ] Daraja sandbox STK push, overpayment ledger credit, C2B auto-matching, and unmatched queue resolution work end-to-end.
    - [ ] Internal AI API enforces service token + signed user context and logs every AI action in `ai_action_logs`.
    # Ikonex Academy – Release 1 Master Integration Plan
    
    - **Document Version:** 1.0 (Release 1 Official Architecture & Execution Plan)
    - **Authoritative Sources:** `Docs/Ikonex.md`, `Docs/REQUIREMENTS.md`, `Docs/ARCHITECTURE.md`, `Docs/API_CONTRACTS.md`, `Docs/DECISIONS.md`, `Docs/SECURITY.md`, `Docs/AGENTS.md`
    - **Release 1 Scope:** Synthetic-data proof of concept (`C51, C52`), English only (`C56`), running against Safaricom Daraja sandbox and test WhatsApp numbers (`A17`). Passes when **Onboarding works**, **Parents can log in (App SMS OTP & WhatsApp 6h OTP session)**, **Badges work stress-free**, and **all Release 1 features work cleanly end-to-end (`C53, C54`)**.
    
    ---
    
    ## 1. Overview of the Five Developer Assignments
    
    | Developer ID | Name | GitHub Tag | Role | Assigned Scope Summary | Task Document |
    |---|---|---|---|---|---|
    | **Developer 1** | **Sos** | `@sospeter-57` | AI Engineer (Python / FastAPI / Mistral) | Owns `ai_service/**`: Service-token + HMAC `X-User-Context` verifier, Parent Assistant (In-App & WhatsApp minimal-fact Q&A), Multi-turn M-Pesa Chat Payment Confirmation flow, CBC Report Card Comment & Re-Release Note Drafter, and AI Audit Logging. Zero direct DB access. | `tasks/developer-01-ai-engineer.md` |
    | **Developer 2** | **Ashy** | `@Ashomondi` | Frontend Engineer (Next.js 14 / TypeScript / Tailwind PWA) | Owns `frontend/**`: Tailwind Design System, SMS OTP Login & Membership Switcher, Principal 5-Step Setup Wizard, Staff & 12h Parent QR Registration + In-Person Verification UI, Dean of Studies Workspace, Teacher Portal (Roll-Call, Marks, Hard-Gated Release/Reopen, AI Comment Drafter), and Parent Portal (Results, CBC Graph, Fee Ledger, STK Push, In-App AI Chat, 2FA Phone Change, R2 Placeholders). | `tasks/developer-02-frontend-engineer.md` |
    | **Developer 3** | **Evans** | `@eojuma` | Backend Engineer A (Go — Core, Auth & Onboarding) | Owns `cmd/server`, `internal/config`, `internal/database`, `internal/middleware`, `internal/auth`, `internal/onboarding`, `internal/consent`, and `migrations/000001`–`000002`: PostgreSQL RLS (`WithTenantTx`), SMS OTP Login, Scoped/Time-Bound Badge Engine, Audit Logger, School Provisioning, Setup Wizard, Staff/Parent QR Codes, In-Person Verification & Document Deletion, Verifier Delegation, Student Admission, and Guardian 2FA Phone Change. | `tasks/developer-03-backend-engineer-a.md` |
    | **Developer 4** | **Ema** | `@ejao-000` | Backend Engineer B (Go — CBC Academics & Results) | Owns `internal/academics` and `migrations/000003`: Dean of Studies Teacher Assignments, Timetables, School Calendar & Term-Date Fee Sync, Versioned CBC Grading Scales, Daily Class-Teacher Roll-Call Attendance, Anonymised Class Means, Subject Marks Entry & Reassignment, Hard-Gated Results Release/Reopen/Re-Release Snapshot State Machine, AI Drafter Proxy, and Student CBC Performance Graph. | `tasks/developer-04-backend-engineer-b.md` |
    | **Developer 5** | **Ian** | `@frihk` | Backend Engineer C (Go — Billing, WhatsApp, Jobs & Internal AI API) | Owns `cmd/worker`, `internal/billing`, `internal/whatsapp`, `internal/notifications`, `internal/jobs`, `internal/internalapi`, and `migrations/000004`: Team-Loaded Fee Structures, Automated Term Charging, Ledger Credit Balances, Mid-Year Fee Changes, Daraja STK Push & Direct C2B Paybill Auto-Matching, Per-School WhatsApp Bot & 6h OTP Sessions, In-App + Transactional SMS Notifications, Asynq Workers, and Internal AI API. | `tasks/developer-05-backend-engineer-c.md` |
    
    ---
    
    ## 2. Complete Release 1 Feature Ownership Matrix
    
    Every Release 1 feature and acceptance criterion (`AC-01` to `AC-29` in `Docs/Ikonex.md` Section 14) has a single accountable backend/AI owner and frontend integration owner:
    
    | Area | Feature / Requirement | Acceptance Criteria (`Docs/Ikonex.md`) | Primary Backend / AI Owner | Frontend Owner |
    |---|---|---|---|---|
    | **Foundation** | Multi-school PostgreSQL RLS isolation (`SET LOCAL app.school_id`) | `AC-27` | **Evans** (`@eojuma` — `internal/database`) | — |
    | **Foundation** | Roles & scoped/time-bound Badges engine (`class_teacher`, `delegated_verifier`) | `C53`, `AC-05`, `AC-14` | **Evans** (`@eojuma` — `internal/middleware`) | **Ashy** (`@Ashomondi` — `auth-context.tsx`) |
    | **Foundation** | Centralized Audit Trail of all sensitive actions | `AC-29` | **Evans** (`@eojuma` — `internal/database/audit.go`) | — |
    | **Foundation** | Phone + SMS OTP App Login & Membership Switcher | `AC-01` | **Evans** (`@eojuma` — `internal/auth`) | **Ashy** (`@Ashomondi` — `/login`) |
    | **Foundation** | Consent Centre Schema (`biometric_attendance`, `photo`, `sms`, `whatsapp`) | Section 6.10 | **Evans** (`@eojuma` — `internal/consent`) | **Ashy** (`@Ashomondi` — `/parent/settings`) |
    | **Onboarding** | Ikonex Team School + Principal creation & Required School Contact Profile | `AC-01`, `C37`, `C57` | **Evans** (`@eojuma` — `internal/onboarding`) | **Ashy** (`@Ashomondi` — `/admin/platform`) |
    | **Onboarding** | Principal 5-Step Setup Wizard (Years, Terms, Grades 7–12, Classes, Dorms, Grading Scale) | `AC-01`, `C38` | **Evans** (`@eojuma` — `internal/onboarding`) | **Ashy** (`@Ashomondi` — `/admin/wizard`) |
    | **Onboarding** | Staff QR Code (8am–6pm adjustable window, revocation, document upload) | `AC-02`, `C62` | **Evans** (`@eojuma` — `internal/onboarding`) | **Ashy** (`@Ashomondi` — `/admin/onboarding`, `/register/[token]`) |
    | **Onboarding** | Parent 12h Multi-Use QR Code & Revocation | `AC-03`, `C28` | **Evans** (`@eojuma` — `internal/onboarding`) | **Ashy** (`@Ashomondi` — `/admin/onboarding`, `/register/[token]`) |
    | **Onboarding** | Pending Submission Isolation (Zero child data access until verified) | `AC-04`, `A8` | **Evans** (`@eojuma` — `internal/onboarding`) | **Ashy** (`@Ashomondi` — `/register/[token]`) |
    | **Onboarding** | Time-Bound Verifier Delegation to Staff (Expires with 12h Parent QR) | `AC-05`, `C28`, `C29` | **Evans** (`@eojuma` — `internal/onboarding`) | **Ashy** (`@Ashomondi` — `/admin/onboarding`) |
    | **Onboarding** | In-Person Document Verification, Immediate Document Deletion & Auto-Purge of Unverified Submissions | `AC-06`, `AC-07`, `C63`, `C64` | **Evans** (`@eojuma` — `internal/onboarding`) + **Ian** (`@frihk` — `internal/jobs` cron) | **Ashy** (`@Ashomondi` — `/admin/verifications`) |
    | **Onboarding** | Dual-Route Student Admission (Class-Teacher Manual & Opening-Day QR Duplicate Flagging) + Re-Enrolment | `C39–C41`, `A12` | **Evans** (`@eojuma` — `internal/onboarding`) | **Ashy** (`@Ashomondi` — `/admin/dean`, `/admin/verifications`) |
    | **Onboarding** | Guardian 2FA Registered Phone Number Change (OTP to new number, SMS to old number, 24h chat payment pause) | `AC-26`, `C25`, `C27` | **Evans** (`@eojuma` — `internal/onboarding`) | **Ashy** (`@Ashomondi` — `/parent/settings`) |
    | **Academics** | Dean of Studies Teacher Class/Subject Assignments & Marks Reassignment | `AC-15`, `AC-17`, `C32`, `C59` | **Ema** (`@ejao-000` — `internal/academics`) | **Ashy** (`@Ashomondi` — `/admin/dean`) |
    | **Academics** | Timetables & School Calendar (including Post-Wizard Term Date Edits & Fee Due-Date Shift) | `AC-22`, `C59–C61` | **Ema** (`@ejao-000` — `internal/academics`) + **Ian** (`@frihk` — `billing` hook) | **Ashy** (`@Ashomondi` — `/admin/dean`, `/teacher`, `/parent`) |
    | **Academics** | Daily Class-Teacher Roll-Call Attendance & In-App Absence Notifications | `AC-14`, `C35`, `C36` | **Ema** (`@ejao-000` — `internal/academics`) | **Ashy** (`@Ashomondi` — `/teacher/attendance`) |
    | **Academics** | Scoped Class Roster vs. Anonymised Other-Class Means (Never student names) | Section 4 | **Ema** (`@ejao-000` — `internal/academics`) | **Ashy** (`@Ashomondi` — `/teacher`) |
    | **Academics** | Subject Teacher Marks Entry & Hard-Gated Class Results Release (100% marks required) | `AC-15`, `AC-16`, `C30`, `C31` | **Ema** (`@ejao-000` — `internal/academics`) | **Ashy** (`@Ashomondi` — `/teacher/marks`, `/teacher/results`) |
    | **Academics** | Results Reopen & Re-Release State Machine (Snapshot preserved while reopened, mandatory teacher note, only affected parents notified) | `AC-18`, `C33`, `C34` | **Ema** (`@ejao-000` — `internal/academics`) | **Ashy** (`@Ashomondi` — `/teacher/results`, `/parent/results`) |
    | **Academics** | Versioned Per-Term CBC Grading Scale & Multi-Term CBC Performance Graph (Zero class rank) | `AC-19`, `C42`, `C43` | **Ema** (`@ejao-000` — `internal/academics`) | **Ashy** (`@Ashomondi` — `/parent/results` — `CbcPerformanceChart.tsx`) |
    | **Fees & M-Pesa** | Team-Loaded Fee Structures & Automated Start-of-Year Term Charging | `AC-20`, `C45–C49` | **Ian** (`@frihk` — `internal/billing`) | **Ashy** (`@Ashomondi` — `/admin/platform`, `/parent/fees`) |
    | **Fees & M-Pesa** | Mid-Year Fee Changes (Mandatory note, audit log, only affected parents notified) | `AC-21`, `C45` | **Ian** (`@frihk` — `internal/billing`) | **Ashy** (`@Ashomondi` — `/admin/platform`, `/parent/fees`) |
    | **Fees & M-Pesa** | M-Pesa Daraja STK Push, Any-Amount Acceptance & Overpayment Ledger Credit | `AC-23`, `AC-24`, `C23`, `C24` | **Ian** (`@frihk` — `internal/billing`) | **Ashy** (`@Ashomondi` — `/parent/fees`) |
    | **Fees & M-Pesa** | Direct Paybill (C2B) Auto-Matching by Admission Number & Unmatched Payments Queue | `AC-25`, `C50` | **Ian** (`@frihk` — `internal/billing`) | **Ashy** (`@Ashomondi` — `/admin/platform`) |
    | **WhatsApp & Notifications** | Per-School WhatsApp Bot Routing, Masked Guardian Numbers, 6h OTP Sessions & Sender-Only 24h Rate Block | `AC-08`, `AC-09`, `AC-10`, `AC-11`, `AC-12` | **Ian** (`@frihk` — `internal/whatsapp`) | — |
    | **WhatsApp & Notifications** | In-App Event Notifications & Transactional SMS Dispatcher (Africa's Talking) | `C58`, `C12`, `C27` | **Ian** (`@frihk` — `internal/notifications`, `internal/jobs`) | **Ashy** (`@Ashomondi` — `NotificationDrawer.tsx`) |
    | **AI Service** | Internal AI API Gateway (`X-Internal-Service-Token` + signed `X-User-Context`) & AI Action Log | `AC-28`, `C21` | **Ian** (`@frihk` — `internal/internalapi`) + **Sos** (`@sospeter-57` — `ai_service/app/security.py`) | — |
    | **AI Service** | Parent Assistant (In-App & WhatsApp Read-Only Q&A, WhatsApp Minimal Facts / No Results on WhatsApp, School Contact Fallback) | `AC-13`, `C22`, `C56`, `C57`, `C66` | **Sos** (`@sospeter-57` — `parent_assistant.py`) + **Ian** (`@frihk` — `internalapi`, `whatsapp`) | **Ashy** (`@Ashomondi` — `/parent/chat`) |
    | **AI Service** | Chat-Initiated M-Pesa Payment (Multi-Turn Slot Extraction & Mandatory Explicit "YES" Repeat-Back Confirmation) | `AC-23`, `AC-26`, `C23`, `A4` | **Sos** (`@sospeter-57` — `chat_payment.py`) + **Ian** (`@frihk` — `internalapi`, `billing`) | **Ashy** (`@Ashomondi` — `/parent/chat`) |
    | **AI Service** | CBC Report Card Comment & Re-Release Explanation Note Drafter (Teacher edits before publishing) | `C5`, `C30`, `C34` | **Sos** (`@sospeter-57` — `report_drafter.py`) + **Ema** (`@ejao-000` — `academics`) | **Ashy** (`@Ashomondi` — `/teacher/results`) |
    
    ---
    
    ## 3. API Contract Ownership and Shared Conventions
    
    To resolve the open API formatting items noted in `Docs/Ikonex.md` Section 10 and `Docs/API_CONTRACTS.md`, **all five developers must follow these exact conventions**:
    
    ### 3.1 Base URLs & Versioning
    - **Frontend -> Go Backend:** `http://localhost:8080/api/v1` (`NEXT_PUBLIC_API_URL`)
    - **Go Backend -> AI Service:** `http://ai_service:8000/api/v1/ai`
    - **AI Service -> Go Backend Internal API:** `http://backend:8080/api/v1/internal` (`GO_BACKEND_INTERNAL_URL`)
    
    ### 3.2 Standard JSON Response Envelope
    - **Success (`200 OK`, `201 Created`):**
      ```json
      {
        "status": "success",
        "data": { ... },
        "meta": {
          "page": 1,
          "per_page": 25,
          "total": 100
        }
      }
      ```
      *(Note: `meta` is required on paginated list endpoints and optional on single-object responses).*
    - **Error (`400`, `401`, `403`, `404`, `409`, `410`, `422`, `429`, `500`):**
      ```json
      {
        "status": "error",
        "error": {
          "code": "MACHINE_READABLE_UPPER_SNAKE_CODE",
          "message": "Clear human-readable explanation for UI display.",
          "details": {}
        }
      }
      ```
    
    ### 3.3 Date, Time, and Currency Conventions
    - **Timestamps:** Always ISO-8601 in East Africa Time (`Africa/Nairobi`, `+03:00`) or UTC with explicit offset, e.g. `"2026-11-05T14:30:00+03:00"`.
    - **Calendar Dates (Term dates, DOB, Attendance date, Due dates):** `YYYY-MM-DD` (`"2026-11-05"`).
    - **Currency (KES):** Sent in JSON as fixed 2-decimal strings (`"12500.00"`) or numeric values formatted to 2 decimal places (`NUMERIC(12,2)` in PostgreSQL; never float32).
    - **Phone Numbers:** Normalized to E.164 format (`"+254712345678"`) at the API boundary.
    
    ### 3.4 Go-to-AI Service Authentication Contract (**Ian** `@frihk`, **Ema** `@ejao-000`, **Sos** `@sospeter-57`)
    Every request between the Go Backend and `ai_service` (in both directions) must include:
    1. `X-Internal-Service-Token: <INTERNAL_SERVICE_TOKEN>`
    2. `X-User-Context: <base64url(json_payload)>.<hex(hmac_sha256(base64url_payload, INTERNAL_SERVICE_TOKEN))>`
       Where `json_payload` is:
       ```json
       {
         "user_id": "uuid",
         "school_id": "uuid",
         "membership_id": "uuid",
         "role": "parent",
         "channel": "whatsapp",
         "session_id": "uuid",
         "exp": 1762300000
       }
       ```
    
    ---
    
    ## 4. Database Ownership and Schema Coordination
    
    To prevent migration merge conflicts, database migrations are partitioned into **four numbered files**, each with a single owner:
    
    | Migration File | Owner | Tables Created / Altered |
    |---|---|---|
    | `migrations/000001_init_schema.up.sql` & `.down.sql` | **Evans** (`@eojuma`, Dev 3) | **Existing baseline (do not edit):** `schools`, `users`, `permissions`, `memberships`, `badges`, `badge_permissions`, `membership_badges`, `academic_years`, `terms`, `grade_levels`, `classes`, `dormitories`, `clubs`, `students`, `student_guardians`, `enrolments`, `consents`, `biometric_templates`, `staff_invites`, `fee_structures`, `fee_structure_items`, `fee_ledgers`, `payments`, `payment_plans`, `nurse_visits`, `library_books`, `book_loans`, `cbc_strands`, `syllabus_coverage`, `student_assessments`, `blog_posts`, `audit_log` + RLS loop. |
    | `migrations/000002_r1_core_onboarding.up.sql` & `.down.sql` | **Evans** (`@eojuma`, Dev 3) | `school_profile`, `otp_requests`, `sender_rate_blocks`, `parent_qr_codes`, `registration_submissions`, `verification_records`; alters `users` (`chat_payment_paused_until`) and `consents` (adds `'whatsapp'` purpose) + RLS policies. |
    | `migrations/000003_r1_academics.up.sql` & `.down.sql` | **Ema** (`@ejao-000`, Dev 4) | `subjects`, `subject_teacher_assignments`, `grading_scale_versions`, `timetable_slots`, `calendar_events`, `attendance_daily`, `student_marks`, `report_comments`, `result_releases`, `released_marks_snapshots` + RLS policies. |
    | `migrations/000004_r1_billing_whatsapp_notifications.up.sql` & `.down.sql` | **Ian** (`@frihk`, Dev 5) | `school_mpesa_configs`, `fee_periods`, `fee_charges`, `fee_change_notes`, `ledger_entries`, `unmatched_payments`, `school_whatsapp_bots`, `chat_sessions`, `in_app_notifications`, `ai_action_logs`; alters `fee_ledgers` (`credit_balance`, `due_date`) + RLS policies. |
    
    **Mandatory RLS Rule for All Migrations:** Every tenant-owned table created in `000002`, `000003`, and `000004` must include `school_id UUID NOT NULL REFERENCES schools(id) ON DELETE CASCADE` and execute:
    ```sql
    ALTER TABLE <table_name> ENABLE ROW LEVEL SECURITY;
    ALTER TABLE <table_name> FORCE ROW LEVEL SECURITY;
    CREATE POLICY <table_name>_tenant_isolation ON <table_name>
        FOR ALL USING (school_id = current_school_id())
        WITH CHECK (school_id = current_school_id());
    ```
    
    ---
    
    ## 5. Cross-Developer Dependencies
    
    ```mermaid
    flowchart TD
        subgraph Dev3_Evans["Dev 3: Evans @eojuma (Core, Auth, Onboarding)"]
            DB["internal/database (WithTenantTx + AuditLogger)"]
            MW["internal/middleware (Auth + Scoped Badges)"]
            AUTH["internal/auth (SMS OTP + JWT)"]
            ONB["internal/onboarding (Wizard, QR, Verify, Admit)"]
        end
    
        subgraph Dev4_Ema["Dev 4: Ema @ejao-000 (CBC Academics)"]
            ACAD["internal/academics (Assignments, Timetable, Calendar, Roll-Call, Marks, Release/Reopen, Graph)"]
        end
    
        subgraph Dev5_Ian["Dev 5: Ian @frihk (Billing, WhatsApp, Notifications, InternalAPI)"]
            BILL["internal/billing (Fee Structures, Ledgers, STK Push, C2B)"]
            WA["internal/whatsapp (Per-School Bot, 6h OTP Sessions)"]
            NOTIF["internal/notifications & internal/jobs (In-App + AT SMS + Asynq)"]
            IAPI["internal/internalapi (Permission-Checked AI Gateway)"]
        end
    
        subgraph Dev1_Sos["Dev 1: Sos @sospeter-57 (Python FastAPI AI Service)"]
            AIS["ai_service (Parent Assistant, Chat Payment Confirmation, Report Comment Drafter)"]
        end
    
        subgraph Dev2_Ashy["Dev 2: Ashy @Ashomondi (Next.js 14 PWA Frontend)"]
            FE["frontend (Login, Wizard, Onboarding, Dean, Teacher, Parent Portals)"]
        end
    
        MW --> DB
        AUTH --> DB
        ONB --> MW
        ACAD --> MW
        BILL --> MW
        WA --> DB
        IAPI --> MW
    
        AUTH -->|SMSDispatcher| NOTIF
        ONB -->|StudentChargeHook| BILL
        ACAD -->|OnTermStartDateChanged| BILL
        ACAD -->|CreateInAppNotification| NOTIF
        IAPI -->|Read Marks/Attendance| ACAD
        IAPI -->|Read Fees / Trigger STK| BILL
        WA -->|Verified Message| AIS
        ACAD -->|Draft Report Comment| AIS
        AIS -->|Read Facts / Confirmed STK / Audit| IAPI
    
        FE -->|REST /api/v1/auth & onboarding| ONB
        FE -->|REST /api/v1/classes, timetables, calendar, marks| ACAD
        FE -->|REST /api/v1/students/fee-balance, payments, chat, notifications| BILL
    ```
    
    ### Cross-Module Go Interfaces (Defined in Day 1 Scaffold)
    To allow **Evans** (`@eojuma`), **Ema** (`@ejao-000`), and **Ian** (`@frihk`) to compile and unit-test their packages in parallel without waiting on each other's implementations, define lightweight Go interfaces at the consumer boundary:
    1. `SMSDispatcher` (implemented by **Ian** `@frihk` in `internal/notifications`, injected into **Evans**'s `@eojuma` `auth` & `onboarding` and **Ian**'s `@frihk` `whatsapp`).
    2. `InAppNotifier` (implemented by **Ian** `@frihk` in `internal/notifications`, injected into **Ema**'s `@ejao-000` `academics` and **Ian**'s `@frihk` `billing`).
    3. `TermDateFeeSyncer` (implemented by **Ian** `@frihk` in `internal/billing`, injected into **Ema**'s `@ejao-000` `academics` for `PUT /api/v1/calendar`).
    4. `StudentBillingHook` (implemented by **Ian** `@frihk` in `internal/billing`, injected into **Evans**'s `@eojuma` `onboarding` when students are admitted/verified).
    
    ---
    
    ## 6. Shared Resources and Their Owners
    
    | Shared Resource | Single Accountable Owner | Modification Process for Other Developers |
    |---|---|---|
    | `backend/internal/database/` & `backend/internal/middleware/` | **Evans** (`@eojuma`, Dev 3) | Frozen after Milestone 1. Any change requires a PR reviewed by Evans (`@eojuma`), Ema (`@ejao-000`), and Ian (`@frihk`). |
    | `backend/cmd/server/main.go` | **Evans** (`@eojuma`, Dev 3) | Ema (`@ejao-000`) and Ian (`@frihk`) export `RegisterRoutes(mux *http.ServeMux, deps ...)` in their own packages; Evans (`@eojuma`) wires the 1-line calls in `main.go`. |
    | `backend/cmd/worker/main.go` & `backend/internal/jobs/` | **Ian** (`@frihk`, Dev 5) | Evans (`@eojuma`) provides `onboarding.PurgeExpiredUnverifiedSubmissions`; Ian (`@frihk`) registers it in the Asynq cron scheduler. |
    | `backend/internal/internalapi/` | **Ian** (`@frihk`, Dev 5) | Coordinates JSON payloads directly with **Sos** (`@sospeter-57`, Dev 1) and calls read methods on **Ema**'s (`@ejao-000`) `academics.Service`. |
    | `backend/go.mod` & `backend/go.sum` | **Evans** (`@eojuma`, Dev 3) | Evans (`@eojuma`) adds initial core deps (`pgx/v5`, `golang-jwt/jwt/v5`, `asynq`, `go-redis/v9`) in Milestone 1 so `go.mod` stabilizes early. |
    | `docker-compose.yml`, `Makefile`, `.env.example` | **Evans** (`@eojuma`, Dev 3) | Changes require PR approval from affected service owner(s). |
    | `Docs/API_CONTRACTS.md` | Shared Documentation | Any contract adjustment must be updated in `Docs/API_CONTRACTS.md` in the same PR. |
    
    ---
    
    ## 7. Recommended Implementation Order
    
    ### Phase 1: Foundation & Contracts (Days 1–3)
    - **Evans (`@eojuma`):** Delivers `config`, `database` (`WithTenantTx` + `AuditLogger`), `middleware` (`RequireAuth`, `RequirePermission`, `RespondJSON`), `migrations/000002_r1_core_onboarding.up.sql`, and `/api/v1/auth/otp/*`.
    - **Ema (`@ejao-000`):** Authors `migrations/000003_r1_academics.up.sql` and builds `assignments.go`, `grading_scale.go`, and `roster_stats.go`.
    - **Ian (`@frihk`):** Authors `migrations/000004_r1_billing_whatsapp_notifications.up.sql`, implements `notifications` (In-App + Sandbox SMS), and builds `context_signer.go` + `internalapi` skeleton.
    - **Sos (`@sospeter-57`):** Implements `ai_service/app/security.py` (`X-User-Context` HMAC verifier), `backend_client.py`, `llm_client.py` (with deterministic fallback), and session lookup rate limiter.
    - **Ashy (`@Ashomondi`):** Sets up the Tailwind Design System (`tailwind.config.ts`), `src/components/ui/*`, `src/lib/api-client.ts`, `auth-context.tsx`, and `/login` screen.
    
    ### Phase 2: Core Domain Features (Days 4–8)
    - **Evans (`@eojuma`):** Implements School creation, Principal Setup Wizard, Staff QR & 12h Parent QR generation/revocation, Verifier badge delegation, In-person verification + immediate document deletion, Student admission, and Guardian 2FA phone change.
    - **Ema (`@ejao-000`):** Implements Daily Class-Teacher Roll-Call Attendance, Subject Teacher Marks Entry, Marks Reassignment, and Timetables/Calendar (with Term Start-Date change hook).
    - **Ian (`@frihk`):** Implements Fee structure loader, Automated start-of-year/term charging, Student fee balance/ledger (with credit balance), Mid-year fee changes, and Per-school WhatsApp webhook + 6h OTP session state machine.
    - **Sos (`@sospeter-57`):** Implements `parent_assistant.py` (In-App & WhatsApp minimal-fact Q&A + school contact fallback) and `chat_payment.py` (multi-turn slot extraction + explicit `"YES"` confirmation).
    - **Ashy (`@Ashomondi`):** Implements Principal Setup Wizard (`/admin/wizard`), QR Onboarding & Verification screens (`/admin/onboarding`, `/register/[token]`, `/admin/verifications`), Dean workspace (`/admin/dean`), and Teacher Attendance & Marks screens (`/teacher/*`).
    
    ### Phase 3: Advanced State Machines & End-to-End Integration (Days 9–12)
    - **Ema (`@ejao-000`):** Completes Hard-Gated Results Release, Snapshot-Preserving Reopen & Targeted Re-Release state machine, AI Report Comment Drafter proxy, and Student CBC Performance Graph endpoint.
    - **Ian (`@frihk`):** Completes Daraja Sandbox STK Push (with 24h phone-change freeze check), Direct Paybill C2B auto-matching, Unmatched payments resolution, and Asynq background workers (`cmd/worker`).
    - **Sos (`@sospeter-57`):** Completes `report_drafter.py` (`term_comment` and `rerelease_note` modes with anti-ranking/anti-diagnosis guardrails) and full `pytest` suite.
    - **Ashy (`@Ashomondi`):** Completes Teacher Results Release/Reopen UI (`/teacher/results`), Parent Portal (`/parent/*` including `CbcPerformanceChart.tsx`, Fee Ledger & STK Push modal, In-App AI Assistant drawer, 2FA Phone Change, and R2 placeholders).
    
    ---
    
    ## 8. Integration Milestones
    
    | Milestone | Target | Verification Gate |
    |---|---|---|
    | **M1: Schema & Auth Spine Ready** | End of Day 3 | `make migrate-up` applies migrations `000001`–`000004` cleanly. `POST /api/v1/auth/otp/request` & `verify` return valid JWTs. `WithTenantTx` enforces RLS (`AC-27`). |
    | **M2: Release 1 Pass Condition Gate (`C53`)** | End of Day 7 | School creation, Principal Setup Wizard (`AC-01`), Staff QR (`AC-02`), 12h Parent QR (`AC-03`), Pending isolation (`AC-04`), Verifier delegation (`AC-05`), Document deletion (`AC-06`), App OTP login, and WhatsApp 6h OTP session (`AC-08`–`AC-12`) all work end-to-end with Ashy's (`@Ashomondi`) UI. |
    | **M3: Academics & Billing Gate** | End of Day 10 | Roll-call (`AC-14`), Marks entry & reassignment (`AC-15`, `AC-17`), Hard-gated release (`AC-16`), Reopen/re-release (`AC-18`), CBC Graph (`AC-19`), Term fee charging (`AC-20`), Mid-year fee changes (`AC-21`), Term-date shift (`AC-22`), Overpayment credit (`AC-24`), and C2B matching (`AC-25`) pass integration tests. |
    | **M4: Full AI, WhatsApp & PWA Sign-Off** | End of Day 12 | In-App & WhatsApp Parent Assistant (`AC-13`), Two-step confirmed Chat STK Push (`AC-23`), 24h Phone-Change Payment Lock (`AC-26`), AI Report Comment Drafter, and AI Audit Logging (`AC-28`, `AC-29`) pass full QA verification on synthetic data. |
    
    ---
    
    ## 9. Testing and Quality Assurance Strategy
    
    1. **Synthetic Data Mandate (`Docs/Ikonex.md` C52, `AGENTS.md` Rule 8):**
       - Never use real student, parent, or staff names, national IDs, or phone numbers in seed scripts, tests, or demo environments.
       - Provide a deterministic synthetic seed script (`backend/cmd/seed` or SQL seed fixture) creating two synthetic schools (*"Ikonex Demo High School A"* and *"Ikonex Demo Junior School B"*) to test cross-tenant RLS isolation (`AC-27`, `AC-12`) and all role/badge combinations.
    2. **Automated Test Suites (`make test`):**
       - **Go Backend (`Evans @eojuma`, `Ema @ejao-000`, `Ian @frihk`):** `cd backend && go test -v -race ./...` must pass all unit and database integration tests.
       - **Python AI Service (`Sos @sospeter-57`):** `cd ai_service && pytest -v` must pass all security, rate-limit, hallucination-guard, WhatsApp minimal-fact, and multi-turn payment confirmation tests.
       - **Next.js Frontend (`Ashy @Ashomondi`):** `cd frontend && npm run lint && npm run build` must succeed with zero errors.
    3. **Acceptance Criteria Traceability:**
       - Every single acceptance criterion `AC-01` through `AC-29` in `Docs/Ikonex.md` Section 14 must be mapped to an automated test or documented QA verification step before Release 1 sign-off.
    
    ---
    
    ## 10. Git Branching and Pull Request Workflow
    
    ### 10.1 Branch Naming Convention
    All work branches off `main` (or integration branch `release-1`) using developer-prefixed feature branches:
    - **Sos (`@sospeter-57`, Dev 1):** `feat/sos-ai-core`, `feat/sos-parent-assistant`, `feat/sos-report-drafter`
    - **Ashy (`@Ashomondi`, Dev 2):** `feat/ashy-design-auth`, `feat/ashy-onboarding-ui`, `feat/ashy-teacher-portal`, `feat/ashy-parent-portal`
    - **Evans (`@eojuma`, Dev 3):** `feat/evans-rls-auth-badges`, `feat/evans-wizard-staff-qr`, `feat/evans-parent-qr-verify`
    - **Ema (`@ejao-000`, Dev 4):** `feat/ema-dean-calendar`, `feat/ema-attendance-stats`, `feat/ema-marks-results-release`
    - **Ian (`@frihk`, Dev 5):** `feat/ian-fee-ledgers`, `feat/ian-mpesa-daraja`, `feat/ian-whatsapp-internalapi`
    
    ### 10.2 Commit & Pull Request Rules
    1. **Conventional Commits:** Use `feat(scope): ...`, `fix(scope): ...`, `test(scope): ...`, `docs(scope): ...` (e.g., `feat(academics): enforce hard gate on class results release`).
    2. **Directory Boundary Check:** A PR must only touch files within the author's owned paths (Section 9 of each developer's task file). If a shared file (`docker-compose.yml`, `main.go`, `API_CONTRACTS.md`) is modified, the owner of that shared resource must approve the PR.
    3. **CI / Pre-Merge Requirement:** Before merging any PR, `make test` and `frontend` build checks must pass, and any new endpoint must match `Docs/API_CONTRACTS.md`.
    
    ---
    
    ## 11. Open Decisions & Assumptions Handling for Release 1
    
    Per `Docs/Ikonex.md` Appendices A & B and `AGENTS.md` Rule 10 (*"Don't guess open decisions"*), developers must implement the following configurable defaults so no open question blocks Release 1 or hardcodes an irreversible assumption:
    
    | ID | Open Question / Assumption | How to Handle Cleanly in Release 1 Code | Owner |
    |---|---|---|---|
    | **O9 / A12** | How a shared Parent QR links each parent to the right child | Allow the parent registration form to optionally accept `student_admission_number` (when the class-teacher route already created the student, flagging duplicates per `A12`) OR new student details, and allow the in-person staff verifier to confirm/select the linked student during `POST /api/v1/registrations/{id}/verify`. | **Evans** (`@eojuma`) & **Ashy** (`@Ashomondi`) |
    | **O16** | Whether Principal/Deputy can override the Dean on assignments, timetables, and calendar | Gate Dean endpoints via permission codes (`academics.assignments.manage`, `academics.timetables.manage`, `academics.calendar.manage`) seeded to `dean_of_studies` by default, with a boolean config flag `ALLOW_PRINCIPAL_DEAN_OVERRIDE=true` in `.env.example` so behavior can be toggled without code changes. | **Ema** (`@ejao-000`) & **Evans** (`@eojuma`) |
    | **O17** | Days before rejected/never-verified submissions are deleted | Controlled via environment variable `UNVERIFIED_SUBMISSION_RETENTION_DAYS=7` in `.env.example` rather than a hardcoded SQL constant. | **Evans** (`@eojuma`) & **Ian** (`@frihk`) |
    | **O20 / A3** | How OTP limits apply to app login & per admission number | Enforce the `1 OTP/min` and `>5 in 30 mins -> 24h sender block` in a configurable rate-limiter middleware used by both WhatsApp (`C18–C20`) and App login (`O20`), configurable via `OTP_COOLDOWN_SECONDS=60` and `OTP_MAX_PER_30_MIN=5`. | **Evans** (`@eojuma`) & **Ian** (`@frihk`) |
    | **O13 / C47** | Fee structure variants (boarder vs. day scholar) | Build `fee_periods` on generic billing periods and include an optional `boarding_status VARCHAR(20) DEFAULT 'all'` filter on `fee_structures` so Release 1 works with standard term fees while supporting boarder/day variants without schema changes. | **Ian** (`@frihk`) |
    
    ---
    
    ## 12. Final Release 1 Integration Checklist
    
    - [ ] **Migrations (`000001`–`000004`):** All tables created, `FORCE ROW LEVEL SECURITY` enabled on every tenant table, and `AC-27` verified.
    - [ ] **Pass Condition 1 — Onboarding (`C53`):** School + Principal provisioning (`AC-01`), Principal Setup Wizard (`AC-01`), Staff QR (`AC-02`), 12h Parent QR (`AC-03`), Pending submission isolation (`AC-04`), Staff verifier delegation (`AC-05`), Immediate document deletion on approval (`AC-06`), and Unverified submission auto-purge (`AC-07`) verified end-to-end.
    - [ ] **Pass Condition 2 — Parent Login & WhatsApp Sessions (`C53`):** App SMS OTP login (`AC-01`), WhatsApp chat initiation from any number (`AC-08`), Masked guardian phone selection (`AC-09`), OTP rate limits & 24h sender-only block (`AC-10`), 6-hour session expiry (`AC-11`), and Per-school WhatsApp bot isolation (`AC-12`) verified end-to-end.
    - [ ] **Pass Condition 3 — Badges & Permissions (`C53`):** Scoped `class_teacher` badge (`AC-14`), Time-bound `delegated_verifier` badge (`AC-05`), Subject teacher assignment & Dean reassignment (`AC-15`, `AC-17`), and Anonymised other-class means verified end-to-end.
    - [ ] **Academics & Results State Machine:** Hard-gated results release (`AC-16`), Snapshot-preserving reopen & targeted re-release with required teacher note (`AC-18`), Versioned grading scale & non-ranking CBC Performance Graph (`AC-19`), and AI report comment drafting verified end-to-end.
    - [ ] **Billing, M-Pesa & AI Assistant:** Start-of-year term fee charging (`AC-20`), Mid-year fee changes with mandatory note (`AC-21`), Term-date change shifting fee due dates (`AC-22`), Two-step confirmed chat STK push (`AC-23`), Overpayment ledger credit (`AC-24`), Direct Paybill C2B matching & unmatched queue (`AC-25`), 24h phone-change chat payment pause (`AC-26`), WhatsApp minimal-fact restriction (`AC-13`), Zero AI DB access & AI action logging (`AC-28`), and complete `audit_log` coverage (`AC-29`) verified end-to-end.
    - [ ] **Full Stack Containerization:** `docker compose up --build` starts `postgres`, `redis`, `backend`, `ai_service`, and `frontend` cleanly with zero port or contract conflicts.
    ✓ • 0ms

## Activity

    $ sed -n '1,145p' tasks/MASTER-INTEGRATION-PLAN.md
    # Ikonex Academy – Release 1 Master Integration Plan
    
    - **Document Version:** 1.0 (Release 1 Official Architecture & Execution Plan)
    - **Authoritative Sources:** `Docs/Ikonex.md`, `Docs/REQUIREMENTS.md`, `Docs/ARCHITECTURE.md`, `Docs/API_CONTRACTS.md`, `Docs/DECISIONS.md`, `Docs/SECURITY.md`, `Docs/AGENTS.md`
    - **Release 1 Scope:** Synthetic-data proof of concept (`C51, C52`), English only (`C56`), running against Safaricom Daraja sandbox and test WhatsApp numbers (`A17`). Passes when **Onboarding works**, **Parents can log in (App SMS OTP & WhatsApp 6h OTP session)**, **Badges work stress-free**, and **all Release 1 features work cleanly end-to-end (`C53, C54`)**.
    
    ---
    
    ## 1. Overview of the Five Developer Assignments
    
    | Developer ID | Name | GitHub Tag | Role | Assigned Scope Summary | Task Document |
    |---|---|---|---|---|---|
    | **Developer 1** | **Sos** | `@sospeter-57` | AI Engineer (Python / FastAPI / Mistral) | Owns `ai_service/**`: Service-token + HMAC `X-User-Context` verifier, Parent Assistant (In-App & WhatsApp minimal-fact Q&A), Multi-turn M-Pesa Chat Payment Confirmation flow, CBC Report Card Comment & Re-Release Note Drafter, and AI Audit Logging. Zero direct DB access. | `tasks/developer-01-ai-engineer.md` |
    | **Developer 2** | **Ashy** | `@Ashomondi` | Frontend Engineer (Next.js 14 / TypeScript / Tailwind PWA) | Owns `frontend/**`: Tailwind Design System, SMS OTP Login & Membership Switcher, Principal 5-Step Setup Wizard, Staff & 12h Parent QR Registration + In-Person Verification UI, Dean of Studies Workspace, Teacher Portal (Roll-Call, Marks, Hard-Gated Release/Reopen, AI Comment Drafter), and Parent Portal (Results, CBC Graph, Fee Ledger, STK Push, In-App AI Chat, 2FA Phone Change, R2 Placeholders). | `tasks/developer-02-frontend-engineer.md` |
    | **Developer 3** | **Evans** | `@eojuma` | Backend Engineer A (Go — Core, Auth & Onboarding) | Owns `cmd/server`, `internal/config`, `internal/database`, `internal/middleware`, `internal/auth`, `internal/onboarding`, `internal/consent`, and `migrations/000001`–`000002`: PostgreSQL RLS (`WithTenantTx`), SMS OTP Login, Scoped/Time-Bound Badge Engine, Audit Logger, School Provisioning, Setup Wizard, Staff/Parent QR Codes, In-Person Verification & Document Deletion, Verifier Delegation, Student Admission, and Guardian 2FA Phone Change. | `tasks/developer-03-backend-engineer-a.md` |
    | **Developer 4** | **Ema** | `@ejao-000` | Backend Engineer B (Go — CBC Academics & Results) | Owns `internal/academics` and `migrations/000003`: Dean of Studies Teacher Assignments, Timetables, School Calendar & Term-Date Fee Sync, Versioned CBC Grading Scales, Daily Class-Teacher Roll-Call Attendance, Anonymised Class Means, Subject Marks Entry & Reassignment, Hard-Gated Results Release/Reopen/Re-Release Snapshot State Machine, AI Drafter Proxy, and Student CBC Performance Graph. | `tasks/developer-04-backend-engineer-b.md` |
    | **Developer 5** | **Ian** | `@frihk` | Backend Engineer C (Go — Billing, WhatsApp, Jobs & Internal AI API) | Owns `cmd/worker`, `internal/billing`, `internal/whatsapp`, `internal/notifications`, `internal/jobs`, `internal/internalapi`, and `migrations/000004`: Team-Loaded Fee Structures, Automated Term Charging, Ledger Credit Balances, Mid-Year Fee Changes, Daraja STK Push & Direct C2B Paybill Auto-Matching, Per-School WhatsApp Bot & 6h OTP Sessions, In-App + Transactional SMS Notifications, Asynq Workers, and Internal AI API. | `tasks/developer-05-backend-engineer-c.md` |
    
    ---
    
    ## 2. Complete Release 1 Feature Ownership Matrix
    
    Every Release 1 feature and acceptance criterion (`AC-01` to `AC-29` in `Docs/Ikonex.md` Section 14) has a single accountable backend/AI owner and frontend integration owner:
    
    | Area | Feature / Requirement | Acceptance Criteria (`Docs/Ikonex.md`) | Primary Backend / AI Owner | Frontend Owner |
    |---|---|---|---|---|
    | **Foundation** | Multi-school PostgreSQL RLS isolation (`SET LOCAL app.school_id`) | `AC-27` | **Evans** (`@eojuma` — `internal/database`) | — |
    | **Foundation** | Roles & scoped/time-bound Badges engine (`class_teacher`, `delegated_verifier`) | `C53`, `AC-05`, `AC-14` | **Evans** (`@eojuma` — `internal/middleware`) | **Ashy** (`@Ashomondi` — `auth-context.tsx`) |
    | **Foundation** | Centralized Audit Trail of all sensitive actions | `AC-29` | **Evans** (`@eojuma` — `internal/database/audit.go`) | — |
    | **Foundation** | Phone + SMS OTP App Login & Membership Switcher | `AC-01` | **Evans** (`@eojuma` — `internal/auth`) | **Ashy** (`@Ashomondi` — `/login`) |
    | **Foundation** | Consent Centre Schema (`biometric_attendance`, `photo`, `sms`, `whatsapp`) | Section 6.10 | **Evans** (`@eojuma` — `internal/consent`) | **Ashy** (`@Ashomondi` — `/parent/settings`) |
    | **Onboarding** | Ikonex Team School + Principal creation & Required School Contact Profile | `AC-01`, `C37`, `C57` | **Evans** (`@eojuma` — `internal/onboarding`) | **Ashy** (`@Ashomondi` — `/admin/platform`) |
    | **Onboarding** | Principal 5-Step Setup Wizard (Years, Terms, Grades 7–12, Classes, Dorms, Grading Scale) | `AC-01`, `C38` | **Evans** (`@eojuma` — `internal/onboarding`) | **Ashy** (`@Ashomondi` — `/admin/wizard`) |
    | **Onboarding** | Staff QR Code (8am–6pm adjustable window, revocation, document upload) | `AC-02`, `C62` | **Evans** (`@eojuma` — `internal/onboarding`) | **Ashy** (`@Ashomondi` — `/admin/onboarding`, `/register/[token]`) |
    | **Onboarding** | Parent 12h Multi-Use QR Code & Revocation | `AC-03`, `C28` | **Evans** (`@eojuma` — `internal/onboarding`) | **Ashy** (`@Ashomondi` — `/admin/onboarding`, `/register/[token]`) |
    | **Onboarding** | Pending Submission Isolation (Zero child data access until verified) | `AC-04`, `A8` | **Evans** (`@eojuma` — `internal/onboarding`) | **Ashy** (`@Ashomondi` — `/register/[token]`) |
    | **Onboarding** | Time-Bound Verifier Delegation to Staff (Expires with 12h Parent QR) | `AC-05`, `C28`, `C29` | **Evans** (`@eojuma` — `internal/onboarding`) | **Ashy** (`@Ashomondi` — `/admin/onboarding`) |
    | **Onboarding** | In-Person Document Verification, Immediate Document Deletion & Auto-Purge of Unverified Submissions | `AC-06`, `AC-07`, `C63`, `C64` | **Evans** (`@eojuma` — `internal/onboarding`) + **Ian** (`@frihk` — `internal/jobs` cron) | **Ashy** (`@Ashomondi` — `/admin/verifications`) |
    | **Onboarding** | Dual-Route Student Admission (Class-Teacher Manual & Opening-Day QR Duplicate Flagging) + Re-Enrolment | `C39–C41`, `A12` | **Evans** (`@eojuma` — `internal/onboarding`) | **Ashy** (`@Ashomondi` — `/admin/dean`, `/admin/verifications`) |
    | **Onboarding** | Guardian 2FA Registered Phone Number Change (OTP to new number, SMS to old number, 24h chat payment pause) | `AC-26`, `C25`, `C27` | **Evans** (`@eojuma` — `internal/onboarding`) | **Ashy** (`@Ashomondi` — `/parent/settings`) |
    | **Academics** | Dean of Studies Teacher Class/Subject Assignments & Marks Reassignment | `AC-15`, `AC-17`, `C32`, `C59` | **Ema** (`@ejao-000` — `internal/academics`) | **Ashy** (`@Ashomondi` — `/admin/dean`) |
    | **Academics** | Timetables & School Calendar (including Post-Wizard Term Date Edits & Fee Due-Date Shift) | `AC-22`, `C59–C61` | **Ema** (`@ejao-000` — `internal/academics`) + **Ian** (`@frihk` — `billing` hook) | **Ashy** (`@Ashomondi` — `/admin/dean`, `/teacher`, `/parent`) |
    | **Academics** | Daily Class-Teacher Roll-Call Attendance & In-App Absence Notifications | `AC-14`, `C35`, `C36` | **Ema** (`@ejao-000` — `internal/academics`) | **Ashy** (`@Ashomondi` — `/teacher/attendance`) |
    | **Academics** | Scoped Class Roster vs. Anonymised Other-Class Means (Never student names) | Section 4 | **Ema** (`@ejao-000` — `internal/academics`) | **Ashy** (`@Ashomondi` — `/teacher`) |
    | **Academics** | Subject Teacher Marks Entry & Hard-Gated Class Results Release (100% marks required) | `AC-15`, `AC-16`, `C30`, `C31` | **Ema** (`@ejao-000` — `internal/academics`) | **Ashy** (`@Ashomondi` — `/teacher/marks`, `/teacher/results`) |
    | **Academics** | Results Reopen & Re-Release State Machine (Snapshot preserved while reopened, mandatory teacher note, only affected parents notified) | `AC-18`, `C33`, `C34` | **Ema** (`@ejao-000` — `internal/academics`) | **Ashy** (`@Ashomondi` — `/teacher/results`, `/parent/results`) |
    | **Academics** | Versioned Per-Term CBC Grading Scale & Multi-Term CBC Performance Graph (Zero class rank) | `AC-19`, `C42`, `C43` | **Ema** (`@ejao-000` — `internal/academics`) | **Ashy** (`@Ashomondi` — `/parent/results` — `CbcPerformanceChart.tsx`) |
    | **Fees & M-Pesa** | Team-Loaded Fee Structures & Automated Start-of-Year Term Charging | `AC-20`, `C45–C49` | **Ian** (`@frihk` — `internal/billing`) | **Ashy** (`@Ashomondi` — `/admin/platform`, `/parent/fees`) |
    | **Fees & M-Pesa** | Mid-Year Fee Changes (Mandatory note, audit log, only affected parents notified) | `AC-21`, `C45` | **Ian** (`@frihk` — `internal/billing`) | **Ashy** (`@Ashomondi` — `/admin/platform`, `/parent/fees`) |
    | **Fees & M-Pesa** | M-Pesa Daraja STK Push, Any-Amount Acceptance & Overpayment Ledger Credit | `AC-23`, `AC-24`, `C23`, `C24` | **Ian** (`@frihk` — `internal/billing`) | **Ashy** (`@Ashomondi` — `/parent/fees`) |
    | **Fees & M-Pesa** | Direct Paybill (C2B) Auto-Matching by Admission Number & Unmatched Payments Queue | `AC-25`, `C50` | **Ian** (`@frihk` — `internal/billing`) | **Ashy** (`@Ashomondi` — `/admin/platform`) |
    | **WhatsApp & Notifications** | Per-School WhatsApp Bot Routing, Masked Guardian Numbers, 6h OTP Sessions & Sender-Only 24h Rate Block | `AC-08`, `AC-09`, `AC-10`, `AC-11`, `AC-12` | **Ian** (`@frihk` — `internal/whatsapp`) | — |
    | **WhatsApp & Notifications** | In-App Event Notifications & Transactional SMS Dispatcher (Africa's Talking) | `C58`, `C12`, `C27` | **Ian** (`@frihk` — `internal/notifications`, `internal/jobs`) | **Ashy** (`@Ashomondi` — `NotificationDrawer.tsx`) |
    | **AI Service** | Internal AI API Gateway (`X-Internal-Service-Token` + signed `X-User-Context`) & AI Action Log | `AC-28`, `C21` | **Ian** (`@frihk` — `internal/internalapi`) + **Sos** (`@sospeter-57` — `ai_service/app/security.py`) | — |
    | **AI Service** | Parent Assistant (In-App & WhatsApp Read-Only Q&A, WhatsApp Minimal Facts / No Results on WhatsApp, School Contact Fallback) | `AC-13`, `C22`, `C56`, `C57`, `C66` | **Sos** (`@sospeter-57` — `parent_assistant.py`) + **Ian** (`@frihk` — `internalapi`, `whatsapp`) | **Ashy** (`@Ashomondi` — `/parent/chat`) |
    | **AI Service** | Chat-Initiated M-Pesa Payment (Multi-Turn Slot Extraction & Mandatory Explicit "YES" Repeat-Back Confirmation) | `AC-23`, `AC-26`, `C23`, `A4` | **Sos** (`@sospeter-57` — `chat_payment.py`) + **Ian** (`@frihk` — `internalapi`, `billing`) | **Ashy** (`@Ashomondi` — `/parent/chat`) |
    | **AI Service** | CBC Report Card Comment & Re-Release Explanation Note Drafter (Teacher edits before publishing) | `C5`, `C30`, `C34` | **Sos** (`@sospeter-57` — `report_drafter.py`) + **Ema** (`@ejao-000` — `academics`) | **Ashy** (`@Ashomondi` — `/teacher/results`) |
    
    ---
    
    ## 3. API Contract Ownership and Shared Conventions
    
    To resolve the open API formatting items noted in `Docs/Ikonex.md` Section 10 and `Docs/API_CONTRACTS.md`, **all five developers must follow these exact conventions**:
    
    ### 3.1 Base URLs & Versioning
    - **Frontend -> Go Backend:** `http://localhost:8080/api/v1` (`NEXT_PUBLIC_API_URL`)
    - **Go Backend -> AI Service:** `http://ai_service:8000/api/v1/ai`
    - **AI Service -> Go Backend Internal API:** `http://backend:8080/api/v1/internal` (`GO_BACKEND_INTERNAL_URL`)
    
    ### 3.2 Standard JSON Response Envelope
    - **Success (`200 OK`, `201 Created`):**
      ```json
      {
        "status": "success",
        "data": { ... },
        "meta": {
          "page": 1,
          "per_page": 25,
          "total": 100
        }
      }
      ```
      *(Note: `meta` is required on paginated list endpoints and optional on single-object responses).*
    - **Error (`400`, `401`, `403`, `404`, `409`, `410`, `422`, `429`, `500`):**
      ```json
      {
        "status": "error",
        "error": {
          "code": "MACHINE_READABLE_UPPER_SNAKE_CODE",
          "message": "Clear human-readable explanation for UI display.",
          "details": {}
        }
      }
      ```
    
    ### 3.3 Date, Time, and Currency Conventions
    - **Timestamps:** Always ISO-8601 in East Africa Time (`Africa/Nairobi`, `+03:00`) or UTC with explicit offset, e.g. `"2026-11-05T14:30:00+03:00"`.
    - **Calendar Dates (Term dates, DOB, Attendance date, Due dates):** `YYYY-MM-DD` (`"2026-11-05"`).
    - **Currency (KES):** Sent in JSON as fixed 2-decimal strings (`"12500.00"`) or numeric values formatted to 2 decimal places (`NUMERIC(12,2)` in PostgreSQL; never float32).
    - **Phone Numbers:** Normalized to E.164 format (`"+254712345678"`) at the API boundary.
    
    ### 3.4 Go-to-AI Service Authentication Contract (**Ian** `@frihk`, **Ema** `@ejao-000`, **Sos** `@sospeter-57`)
    Every request between the Go Backend and `ai_service` (in both directions) must include:
    1. `X-Internal-Service-Token: <INTERNAL_SERVICE_TOKEN>`
    2. `X-User-Context: <base64url(json_payload)>.<hex(hmac_sha256(base64url_payload, INTERNAL_SERVICE_TOKEN))>`
       Where `json_payload` is:
       ```json
       {
         "user_id": "uuid",
         "school_id": "uuid",
         "membership_id": "uuid",
         "role": "parent",
         "channel": "whatsapp",
         "session_id": "uuid",
         "exp": 1762300000
       }
       ```
    
    ---
    
    ## 4. Database Ownership and Schema Coordination
    
    To prevent migration merge conflicts, database migrations are partitioned into **four numbered files**, each with a single owner:
    
    | Migration File | Owner | Tables Created / Altered |
    |---|---|---|
    | `migrations/000001_init_schema.up.sql` & `.down.sql` | **Evans** (`@eojuma`, Dev 3) | **Existing baseline (do not edit):** `schools`, `users`, `permissions`, `memberships`, `badges`, `badge_permissions`, `membership_badges`, `academic_years`, `terms`, `grade_levels`, `classes`, `dormitories`, `clubs`, `students`, `student_guardians`, `enrolments`, `consents`, `biometric_templates`, `staff_invites`, `fee_structures`, `fee_structure_items`, `fee_ledgers`, `payments`, `payment_plans`, `nurse_visits`, `library_books`, `book_loans`, `cbc_strands`, `syllabus_coverage`, `student_assessments`, `blog_posts`, `audit_log` + RLS loop. |
    | `migrations/000002_r1_core_onboarding.up.sql` & `.down.sql` | **Evans** (`@eojuma`, Dev 3) | `school_profile`, `otp_requests`, `sender_rate_blocks`, `parent_qr_codes`, `registration_submissions`, `verification_records`; alters `users` (`chat_payment_paused_until`) and `consents` (adds `'whatsapp'` purpose) + RLS policies. |
    | `migrations/000003_r1_academics.up.sql` & `.down.sql` | **Ema** (`@ejao-000`, Dev 4) | `subjects`, `subject_teacher_assignments`, `grading_scale_versions`, `timetable_slots`, `calendar_events`, `attendance_daily`, `student_marks`, `report_comments`, `result_releases`, `released_marks_snapshots` + RLS policies. |
    | `migrations/000004_r1_billing_whatsapp_notifications.up.sql` & `.down.sql` | **Ian** (`@frihk`, Dev 5) | `school_mpesa_configs`, `fee_periods`, `fee_charges`, `fee_change_notes`, `ledger_entries`, `unmatched_payments`, `school_whatsapp_bots`, `chat_sessions`, `in_app_notifications`, `ai_action_logs`; alters `fee_ledgers` (`credit_balance`, `due_date`) + RLS policies. |
    
    **Mandatory RLS Rule for All Migrations:** Every tenant-owned table created in `000002`, `000003`, and `000004` must include `school_id UUID NOT NULL REFERENCES schools(id) ON DELETE CASCADE` and execute:
    ```sql
    ALTER TABLE <table_name> ENABLE ROW LEVEL SECURITY;
    ALTER TABLE <table_name> FORCE ROW LEVEL SECURITY;
    CREATE POLICY <table_name>_tenant_isolation ON <table_name>
        FOR ALL USING (school_id = current_school_id())
        WITH CHECK (school_id = current_school_id());
    ```
    
    ---
    
    ## 5. Cross-Developer Dependencies
    
    ```mermaid
    ✓ • 1m 33s

## Activity

    $ sed -n '130,270p' tasks/developer-05-backend-engineer-c.md
    Create `000004_r1_billing_whatsapp_notifications.up.sql` and `.down.sql` with all Release 1 Billing, WhatsApp, Notification, and AI Audit tables and RLS policies:
    - `school_mpesa_configs`: `school_id (PK/FK)`, `shortcode VARCHAR(30) UNIQUE NOT NULL`, `environment VARCHAR(20) DEFAULT 'sandbox'`, `passkey_encrypted TEXT`, `test_payment_verified BOOLEAN DEFAULT FALSE`, `updated_at`.
    - `fee_periods`: `id`, `school_id`, `academic_year_id`, `term_id`, `period_type VARCHAR(30) NOT NULL DEFAULT 'term'`, `name VARCHAR(100) NOT NULL`, `due_date DATE NOT NULL`, `created_at`, `UNIQUE(school_id, academic_year_id, term_id)`.
    - `fee_charges`: `id`, `school_id`, `student_id`, `fee_period_id`, `fee_structure_id`, `amount NUMERIC(12,2) NOT NULL`, `due_date DATE NOT NULL`, `created_at`, `UNIQUE(school_id, student_id, fee_period_id)`.
    - `fee_change_notes`: `id`, `school_id`, `term_id`, `fee_period_id`, `change_type VARCHAR(50) NOT NULL`, `amount_delta NUMERIC(12,2) DEFAULT 0`, `old_due_date DATE`, `new_due_date DATE`, `note TEXT NOT NULL`, `created_by UUID NOT NULL`, `affected_students_count INT NOT NULL`, `created_at`.
    - `ledger_entries`: `id`, `school_id`, `student_id`, `term_id`, `entry_type VARCHAR(30) NOT NULL CHECK (entry_type IN ('charge', 'payment', 'credit_applied', 'adjustment'))`, `amount NUMERIC(12,2) NOT NULL`, `reference VARCHAR(100)`, `description TEXT`, `created_at`.
    - Alter `fee_ledgers`: Add `credit_balance NUMERIC(12,2) NOT NULL DEFAULT 0.00`, `due_date DATE`.
    - `unmatched_payments`: `id`, `school_id`, `payment_id UUID NOT NULL REFERENCES payments(id) ON DELETE CASCADE`, `transaction_ref VARCHAR(100) NOT NULL`, `amount NUMERIC(12,2) NOT NULL`, `payer_phone VARCHAR(20)`, `payer_name VARCHAR(255)`, `raw_account_ref VARCHAR(100)`, `status VARCHAR(30) NOT NULL DEFAULT 'unmatched'`, `resolved_student_id UUID REFERENCES students(id)`, `resolved_by UUID REFERENCES users(id)`, `resolution_note TEXT`, `resolved_at TIMESTAMPTZ`, `created_at`.
    - `school_whatsapp_bots`: `id`, `school_id`, `phone_number_id VARCHAR(100) UNIQUE NOT NULL`, `display_phone_number VARCHAR(30) UNIQUE NOT NULL`, `is_active BOOLEAN DEFAULT TRUE`, `created_at`.
    - `chat_sessions`: `id`, `school_id`, `channel VARCHAR(20) NOT NULL CHECK (channel IN ('whatsapp', 'in_app'))`, `sender_wa_number VARCHAR(30)`, `student_id UUID REFERENCES students(id)`, `guardian_user_id UUID REFERENCES users(id)`, `state VARCHAR(40) NOT NULL DEFAULT 'AWAITING_ADMISSION_NUMBER'`, `verified_at TIMESTAMPTZ`, `expires_at TIMESTAMPTZ`, `lookup_count INT NOT NULL DEFAULT 0`, `payment_init_count INT NOT NULL DEFAULT 0`, `created_at`, `updated_at`.
    - `in_app_notifications`: `id`, `school_id`, `recipient_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE`, `category VARCHAR(50) NOT NULL`, `title VARCHAR(255) NOT NULL`, `body TEXT NOT NULL`, `metadata JSONB`, `is_read BOOLEAN NOT NULL DEFAULT FALSE`, `read_at TIMESTAMPTZ`, `created_at`.
    - `ai_action_logs`: `id`, `school_id`, `user_id UUID`, `session_id UUID`, `channel VARCHAR(20) NOT NULL`, `feature VARCHAR(50) NOT NULL`, `intent VARCHAR(50)`, `endpoints_read JSONB`, `action_taken VARCHAR(100)`, `metadata JSONB`, `created_at`.
    - Apply `ENABLE ROW LEVEL SECURITY`, `FORCE ROW LEVEL SECURITY`, and `tenant_isolation` policies on all tenant tables above.
    
    ### 4.2 Owned Go Files & Package Structure
    ```text
    backend/
    ├── cmd/
    │   └── worker/
    │       └── main.go                      # Asynq worker binary entry point
    └── internal/
        ├── billing/
        │   ├── billing.go                   # Service struct, RegisterRoutes(mux, deps), cross-module hooks
        │   ├── structures.go                # Team fee structure loader & auto start-of-year/term charging
        │   ├── ledgers.go                   # Student balance, credit_balance carry-forward, ledger history
        │   ├── fee_changes.go               # Mid-year fee changes (note required) & term start-date hook
        │   ├── mpesa_stk.go                 # Daraja STK push, 24h phone-change lock check, callback handler
        │   ├── mpesa_c2b.go                 # Direct Paybill C2B auto-matching by school + admission number
        │   ├── unmatched.go                 # Unmatched payments list & team manual match resolution
        │   └── billing_test.go
        ├── whatsapp/
        │   ├── whatsapp.go                  # Webhook handler, school_whatsapp_bots lookup, session state machine
        │   ├── masking.go                   # Phone masking helper ("+254712345678" -> "07•• ••• 678")
        │   └── whatsapp_test.go
        ├── notifications/
        │   ├── notifications.go             # In-app notifications CRUD + Africa's Talking SMS client
        │   └── notifications_test.go
        ├── jobs/
        │   └── jobs.go                      # Asynq task definitions & handlers (SMS, STK poll, purge)
        └── internalapi/
            ├── internalapi.go               # Service token + HMAC X-User-Context middleware & internal routes
            ├── context_signer.go            # Signs X-User-Context for outbound calls to Sos's AI service
            └── internalapi_test.go
    ```
    
    ---
    
    ## 5. Exact Deliverables
    
    1. **Automated Term Charging & Credit Carry-Forward (`AC-20`, `AC-24`):** Loading a school's fee structure charges enrolled students once at the start of the year divided into terms (`AC-20`), sets due dates to `terms.start_date`, and holds any overpayment as `credit_balance` on `fee_ledgers` (`AC-24`).
    2. **Mid-Year Fee Changes & Term-Date Due-Date Shifts (`AC-21`, `AC-22`):**
       - `POST /api/v1/fee-changes` without a `note` fails with `422`; with a `note`, it updates affected ledgers, writes to `fee_change_notes` and `audit_log`, and notifies **only** affected parents (`AC-21`).
       - Invoking `OnTermStartDateChanged` shifts term fee due dates and sends the fee-change notice (`AC-22`).
    3. **STK Push, 24h Phone-Change Lock & Direct Paybill Matching (`AC-23`, `AC-25`, `AC-26`):**
       - STK push accepts any positive amount (`AC-23`) and blocks chat-initiated payments for 24 hours after a guardian phone number change (`AC-26`).
       - Direct Paybill C2B payments with a valid school admission number auto-credit the student's ledger (`AC-25`); invalid/missing references land in `unmatched_payments` and can be resolved via `POST /api/v1/payments/{id}/match`.
    4. **Per-School WhatsApp Bot & 6-Hour OTP Sessions (`AC-08` to `AC-12`):**
       - Unregistered WhatsApp numbers can start a chat; multiple guardians are shown as masked numbers (`07•• ••• 482`), and the SMS OTP is sent **only** to the selected registered number (`AC-08`, `AC-09`).
       - Rate limiter refuses a 2nd OTP within 1 minute and blocks the requesting WhatsApp number for 24 hours on the 6th request in 30 minutes without locking the guardian's account (`AC-10`).
       - Verified sessions expire after 6 hours (`AC-11`), and School A's bot cannot resolve School B's students (`AC-12`).
    5. **Internal AI API & Action Logging (`AC-28`):** Every call from Sos's AI service is authenticated with `X-Internal-Service-Token` + HMAC-signed `X-User-Context`, scoped to the parent's own children, and logged in `ai_action_logs`.
    
    ---
    
    ## 6. Acceptance Criteria
    
    - [ ] **AC-BE-C01 (aligns with AC-08, AC-09, AC-11, AC-12):** WhatsApp session state machine masks guardian phones, sends OTP only to the chosen registered number, expires after 6 hours, and isolates admission numbers strictly by `school_whatsapp_bots.school_id`.
    - [ ] **AC-BE-C02 (aligns with AC-10):** Requesting a 2nd WhatsApp OTP within 60s returns `429`; the 6th request within 30 minutes blocks `sender_wa_number` for 24h while the guardian can still log in from the web app or their own number.
    - [ ] **AC-BE-C03 (aligns with AC-20, AC-21, AC-22):** Term charges are created once per year divided into terms; `POST /api/v1/fee-changes` rejects missing notes (`422`) and notifies only affected guardians when valid.
    - [ ] **AC-BE-C04 (aligns with AC-23, AC-24, AC-25, AC-26):** Overpaying a fee balance records `current_balance = 0` and `credit_balance = excess`; valid C2B Paybill references auto-match while invalid references create `unmatched_payments` rows; chat payments within 24h of a phone change return `403 CHAT_PAYMENT_PAUSED_PHONE_CHANGE`.
    - [ ] **AC-BE-C05 (aligns with AC-28):** `internalapi` rejects requests with missing/invalid `X-Internal-Service-Token` or expired `X-User-Context`, rejects parent queries for a `student_id` not linked to `X-User-Context.user_id`, and writes to `ai_action_logs`.
    
    ---
    
    ## 7. Testing Requirements
    
    Write automated Go tests across your packages:
    1. **Billing & M-Pesa Tests (`internal/billing/billing_test.go`):**
       - Start-of-year term charging (`AC-20`), overpayment credit calculation (`AC-24`), and credit application to next term (`A22`).
       - Mid-year fee change missing note rejection (`422`) and targeted parent notification (`AC-21`).
       - Term start-date change shifting fee due dates (`AC-22`).
       - Direct Paybill C2B valid admission number auto-match vs. invalid reference landing in `unmatched_payments` (`AC-25`).
       - 24h phone-change chat payment block (`AC-26`).
    2. **WhatsApp Session & OTP Rate Limit Tests (`internal/whatsapp/whatsapp_test.go`):**
       - Unknown admission number returns generic reply and sends zero OTPs (`C14`).
       - Multiple guardians return masked numbers (`07•• ••• 482`) and OTP dispatches only to chosen registered number (`AC-08`, `AC-09`).
       - 1-minute cooldown and 6-in-30-minutes 24h sender block leaving guardian account unblocked (`AC-10`).
       - 6-hour session expiry (`AC-11`) and cross-school bot isolation (`AC-12`).
    3. **Internal AI API Security Tests (`internal/internalapi/internalapi_test.go`):**
       - HMAC signature verification on `X-User-Context`, guardian-child ownership check (`403` when Parent A queries Parent B's child), and AI audit logging (`AC-28`).
    
    ---
    
    ## 8. Integration Responsibilities
    
    - **Depends on Developer 3 (Evans — @eojuma – Backend Engineer A):**
      - Uses `database.WithTenantTx`, `database.AuditLogger`, and `middleware` helpers.
      - Exposes `notifications.SMSDispatcher` and `billing.StudentChargeHook` interfaces for Evans's (@eojuma) `auth` and `onboarding` packages.
    - **Integrates with Developer 4 (Ema — @ejao-000 – Backend Engineer B):**
      - Exposes `notifications.Service.CreateInAppNotification` and `billing.Service.OnTermStartDateChanged` for Ema's (@ejao-000) `academics` package.
      - Calls Ema's (@ejao-000) `academics.Service.GetStudentAttendanceSummary` and `GetStudentMarksSummary` inside `internalapi` to serve Sos's (@sospeter-57) AI service.
    - **Integrates with Developer 1 (Sos — @sospeter-57 – AI Engineer):**
      - Signs `X-User-Context` (`context_signer.go`) and forwards verified WhatsApp and in-app chat messages to Sos's (@sospeter-57) `POST /api/v1/ai/parent-query`, and serves `/api/v1/internal/*` read/STK endpoints to Sos (@sospeter-57).
    
    ---
    
    ## 9. File Ownership and Conflict Prevention
    
    | Category | Paths | Rule |
    |---|---|---|
    | **Exclusively Owned by Ian (@frihk, Dev 5)** | `backend/cmd/worker/main.go`, `backend/internal/billing/**`, `backend/internal/whatsapp/**`, `backend/internal/notifications/**`, `backend/internal/jobs/**`, `backend/internal/internalapi/**`, `migrations/000004_r1_billing_whatsapp_notifications.*` | Full read/write ownership. No other developer edits these packages or migration files. |
    | **Shared Files (Coordinate via PR)** | `backend/cmd/server/main.go` (owned by Evans — @eojuma; Ian exports `RegisterRoutes` in `billing`, `whatsapp`, `notifications`, `internalapi`), `Docs/API_CONTRACTS.md` | Coordinate route mounting with Evans (@eojuma). |
    | **Strictly Do Not Modify** | `backend/internal/auth/**`, `backend/internal/onboarding/**`, `backend/internal/middleware/**` (Evans — @eojuma), `backend/internal/academics/**` (Ema — @ejao-000), `ai_service/**` (Sos — @sospeter-57), `frontend/**` (Ashy — @Ashomondi) | Owned by other developers. |
    
    ---
    
    ## 10. Definition of Done
    
    - [ ] `migrations/000004_r1_billing_whatsapp_notifications.up.sql` and `.down.sql` apply and roll back cleanly with `FORCE ROW LEVEL SECURITY` on all tenant tables.
    - [ ] `billing`, `whatsapp`, `notifications`, `jobs`, `internalapi`, and `cmd/worker` compile and pass `go test -v ./...`.
    - [ ] Per-school WhatsApp webhook, masked guardian OTP verification, 6-hour session expiry, and sender-only 24h rate blocking work end-to-end.
    - [ ] Daraja sandbox STK push, overpayment ledger credit, C2B auto-matching, and unmatched queue resolution work end-to-end.
    - [ ] Internal AI API enforces service token + signed user context and logs every AI action in `ai_action_logs`.
    ✓ • 106ms

## Assistant

The repository is on `main` at `b2b2e3b6f15e60f25947f04c39b8f6b12b5a268f`. There is one existing modification, `Docs/DECISIONS.md`, which I’ll preserve. The sandbox failed to start, so the read-only inspection required execution approval. I’m now checking the documented contracts against the actual scaffold and migrations.

## Activity

    $ sed -n '145,320p' tasks/MASTER-INTEGRATION-PLAN.md
    ```mermaid
    flowchart TD
        subgraph Dev3_Evans["Dev 3: Evans @eojuma (Core, Auth, Onboarding)"]
            DB["internal/database (WithTenantTx + AuditLogger)"]
            MW["internal/middleware (Auth + Scoped Badges)"]
            AUTH["internal/auth (SMS OTP + JWT)"]
            ONB["internal/onboarding (Wizard, QR, Verify, Admit)"]
        end
    
        subgraph Dev4_Ema["Dev 4: Ema @ejao-000 (CBC Academics)"]
            ACAD["internal/academics (Assignments, Timetable, Calendar, Roll-Call, Marks, Release/Reopen, Graph)"]
        end
    
        subgraph Dev5_Ian["Dev 5: Ian @frihk (Billing, WhatsApp, Notifications, InternalAPI)"]
            BILL["internal/billing (Fee Structures, Ledgers, STK Push, C2B)"]
            WA["internal/whatsapp (Per-School Bot, 6h OTP Sessions)"]
            NOTIF["internal/notifications & internal/jobs (In-App + AT SMS + Asynq)"]
            IAPI["internal/internalapi (Permission-Checked AI Gateway)"]
        end
    
        subgraph Dev1_Sos["Dev 1: Sos @sospeter-57 (Python FastAPI AI Service)"]
            AIS["ai_service (Parent Assistant, Chat Payment Confirmation, Report Comment Drafter)"]
        end
    
        subgraph Dev2_Ashy["Dev 2: Ashy @Ashomondi (Next.js 14 PWA Frontend)"]
            FE["frontend (Login, Wizard, Onboarding, Dean, Teacher, Parent Portals)"]
        end
    
        MW --> DB
        AUTH --> DB
        ONB --> MW
        ACAD --> MW
        BILL --> MW
        WA --> DB
        IAPI --> MW
    
        AUTH -->|SMSDispatcher| NOTIF
        ONB -->|StudentChargeHook| BILL
        ACAD -->|OnTermStartDateChanged| BILL
        ACAD -->|CreateInAppNotification| NOTIF
        IAPI -->|Read Marks/Attendance| ACAD
        IAPI -->|Read Fees / Trigger STK| BILL
        WA -->|Verified Message| AIS
        ACAD -->|Draft Report Comment| AIS
        AIS -->|Read Facts / Confirmed STK / Audit| IAPI
    
        FE -->|REST /api/v1/auth & onboarding| ONB
        FE -->|REST /api/v1/classes, timetables, calendar, marks| ACAD
        FE -->|REST /api/v1/students/fee-balance, payments, chat, notifications| BILL
    ```
    
    ### Cross-Module Go Interfaces (Defined in Day 1 Scaffold)
    To allow **Evans** (`@eojuma`), **Ema** (`@ejao-000`), and **Ian** (`@frihk`) to compile and unit-test their packages in parallel without waiting on each other's implementations, define lightweight Go interfaces at the consumer boundary:
    1. `SMSDispatcher` (implemented by **Ian** `@frihk` in `internal/notifications`, injected into **Evans**'s `@eojuma` `auth` & `onboarding` and **Ian**'s `@frihk` `whatsapp`).
    2. `InAppNotifier` (implemented by **Ian** `@frihk` in `internal/notifications`, injected into **Ema**'s `@ejao-000` `academics` and **Ian**'s `@frihk` `billing`).
    3. `TermDateFeeSyncer` (implemented by **Ian** `@frihk` in `internal/billing`, injected into **Ema**'s `@ejao-000` `academics` for `PUT /api/v1/calendar`).
    4. `StudentBillingHook` (implemented by **Ian** `@frihk` in `internal/billing`, injected into **Evans**'s `@eojuma` `onboarding` when students are admitted/verified).
    
    ---
    
    ## 6. Shared Resources and Their Owners
    
    | Shared Resource | Single Accountable Owner | Modification Process for Other Developers |
    |---|---|---|
    | `backend/internal/database/` & `backend/internal/middleware/` | **Evans** (`@eojuma`, Dev 3) | Frozen after Milestone 1. Any change requires a PR reviewed by Evans (`@eojuma`), Ema (`@ejao-000`), and Ian (`@frihk`). |
    | `backend/cmd/server/main.go` | **Evans** (`@eojuma`, Dev 3) | Ema (`@ejao-000`) and Ian (`@frihk`) export `RegisterRoutes(mux *http.ServeMux, deps ...)` in their own packages; Evans (`@eojuma`) wires the 1-line calls in `main.go`. |
    | `backend/cmd/worker/main.go` & `backend/internal/jobs/` | **Ian** (`@frihk`, Dev 5) | Evans (`@eojuma`) provides `onboarding.PurgeExpiredUnverifiedSubmissions`; Ian (`@frihk`) registers it in the Asynq cron scheduler. |
    | `backend/internal/internalapi/` | **Ian** (`@frihk`, Dev 5) | Coordinates JSON payloads directly with **Sos** (`@sospeter-57`, Dev 1) and calls read methods on **Ema**'s (`@ejao-000`) `academics.Service`. |
    | `backend/go.mod` & `backend/go.sum` | **Evans** (`@eojuma`, Dev 3) | Evans (`@eojuma`) adds initial core deps (`pgx/v5`, `golang-jwt/jwt/v5`, `asynq`, `go-redis/v9`) in Milestone 1 so `go.mod` stabilizes early. |
    | `docker-compose.yml`, `Makefile`, `.env.example` | **Evans** (`@eojuma`, Dev 3) | Changes require PR approval from affected service owner(s). |
    | `Docs/API_CONTRACTS.md` | Shared Documentation | Any contract adjustment must be updated in `Docs/API_CONTRACTS.md` in the same PR. |
    
    ---
    
    ## 7. Recommended Implementation Order
    
    ### Phase 1: Foundation & Contracts (Days 1–3)
    - **Evans (`@eojuma`):** Delivers `config`, `database` (`WithTenantTx` + `AuditLogger`), `middleware` (`RequireAuth`, `RequirePermission`, `RespondJSON`), `migrations/000002_r1_core_onboarding.up.sql`, and `/api/v1/auth/otp/*`.
    - **Ema (`@ejao-000`):** Authors `migrations/000003_r1_academics.up.sql` and builds `assignments.go`, `grading_scale.go`, and `roster_stats.go`.
    - **Ian (`@frihk`):** Authors `migrations/000004_r1_billing_whatsapp_notifications.up.sql`, implements `notifications` (In-App + Sandbox SMS), and builds `context_signer.go` + `internalapi` skeleton.
    - **Sos (`@sospeter-57`):** Implements `ai_service/app/security.py` (`X-User-Context` HMAC verifier), `backend_client.py`, `llm_client.py` (with deterministic fallback), and session lookup rate limiter.
    - **Ashy (`@Ashomondi`):** Sets up the Tailwind Design System (`tailwind.config.ts`), `src/components/ui/*`, `src/lib/api-client.ts`, `auth-context.tsx`, and `/login` screen.
    
    ### Phase 2: Core Domain Features (Days 4–8)
    - **Evans (`@eojuma`):** Implements School creation, Principal Setup Wizard, Staff QR & 12h Parent QR generation/revocation, Verifier badge delegation, In-person verification + immediate document deletion, Student admission, and Guardian 2FA phone change.
    - **Ema (`@ejao-000`):** Implements Daily Class-Teacher Roll-Call Attendance, Subject Teacher Marks Entry, Marks Reassignment, and Timetables/Calendar (with Term Start-Date change hook).
    - **Ian (`@frihk`):** Implements Fee structure loader, Automated start-of-year/term charging, Student fee balance/ledger (with credit balance), Mid-year fee changes, and Per-school WhatsApp webhook + 6h OTP session state machine.
    - **Sos (`@sospeter-57`):** Implements `parent_assistant.py` (In-App & WhatsApp minimal-fact Q&A + school contact fallback) and `chat_payment.py` (multi-turn slot extraction + explicit `"YES"` confirmation).
    - **Ashy (`@Ashomondi`):** Implements Principal Setup Wizard (`/admin/wizard`), QR Onboarding & Verification screens (`/admin/onboarding`, `/register/[token]`, `/admin/verifications`), Dean workspace (`/admin/dean`), and Teacher Attendance & Marks screens (`/teacher/*`).
    
    ### Phase 3: Advanced State Machines & End-to-End Integration (Days 9–12)
    - **Ema (`@ejao-000`):** Completes Hard-Gated Results Release, Snapshot-Preserving Reopen & Targeted Re-Release state machine, AI Report Comment Drafter proxy, and Student CBC Performance Graph endpoint.
    - **Ian (`@frihk`):** Completes Daraja Sandbox STK Push (with 24h phone-change freeze check), Direct Paybill C2B auto-matching, Unmatched payments resolution, and Asynq background workers (`cmd/worker`).
    - **Sos (`@sospeter-57`):** Completes `report_drafter.py` (`term_comment` and `rerelease_note` modes with anti-ranking/anti-diagnosis guardrails) and full `pytest` suite.
    - **Ashy (`@Ashomondi`):** Completes Teacher Results Release/Reopen UI (`/teacher/results`), Parent Portal (`/parent/*` including `CbcPerformanceChart.tsx`, Fee Ledger & STK Push modal, In-App AI Assistant drawer, 2FA Phone Change, and R2 placeholders).
    
    ---
    
    ## 8. Integration Milestones
    
    | Milestone | Target | Verification Gate |
    |---|---|---|
    | **M1: Schema & Auth Spine Ready** | End of Day 3 | `make migrate-up` applies migrations `000001`–`000004` cleanly. `POST /api/v1/auth/otp/request` & `verify` return valid JWTs. `WithTenantTx` enforces RLS (`AC-27`). |
    | **M2: Release 1 Pass Condition Gate (`C53`)** | End of Day 7 | School creation, Principal Setup Wizard (`AC-01`), Staff QR (`AC-02`), 12h Parent QR (`AC-03`), Pending isolation (`AC-04`), Verifier delegation (`AC-05`), Document deletion (`AC-06`), App OTP login, and WhatsApp 6h OTP session (`AC-08`–`AC-12`) all work end-to-end with Ashy's (`@Ashomondi`) UI. |
    | **M3: Academics & Billing Gate** | End of Day 10 | Roll-call (`AC-14`), Marks entry & reassignment (`AC-15`, `AC-17`), Hard-gated release (`AC-16`), Reopen/re-release (`AC-18`), CBC Graph (`AC-19`), Term fee charging (`AC-20`), Mid-year fee changes (`AC-21`), Term-date shift (`AC-22`), Overpayment credit (`AC-24`), and C2B matching (`AC-25`) pass integration tests. |
    | **M4: Full AI, WhatsApp & PWA Sign-Off** | End of Day 12 | In-App & WhatsApp Parent Assistant (`AC-13`), Two-step confirmed Chat STK Push (`AC-23`), 24h Phone-Change Payment Lock (`AC-26`), AI Report Comment Drafter, and AI Audit Logging (`AC-28`, `AC-29`) pass full QA verification on synthetic data. |
    
    ---
    
    ## 9. Testing and Quality Assurance Strategy
    
    1. **Synthetic Data Mandate (`Docs/Ikonex.md` C52, `AGENTS.md` Rule 8):**
       - Never use real student, parent, or staff names, national IDs, or phone numbers in seed scripts, tests, or demo environments.
       - Provide a deterministic synthetic seed script (`backend/cmd/seed` or SQL seed fixture) creating two synthetic schools (*"Ikonex Demo High School A"* and *"Ikonex Demo Junior School B"*) to test cross-tenant RLS isolation (`AC-27`, `AC-12`) and all role/badge combinations.
    2. **Automated Test Suites (`make test`):**
       - **Go Backend (`Evans @eojuma`, `Ema @ejao-000`, `Ian @frihk`):** `cd backend && go test -v -race ./...` must pass all unit and database integration tests.
       - **Python AI Service (`Sos @sospeter-57`):** `cd ai_service && pytest -v` must pass all security, rate-limit, hallucination-guard, WhatsApp minimal-fact, and multi-turn payment confirmation tests.
       - **Next.js Frontend (`Ashy @Ashomondi`):** `cd frontend && npm run lint && npm run build` must succeed with zero errors.
    3. **Acceptance Criteria Traceability:**
       - Every single acceptance criterion `AC-01` through `AC-29` in `Docs/Ikonex.md` Section 14 must be mapped to an automated test or documented QA verification step before Release 1 sign-off.
    
    ---
    
    ## 10. Git Branching and Pull Request Workflow
    
    ### 10.1 Branch Naming Convention
    All work branches off `main` (or integration branch `release-1`) using developer-prefixed feature branches:
    - **Sos (`@sospeter-57`, Dev 1):** `feat/sos-ai-core`, `feat/sos-parent-assistant`, `feat/sos-report-drafter`
    - **Ashy (`@Ashomondi`, Dev 2):** `feat/ashy-design-auth`, `feat/ashy-onboarding-ui`, `feat/ashy-teacher-portal`, `feat/ashy-parent-portal`
    - **Evans (`@eojuma`, Dev 3):** `feat/evans-rls-auth-badges`, `feat/evans-wizard-staff-qr`, `feat/evans-parent-qr-verify`
    - **Ema (`@ejao-000`, Dev 4):** `feat/ema-dean-calendar`, `feat/ema-attendance-stats`, `feat/ema-marks-results-release`
    - **Ian (`@frihk`, Dev 5):** `feat/ian-fee-ledgers`, `feat/ian-mpesa-daraja`, `feat/ian-whatsapp-internalapi`
    
    ### 10.2 Commit & Pull Request Rules
    1. **Conventional Commits:** Use `feat(scope): ...`, `fix(scope): ...`, `test(scope): ...`, `docs(scope): ...` (e.g., `feat(academics): enforce hard gate on class results release`).
    2. **Directory Boundary Check:** A PR must only touch files within the author's owned paths (Section 9 of each developer's task file). If a shared file (`docker-compose.yml`, `main.go`, `API_CONTRACTS.md`) is modified, the owner of that shared resource must approve the PR.
    3. **CI / Pre-Merge Requirement:** Before merging any PR, `make test` and `frontend` build checks must pass, and any new endpoint must match `Docs/API_CONTRACTS.md`.
    
    ---
    
    ## 11. Open Decisions & Assumptions Handling for Release 1
    
    Per `Docs/Ikonex.md` Appendices A & B and `AGENTS.md` Rule 10 (*"Don't guess open decisions"*), developers must implement the following configurable defaults so no open question blocks Release 1 or hardcodes an irreversible assumption:
    
    | ID | Open Question / Assumption | How to Handle Cleanly in Release 1 Code | Owner |
    |---|---|---|---|
    | **O9 / A12** | How a shared Parent QR links each parent to the right child | Allow the parent registration form to optionally accept `student_admission_number` (when the class-teacher route already created the student, flagging duplicates per `A12`) OR new student details, and allow the in-person staff verifier to confirm/select the linked student during `POST /api/v1/registrations/{id}/verify`. | **Evans** (`@eojuma`) & **Ashy** (`@Ashomondi`) |
    | **O16** | Whether Principal/Deputy can override the Dean on assignments, timetables, and calendar | Gate Dean endpoints via permission codes (`academics.assignments.manage`, `academics.timetables.manage`, `academics.calendar.manage`) seeded to `dean_of_studies` by default, with a boolean config flag `ALLOW_PRINCIPAL_DEAN_OVERRIDE=true` in `.env.example` so behavior can be toggled without code changes. | **Ema** (`@ejao-000`) & **Evans** (`@eojuma`) |
    | **O17** | Days before rejected/never-verified submissions are deleted | Controlled via environment variable `UNVERIFIED_SUBMISSION_RETENTION_DAYS=7` in `.env.example` rather than a hardcoded SQL constant. | **Evans** (`@eojuma`) & **Ian** (`@frihk`) |
    | **O20 / A3** | How OTP limits apply to app login & per admission number | Enforce the `1 OTP/min` and `>5 in 30 mins -> 24h sender block` in a configurable rate-limiter middleware used by both WhatsApp (`C18–C20`) and App login (`O20`), configurable via `OTP_COOLDOWN_SECONDS=60` and `OTP_MAX_PER_30_MIN=5`. | **Evans** (`@eojuma`) & **Ian** (`@frihk`) |
    | **O13 / C47** | Fee structure variants (boarder vs. day scholar) | Build `fee_periods` on generic billing periods and include an optional `boarding_status VARCHAR(20) DEFAULT 'all'` filter on `fee_structures` so Release 1 works with standard term fees while supporting boarder/day variants without schema changes. | **Ian** (`@frihk`) |
    
    ---
    
    ## 12. Final Release 1 Integration Checklist
    
    - [ ] **Migrations (`000001`–`000004`):** All tables created, `FORCE ROW LEVEL SECURITY` enabled on every tenant table, and `AC-27` verified.
    - [ ] **Pass Condition 1 — Onboarding (`C53`):** School + Principal provisioning (`AC-01`), Principal Setup Wizard (`AC-01`), Staff QR (`AC-02`), 12h Parent QR (`AC-03`), Pending submission isolation (`AC-04`), Staff verifier delegation (`AC-05`), Immediate document deletion on approval (`AC-06`), and Unverified submission auto-purge (`AC-07`) verified end-to-end.
    - [ ] **Pass Condition 2 — Parent Login & WhatsApp Sessions (`C53`):** App SMS OTP login (`AC-01`), WhatsApp chat initiation from any number (`AC-08`), Masked guardian phone selection (`AC-09`), OTP rate limits & 24h sender-only block (`AC-10`), 6-hour session expiry (`AC-11`), and Per-school WhatsApp bot isolation (`AC-12`) verified end-to-end.
    - [ ] **Pass Condition 3 — Badges & Permissions (`C53`):** Scoped `class_teacher` badge (`AC-14`), Time-bound `delegated_verifier` badge (`AC-05`), Subject teacher assignment & Dean reassignment (`AC-15`, `AC-17`), and Anonymised other-class means verified end-to-end.
    - [ ] **Academics & Results State Machine:** Hard-gated results release (`AC-16`), Snapshot-preserving reopen & targeted re-release with required teacher note (`AC-18`), Versioned grading scale & non-ranking CBC Performance Graph (`AC-19`), and AI report comment drafting verified end-to-end.
    - [ ] **Billing, M-Pesa & AI Assistant:** Start-of-year term fee charging (`AC-20`), Mid-year fee changes with mandatory note (`AC-21`), Term-date change shifting fee due dates (`AC-22`), Two-step confirmed chat STK push (`AC-23`), Overpayment ledger credit (`AC-24`), Direct Paybill C2B matching & unmatched queue (`AC-25`), 24h phone-change chat payment pause (`AC-26`), WhatsApp minimal-fact restriction (`AC-13`), Zero AI DB access & AI action logging (`AC-28`), and complete `audit_log` coverage (`AC-29`) verified end-to-end.
    - [ ] **Full Stack Containerization:** `docker compose up --build` starts `postgres`, `redis`, `backend`, `ai_service`, and `frontend` cleanly with zero port or contract conflicts.
    ✓ • 759ms

## Activity

    $ rg --files -g '!package-lock.json' -g '!*.svg' -g '!*.png' -g '!*.jpg' -g '!*.pdf' -g '!*.docx' && cat AGENTS.md Docs/AGENTS.md backend/go.mod
    ai_service/requirements.txt
    ai_service/Dockerfile
    ai_service/app/clients/backend_client.py
    ai_service/app/clients/llm_client.py
    ai_service/app/api/reconciliation.py
    ai_service/app/api/payment_reminders.py
    ai_service/app/api/parent_assistant.py
    ai_service/app/api/chat_payment.py
    ai_service/app/api/syllabus_mapper.py
    ai_service/app/api/blog_safety.py
    ai_service/app/api/report_drafter.py
    ai_service/app/main.py
    ai_service/app/config.py
    Docs/AGENTS.md
    Docs/ARCHITECTURE.md
    Docs/DECISIONS.md
    Docs/REQUIREMENTS.md
    Docs/Ikonex.md
    Docs/API_CONTRACTS.md
    Docs/SECURITY.md
    README.md
    tasks/MASTER-INTEGRATION-PLAN.md
    tasks/developer-02-frontend-engineer.md
    tasks/developer-04-backend-engineer-b.md
    tasks/developer-05-backend-engineer-c.md
    tasks/developer-01-ai-engineer.md
    tasks/developer-03-backend-engineer-a.md
    frontend/tailwind.config.ts
    frontend/postcss.config.mjs
    frontend/package.json
    frontend/tsconfig.json
    Makefile
    migrations/000001_init_schema.down.sql
    migrations/000001_init_schema.up.sql
    frontend/public/manifest.json
    frontend/next.config.mjs
    AGENTS.md
    docker-compose.yml
    backend/cmd/worker/main.go
    backend/cmd/server/main.go
    backend/Dockerfile
    backend/go.mod
    backend/internal/academics/academics.go
    backend/internal/auth/auth.go
    backend/internal/nurse/nurse.go
    backend/internal/middleware/middleware.go
    backend/internal/internalapi/internalapi.go
    backend/internal/consent/consent.go
    backend/internal/notifications/notifications.go
    backend/internal/database/database.go
    backend/internal/jobs/jobs.go
    backend/internal/billing/billing.go
    backend/internal/onboarding/onboarding.go
    backend/internal/clubs/clubs.go
    backend/internal/library/library.go
    backend/internal/config/config.go
    frontend/src/app/page.tsx
    frontend/src/app/layout.tsx
    frontend/src/app/globals.css
    frontend/src/app/admin/page.tsx
    frontend/src/app/nurse/page.tsx
    frontend/src/app/bursar/page.tsx
    frontend/src/app/librarian/page.tsx
    frontend/src/app/parent/page.tsx
    frontend/src/app/teacher/page.tsx
    frontend/src/app/(auth)/login/page.tsx
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
    filePath: /home/banker/code/open/Ikonex-Academy/AGENTS.md# AGENTS.md
    
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
    module github.com/Ashomondi/ikonex-academy/backend
    
    go 1.22.2
    ✓ • 0ms

## Activity

    $ sed -n '1,310p' migrations/000001_init_schema.up.sql
    -- ==============================================================================
    -- 000001_init_schema.up.sql
    -- Ikonex Academy - Master PostgreSQL Schema with Row-Level Security (RLS)
    -- ==============================================================================
    
    CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
    CREATE EXTENSION IF NOT EXISTS "pgcrypto";
    
    -- Function to return current tenant school_id from session context
    CREATE OR REPLACE FUNCTION current_school_id() RETURNS uuid AS $$
        SELECT NULLIF(current_setting('app.school_id', true), '')::uuid;
    $$ LANGUAGE sql STABLE;
    
    -- ------------------------------------------------------------------------------
    -- Global Tables (Non-Tenant Scoped)
    -- ------------------------------------------------------------------------------
    
    CREATE TABLE schools (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
        code VARCHAR(50) UNIQUE NOT NULL,
        name VARCHAR(255) NOT NULL,
        county VARCHAR(100),
        logo_url TEXT,
        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
        updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
    );
    
    CREATE TABLE users (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
        phone_number VARCHAR(20) UNIQUE NOT NULL,
        full_name VARCHAR(255) NOT NULL,
        national_id VARCHAR(50),
        email VARCHAR(255),
        is_active BOOLEAN NOT NULL DEFAULT TRUE,
        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
        updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
    );
    
    CREATE TABLE permissions (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
        code VARCHAR(100) UNIQUE NOT NULL,
        name VARCHAR(255) NOT NULL,
        category VARCHAR(100) NOT NULL,
        description TEXT
    );
    
    -- ------------------------------------------------------------------------------
    -- Tenant Tables (Isolated via school_id and RLS)
    -- ------------------------------------------------------------------------------
    
    CREATE TABLE memberships (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
        school_id UUID NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
        user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
        role VARCHAR(50) NOT NULL, -- e.g. principal, deputy, teacher, bursar, nurse, librarian, dean_of_studies, parent
        status VARCHAR(50) NOT NULL DEFAULT 'active',
        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
        updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
        UNIQUE (school_id, user_id, role)
    );
    
    CREATE TABLE badges (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
        school_id UUID NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
        name VARCHAR(100) NOT NULL,
        slug VARCHAR(100) NOT NULL,
        scope_type VARCHAR(50) NOT NULL DEFAULT 'none', -- none, class, club, term, window
        description TEXT,
        UNIQUE (school_id, slug)
    );
    
    CREATE TABLE badge_permissions (
        badge_id UUID NOT NULL REFERENCES badges(id) ON DELETE CASCADE,
        permission_id UUID NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
        PRIMARY KEY (badge_id, permission_id)
    );
    
    CREATE TABLE membership_badges (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
        membership_id UUID NOT NULL REFERENCES memberships(id) ON DELETE CASCADE,
        badge_id UUID NOT NULL REFERENCES badges(id) ON DELETE CASCADE,
        scope_id UUID, -- class_id, club_id, etc.
        valid_from TIMESTAMPTZ,
        valid_until TIMESTAMPTZ,
        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
    );
    
    CREATE TABLE academic_years (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
        school_id UUID NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
        year INT NOT NULL,
        is_current BOOLEAN NOT NULL DEFAULT FALSE,
        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
        UNIQUE (school_id, year)
    );
    
    CREATE TABLE terms (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
        school_id UUID NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
        academic_year_id UUID NOT NULL REFERENCES academic_years(id) ON DELETE CASCADE,
        term_number INT NOT NULL CHECK (term_number BETWEEN 1 AND 3),
        start_date DATE NOT NULL,
        end_date DATE NOT NULL,
        is_current BOOLEAN NOT NULL DEFAULT FALSE,
        UNIQUE (school_id, academic_year_id, term_number)
    );
    
    CREATE TABLE grade_levels (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
        school_id UUID NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
        number INT NOT NULL CHECK (number BETWEEN 7 AND 12), -- Kenya CBC Junior (7-9) & Senior (10-12)
        phase VARCHAR(20) NOT NULL CHECK (phase IN ('junior', 'senior')),
        name VARCHAR(50) NOT NULL,
        UNIQUE (school_id, number)
    );
    
    CREATE TABLE classes (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
        school_id UUID NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
        grade_level_id UUID NOT NULL REFERENCES grade_levels(id) ON DELETE RESTRICT,
        stream_name VARCHAR(100) NOT NULL,
        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
        UNIQUE (school_id, grade_level_id, stream_name)
    );
    
    CREATE TABLE dormitories (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
        school_id UUID NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
        name VARCHAR(100) NOT NULL,
        gender VARCHAR(20) NOT NULL,
        capacity INT NOT NULL DEFAULT 0
    );
    
    CREATE TABLE clubs (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
        school_id UUID NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
        name VARCHAR(100) NOT NULL,
        description TEXT
    );
    
    CREATE TABLE students (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
        school_id UUID NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
        admission_number VARCHAR(100) NOT NULL,
        full_name VARCHAR(255) NOT NULL,
        dob DATE NOT NULL,
        gender VARCHAR(20) NOT NULL,
        current_class_id UUID REFERENCES classes(id) ON DELETE SET NULL,
        current_dormitory_id UUID REFERENCES dormitories(id) ON DELETE SET NULL,
        status VARCHAR(50) NOT NULL DEFAULT 'active',
        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
        updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
        UNIQUE (school_id, admission_number)
    );
    
    CREATE TABLE student_guardians (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
        school_id UUID NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
        student_id UUID NOT NULL REFERENCES students(id) ON DELETE CASCADE,
        user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
        relationship VARCHAR(50) NOT NULL,
        is_primary BOOLEAN NOT NULL DEFAULT FALSE,
        is_fee_payer BOOLEAN NOT NULL DEFAULT TRUE,
        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
        UNIQUE (school_id, student_id, user_id)
    );
    
    CREATE TABLE enrolments (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
        school_id UUID NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
        student_id UUID NOT NULL REFERENCES students(id) ON DELETE CASCADE,
        academic_year_id UUID NOT NULL REFERENCES academic_years(id) ON DELETE RESTRICT,
        grade_level_id UUID NOT NULL REFERENCES grade_levels(id) ON DELETE RESTRICT,
        class_id UUID NOT NULL REFERENCES classes(id) ON DELETE RESTRICT,
        status VARCHAR(50) NOT NULL DEFAULT 'active',
        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
        UNIQUE (school_id, student_id, academic_year_id)
    );
    
    CREATE TABLE consents (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
        school_id UUID NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
        student_id UUID NOT NULL REFERENCES students(id) ON DELETE CASCADE,
        guardian_user_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
        purpose VARCHAR(50) NOT NULL CHECK (purpose IN ('biometric_attendance', 'photo', 'sms')),
        granted BOOLEAN NOT NULL DEFAULT FALSE,
        timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
        revoked_at TIMESTAMPTZ
    );
    
    CREATE TABLE biometric_templates (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
        school_id UUID NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
        student_id UUID NOT NULL REFERENCES students(id) ON DELETE CASCADE,
        template_data BYTEA NOT NULL, -- Raw images never stored, encrypted templates only
        algorithm VARCHAR(50) NOT NULL,
        key_version INT NOT NULL DEFAULT 1,
        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
        updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
        UNIQUE (school_id, student_id)
    );
    
    CREATE TABLE staff_invites (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
        school_id UUID NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
        phone_number VARCHAR(20) NOT NULL,
        role VARCHAR(50) NOT NULL,
        qr_token VARCHAR(255) UNIQUE NOT NULL,
        valid_from TIMESTAMPTZ NOT NULL,
        valid_to TIMESTAMPTZ NOT NULL,
        status VARCHAR(50) NOT NULL DEFAULT 'pending',
        created_by UUID REFERENCES users(id),
        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
    );
    
    CREATE TABLE fee_structures (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
        school_id UUID NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
        academic_year_id UUID NOT NULL REFERENCES academic_years(id) ON DELETE RESTRICT,
        grade_level_id UUID NOT NULL REFERENCES grade_levels(id) ON DELETE RESTRICT,
        term_id UUID NOT NULL REFERENCES terms(id) ON DELETE RESTRICT,
        total_amount NUMERIC(12, 2) NOT NULL DEFAULT 0.00,
        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
        UNIQUE (school_id, academic_year_id, grade_level_id, term_id)
    );
    
    CREATE TABLE fee_structure_items (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
        fee_structure_id UUID NOT NULL REFERENCES fee_structures(id) ON DELETE CASCADE,
        title VARCHAR(255) NOT NULL,
        amount NUMERIC(12, 2) NOT NULL,
        is_compulsory BOOLEAN NOT NULL DEFAULT TRUE
    );
    
    CREATE TABLE fee_ledgers (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
        school_id UUID NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
        student_id UUID NOT NULL REFERENCES students(id) ON DELETE CASCADE,
        academic_year_id UUID NOT NULL REFERENCES academic_years(id) ON DELETE RESTRICT,
        term_id UUID NOT NULL REFERENCES terms(id) ON DELETE RESTRICT,
        opening_balance NUMERIC(12, 2) NOT NULL DEFAULT 0.00,
        billed_amount NUMERIC(12, 2) NOT NULL DEFAULT 0.00,
        paid_amount NUMERIC(12, 2) NOT NULL DEFAULT 0.00,
        current_balance NUMERIC(12, 2) NOT NULL DEFAULT 0.00,
        updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
        UNIQUE (school_id, student_id, academic_year_id, term_id)
    );
    
    CREATE TABLE payments (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
        school_id UUID NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
        student_id UUID REFERENCES students(id) ON DELETE SET NULL,
        transaction_ref VARCHAR(100) UNIQUE NOT NULL,
        provider VARCHAR(50) NOT NULL, -- mpesa_stk, mpesa_c2b, bank
        amount NUMERIC(12, 2) NOT NULL,
        payer_phone VARCHAR(20),
        payer_name VARCHAR(255),
        matched_status VARCHAR(50) NOT NULL DEFAULT 'unmatched', -- unmatched, matched, ai_suggested
        matched_by UUID REFERENCES users(id),
        payment_date TIMESTAMPTZ NOT NULL DEFAULT NOW(),
        raw_payload JSONB
    );
    
    CREATE TABLE payment_plans (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
        school_id UUID NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
        student_id UUID NOT NULL REFERENCES students(id) ON DELETE CASCADE,
        guardian_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
        pledged_amount NUMERIC(12, 2) NOT NULL,
        pledge_frequency VARCHAR(50) NOT NULL DEFAULT 'monthly',
        next_due_date DATE NOT NULL,
        default_count INT NOT NULL DEFAULT 0,
        status VARCHAR(50) NOT NULL DEFAULT 'active'
    );
    
    CREATE TABLE nurse_visits (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
        school_id UUID NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
        student_id UUID NOT NULL REFERENCES students(id) ON DELETE CASCADE,
        visit_date TIMESTAMPTZ NOT NULL DEFAULT NOW(),
        clinical_notes TEXT NOT NULL, -- Strictly restricted to nurse role; never shared with teachers
        recommendation TEXT,
        notified_class_teacher BOOLEAN NOT NULL DEFAULT FALSE,
        logged_by UUID NOT NULL REFERENCES users(id)
    );
    
    CREATE TABLE library_books (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
        school_id UUID NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
        isbn VARCHAR(50),
        title VARCHAR(255) NOT NULL,
        author VARCHAR(255) NOT NULL,
        total_copies INT NOT NULL DEFAULT 1,
        available_copies INT NOT NULL DEFAULT 1
    );
    
    CREATE TABLE book_loans (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
        school_id UUID NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
        book_id UUID NOT NULL REFERENCES library_books(id) ON DELETE CASCADE,
        student_id UUID NOT NULL REFERENCES students(id) ON DELETE CASCADE,
        borrowed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
        due_date DATE NOT NULL,
        returned_at TIMESTAMPTZ,
        status VARCHAR(50) NOT NULL DEFAULT 'borrowed'
    );
    
    CREATE TABLE cbc_strands (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
        school_id UUID NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
    ✓ • 18ms

## Activity

    $ cat backend/internal/database/database.go backend/internal/middleware/middleware.go backend/internal/config/config.go backend/cmd/server/main.go backend/cmd/worker/main.go backend/internal/billing/billing.go backend/internal/notifications/notifications.go backend/internal/jobs/jobs.go backend/internal/internalapi/internalapi.go backend/internal/auth/auth.go backend/internal/onboarding/onboarding.go backend/internal/academics/academics.go
    package database
    
    // DB handles PostgreSQL connection pooling and Row-Level Security (RLS) enforcement.
    type DB struct{}
    package middleware
    
    // Middleware definitions for tenant RLS injection, authentication, and badge permissions.
    type Middleware struct{}
    package config
    
    // Config placeholder for environment and application configuration
    type Config struct{}
    package main
    
    // Server entry point placeholder
    func main() {}
    package main
    
    // Background worker entry point placeholder
    func main() {}
    package billing
    
    // Service handles fee structures, student ledgers, M-Pesa STK push payments, and reconciliation.
    type Service struct{}
    package notifications
    
    // Service handles SMS dispatch (Africa's Talking), WhatsApp integration, and in-app notifications.
    type Service struct{}
    package jobs
    
    // Worker and queue manager for Asynq tasks (SMS jobs, STK polling, AI reminders).
    type QueueManager struct{}
    package internalapi
    
    // Handler exposes permission-checked endpoints accessible strictly by the AI service.
    // The AI service calls these endpoints to query facts and trigger actions without direct database access.
    type Handler struct{}
    package auth
    
    // Auth handles phone number OTP authentication, sessions, and user memberships.
    type Service struct{}
    package onboarding
    
    // Service handles staff invites via QR codes, principal/deputy onboarding, student admissions, and parent self-registration.
    type Service struct{}
    package academics
    
    // Service handles CBC curricula (Grade 7-12), timetable, syllabus coverage, and class performance stats.
    type Service struct{}
    ✓ • 10.58s

## Activity

    $ git diff -- Docs/DECISIONS.md && git log -5 --oneline && git symbolic-ref refs/remotes/origin/HEAD && git diff --cached --name-only
    diff --git a/Docs/DECISIONS.md b/Docs/DECISIONS.md
    index a9a94b6..ac82a09 100644
    --- a/Docs/DECISIONS.md
    +++ b/Docs/DECISIONS.md
    @@ -1,3 +1,4 @@
    +
     # DECISIONS.md
    
     A short, dated log of decisions worth remembering — so nobody (human or AI
    b2b2e3b Merge pull request #6 from Ashomondi/docs/add-agents-md
    f244d74 docs: add root AGENTS.md for AI coding agents
    cd3374b detailed working methodology
    de5816f updated docs
    1147e5b Intial commit.
    fatal: ref refs/remotes/origin/HEAD is not a symbolic ref
    ✗ (128) • 0ms

## Activity

    $ cat Docs/API_CONTRACTS.md Docs/ARCHITECTURE.md Docs/SECURITY.md
    # API_CONTRACTS.md
    
    How the frontend, Go backend, and AI service agree on interfaces. Update this
    as real endpoints are built — this is meant to track actual contracts, not
    aspirational ones. Once the project has a real number of endpoints, consider
    generating this from an OpenAPI spec instead of maintaining it by hand.
    
    ## Auth model
    
    - Every authenticated request carries the caller's identity (user + active
      membership) and resolves to a `school_id`.
    - The Go backend sets `app.school_id` for the request's database transaction
      so RLS does the enforcement — see `ARCHITECTURE.md`.
    - Every endpoint declares which permission(s) it requires. A request without
      the right permission (via the caller's badges) is rejected before it
      reaches any handler logic — not filtered afterwards.
    - Parent-facing endpoints additionally scope to "children this parent is a
      guardian of" — never just "this school."
    
    ## Rule: the AI service never touches the database directly
    
    The AI service (Python/FastAPI) has no direct database credentials. It:
    
    1. Receives a request already carrying the caller's identity and permission
       context (forwarded by the Go backend, not re-derived by the AI service).
    2. Calls specific, permission-checked Go endpoints to fetch facts (fee
       balance, attendance record, syllabus status, etc.).
    3. Uses the model only to phrase, translate, summarise, or map — never to
       originate a fact that should come from the database.
    4. Any action with a real-world effect (starting an STK push, updating
       syllabus coverage, posting a reminder) calls a Go endpoint to perform it;
       the AI service does not write to storage on its own.
    
    This is the enforcement point for "AI never invents a fee balance" and "AI
    never handles money directly" from `AGENTS.md` and `SECURITY.md`.
    
    ## AI service authentication (decided 2026-10-29)
    
    The AI service authenticates to the Go backend with a **service token plus a
    short-lived signed user context** (user, school, membership, and, for
    WhatsApp, the verified session). Every call carries both; a valid token
    without a valid, unexpired context is rejected. A hybrid with mutual TLS may
    be added later.
    
    ## Rate limits
    
    - **OTP:** one request per minute; more than five requests in thirty minutes
      blocks the *requesting WhatsApp number* for 24 hours (not the guardian's
      account). Applying the same limits to app login is undecided.
    - **AI-facing endpoints:** limits are on backend lookups per verified
      session (fee-balance, attendance, payment lookups), with a tighter limit on
      payment starts. Numbers are set by the team and QA. Assistant messages
      themselves are not capped.
    
    ## Endpoint groups (fill in as built)
    
    > Each group below should end up as a real section with method, path,
    > request/response shape, and required permission. Skeleton only for now.
    > No endpoint exists yet; this tracks the contracts agreed so far.
    
    ### Auth & onboarding
    - `POST /auth/otp/request`, `POST /auth/otp/verify`
    - `POST /staff-invites` (permission: `staff.invites.create`)
    - `POST /staff-applications/{id}/review` (permission:
      `staff.applications.review`; verifier checks documents in person)
    - `POST /parent-qr` (generate, valid 12 hours) and
      `POST /parent-qr/{id}/revoke` (permission: `parents.qr.manage`; dean,
      principal, deputy)
    - `POST /registrations` (parent or staff submission via QR; creates a
      *pending* record with no data access)
    - `POST /registrations/{id}/verify` (permission: `registrations.verify`;
      held by staff verifiers, delegable for the QR's active window)
    - `POST /verifier-grants` (permission: `verifier.grant`; staff only,
      expires with the QR)
    - `POST /students/admit` (permission: `students.admit`)
    - `POST /guardians/me/phone-change` (2FA; notifies the old number; pauses
      chat payments for 24 hours)
    
    ### WhatsApp & parent sessions
    - `POST /whatsapp/webhook` (receiving number resolves the school)
    - `POST /chat/sessions` (admission number → masked guardian numbers)
    - `POST /chat/sessions/{id}/otp` (OTP sent only to the chosen registered
      number); sessions last 6 hours
    
    ### Academics
    - `GET /classes/{id}/roster` (permission: `class.students.view`)
    - `GET /classes/{id}/stats` (permission: `class.stats.view` — means only, no
      names)
    - `POST /classes/{id}/attendance` (class teacher; one daily roll-call)
    - `PUT /classes/{id}/subjects/{sid}/marks` (assigned subject teacher only)
    - `POST /classes/{id}/marks-assignments` (permission:
      `marks.reassign`; dean, principal, deputy)
    - `POST /classes/{id}/results/release` (class teacher; rejected unless every
      subject has marks for every student)
    - `POST /classes/{id}/results/reopen` and `.../re-release` (class teacher;
      re-release requires a teacher's note per corrected mark; notifies only
      affected parents)
    - `POST /report-comments/draft` (AI service drafts; teacher edits and
      publishes)
    - `GET/PUT /timetables`, `GET/PUT /calendar` (dean; term-date changes also
      trigger the fee-change notice)
    - `POST /syllabus/coverage` — called by AI service after mapping a teacher's
      natural-language input to CBC strands/sub-strands (Release 2)
    - `GET /students/{id}/performance-history` (parent/teacher, scoped; returns
      CBC levels per subject per term, never rank)
    
    ### Fees & payments
    - `GET /students/{id}/fee-balance` (parent-scoped, or Ikonex team in
      Release 1)
    - `POST /payments/stk-push` — AI service calls this only after the parent has
      confirmed child, amount, and the typed M-Pesa number; any amount accepted,
      excess held as credit
    - `POST /fee-changes` (Ikonex team in Release 1; note required; notifies
      only affected parents)
    - `GET /payments/unmatched` (Ikonex team in Release 1; bursar in Release 2)
      — feeds the AI reconciliation feature in Release 2
    - `POST /payments/{id}/match` (team in Release 1; bursar confirms an
      AI-suggested match in Release 2)
    - `POST /payment-plans/{id}/reminder-check` — Release 2; AI service reads
      pledge vs. actual payments and arrears, triggers a reminder job (does not
      send directly — enqueues via the notification system)
    
    ### Notifications
    - `POST /notifications/send` (internal use by jobs, not directly callable by
      AI service without going through a job). Release 1 event notifications are
      in-app only; OTPs and the old-number SMS are transactional.
    
    ### AI-service-facing (internal only — never exposed to frontend directly)
    - `POST /internal/ai/parent-query` — resolves a parent's natural-language
      question by calling the appropriate read endpoints above. On WhatsApp it
      returns only minimal facts and a link for anything else.
    - `POST /internal/ai/blog-safety-check` — checks a draft post for children's
      personal details before a patron publishes (Release 2)
    
    ## Response conventions
    
    > Fill in once decided: error shape, pagination style, date/time format
    > (recommend ISO 8601, always with the school's timezone made explicit),
    > versioning approach.
    
    ## What's not yet decided
    
    - How OTP limits and blocks apply to app login.
    - Response conventions above.
    - Exact lookup-limit numbers for AI-facing endpoints.
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
    ✓ • 3.09s

## Activity

    $ sed -n '310,520p' migrations/000001_init_schema.up.sql && cat migrations/000001_init_schema.down.sql Makefile docker-compose.yml .env.example frontend/package.json ai_service/app/clients/backend_client.py ai_service/app/api/parent_assistant.py && rg -n 'WithTenantTx|AuditLogger|SMSDispatcher|StudentChargeHook|StudentBillingHook|sender_rate_blocks|otp_requests|PurgeExpired|OnTermStartDateChanged' tasks/developer-03-backend-engineer-a.md tasks/developer-04-backend-engineer-b.md
        school_id UUID NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
        grade_level_id UUID NOT NULL REFERENCES grade_levels(id) ON DELETE CASCADE,
        subject VARCHAR(100) NOT NULL,
        strand_name VARCHAR(255) NOT NULL,
        sub_strand_name VARCHAR(255) NOT NULL,
        learning_outcomes TEXT
    );
    
    CREATE TABLE syllabus_coverage (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
        school_id UUID NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
        class_id UUID NOT NULL REFERENCES classes(id) ON DELETE CASCADE,
        strand_id UUID NOT NULL REFERENCES cbc_strands(id) ON DELETE CASCADE,
        teacher_user_id UUID NOT NULL REFERENCES users(id),
        status VARCHAR(50) NOT NULL DEFAULT 'in_progress',
        notes TEXT,
        updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
        UNIQUE (school_id, class_id, strand_id)
    );
    
    CREATE TABLE student_assessments (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
        school_id UUID NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
        student_id UUID NOT NULL REFERENCES students(id) ON DELETE CASCADE,
        term_id UUID NOT NULL REFERENCES terms(id) ON DELETE RESTRICT,
        strand_id UUID NOT NULL REFERENCES cbc_strands(id) ON DELETE CASCADE,
        performance_level VARCHAR(10) NOT NULL CHECK (performance_level IN ('EE', 'ME', 'AE', 'BE')), -- CBC
        teacher_comments TEXT,
        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
    );
    
    CREATE TABLE blog_posts (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
        school_id UUID NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
        club_id UUID REFERENCES clubs(id) ON DELETE SET NULL,
        author_membership_id UUID NOT NULL REFERENCES memberships(id) ON DELETE CASCADE,
        title VARCHAR(255) NOT NULL,
        content TEXT NOT NULL,
        safety_checked BOOLEAN NOT NULL DEFAULT FALSE,
        safety_flags JSONB,
        patron_reviewed_by UUID REFERENCES users(id),
        published_at TIMESTAMPTZ,
        status VARCHAR(50) NOT NULL DEFAULT 'draft'
    );
    
    CREATE TABLE audit_log (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
        school_id UUID REFERENCES schools(id) ON DELETE SET NULL,
        user_id UUID REFERENCES users(id) ON DELETE SET NULL,
        action VARCHAR(100) NOT NULL,
        resource_type VARCHAR(100) NOT NULL,
        resource_id VARCHAR(100),
        ip_address VARCHAR(50),
        user_agent TEXT,
        details JSONB,
        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
    );
    
    -- ------------------------------------------------------------------------------
    -- Row-Level Security (RLS) Configuration on all tenant-owned tables
    -- ------------------------------------------------------------------------------
    
    DO $$
    DECLARE
        tbl TEXT;
        tenant_tables TEXT[] := ARRAY[
            'memberships', 'badges', 'academic_years', 'terms', 'grade_levels',
            'classes', 'dormitories', 'clubs', 'students', 'student_guardians',
            'enrolments', 'consents', 'biometric_templates', 'staff_invites',
            'fee_structures', 'fee_ledgers', 'payments', 'payment_plans',
            'nurse_visits', 'library_books', 'book_loans', 'cbc_strands',
            'syllabus_coverage', 'student_assessments', 'blog_posts'
        ];
    BEGIN
        FOREACH tbl IN ARRAY tenant_tables LOOP
            EXECUTE format('ALTER TABLE %I ENABLE ROW LEVEL SECURITY;', tbl);
            EXECUTE format('ALTER TABLE %I FORCE ROW LEVEL SECURITY;', tbl);
            EXECUTE format('CREATE POLICY %I_tenant_isolation ON %I FOR ALL USING (school_id = current_school_id()) WITH CHECK (school_id = current_school_id());', tbl, tbl);
        END LOOP;
    END $$;
    -- ==============================================================================
    -- 000001_init_schema.down.sql
    -- Rollback master schema
    -- ==============================================================================
    
    DROP TABLE IF EXISTS audit_log CASCADE;
    DROP TABLE IF EXISTS blog_posts CASCADE;
    DROP TABLE IF EXISTS student_assessments CASCADE;
    DROP TABLE IF EXISTS syllabus_coverage CASCADE;
    DROP TABLE IF EXISTS cbc_strands CASCADE;
    DROP TABLE IF EXISTS book_loans CASCADE;
    DROP TABLE IF EXISTS library_books CASCADE;
    DROP TABLE IF EXISTS nurse_visits CASCADE;
    DROP TABLE IF EXISTS payment_plans CASCADE;
    DROP TABLE IF EXISTS payments CASCADE;
    DROP TABLE IF EXISTS fee_ledgers CASCADE;
    DROP TABLE IF EXISTS fee_structure_items CASCADE;
    DROP TABLE IF EXISTS fee_structures CASCADE;
    DROP TABLE IF EXISTS staff_invites CASCADE;
    DROP TABLE IF EXISTS biometric_templates CASCADE;
    DROP TABLE IF EXISTS consents CASCADE;
    DROP TABLE IF EXISTS enrolments CASCADE;
    DROP TABLE IF EXISTS student_guardians CASCADE;
    DROP TABLE IF EXISTS students CASCADE;
    DROP TABLE IF EXISTS clubs CASCADE;
    DROP TABLE IF EXISTS dormitories CASCADE;
    DROP TABLE IF EXISTS classes CASCADE;
    DROP TABLE IF EXISTS grade_levels CASCADE;
    DROP TABLE IF EXISTS terms CASCADE;
    DROP TABLE IF EXISTS academic_years CASCADE;
    DROP TABLE IF EXISTS membership_badges CASCADE;
    DROP TABLE IF EXISTS badge_permissions CASCADE;
    DROP TABLE IF EXISTS badges CASCADE;
    DROP TABLE IF EXISTS memberships CASCADE;
    DROP TABLE IF EXISTS permissions CASCADE;
    DROP TABLE IF EXISTS users CASCADE;
    DROP TABLE IF EXISTS schools CASCADE;
    
    DROP FUNCTION IF EXISTS current_school_id();
    .PHONY: help dev up down test backend-run frontend-run ai-run migrate-up migrate-down clean
    
    help:
        @echo "Ikonex Academy - Development Commands:"
        @echo "  make dev          - Run all services via Docker Compose"
        @echo "  make up           - Start Postgres and Redis containers in the background"
        @echo "  make down         - Stop all Docker containers"
        @echo "  make backend-run  - Run Go backend locally"
        @echo "  make frontend-run - Run Next.js frontend locally"
        @echo "  make ai-run       - Run Python AI service locally"
        @echo "  make test         - Run tests across backend and AI service"
        @echo "  make migrate-up   - Apply SQL migrations"
        @echo "  make clean        - Clean temporary build and cache files"
    
    up:
        docker compose up -d postgres redis
    
    down:
        docker compose down
    
    dev:
        docker compose up --build
    
    backend-run:
        cd backend && go run cmd/server/main.go
    
    frontend-run:
        cd frontend && npm run dev
    
    ai-run:
        cd ai_service && uvicorn app.main:app --host 0.0.0.0 --port 8000 --reload
    
    test:
        @echo "Running Go Backend tests..."
        cd backend && go test -v ./...
        @echo "Running Python AI Service tests..."
        cd ai_service && python3 -m pytest tests/ || true
    
    clean:
        rm -rf backend/bin frontend/.next frontend/node_modules ai_service/__pycache__
    services:
      postgres:
        image: postgres:16-alpine
        container_name: ikonex_postgres
        restart: unless-stopped
        environment:
          POSTGRES_USER: ${DB_USER:-ikonex_admin}
          POSTGRES_PASSWORD: ${DB_PASSWORD:-ikonex_secure_pass}
          POSTGRES_DB: ${DB_NAME:-ikonex_academy}
        ports:
          - "${DB_PORT:-5432}:5432"
        volumes:
          - postgres_data:/var/lib/postgresql/data
        healthcheck:
          test: ["CMD-SHELL", "pg_isready -U ${DB_USER:-ikonex_admin} -d ${DB_NAME:-ikonex_academy}"]
          interval: 5s
          timeout: 5s
          retries: 5
    
      redis:
        image: redis:7-alpine
        container_name: ikonex_redis
        restart: unless-stopped
        ports:
          - "${REDIS_PORT:-6379}:6379"
        volumes:
          - redis_data:/data
        healthcheck:
          test: ["CMD", "redis-cli", "ping"]
          interval: 5s
          timeout: 3s
          retries: 5
    
      backend:
        build:
          context: ./backend
          dockerfile: Dockerfile
        container_name: ikonex_backend
        restart: unless-stopped
        depends_on:
          postgres:
            condition: service_healthy
          redis:
            condition: service_healthy
        environment:
          - ENVIRONMENT=development
          - DB_HOST=postgres
          - DB_PORT=5432
          - DB_USER=${DB_USER:-ikonex_admin}
          - DB_PASSWORD=${DB_PASSWORD:-ikonex_secure_pass}
          - DB_NAME=${DB_NAME:-ikonex_academy}
          - REDIS_HOST=redis
          - REDIS_PORT=6379
          - JWT_SECRET=${JWT_SECRET:-dev_secret_key_needs_replacement_32_chars_min}
          - INTERNAL_SERVICE_TOKEN=${INTERNAL_SERVICE_TOKEN:-dev_internal_secret_token}
          - BACKEND_PORT=8080
        ports:
          - "8080:8080"
    
      ai_service:
        build:
          context: ./ai_service
          dockerfile: Dockerfile
        container_name: ikonex_ai_service
        restart: unless-stopped
        depends_on:
          - backend
        environment:
          - AI_SERVICE_PORT=8000
          - GO_BACKEND_INTERNAL_URL=http://backend:8080/api/v1/internal
          - AI_SERVICE_SECRET_TOKEN=${INTERNAL_SERVICE_TOKEN:-dev_internal_secret_token}
          - MISTRAL_API_URL=${MISTRAL_API_URL:-http://host.docker.internal:11434/v1}
          - MISTRAL_MODEL_NAME=${MISTRAL_MODEL_NAME:-mistral-7b-instruct}
        ports:
          - "8000:8000"
    
      frontend:
        build:
          context: ./frontend
          dockerfile: Dockerfile
        container_name: ikonex_frontend
        restart: unless-stopped
        depends_on:
          - backend
        environment:
          - NEXT_PUBLIC_API_URL=http://localhost:8080/api/v1
        ports:
          - "3000:3000"
    
    volumes:
      postgres_data:
      redis_data:
    # ==============================================================================
    # Ikonex Academy - Master Environment Configuration Template
    # ==============================================================================
    
    # System Environment
    ENVIRONMENT=development
    LOG_LEVEL=debug
    
    # ------------------------------------------------------------------------------
    # PostgreSQL Database (Tenancy enforced via Row-Level Security)
    # ------------------------------------------------------------------------------
    DB_HOST=localhost
    DB_PORT=5432
    DB_USER=ikonex_admin
    DB_PASSWORD=ikonex_secure_pass
    DB_NAME=ikonex_academy
    DB_SSLMODE=disable
    DB_MAX_OPEN_CONNS=25
    DB_MAX_IDLE_CONNS=10
    
    # ------------------------------------------------------------------------------
    # Redis & Asynq Job Queue (SMS, STK push polling, payment reminders)
    # ------------------------------------------------------------------------------
    REDIS_HOST=localhost
    REDIS_PORT=6379
    REDIS_PASSWORD=
    
    # ------------------------------------------------------------------------------
    # Go Backend Server
    # ------------------------------------------------------------------------------
    BACKEND_PORT=8080
    BACKEND_HOST=0.0.0.0
    JWT_SECRET=replace_with_a_secure_random_string_at_least_32_bytes
    JWT_EXPIRY_HOURS=24
    INTERNAL_SERVICE_TOKEN=internal_service_secret_token_between_go_and_ai
    
    # ------------------------------------------------------------------------------
    # Next.js Frontend
    # ------------------------------------------------------------------------------
    FRONTEND_PORT=3000
    NEXT_PUBLIC_API_URL=http://localhost:8080/api/v1
    NEXT_PUBLIC_APP_ENV=development
    
    # ------------------------------------------------------------------------------
    # Python AI Service (Mistral LLM integration, zero direct DB access)
    # ------------------------------------------------------------------------------
    AI_SERVICE_PORT=8000
    AI_SERVICE_HOST=0.0.0.0
    GO_BACKEND_INTERNAL_URL=http://localhost:8080/api/v1/internal
    MISTRAL_API_URL=http://localhost:11434/v1 # Ollama or vLLM local endpoint
    MISTRAL_MODEL_NAME=mistral-7b-instruct
    AI_SERVICE_SECRET_TOKEN=internal_service_secret_token_between_go_and_ai
    
    # ------------------------------------------------------------------------------
    # Africa's Talking (SMS / USSD)
    # ------------------------------------------------------------------------------
    AFRICASTALKING_USERNAME=sandbox
    AFRICASTALKING_API_KEY=replace_with_at_api_key
    AFRICASTALKING_SENDER_ID=IKONEX
    
    # ------------------------------------------------------------------------------
    # Safaricom Daraja API (M-Pesa STK Push / C2B)
    # ------------------------------------------------------------------------------
    DARAJA_ENVIRONMENT=sandbox
    DARAJA_CONSUMER_KEY=replace_with_daraja_consumer_key
    DARAJA_CONSUMER_SECRET=replace_with_daraja_consumer_secret
    DARAJA_PASSKEY=replace_with_daraja_passkey
    DARAJA_CALLBACK_URL=https://your-domain.co.ke/api/v1/billing/mpesa/callback
    
    # ------------------------------------------------------------------------------
    # Biometric Encryption (Per-school envelope encryption key management)
    # ------------------------------------------------------------------------------
    BIOMETRIC_MASTER_KEY=base64_encoded_32_byte_aes_gcm_master_key
    {
      "name": "ikonex-academy-frontend",
      "version": "0.1.0",
      "private": true,
      "scripts": {
        "dev": "next dev",
        "build": "next build",
        "start": "next start",
        "lint": "next lint"
      },
      "dependencies": {
        "next": "^14.2.5",
        "react": "^18.3.1",
        "react-dom": "^18.3.1",
        "lucide-react": "^0.400.0",
        "clsx": "^2.1.1",
        "tailwind-merge": "^2.3.0"
      },
      "devDependencies": {
        "@types/node": "^20",
        "@types/react": "^18",
        "@types/react-dom": "^18",
        "autoprefixer": "^10.4.19",
        "postcss": "^8.4.38",
        "tailwindcss": "^3.4.4",
        "typescript": "^5"
      }
    }
    # Go backend client placeholder (strictly calls permission-checked internal Go API; zero direct DB access)
    # Parent assistant router placeholder (answers fee/attendance/exam questions in Swahili or English)
    tasks/developer-04-backend-engineer-b.md:59:     - Within the transaction, call Ian's (@frihk) `billing.OnTermStartDateChanged(ctx, tx, schoolID, termID, newStartDate, deanUserID, feeChangeNote)` hook, which shifts the term's fee due dates, records a `fee_change_notes` row, and notifies affected parents (`AC-22`).
    tasks/developer-04-backend-engineer-b.md:222:  - Uses `database.WithTenantTx`, `database.AuditLogger`, `middleware.RequireAuth`, and `middleware.RequirePermission` (including scoped `class_teacher` badge check).
    tasks/developer-04-backend-engineer-b.md:225:  - Calls Ian's (@frihk) `billing.Service.OnTermStartDateChanged(ctx, tx, ...)` when the Dean changes a term's start date (`C61, AC-22`).
    tasks/developer-03-backend-engineer-a.md:41:   - Provide `db.WithTenantTx(ctx, schoolID, func(tx pgx.Tx) error)` in `backend/internal/database` so Ema (@ejao-000) and Ian (@frihk) never have to write raw session-variable boilerplate.
    tasks/developer-03-backend-engineer-a.md:44:   - `POST /api/v1/auth/otp/request`: Validates Kenyan phone number (`+254...`), enforces rate limits (`1 OTP request per minute` per phone number; `> 5 requests in 30 minutes` blocks the requesting phone/IP for 24 hours—see `Docs/Ikonex.md` C18–C20, O20), generates a 6-digit cryptographic OTP, stores its SHA-256 hash in `otp_requests` (5-minute expiry), and enqueues a transactional SMS via Ian's (@frihk) `notifications.SMSDispatcher` interface.
    tasks/developer-03-backend-engineer-a.md:97:- `otp_requests`: `id`, `phone_number`, `admission_number`, `sender_identifier` (phone or WhatsApp number), `otp_hash`, `purpose` (`app_login`, `whatsapp_session`, `phone_change`), `expires_at`, `consumed_at`, `created_at`.
    tasks/developer-03-backend-engineer-a.md:98:- `sender_rate_blocks`: `id`, `sender_identifier UNIQUE`, `blocked_until TIMESTAMPTZ NOT NULL`, `reason`, `created_at`.
    tasks/developer-03-backend-engineer-a.md:116:│   │   ├── database.go                  # pgxpool connection, WithTenantTx(ctx, schoolID, fn)
    tasks/developer-03-backend-engineer-a.md:117:│   │   └── audit.go                     # Shared AuditLogger writing to audit_log
    tasks/developer-03-backend-engineer-a.md:160:1. **Database RLS Isolation (`AC-27`):** Queries run via `WithTenantTx` return only rows for `app.school_id`; queries without `app.school_id` set return zero rows.
    tasks/developer-03-backend-engineer-a.md:202:  - `database.DB.WithTenantTx(ctx, schoolID, fn)` and `database.AuditLogger`.
    tasks/developer-03-backend-engineer-a.md:206:  - Calls Ian's (@frihk) `notifications.SMSDispatcher` interface to send login OTPs and the old-number phone-change SMS alert.
    tasks/developer-03-backend-engineer-a.md:207:  - Calls Ian's (@frihk) `billing.StudentChargeHook` interface when a student is admitted or verified so their initial term fee charges are created automatically.
    tasks/developer-03-backend-engineer-a.md:208:  - Provides a clean cleanup query/method `onboarding.PurgeExpiredUnverifiedSubmissions(ctx)` that Ian's (@frihk) Asynq background worker invokes daily (`AC-07`).
    ✓ • 760ms

## Activity

    $ git ls-remote --heads origin && git diff --cached --name-only && command -v go psql postgres initdb docker redis-server node npm python3 && go version
    5c7a1d51e2290cce947d45a7c157b441e5462d00    refs/heads/Docs
    0b502815e5ff2ef2f3f755272d330fc2f4ff4424    refs/heads/dev
    f244d74cf54349fcb2e0500933279b647d306bf7    refs/heads/docs/add-agents-md
    01bfde8fb25d5ece2bf5d1500a4e6cc625ab001e    refs/heads/feat/ema-academics-b1
    7323eccd263eca2377d66401f64befd01f7dc4b2    refs/heads/feat/ema-academics-b2
    4c339754d7545c92ea438dfc16cc4df42265b3be    refs/heads/feat/ema-academics-b3
    6016c85c46ad2b290c1688f09c444b32a4c4b3a3    refs/heads/feat/finances
    b2b2e3b6f15e60f25947f04c39b8f6b12b5a268f    refs/heads/main
    /usr/bin/go
    /usr/local/bin/docker
    /usr/local/lib/nodejs/bin/node
    /usr/local/lib/nodejs/bin/npm
    /usr/bin/python3
    ✗ (1) • 1.16s

## Activity

    $ sed -n '1,240p' Docs/REQUIREMENTS.md && sed -n '1,220p' Docs/Ikonex.md
    # REQUIREMENTS.md
    
    The feature register. This is the source of truth for what's in scope, by
    role and by release phase. If a task isn't covered here, check with the
    product owner before building it — don't infer scope from what would be
    technically interesting to add.
    
    Updated 2026-10-29 to match the decisions recorded in `ikonex.md`. Where this
    file is silent on detail, `ikonex.md` has the behaviour and acceptance
    criteria.
    
    Legend: **[M]** must have · **[S]** should have · **[C]** could have ·
    **(+)** added during scoping discussion, not in the original brief ·
    **[AI]** implemented using the AI service · **(R1/R2/R3)** release phase
    
    Release 1 is a **test of the idea**, built and tested with **synthetic data
    only** (no real children's records), and every Release 1 feature must work
    well and stress free. See `ikonex.md` section 3.
    
    ## A. Platform foundation
    
    - [M] Multi-school hosting with fully isolated data per school
    - [M] Roles plus a badge system (badges are permissions, can be tied to a
      class, club, or time period)
    - [M] Login by phone number with SMS OTP (+)
    - [M] Notifications: in-app [M] (Release 1 event notifications are in-app
      only); SMS [M] for OTP and security messages in Release 1, SMS alerts in
      Release 2; WhatsApp [M] for the parent assistant, moved to Release 1 (+)
    - [M] Audit trail of sensitive actions (+)
    - [S] English (R1) and Swahili (R2) (+)
    - [S] Excel import for students, staff, and fee data (+) — not used for the
      Release 1 first load (class teachers or parents register students; the
      Ikonex team loads fee data)
    - [M] School setup wizard (+): after the Ikonex team creates the school and
      principal, the principal sets up the calendar, classes, dormitories, and
      grading scale
    - [M] Parent consent centre for biometrics, photos, SMS, and WhatsApp (+);
      designed now, collected from real parents from Release 2
    - [M] Editable standard grading scale per school, aligned to CBC; versioned
      per term (a term's levels use the scale in force when released)
    
    ## B. Onboarding
    
    - [M] School and principal created by the Ikonex team (first stage); the
      principal and deputy are onboarded first
    - [M] Staff QR code (expires; valid 8am–6pm by default; principal/DP can
      change the window); applicants upload documents, verified in person
    - [M] Applications approved by principal/DP, or a delegate (e.g. dean of
      studies)
    - [M] Student records: per school, either entered by class teachers (parents
      notified on their existing numbers) or registered through parents on
      opening day; existing students carry forward automatically on re-enrolment
    - [M] Student admission with class, dormitory, admission number
    - [M] Parent self-registration by QR code (valid 12 hours, generated by the
      dean, principal, or deputy, revocable), then in-person verification by
      staff; verification can be delegated to other staff while the QR is active
    - [M] Uploaded documents are deleted after approval (a verification record
      remains); rejected or never-verified submissions are deleted
      automatically after a fixed number of days
    - [S] Biometric capture only where the school can afford it, with a manual
      fallback (Release 2)
    
    ## C. Parent portal
    
    - [M] Child's results, attendance, fee structure, balance sheet
    - [M] Performance graph from the day the child joined: CBC performance
      levels per subject per term (never a class position)
    - [M] Upcoming events, clubs the child is in, books borrowed (clubs and
      books show placeholders in Release 1; real data in Release 2)
    - [M] Student blogs and newsletters (placeholder in Release 1; Release 2)
    - [M] Pay fees by M-Pesa STK push
    - [AI] Parent assistant (R1): answers fee and attendance questions in
      English via WhatsApp and the app, reading only from the database (never
      invents figures), scoped to the requesting parent's own children. On
      WhatsApp it carries only minimal facts (fee balances, attendance status,
      payment prompts); results and report content stay in the app, with a link.
      When it cannot or should not answer, it refuses politely and gives the
      school's contact details.
    - [AI] Chat-initiated fee payment (R1): parent states an amount and child and
      types the M-Pesa number to charge; the AI repeats child, amount, and number
      and starts an STK push only after an explicit yes; parent authorises with
      their own M-Pesa PIN — AI never handles money directly. Any amount is
      accepted; excess is held as a credit.
    - [AI] Plain-language CBC performance-level explanations and report card
      summaries, teacher-approved before sending (R2)
    - [AI] Payment-default reminders (R2): cross-checks a parent's payment
      pledge/package against actual payments and arrears; escalates tone across
      three stages (reminder → firmer reminder → final warning) based on default
      count. Pledges come from the bursar or from the parent; a custom pledge
      waits for bursar acceptance, a pledge that picks a predefined scheme is
      accepted automatically.
    - [S] Sports and activities in the graph (scoring method still to be decided)
    - [S] Absence and sickbay SMS alerts (R2; in Release 1 absences appear as
      in-app notifications; sickbay alerts state only that a visit happened,
      never clinical detail) (+)
    - [C] Notification preferences and weekly digest (+)
    - [C] Visiting day and parent-teacher meeting booking (+)
    - [C] Digital permission slips, and exeat requests for boarders (+)
    
    ## D. Teacher portal
    
    - [M] Personal timetable, events affecting them (e.g. exams), syllabus
      coverage per unit taught (syllabus coverage in Release 2)
    - [M] Class teacher badge: sees every student in own class; takes the daily
      roll-call; edits report comments; approves, reopens, and re-releases results
    - [M] Marks entry: each subject teacher enters marks for their assigned class
      and subject; release is blocked until every subject has marks for every
      student; the dean reassigns marks entry when a teacher cannot
    - [M] Corrections after release: class teacher reopens, subject teacher
      corrects, class teacher re-releases with a teacher's note; only parents of
      affected students are notified
    - [M] Any teacher sees other classes as class/subject means only, never
      individual student names
    - [M] Duty badge: sees the week's expectations/events, can post to the blog
      for parents (Release 2)
    - [AI] Syllabus coverage from natural language: teacher types or dictates
      what was taught; AI maps it to CBC strands/sub-strands and updates coverage
      (R2)
    - [AI] Report comment drafting from marks and trends, teacher edits before
      publishing (R1)
    - [S] Behind-schedule syllabus alerts (+)
    - [C] Timetable clash checker and duty roster generator (+)
    
    ## E. Other staff modules
    
    **Dean of studies** (Release 1)
    - [M] Assigns teachers to classes and subjects; owns all timetables (teachers
      and students) and the school calendar, including term dates after setup;
      changes notify parents and teachers in-app
    - [M] Admissions, class and dormitory assignment, admission numbers
    - [M] Generates parent QR codes and verifies registrations; can delegate
      verification to other staff
    
    **Principal / deputy**
    - [M] Staff onboarding and approvals; leave approvals (Release 2)
    - [M] Setup wizard; sets term dates at setup
    - [S] One dashboard: fees, attendance, syllabus coverage (+)
    
    **Nurse** (Release 2)
    - [M] Record consultations; automatic notification to the class teacher
      (notification only, no clinical detail)
    
    **Librarian** (Release 2)
    - [M] Book catalogue and who holds each book; borrowed books show on the
      parent side
    - [S] Due dates and overdue reminders (+)
    
    **Leave management** (Release 2)
    - [M] Teachers apply for time off in emergencies, with approval
    - [S] Cover arrangements for missed lessons (+)
    
    **Journalism club** (Release 2)
    - [M] Patron reviews and publishes blog posts and newsletters
    - [AI] Blog/newsletter safety check: flags children's personal details or
      inappropriate content before the patron publishes
    - [S] Other clubs get their own patron posting
    
    **Bursar** (Release 2; there is no bursar role in Release 1)
    - Release 1: fee structures are loaded by the Ikonex team, charges are
      applied automatically once at the start of the year, mid-year changes are
      applied by the team with a note and a parent notification, and direct
      Paybill payments with a valid admission number are matched automatically;
      unmatched payments are resolved by the team
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
    - Legacy 8-4-4 (Form 3/4) classes
    - A bursar role and fee-management screens in Release 1
    - A human hand-off inbox for the parent assistant in Release 1
    
    ## Release phases
    
    1. **Release 1 (a test with synthetic data; passes when onboarding works,
       parents can log in, and badges work well):** foundation, setup wizard,
       onboarding (staff and parent QR, in-person verification), dean's timetable,
       calendar and teacher assignment, daily attendance, marks entry and results
       release/reopen, performance graph, parent portal core (fees, results,
       attendance, events) with placeholders for clubs, books, and blogs, fee
       charging and STK push with team-loaded fee structures, **WhatsApp and
       in-app parent assistant** (read-only) and chat-initiated payment, AI report
       comment drafting. English only; in-app event notifications only.
    2. **Release 2:** nurse, librarian, leave management, blogs and clubs,
       syllabus coverage (including AI mapping), biometric attendance, duty
       badge, bursar role with payables and signatories, AI M-Pesa reconciliation,
       AI default reminders and payment plans, parent-facing plain-language
       explanations, SMS alerts, Swahili, consent collection with real parents.
       The first release with real data must clear the launch gates in
       `SECURITY.md`.
    3. **Release 3:** store, caterer, exeat and permission slips, dashboards,
       USSD.
    
    ## Open decisions
    
    - Hosting location: Kenya only vs regional Africa, and where the AI model
      runs (to be discussed with the group)
    - Who pays for WhatsApp and SMS messages
    - Sports/activities scoring: patron-entered scores vs. participation records
      only
    - Fee structure variants (boarder/day, optional items)
    - What happens to a leftover credit when a student leaves
    - How a shared parent QR links each parent to the right child
    - Whether the principal/deputy can override the dean on assignments,
      timetables, and calendar
    - Later-stage route for onboarding a school's first principal
    - Store/caterer module: not yet designed
    
    Settled since the last version: guardian fee visibility (all linked guardians
    see fees, may be revisited); legacy 8-4-4 support (out of scope). Both are
    recorded in `DECISIONS.md`.
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
    ✓ • 14.13s

## Activity

    $ command -v go psql postgres initdb docker redis-server node npm python3; go version; git diff --cached --name-only; git status --short
    /usr/bin/go
    /usr/local/bin/docker
    /usr/local/lib/nodejs/bin/node
    /usr/local/lib/nodejs/bin/npm
    /usr/bin/python3
    go version go1.22.2 linux/amd64
     M Docs/DECISIONS.md
    ✓ • 733ms

## Activity

    $ sed -n '220,500p' Docs/Ikonex.md && cat Docs/DECISIONS.md && sed -n '75,140p' Docs/SECURITY.md
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
    
    # DECISIONS.md
    
    A short, dated log of decisions worth remembering — so nobody (human or AI
    assistant) re-litigates a settled question without knowing why it was made.
    Add a new entry any time a real decision is made; don't edit old entries,
    add a new one that supersedes it if the decision changes.
    
    Format: date, decision, why, alternatives considered.
    
    ---
    
    **2026-09-22 — Shared PostgreSQL database with row-level security, not
    per-school databases**
    Every tenant-owned table carries `school_id`; isolation is enforced with
    Postgres RLS (`FORCE ROW LEVEL SECURITY`), not only application-level
    filtering.
    *Why:* cheapest to run and easiest to back up and migrate across many
    schools; RLS gives a database-level safety net against a missed `WHERE
    school_id = ?` in application code, which matters given the sensitivity of
    the data (children's records).
    *Alternative considered:* one database per school — rejected as
    operationally heavy past a handful of schools, though the shared-table
    design keeps this open as a later migration path for a school that
    specifically needs it (e.g. a contractual requirement).
    
    **2026-09-22 — Global `users` table, separate per-school `memberships`**
    A person has one login; their role(s) at each school live in `memberships`.
    *Why:* supports a parent with children at two schools, or a teacher working
    at two schools, without duplicate accounts.
    
    **2026-09-22 — Badges are permissions, not roles or table columns**
    "Class teacher," "on duty," and "club patron" are badges — rows in
    `membership_badges`, optionally scoped to a class, club, or time window —
    not a role type or a column on `classes`.
    *Why:* one membership can hold several badges at once (a teacher can be a
    class teacher, on duty, and a club patron simultaneously); keeping this as
    one badge system avoids one-off columns for every new permission concept.
    
    **2026-09-22 — CBC only (Grade 7–12), not 8-4-4**
    `grade_levels.number` is constrained to 7–12 with a `junior`/`senior` phase.
    *Why:* stated scope is CBC, both junior and senior school.
    *Open:* whether legacy Form 3/4 classes need support during the transition
    period is unresolved — see `REQUIREMENTS.md` open decisions.
    
    **2026-09-22 — AI service is separate from the Go backend, with no direct
    database access**
    The AI service (Python/FastAPI) calls permission-checked Go endpoints only;
    it never queries the database directly and never originates a fact.
    *Why:* keeps the "AI never invents a fee balance / never handles money"
    guarantees enforceable at the API boundary rather than relying on prompt
    instructions; also lets the AI service be hosted, scaled, or swapped
    independently of the main application.
    
    **2026-09-22 — Payments: system records and approves; it doesn't disburse to
    suppliers**
    Supplier/bill payments go through a bursar-drafted, signatory-approved
    voucher; the actual bank transfer happens outside the system. The one
    exception is parent-initiated fee payment, where the AI service can start
    an M-Pesa STK push, but only after the parent has confirmed the child and
    amount, and the parent still authorises the transaction with their own PIN.
    *Why:* automated disbursement carries materially more regulatory and
    operational risk than recording an approved payment; parent-initiated STK
    push is different because the parent remains the one authorising the actual
    transfer.
    *Open:* whether automated supplier disbursement is ever added later is an
    open question, not a rejected one.
    
    **2026-09-22 — Biometrics: encrypted templates only, per-school keys, manual
    fallback always offered**
    Raw fingerprint images are never stored; only the encrypted template.
    Schools that can't afford biometric hardware use manual attendance.
    *Why:* Data Protection Act treatment of biometric data as sensitive
    personal data of children; minimising what's stored reduces both legal risk
    and breach impact.
    
    **2026-09-22 — Payroll is out of scope**
    *Why:* stated explicitly as out of scope; revisit only if a clear
    justification and owner emerge.
    
    **2026-09-22 — Hosting: Kenya or regional-Africa cloud, not a default global
    region**
    *Why:* supports the Data Protection Act cross-border transfer rules and the
    "data stays in Kenya" product/compliance claim; see `SECURITY.md`.
    
    ---
    
    **2026-10-29 — WhatsApp moves into Release 1; one WhatsApp bot per school**
    The parent assistant and chat payment ship on WhatsApp and in-app in Release
    1. SMS and USSD come later. Each school has its own bot so admission numbers
    never collide.
    *Why:* WhatsApp reaches parents where they already are.
    *Supersedes:* WhatsApp as a Release 3 could-have.
    
    **2026-10-29 — WhatsApp carries minimal facts only**
    Fee balances, attendance status, and payment prompts. Results and report
    content stay in the app with a link. Launch gates before real data: a
    `whatsapp` consent purpose, an impact assessment, a processing agreement with
    Meta, ODPC notification, lawyer review.
    *Why:* WhatsApp passes through Meta, a transfer outside Kenya; less child data
    in it means less exposure.
    *Alternatives considered:* full answers with consent (more legal exposure);
    per-school settings (more cases to test).
    
    **2026-10-29 — Parent identity in chat: admission number, masked guardian
    numbers, OTP to the registered number, 6-hour session**
    Any WhatsApp number may start a chat; OTP requests are limited to one per
    minute, and more than five in thirty minutes blocks the requesting number for
    24 hours (not the guardian).
    *Why:* blocking the guardian would let a stranger lock a real parent out.
    
    **2026-10-29 — AI service authenticates with a service token plus a signed
    user context**
    *Alternative kept open:* adding mutual TLS later.
    
    **2026-10-29 — Chat payments: parent types the M-Pesa number, any amount,
    explicit confirmation**
    Excess is held as a ledger credit; a limiter for the top-most grade is planned
    later, not agreed.
    
    **2026-10-29 — Release 1 is a test with synthetic data only**
    It passes when onboarding works, parents can log in, and badges work well;
    every Release 1 feature must work well and stress free. A QA developer owns
    verification.
    *Why:* nothing real is at stake, so real-data compliance gates apply to the
    first release with real data.
    
    **2026-10-29 — Release 1 has no bursar; the Ikonex team loads fee structures**
    Fee structure is fixed to terms, set by the school's own structure, loaded as
    data by the team, charged automatically once at the start of the year. Mid-year
    changes need a note and notify only affected parents. Due dates follow term
    start dates.
    *Supersedes:* "bursar fee ledger" in Release 1.
    
    **2026-10-29 — Direct Paybill payments**
    Auto-matched when the account number is a valid admission number for the
    school; the rest go to an unmatched list the team resolves.
    
    **2026-10-29 — Results: marks by subject teachers, hard-gated release, reopen
    to correct**
    Class teacher releases only when every subject has marks for every student.
    The dean reassigns marks entry. Corrections reopen the class; parents keep
    seeing the old results until re-release, then only affected parents are
    notified with a teacher's note.
    
    **2026-10-29 — Performance graph plots CBC levels; grading scale versioned
    per term**
    
    **2026-10-29 — Daily class-teacher roll-call in Release 1; in-app absence
    notifications; SMS in Release 2**
    
    **2026-10-29 — Onboarding**
    The team creates the school and principal; the principal runs a setup wizard.
    Parents register by QR (12 hours, generated by dean/principal/deputy,
    revocable) and are verified in person; verification is delegable to staff for
    the QR's active window. Staff upload documents and are verified in person.
    Documents are deleted after approval; unverified applications are deleted
    automatically after a fixed period.
    
    **2026-10-29 — The dean owns teacher assignment, timetables, and the school
    calendar (including term dates after setup)**
    A term-date change moves fee due dates and triggers the fee-change notice.
    
    **2026-10-29 — Guardian fee visibility: all linked guardians see fees**
    May be revisited. Closes the open decision in `REQUIREMENTS.md`.
    
    **2026-10-29 — Legacy 8-4-4 (Form 3/4) is out of scope**
    *Why:* 8-4-4 ends in under a year. Closes the open question under the CBC
    entry above.
    
    **2026-10-29 — Release 1 is English only; Swahili in Release 2**
    
    **2026-10-29 — Hosting location is unresolved**
    This note does not change the earlier hosting entry; it flags that the earlier
    entry ("Kenya or regional-Africa cloud") conflicts with the Kenya-only rule in
    `AGENTS.md` and `SECURITY.md`, and that regional hosting would count as a
    cross-border transfer. To be discussed with the project group. Release 1 is
    unaffected (synthetic data).
    
    **2026-10-29 — Retention at student exit is set per record type**
    Periods to be set with a data protection lawyer. Biometric templates are
    deleted on exit (unchanged).
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
    ✓ • 0ms

## Assistant

The Go backend is a placeholder scaffold: `WithTenantTx`, authentication, audit logging, academics summaries, and cleanup are all absent. `go.mod` has no dependencies, including the planned pgx and Asynq libraries, and only migration `000001` exists.

Two schema details affect the design: ledgers are per term, and fee items lack `school_id` and RLS. I’ll account for both explicitly and document the missing integration contracts inside the owned directories. Adding pgx/Asynq would require a change to the protected module files; I’ll keep those untouched and start with work that can compile independently.

## Activity

    $ sed -n '90,112p' tasks/developer-03-backend-engineer-a.md && sed -n '195,295p' Docs/Ikonex.md && sed -n '345,388p' Docs/Ikonex.md
    ---
    
    ## 4. Technical Implementation
    
    ### 4.1 Database Migration Owned by Evans (`migrations/000002_r1_core_onboarding.up.sql`)
    Do not modify `000001_init_schema.up.sql`. Create `000002_r1_core_onboarding.up.sql` and `.down.sql` adding the Release 1 Foundation & Onboarding tables and RLS policies:
    - `school_profile`: `school_id (PK/FK)`, `contact_phone NOT NULL`, `contact_email NOT NULL`, `office_hours`, `physical_address`, `student_onboarding_route VARCHAR(30) DEFAULT 'both'`, `wizard_completed BOOLEAN DEFAULT FALSE`, `updated_at`.
    - `otp_requests`: `id`, `phone_number`, `admission_number`, `sender_identifier` (phone or WhatsApp number), `otp_hash`, `purpose` (`app_login`, `whatsapp_session`, `phone_change`), `expires_at`, `consumed_at`, `created_at`.
    - `sender_rate_blocks`: `id`, `sender_identifier UNIQUE`, `blocked_until TIMESTAMPTZ NOT NULL`, `reason`, `created_at`.
    - `parent_qr_codes`: `id`, `school_id`, `qr_token UNIQUE`, `created_by`, `valid_from`, `valid_until` (12h), `revoked_at`, `revoked_by`, `created_at`.
    - `registration_submissions`: `id`, `school_id`, `submission_type` (`parent`, `staff`), `parent_qr_id`, `staff_invite_id`, `applicant_phone`, `applicant_name`, `national_id`, `payload JSONB`, `document_type`, `document_blob BYTEA`, `document_deleted_at TIMESTAMPTZ`, `matched_student_id UUID`, `is_duplicate_flag BOOLEAN DEFAULT FALSE`, `status` (`pending`, `approved`, `rejected`), `expires_for_purge_at TIMESTAMPTZ`, `created_at`.
    - `verification_records`: `id`, `school_id`, `submission_id`, `verified_user_id`, `student_id`, `verifier_user_id`, `document_type VARCHAR(100) NOT NULL`, `notes`, `verified_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`.
    - Alter `users`: Add `chat_payment_paused_until TIMESTAMPTZ`, `phone_changed_at TIMESTAMPTZ`.
    - Alter `consents`: Drop and recreate `consents_purpose_check` to allow `('biometric_attendance', 'photo', 'sms', 'whatsapp')`.
    - Apply `ENABLE ROW LEVEL SECURITY`, `FORCE ROW LEVEL SECURITY`, and `tenant_isolation` policies on `school_profile`, `parent_qr_codes`, `registration_submissions`, and `verification_records`.
    
    ### 4.2 Owned Go Files & Package Structure
    ```text
    backend/
    ├── cmd/
    │   └── server/
    │       └── main.go                      # HTTP server bootstrap, dependency injection, route mounting
    ├── internal/
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
    ✓ • 20ms

## Activity

    $ docker info --format '{{.ServerVersion}}' && docker image ls --format '{{.Repository}}:{{.Tag}}'
    27.3.1
    neo4j:5
    ✓ • 1.54s

## Activity

    $ git switch -c feat/ian-backend-engineer-c main
    Switched to a new branch 'feat/ian-backend-engineer-c'
    M    Docs/DECISIONS.md
    ✓ • 171ms

## Activity

    $ go test -v -race ./...
    ?       github.com/Ashomondi/ikonex-academy/backend/cmd/server    [no test files]
    ?       github.com/Ashomondi/ikonex-academy/backend/cmd/worker    [no test files]
    ?       github.com/Ashomondi/ikonex-academy/backend/internal/academics    [no test files]
    ?       github.com/Ashomondi/ikonex-academy/backend/internal/auth    [no test files]
    ?       github.com/Ashomondi/ikonex-academy/backend/internal/billing    [no test files]
    ?       github.com/Ashomondi/ikonex-academy/backend/internal/clubs    [no test files]
    ?       github.com/Ashomondi/ikonex-academy/backend/internal/config    [no test files]
    ?       github.com/Ashomondi/ikonex-academy/backend/internal/consent    [no test files]
    ?       github.com/Ashomondi/ikonex-academy/backend/internal/database    [no test files]
    ?       github.com/Ashomondi/ikonex-academy/backend/internal/internalapi    [no test files]
    ?       github.com/Ashomondi/ikonex-academy/backend/internal/jobs    [no test files]
    ?       github.com/Ashomondi/ikonex-academy/backend/internal/library    [no test files]
    ?       github.com/Ashomondi/ikonex-academy/backend/internal/middleware    [no test files]
    ?       github.com/Ashomondi/ikonex-academy/backend/internal/notifications    [no test files]
    ?       github.com/Ashomondi/ikonex-academy/backend/internal/nurse    [no test files]
    ?       github.com/Ashomondi/ikonex-academy/backend/internal/onboarding    [no test files]
    ✓ • 47.25s

## Assistant

Phase Zero is complete, and `feat/ian-backend-engineer-c` has been created from `main`. The remote has no `release-1` branch, and remote `main` matches the starting commit. The existing documentation edit is preserved.

The implementation will expose route registration and consumer interfaces for Evans, Ema, and Sos. Real database and queue builds need pgx and Asynq; full integration also needs migrations `000002`/`000003`, authentication, cleanup, and academics implementations. Docker is available for disposable database tests.

I’ll record the detailed plan in `backend/internal/billing/IMPLEMENTATION_PLAN.md`, then start with the notification contracts and transactional SMS dispatcher. The execution order is foundation → notifications → migration → billing → fee changes → M-Pesa → WhatsApp → internal AI API → workers → final verification, with tests at each gate.

## Activity

    file changes: Completed · 1 changes
    Add: /home/fian/frihk/Ikonex-Academy/backend/internal/billing/IMPLEMENTATION_PLAN.md
    # Developer 05 implementation plan
    
    ## Repository assessment — 2026-11-09
    
    - Branch: `feat/ian-backend-engineer-c`, created from `main`.
    - Starting commit: `b2b2e3b6f15e60f25947f04c39b8f6b12b5a268f`; remote main agrees. No local or remote `release-1` or existing Ian branch.
    - Pre-existing working-tree change: one leading blank line in `Docs/DECISIONS.md`. Preserve it; exclude it from every commit. Index initially empty.
    - Read both assignment documents in full, baseline UP/DOWN migrations, repository instructions, API/architecture/security/requirements documents, and relevant collaborator contracts.
    - Backend: Go 1.22.2, standard-library placeholder packages, empty server and worker entry points. No third-party dependencies or go.sum. No existing tests. Frontend and Python services are scaffolds.
    - Schema: only migration 000001 exists. Tenant policies use `current_school_id()` and transaction-local `app.school_id`. Users and schools are global. Fee structures are unique per school/year/grade/term; ledgers are per school/student/year/term. Payments have globally unique transaction references.
    - Schema gaps: fee items have neither school_id nor RLS; baseline foreign keys do not enforce matching school IDs; audit_log is not RLS protected. All owned queries must explicitly scope and validate related records. Never bypass RLS to route webhooks: inject trusted identifier-to-school routing, then verify configuration within that tenant.
    - Missing migration 000002 supplies school_profile, otp_requests, sender_rate_blocks, and users.chat_payment_paused_until. Missing 000003 supplies academics data. Do not create replacements for these migrations.
    - Documentation differences: old API_CONTRACTS.md leaves envelopes unspecified; the master plan defines them. Chat state is called both state/uppercase and status/lowercase: use the specified `state` column and uppercase state machine. StudentChargeHook/StudentBillingHook name the same admission capability. Fee variants are unresolved product policy; preserve the plan's default `all`, without inventing boarder classification. Term ledgers require explicit credit transfer, not repeated summing of rolled-forward balances.
    - Tooling: Go 1.22.2 and Docker 27.3.1 available; PostgreSQL/Redis executables absent, images not cached. Live provider credentials and deployed AI service have not been verified. No real SMS/payment requests are authorized by test execution.
    
    ## Dependency map
    
    | Dependency | Owner | Available | Required contract / approach |
    |---|---|---|---|
    | Tenant transactions | Evans | No | `WithTenantTx(ctx, schoolID, func(pgx.Tx) error) error`; injected consumer contract, test double; no edits to database package. |
    | Audit logger | Evans | No | Transaction-bound audit writer; injected contract; failures roll back financial changes. |
    | Authentication | Evans | No | Trusted user, membership, school, role and channel accessor; deny absent identity; no caller-controlled identity headers. |
    | SMSDispatcher | Ian | No | Context-aware transactional SMS message with purpose; local sandbox and Africa's Talking implementations. |
    | InAppNotifier | Ian | No | `CreateInAppNotification` participating in the caller's tenant transaction. |
    | Admission billing | Ian / Evans consumer | No | StudentChargeHook (StudentBillingHook alias if needed), caller transaction, tenant/student/year and actor. |
    | Term-date sync | Ian / Ema consumer | No | `OnTermStartDateChanged(ctx, tx, schoolID, termID, newStartDate, actorUserID, note) error`. |
    | Academic summaries | Ema | No | GetStudentAttendanceSummary / GetStudentMarksSummary consumer interface; mock only, no substitute academics system. |
    | Cleanup | Evans | No | `PurgeExpiredUnverifiedSubmissions(ctx)`; configuration retained by adapter; default seven days, scheduled daily. |
    | AI | Sos | Placeholder | HTTP parent-query; token plus raw base64url JSON + dot + hex HMAC-SHA256 of encoded payload, keyed by service token. |
    | Frontend | Ashy | Placeholder | Standard success/error envelopes, fixed two-decimal money strings, paginated lists, explicit offsets. |
    | pgx / Asynq | Evans module files | Absent | Native implementations cannot compile in the root module until dependencies are added by owner or explicitly authorized. |
    | Redis / PostgreSQL | Infrastructure | Docker available | Disposable synthetic services for migration, concurrency and worker verification. |
    
    ## Protected-file coordination proposal
    
    The minimal dependency change needed for the prescribed pgx transaction contract and real Asynq worker is the following direct requirements in `backend/go.mod`, plus their generated indirect requirements and `backend/go.sum`:
    
    ```go
    require (
        github.com/hibiken/asynq v0.25.1
        github.com/jackc/pgx/v5 v5.7.2
    )
    ```
    
    Both versions support this Go toolchain. This is a proposal, not authorization to edit protected files. Do not add nested modules, vendor libraries, replace Asynq, or hide unresolved imports behind build tags. Standard-library tasks can proceed independently. Migration 000002's global sender-block semantics and OTP school/session binding also need coordination; use boundary tests until the real schema exists.
    
    ## Execution and verification gates
    
    Paths below are relative to the repository root. `B`, `N`, `W`, `I`, `J` mean `backend/internal/billing`, `notifications`, `whatsapp`, `internalapi`, `jobs` respectively. `M` means `migrations/000004_r1_billing_whatsapp_notifications`. Tests live beside their implementation. Each row is one task: inspect, implement its ordered steps, gofmt only affected files, run verification, review the diff and ownership, then update status. Never advance a dependent task past a failed gate.
    
    `G(package)` means `cd backend && go test -v -race ./internal/<package>` and `go vet ./internal/<package>`; both must pass. `DB` means disposable PostgreSQL 16: baseline then owned UP, inspect schema/constraints/indexes, use a non-superuser role to test RLS, owned DOWN, verify baseline unchanged, UP again. Missing external integration is reported separately from a successful unit test. Commit tested coherent tasks only; no push or merge.
    
    | ID | Feature / objective | Dependencies | Exact files | Ordered implementation / expected behavior | Tests and verification | Acceptance / status |
    |---|---|---|---|---|---|---|
    | INF-01 | C1–C3 baseline and architecture | None | B/IMPLEMENTATION_PLAN.md | Read documents; inventory actual code/schema; preserve edits; inspect remote; create branch; map gaps | Git status/branches/HEAD, baseline race suite | Analysis and branch complete; baseline suite running |
    | NOTIF-01 | C3 SMS boundary and normalization | INF-01 | N/sms.go, N/sms_test.go | Define SMSDispatcher and allowed transactional purposes; normalize Kenyan/E.164 phones; validate without leaking content | Table tests for accepted/rejected numbers, purposes and missing content; G(notifications) | No event SMS; exact normalized recipient; Pending |
    | NOTIF-02 | C3 deterministic sandbox SMS | NOTIF-01 | N/sms.go, N/sms_test.go | Select sandbox by username; log masked recipient and safe metadata, never message/OTP; propagate cancellation | Capture slog output, assert no OTP/full phone/API key; concurrent dispatch; G(notifications) | No network for sandbox; Pending |
    | NOTIF-03 | C3 Africa's Talking transport | NOTIF-01 | N/sms.go, N/sms_test.go | Context-bound HTTP client; encoded form and apiKey; validate per-recipient response; classify retryable/permanent failures | httptest request contract, provider rejection, timeout, malformed response, cancellation; G(notifications) | No false delivery success or secret-bearing errors; Pending |
    | INF-02 | C1–C3 transaction/auth/error contracts | INF-01, pgx | I/contracts/contracts.go, I/contracts/contracts_test.go | Small consumer contracts for pgx tenant transactions/audit/identity; APIError and standard envelopes; deny missing identity | Contract doubles, malformed identity, error envelope and safe error mapping; G(internalapi/...) | Compatible injection, no substitute auth; Blocked on module dependency |
    | NOTIF-04 | C3 transaction-bound notification creation | INF-02 | N/notifications.go, N/notifications_test.go | Validate content/recipient; scoped insertion within supplied transaction; define InAppNotifier | Rollback, tenant/user checks; PostgreSQL insertion and RLS; G(notifications) | Usable by Ema and billing without separate commit; Pending |
    | NOTIF-05 | C3 notification HTTP routes | NOTIF-04, DB-03 | N/http.go, N/notifications_test.go | Register GET list / PATCH read; bounded pagination; recipient+school scope; preserve read_at on repeat | HTTP authorization, paging, ownership, read idempotency; G(notifications) | Both endpoints and envelopes verified; Pending |
    | DB-01 | C1–C3 migration design and UP | INF-01 | M.up.sql, B/migration_test.go | Add ten assigned tables; ledger credit/date; necessary financial constraints, indexes and tenant policies; enforce reference ownership | Review against baseline; DB schema and cross-tenant FK tests | No duplicated baseline/core tables; Pending |
    | DB-02 | C1–C3 reversible DOWN | DB-01 | M.down.sql, B/migration_test.go | Reverse owned changes in dependency order; no CASCADE into baseline; preserve unrelated records | DB UP/DOWN/UP and baseline schema/data comparison | Rollback verified; Pending |
    | DB-03 | C1–C3 RLS verification | DB-02 | B/migration_test.go, B/testdata/schema_checks.sql | Two synthetic schools; non-superuser sessions; no-tenant, wrong-tenant read/write and transaction-local context | DB policies, indexes, constraints, rollback | Actual PostgreSQL isolation evidence; Pending |
    | BILL-01 | C1 exact money and ledger arithmetic | INF-01 | B/money.go, B/money_test.go | Integer minor units; strict decimal parser/JSON strings/NUMERIC limits; compute nonnegative balance and credit | Precision, overflow, negative/zero, overpayment, serialization, conservation; G(billing) | No float arithmetic; Pending |
    | BILL-02 | C1 fee-structure validation | BILL-01, INF-02 | B/structures.go, B/structures_test.go | Validate team role, references, term/year agreement and items; check sums and duplicates | Invalid/cross-school/year/grade/items and duplicate request tests; G(billing) | Admin-only valid consistent structure; Pending |
    | BILL-03 | C1 fee periods and structure loading | BILL-02, DB-03 | B/billing.go, B/structures.go, B/structures_test.go | Transactionally persist itemized structure and term period with due_date=start_date | PostgreSQL reference/rollback/duplicate tests; G(billing) | No partial structure; Pending |
    | BILL-04 | C1 enrolled-student term charging | BILL-03 | B/charges.go, B/charges_test.go | Select applicable enrolled students; serialize student financial operations; insert unique charge and ledger entry | Concurrent repeat charge, inactive/other-grade exclusions, dates; G(billing) | One charge per student/period; Pending |
    | BILL-05 | C1 admission hook | BILL-04 | B/billing.go, B/charges_test.go | Export caller-transaction StudentChargeHook; load applicable structures; reuse charging logic | New admission, repeat hook, absent structures, rollback; G(billing) | Onboarding can inject without importing its code; Pending |
    | BILL-06 | C1 credit carry-forward | BILL-04 | B/ledgers.go, B/ledgers_test.go | Serialize student ledger changes; consume existing credit once; preserve term history; record credit_applied | Multi-term/year, partial/exact/excess credit and concurrency; G(billing) | Financial conservation and no double credit; Pending |
    | BILL-07 | C1 balance and ledger endpoints | BILL-06 | B/http.go, B/ledgers.go, B/ledgers_test.go | Guardian/team authorization; retrieve balances/items/history/notes; fixed money and date formats | Own/unrelated guardian, cross-tenant, item/entry accuracy; G(billing) | Two GET routes conform to envelope; Pending |
    | BILL-08 | C1 explained fee adjustments | BILL-07, NOTIF-04 | B/fee_changes.go, B/fee_changes_test.go | Require note; select affected charges; apply deltas, preserve entries; create note and audit in transaction | 422 code, positive/negative changes, rollback and history; G(billing) | Atomic adjustment with audit; Pending |
    | BILL-09 | C1 targeted fee notices | BILL-08 | B/fee_changes.go, B/fee_changes_test.go | Deduplicate affected guardians; create notifications inside same transaction | Shared guardian, unaffected guardian, notifier failure rollback; G(billing) | Only affected guardians receive one notice; Pending |
    | BILL-10 | C1 term-date synchronization | BILL-09 | B/fee_changes.go, B/fee_changes_test.go | Export exact date hook; shift periods/charges/ledger due dates; note/audit/notify | Same/date-shift, rollback, targeted notices; G(billing) | AC-22 hook verified; Pending |
    | MPESA-01 | C2 Daraja client and encrypted config | BILL-01 | B/daraja.go, B/daraja_test.go | School credential resolver; OAuth; safe HTTP timeout/errors; exact request encoding | httptest OAuth/STK/query payloads, secrets, provider failures; G(billing) | No hardcoded credentials or live calls; Pending |
    | MPESA-02 | C2 STK authorization and validation | MPESA-01, BILL-07 | B/mpesa_stk.go, B/mpesa_stk_test.go | Linked guardian/registered phone; positive amount; chat-only freeze; internal parent confirmation | Under/equal/over balance, 0.01, unauthorized, freeze boundary/web, missing confirmation; G(billing) | AC-23/26 enforced before provider call; Pending |
    | MPESA-03 | C2 durable STK initiation | MPESA-02 | B/mpesa_stk.go, B/mpesa_stk_test.go, M.up.sql, M.down.sql | Persist correlation before provider side effect; admission AccountReference; schedule 45s polling; recover queue failures | Provider failure, delayed callback, correlation and queue tests; G(billing), DB if schema changes | Pending STK survives process restart; Pending |
    | MPESA-04 | C2 idempotent callback credit | MPESA-03, BILL-06, NOTIF-04 | B/mpesa_stk.go, B/payments.go, B/payments_test.go | Authenticate callback route context; validate response against pending request; persist receipt/payment/entry/credit/notice atomically | Failed/malformed/duplicate/concurrent callback, wrong amount/receipt/school; G(billing) | Exactly one financial credit; Pending |
    | MPESA-05 | C2 direct Paybill and unmatched storage | MPESA-04 | B/mpesa_c2b.go, B/mpesa_c2b_test.go | Resolve receiving shortcode through trusted boundary; school-scoped normalized admission; persist match or full unmatched facts | Two schools/same admission, invalid/blank/ambiguous reference, duplicate receipt; G(billing) | No lost incoming payment; Pending |
    | MPESA-06 | C2 unmatched list and manual resolution | MPESA-05 | B/unmatched.go, B/unmatched_test.go | Team-only list; lock unresolved payment; validate target; credit once; resolve with actor/audit | Forbidden/cross-school/duplicate/concurrent resolution and overpayment; G(billing) | No fuzzy matching; Pending |
    | WA-01 | C3 guardian phone masking | NOTIF-01 | W/masking.go, W/masking_test.go | Normalize and mask; reject invalid values without echoing private input | Exact +254712345678 -> 07•• ••• 678; malformed cases; G(whatsapp) | No full guardian number returned; Pending |
    | WA-02 | C3 session persistence and bot routing | INF-02, DB-03 | W/whatsapp.go, W/repository.go, W/whatsapp_test.go | Trusted bot mapping then tenant verification; create/reuse sender-bound sessions; scoped admission/guardian lookup | Unknown admission zero SMS; two schools; masked multi/single selection; G(whatsapp) | No cross-school lookup or identity assumption; Pending |
    | WA-03 | C3 secure OTP and sender limits | WA-02, NOTIF-03, migration 000002 contract | W/otp.go, W/otp_test.go | Cryptographic 6-digit OTP/hash/expiry; bind guardian/session; serialize global sender and school/admission limits; persist sixth-request 24h block | 59/60s, 5/6 requests, 30min/24h boundaries, concurrent senders, selected recipient only; G(whatsapp) | Guardian account never blocked; Pending |
    | WA-04 | C3 verification and six-hour sessions | WA-03 | W/whatsapp.go, W/whatsapp_test.go | Atomic consume; bounded wrong-code attempts; verified_at/expires_at; deny at exact expiry; restart identity flow | Replay, wrong/expired code, exact six-hour boundary, ownership; G(whatsapp) | Fresh OTP after expiry; Pending |
    | WA-05 | C3 webhook and REST adapters | WA-04 | W/http.go, W/cloud.go, W/http_test.go | Verify Meta challenge and request signature; route events to same state machine; bind REST sender; bounded payloads and delivery retries | Invalid signature/session, duplicate events, REST 429, Cloud API HTTP tests; G(whatsapp) | One state machine across transports; Pending |
    | API-01 | C3 signed user context | INF-01 | I/context_signer.go, I/context_signer_test.go | Implement prescribed base64url/hex HMAC; constant-time comparisons; strict fields/channels/expiry; injectable clock | Fixed cross-language vector, tampering, malformed/expired/missing context/token; G(internalapi) | Verify signature before decoding/trusting payload; Pending |
    | API-02 | C3 internal authentication and authorization | API-01, INF-02, WA-04 | I/internalapi.go, I/internalapi_test.go | Authenticate both headers; recheck active membership/session; guardian-child tenant guard; inject academics interface | Cross-parent/school, revoked/expired session, bad membership; G(internalapi) | No AI trust for authorization; Pending |
    | API-03 | C3 five internal read endpoints | API-02, BILL-07 | I/reads.go, I/reads_test.go | School profile, own children, fees, attendance, marks; minimal WhatsApp facts and app link for marks | Route/data/envelope tests; academics mocks labeled; WhatsApp results denial; G(internalapi) | Backend enforces own-child and channel restrictions; Pending |
    | API-04 | C3 internal confirmed STK and audit | API-03, MPESA-03 | I/actions.go, I/actions_test.go | Force chat source/confirmation; call billing; persist safe action/read metadata; configurable lookup/start limiter | Missing confirmation/freeze, audit failures, secret metadata, rate limits; G(internalapi) | AC-28 and payment boundaries; Pending |
    | API-05 | C3 chat proxy and WhatsApp forwarding | API-04, WA-05 | I/client.go, I/client_test.go, W/whatsapp.go | Sign outbound identity; call AI parent-query with both headers; forward validated response; safe failure/fallback | httptest contract, context TTL<=session expiry, canceled/failed AI, channel limits; G(internalapi), G(whatsapp) | No AI end-to-end claim from mocks; Pending |
    | JOB-01 | C3 Asynq client and task payloads | Asynq dependency, NOTIF-03, MPESA-03 | J/jobs.go, J/jobs_test.go | Typed sms:transactional, mpesa:stk_poll, onboarding:purge_unverified tasks; validate/serialize; stable task IDs | Serialization, invalid payload, deduplication and 45s delay; G(jobs) | Actual Asynq, no substitute queue; Blocked on module dependency |
    | JOB-02 | C3 transactional SMS worker | JOB-01 | J/sms.go, J/sms_test.go | Context-bound dispatch; retry temporary errors; skip permanent/expired OTP; delivery deduplication | Redis-backed retry, permanent failure, cancellation, duplicate task; G(jobs) | No event SMS or unbounded OTP retries; Pending |
    | JOB-03 | C2 STK polling worker | JOB-01, MPESA-04 | J/mpesa.go, J/mpesa_test.go, B/mpesa_stk.go | Check pending; query provider; terminal failure/timeout; success requires verified receipt facts; same idempotent settlement | Callback/poll race, retry, pending, success without receipt; G(jobs), G(billing) | Query response never fabricates a receipt; Pending |
    | JOB-04 | C3 cleanup schedule and lifecycle | JOB-02, JOB-03 | J/worker.go, J/worker_test.go, backend/cmd/worker/main.go | Redis configuration; inject Evans cleanup; retention default 7; daily scheduler; signals/graceful shutdown | Test cleanup double, retention/env, handler registration/start/shutdown; G(jobs), go build ./cmd/worker | Missing cleanup explicitly reported, never silently skipped; Pending |
    | QA-01 | C1–C3 final verification and report | All above | B/VERIFICATION.md | Format owned Go; test/build/vet/race; DB cycle; inspect security/concurrency/contracts; map ACs; ownership audit; conventional commits | go test -v -race ./...; go build ./...; go vet ./...; make test; frontend lint/build; git status/diff/cached paths | Report evidence and blockers separately; Pending |
    
    ## Acceptance evidence tracker
    
    | Criterion | Required coverage | Evidence | Status |
    |---|---|---|---|
    | AC-BE-C01 | School routing, masked selected guardian, OTP, exact 6h expiry | WA-01–05, PostgreSQL/HTTP tests | Not implemented |
    | AC-BE-C02 | Sender/admission cooldown, sixth-request sender-only 24h block | WA-03 concurrent and boundary tests | Not implemented |
    | AC-BE-C03 | Idempotent term charge, credit, explained targeted fee/date changes | BILL-02–10, PostgreSQL tests | Not implemented |
    | AC-BE-C04 | STK rules, receipt idempotency, overpayment, school-scoped C2B/manual matching | MPESA-01–06, JOB-03 | Not implemented |
    | AC-BE-C05 | Token/HMAC/expiry, own-child permissions, AI audit | API-01–05 | Not implemented |
    
    ## Outstanding integration, not substitute implementations
    
    Evans must mount exported routes, supply authenticated identities/tenant transactions/audit/cleanup, and land core migration/dependencies. Ema supplies academic reads and invokes term-date synchronization. Sos supplies AI service and compatible signed-context verification; Ashy consumes documented owned endpoint shapes. Provider credentials, a reachable authenticated callback ingress, and provider sandbox end-to-end tests remain separate verification gates. Full integration of migrations 000001–000004 cannot be claimed before 000002/000003 exist.

## Assistant

The repository has no Go dependencies. The required pgx transaction contract and real Asynq worker need github.com/jackc/pgx/v5 v5.7.2 and github.com/hibiken/asynq v0.25.1, plus their indirect dependencies. I recorded this exact proposal in backend/internal/billing/IMPLEMENTATION_PLAN.md. Your ownership rule explicitly forbids editing backend/go.mod and backend/go.sum, so I need an exception before adding them. Which approach should I take? I’ll continue the independent SMS work meanwhile.
- Allow only these dependency additions to go.mod/go.sum
- Keep the boundary; leave dependency-bound tasks blocked for Evans

## Assistant

The detailed plan is now in `backend/internal/billing/IMPLEMENTATION_PLAN.md`, including dependencies, exact files, test gates, and acceptance tracking.

I’m starting `NOTIF-01`: define `SMSDispatcher`, restrict SMS to the agreed transactional purposes, and normalize recipient numbers. This touches only `notifications/sms.go` and its tests and has no dependency on the missing database or authentication code.

## Activity

    file changes: Completed · 2 changes
    Add: /home/fian/frihk/Ikonex-Academy/backend/internal/notifications/sms.go
    package notifications
    
    import (
    	"context"
    	"errors"
    	"strings"
    	"unicode/utf8"
    )
    
    type SMSPurpose string
    
    const (
    	AppLoginOTP       SMSPurpose = "app_login"
    	WhatsAppSessionOTP SMSPurpose = "whatsapp_session"
    	PhoneChangeOTP    SMSPurpose = "phone_change"
    	PhoneChangeNotice SMSPurpose = "phone_change_notice"
    )
    
    var (
    	ErrInvalidPhone = errors.New("invalid recipient phone number")
    	ErrInvalidSMS   = errors.New("invalid transactional SMS")
    )
    
    // SMSMessage contains sensitive data; dispatchers must not log Body or To.
    type SMSMessage struct {
    	To      string     `json:"to"`
    	Body    string     `json:"body"`
    	Purpose SMSPurpose `json:"purpose"`
    }
    
    // SMSDispatcher accepts only transactional messages. A nil error means the
    // dispatcher accepted the message, not that the handset received it.
    type SMSDispatcher interface {
    	DispatchSMS(context.Context, SMSMessage) error
    }
    
    // NormalizePhone accepts Kenyan local numbers and international E.164 numbers.
    // Validation errors intentionally exclude the input to avoid exposing numbers.
    func NormalizePhone(phone string) (string, error) {
    	phone = strings.TrimSpace(phone)
    	var normalized strings.Builder
    	for _, c := range phone {
    		switch {
    		case c >= '0' && c <= '9':
    			normalized.WriteRune(c)
    		case c == '+' && normalized.Len() == 0:
    			normalized.WriteRune(c)
    		case c == ' ' || c == '-' || c == '(' || c == ')':
    		default:
    			return "", ErrInvalidPhone
    		}
    	}
    	phone = normalized.String()
    	if len(phone) == 10 && (strings.HasPrefix(phone, "07") || strings.HasPrefix(phone, "01")) {
    		phone = "+254" + phone[1:]
    	} else if len(phone) == 12 && strings.HasPrefix(phone, "254") {
    		phone = "+" + phone
    	}
    	if len(phone) < 9 || len(phone) > 16 || phone[0] != '+' || phone[1] == '0' {
    		return "", ErrInvalidPhone
    	}
    	for _, c := range phone[1:] {
    		if c < '0' || c > '9' {
    			return "", ErrInvalidPhone
    		}
    	}
    	if strings.HasPrefix(phone, "+254") && (len(phone) != 13 || (phone[4] != '7' && phone[4] != '1')) {
    		return "", ErrInvalidPhone
    	}
    	return phone, nil
    }
    
    func (m SMSMessage) normalized() (SMSMessage, error) {
    	switch m.Purpose {
    	case AppLoginOTP, WhatsAppSessionOTP, PhoneChangeOTP, PhoneChangeNotice:
    	default:
    		return SMSMessage{}, ErrInvalidSMS
    	}
    	if strings.TrimSpace(m.Body) == "" || !utf8.ValidString(m.Body) {
    		return SMSMessage{}, ErrInvalidSMS
    	}
    	phone, err := NormalizePhone(m.To)
    	if err != nil {
    		return SMSMessage{}, err
    	}
    	m.To = phone
    	return m, nil
    }
    Add: /home/fian/frihk/Ikonex-Academy/backend/internal/notifications/sms_test.go
    package notifications
    
    import (
    	"errors"
    	"strings"
    	"testing"
    )
    
    func TestNormalizePhone(t *testing.T) {
    	for _, tc := range []struct{ input, want string }{
    		{"0712345678", "+254712345678"},
    		{"0112345678", "+254112345678"},
    		{"254712345678", "+254712345678"},
    		{" +254 (712) 345-678 ", "+254712345678"},
    		{"+442079460000", "+442079460000"},
    		{"", ""}, {"+", ""}, {"071234567", ""},
    		{"+2547123456789", ""}, {"+254012345678", ""},
    		{"+254212345678", ""}, {"+0712345678", ""},
    		{"0712345678\napiKey=secret", ""}, {"+254７12345678", ""},
    		{"+254+712345678", ""}, {"712345678", ""},
    		{"+1234567890123456", ""},
    	} {
    		t.Run(tc.input, func(t *testing.T) {
    			got, err := NormalizePhone(tc.input)
    			if got != tc.want || (err != nil) != (tc.want == "") {
    				t.Fatalf("got %q, %v; want %q", got, err, tc.want)
    			}
    			if err != nil && !errors.Is(err, ErrInvalidPhone) {
    				t.Fatalf("unexpected error: %v", err)
    			}
    		})
    	}
    }
    
    func TestTransactionalSMSValidation(t *testing.T) {
    	for _, purpose := range []SMSPurpose{AppLoginOTP, WhatsAppSessionOTP, PhoneChangeOTP, PhoneChangeNotice} {
    		msg, err := (SMSMessage{To: "0712345678", Body: "Synthetic test message", Purpose: purpose}).normalized()
    		if err != nil || msg.To != "+254712345678" || msg.Purpose != purpose {
    			t.Fatalf("purpose %s: %+v %v", purpose, msg, err)
    		}
    	}
    	for _, msg := range []SMSMessage{
    		{To: "0712345678", Body: "Absence alert", Purpose: "absence"},
    		{To: "0712345678", Body: "Sickbay alert", Purpose: "sickbay"},
    		{To: "0712345678", Body: "hello"},
    		{To: "0712345678", Body: " \n ", Purpose: AppLoginOTP},
    		{To: "0712345678", Body: string([]byte{0xff}), Purpose: AppLoginOTP},
    		{To: "private-invalid-number", Body: "654321", Purpose: AppLoginOTP},
    	} {
    		_, err := msg.normalized()
    		if err == nil {
    			t.Fatal("invalid message accepted")
    		}
    		if strings.Contains(err.Error(), msg.To) || strings.Contains(err.Error(), "654321") {
    			t.Fatal("validation error exposed private message data")
    		}
    	}
    }

## Activity

    $ git status --short && git diff --name-only && git diff --cached --name-only && git diff --check
     M Docs/DECISIONS.md
    ?? backend/internal/billing/IMPLEMENTATION_PLAN.md
    ?? backend/internal/notifications/sms.go
    ?? backend/internal/notifications/sms_test.go
    Docs/DECISIONS.md
    ✓ • 112ms

## Activity

    $ gofmt -w internal/notifications/sms.go internal/notifications/sms_test.go && go test -v -race ./internal/notifications && go vet ./internal/notifications
    === RUN   TestNormalizePhone
    === RUN   TestNormalizePhone/0712345678
    === RUN   TestNormalizePhone/0112345678
    === RUN   TestNormalizePhone/254712345678
    === RUN   TestNormalizePhone/_+254_(712)_345-678_
    === RUN   TestNormalizePhone/+442079460000
    === RUN   TestNormalizePhone/#00
    === RUN   TestNormalizePhone/+
    === RUN   TestNormalizePhone/071234567
    === RUN   TestNormalizePhone/+2547123456789
    === RUN   TestNormalizePhone/+254012345678
    === RUN   TestNormalizePhone/+254212345678
    === RUN   TestNormalizePhone/+0712345678
    === RUN   TestNormalizePhone/0712345678_apiKey=secret
    === RUN   TestNormalizePhone/+254７12345678
    === RUN   TestNormalizePhone/+254+712345678
    === RUN   TestNormalizePhone/712345678
    === RUN   TestNormalizePhone/+1234567890123456
    --- PASS: TestNormalizePhone (0.00s)
        --- PASS: TestNormalizePhone/0712345678 (0.00s)
        --- PASS: TestNormalizePhone/0112345678 (0.00s)
        --- PASS: TestNormalizePhone/254712345678 (0.00s)
        --- PASS: TestNormalizePhone/_+254_(712)_345-678_ (0.00s)
        --- PASS: TestNormalizePhone/+442079460000 (0.00s)
        --- PASS: TestNormalizePhone/#00 (0.00s)
        --- PASS: TestNormalizePhone/+ (0.00s)
        --- PASS: TestNormalizePhone/071234567 (0.00s)
        --- PASS: TestNormalizePhone/+2547123456789 (0.00s)
        --- PASS: TestNormalizePhone/+254012345678 (0.00s)
        --- PASS: TestNormalizePhone/+254212345678 (0.00s)
        --- PASS: TestNormalizePhone/+0712345678 (0.00s)
        --- PASS: TestNormalizePhone/0712345678_apiKey=secret (0.00s)
        --- PASS: TestNormalizePhone/+254７12345678 (0.00s)
        --- PASS: TestNormalizePhone/+254+712345678 (0.00s)
        --- PASS: TestNormalizePhone/712345678 (0.00s)
        --- PASS: TestNormalizePhone/+1234567890123456 (0.00s)
    === RUN   TestTransactionalSMSValidation
    --- PASS: TestTransactionalSMSValidation (0.00s)
    PASS
    ok      github.com/Ashomondi/ikonex-academy/backend/internal/notifications    1.071s
    ✓ • 1m 08s

## Assistant

The baseline race suite passed; every package reported “no test files.” The new SMS normalization and purpose-validation tests also passed under the race detector. The remaining check for this task is `go vet`; after that I’ll add the sandbox dispatcher and test that its logs expose neither OTPs nor full phone numbers.

## Activity

    file changes: Completed · 1 changes
    Update { move_path: None }: /home/fian/frihk/Ikonex-Academy/backend/internal/billing/IMPLEMENTATION_PLAN.md
    @@ -54,4 +54,4 @@
     |---|---|---|---|---|---|---|
    -| INF-01 | C1–C3 baseline and architecture | None | B/IMPLEMENTATION_PLAN.md | Read documents; inventory actual code/schema; preserve edits; inspect remote; create branch; map gaps | Git status/branches/HEAD, baseline race suite | Analysis and branch complete; baseline suite running |
    -| NOTIF-01 | C3 SMS boundary and normalization | INF-01 | N/sms.go, N/sms_test.go | Define SMSDispatcher and allowed transactional purposes; normalize Kenyan/E.164 phones; validate without leaking content | Table tests for accepted/rejected numbers, purposes and missing content; G(notifications) | No event SMS; exact normalized recipient; Pending |
    +| INF-01 | C1–C3 baseline and architecture | None | B/IMPLEMENTATION_PLAN.md | Read documents; inventory actual code/schema; preserve edits; inspect remote; create branch; map gaps | Git status/branches/HEAD, baseline race suite | Completed; baseline race suite passed (no pre-existing tests) |
    +| NOTIF-01 | C3 SMS boundary and normalization | INF-01 | N/sms.go, N/sms_test.go | Define SMSDispatcher and allowed transactional purposes; normalize Kenyan/E.164 phones; validate without leaking content | Table tests for accepted/rejected numbers, purposes and missing content; G(notifications) | Completed; race tests and vet passed |
     | NOTIF-02 | C3 deterministic sandbox SMS | NOTIF-01 | N/sms.go, N/sms_test.go | Select sandbox by username; log masked recipient and safe metadata, never message/OTP; propagate cancellation | Capture slog output, assert no OTP/full phone/API key; concurrent dispatch; G(notifications) | No network for sandbox; Pending |

## Activity

    $ git status --short && git diff --name-only && git diff --cached --name-only && git diff --check && git add backend/internal/billing/IMPLEMENTATION_PLAN.md backend/internal/notifications/sms.go backend/internal/notifications/sms_test.go && git diff --cached --name-only && git commit -m 'feat(notifications): define transactional SMS contract'
     M Docs/DECISIONS.md
    ?? backend/internal/billing/IMPLEMENTATION_PLAN.md
    ?? backend/internal/notifications/sms.go
    ?? backend/internal/notifications/sms_test.go
    Docs/DECISIONS.md
    backend/internal/billing/IMPLEMENTATION_PLAN.md
    backend/internal/notifications/sms.go
    backend/internal/notifications/sms_test.go
    [feat/ian-backend-engineer-c 5925aa5] feat(notifications): define transactional SMS contract
     3 files changed, 255 insertions(+)
     create mode 100644 backend/internal/billing/IMPLEMENTATION_PLAN.md
     create mode 100644 backend/internal/notifications/sms.go
     create mode 100644 backend/internal/notifications/sms_test.go
    ✓ • 78ms

## Activity

    file changes: Completed · 2 changes
    Update { move_path: None }: /home/fian/frihk/Ikonex-Academy/backend/internal/notifications/sms.go
    @@ -5,2 +5,3 @@
     	"errors"
    +	"log/slog"
     	"strings"
    @@ -88 +89,30 @@
     }
    +
    +type sandboxSMSDispatcher struct {
    +	logger *slog.Logger
    +}
    +
    +// NewSandboxSMSDispatcher records only redacted delivery metadata, never sends
    +// a network request, and never logs the message body or the full recipient.
    +func NewSandboxSMSDispatcher(logger *slog.Logger) SMSDispatcher {
    +	if logger == nil {
    +		logger = slog.Default()
    +	}
    +	return &sandboxSMSDispatcher{logger: logger}
    +}
    +
    +func (d *sandboxSMSDispatcher) DispatchSMS(ctx context.Context, message SMSMessage) error {
    +	if err := ctx.Err(); err != nil {
    +		return err
    +	}
    +	message, err := message.normalized()
    +	if err != nil {
    +		return err
    +	}
    +	d.logger.InfoContext(ctx, "sandbox transactional SMS accepted",
    +		"recipient", "***"+message.To[len(message.To)-3:],
    +		"purpose", string(message.Purpose),
    +		"characters", utf8.RuneCountInString(message.Body),
    +	)
    +	return nil
    +}
    Update { move_path: None }: /home/fian/frihk/Ikonex-Academy/backend/internal/notifications/sms_test.go
    @@ -3,4 +3,8 @@
     import (
    +	"bytes"
    +	"context"
     	"errors"
    +	"log/slog"
     	"strings"
    +	"sync"
     	"testing"
    @@ -34,2 +38,63 @@
     
    +func TestSandboxSMSRedactsSensitiveData(t *testing.T) {
    +	var output bytes.Buffer
    +	logger := slog.New(slog.NewJSONHandler(&output, &slog.HandlerOptions{
    +		ReplaceAttr: func(_ []string, attr slog.Attr) slog.Attr {
    +			if attr.Key == slog.TimeKey {
    +				return slog.Attr{}
    +			}
    +			return attr
    +		},
    +	}))
    +	dispatcher := NewSandboxSMSDispatcher(logger)
    +	message := SMSMessage{To: "0712345678", Body: "Your synthetic OTP is 654321", Purpose: WhatsAppSessionOTP}
    +	if err := dispatcher.DispatchSMS(context.Background(), message); err != nil {
    +		t.Fatal(err)
    +	}
    +	first := output.String()
    +	for _, secret := range []string{message.To, "+254712345678", message.Body, "654321"} {
    +		if strings.Contains(first, secret) {
    +			t.Fatal("sandbox exposed sensitive message data")
    +		}
    +	}
    +	if !strings.Contains(first, `"recipient":"***678"`) || !strings.Contains(first, `"purpose":"whatsapp_session"`) {
    +		t.Fatalf("safe dispatch metadata missing: %s", first)
    +	}
    +	output.Reset()
    +	if err := dispatcher.DispatchSMS(context.Background(), message); err != nil || output.String() != first {
    +		t.Fatalf("nondeterministic sandbox output: %v", err)
    +	}
    +	output.Reset()
    +	ctx, cancel := context.WithCancel(context.Background())
    +	cancel()
    +	if err := dispatcher.DispatchSMS(ctx, message); !errors.Is(err, context.Canceled) || output.Len() != 0 {
    +		t.Fatalf("canceled dispatch: %v, %s", err, output.String())
    +	}
    +	message.Purpose = "absence"
    +	if err := dispatcher.DispatchSMS(context.Background(), message); !errors.Is(err, ErrInvalidSMS) || output.Len() != 0 {
    +		t.Fatalf("invalid dispatch: %v, %s", err, output.String())
    +	}
    +}
    +
    +func TestSandboxSMSConcurrentDispatch(t *testing.T) {
    +	var output bytes.Buffer
    +	dispatcher := NewSandboxSMSDispatcher(slog.New(slog.NewJSONHandler(&output, nil)))
    +	var group sync.WaitGroup
    +	for i := 0; i < 20; i++ {
    +		group.Add(1)
    +		go func() {
    +			defer group.Done()
    +			if err := dispatcher.DispatchSMS(context.Background(), SMSMessage{
    +				To: "0712345678", Body: "Synthetic test", Purpose: AppLoginOTP,
    +			}); err != nil {
    +				t.Error(err)
    +			}
    +		}()
    +	}
    +	group.Wait()
    +	if strings.Count(output.String(), "sandbox transactional SMS accepted") != 20 {
    +		t.Fatal("sandbox lost dispatch metadata")
    +	}
    +}
    +
     func TestTransactionalSMSValidation(t *testing.T) {

## Activity

    $ gofmt -w internal/notifications/sms.go internal/notifications/sms_test.go && go test -v -race ./internal/notifications && go vet ./internal/notifications
    === RUN   TestNormalizePhone
    === RUN   TestNormalizePhone/0712345678
    === RUN   TestNormalizePhone/0112345678
    === RUN   TestNormalizePhone/254712345678
    === RUN   TestNormalizePhone/_+254_(712)_345-678_
    === RUN   TestNormalizePhone/+442079460000
    === RUN   TestNormalizePhone/#00
    === RUN   TestNormalizePhone/+
    === RUN   TestNormalizePhone/071234567
    === RUN   TestNormalizePhone/+2547123456789
    === RUN   TestNormalizePhone/+254012345678
    === RUN   TestNormalizePhone/+254212345678
    === RUN   TestNormalizePhone/+0712345678
    === RUN   TestNormalizePhone/0712345678_apiKey=secret
    === RUN   TestNormalizePhone/+254７12345678
    === RUN   TestNormalizePhone/+254+712345678
    === RUN   TestNormalizePhone/712345678
    === RUN   TestNormalizePhone/+1234567890123456
    --- PASS: TestNormalizePhone (0.01s)
        --- PASS: TestNormalizePhone/0712345678 (0.00s)
        --- PASS: TestNormalizePhone/0112345678 (0.00s)
        --- PASS: TestNormalizePhone/254712345678 (0.00s)
        --- PASS: TestNormalizePhone/_+254_(712)_345-678_ (0.00s)
        --- PASS: TestNormalizePhone/+442079460000 (0.00s)
        --- PASS: TestNormalizePhone/#00 (0.00s)
        --- PASS: TestNormalizePhone/+ (0.00s)
        --- PASS: TestNormalizePhone/071234567 (0.00s)
        --- PASS: TestNormalizePhone/+2547123456789 (0.00s)
        --- PASS: TestNormalizePhone/+254012345678 (0.00s)
        --- PASS: TestNormalizePhone/+254212345678 (0.00s)
        --- PASS: TestNormalizePhone/+0712345678 (0.00s)
        --- PASS: TestNormalizePhone/0712345678_apiKey=secret (0.00s)
        --- PASS: TestNormalizePhone/+254７12345678 (0.00s)
        --- PASS: TestNormalizePhone/+254+712345678 (0.00s)
        --- PASS: TestNormalizePhone/712345678 (0.00s)
        --- PASS: TestNormalizePhone/+1234567890123456 (0.00s)
    === RUN   TestSandboxSMSRedactsSensitiveData
    --- PASS: TestSandboxSMSRedactsSensitiveData (0.00s)
    === RUN   TestSandboxSMSConcurrentDispatch
    --- PASS: TestSandboxSMSConcurrentDispatch (0.00s)
    === RUN   TestTransactionalSMSValidation
    --- PASS: TestTransactionalSMSValidation (0.00s)
    PASS
    ok      github.com/Ashomondi/ikonex-academy/backend/internal/notifications    1.041s
    ✓ • 7.30s

## Assistant

`NOTIF-01` passed tests and vet and is committed as `5925aa5`. The sandbox dispatcher is now implemented; its checks cover deterministic log metadata, redaction, cancellation, rejection of event SMS, and concurrent dispatch.

The dependency exception remains pending. I’m keeping `go.mod` and `go.sum` unchanged.

## Activity

    file changes: Completed · 1 changes
    Update { move_path: None }: /home/fian/frihk/Ikonex-Academy/backend/internal/billing/IMPLEMENTATION_PLAN.md
    @@ -56,3 +56,3 @@
     | NOTIF-01 | C3 SMS boundary and normalization | INF-01 | N/sms.go, N/sms_test.go | Define SMSDispatcher and allowed transactional purposes; normalize Kenyan/E.164 phones; validate without leaking content | Table tests for accepted/rejected numbers, purposes and missing content; G(notifications) | Completed; race tests and vet passed |
    -| NOTIF-02 | C3 deterministic sandbox SMS | NOTIF-01 | N/sms.go, N/sms_test.go | Select sandbox by username; log masked recipient and safe metadata, never message/OTP; propagate cancellation | Capture slog output, assert no OTP/full phone/API key; concurrent dispatch; G(notifications) | No network for sandbox; Pending |
    +| NOTIF-02 | C3 deterministic sandbox SMS | NOTIF-01 | N/sms.go, N/sms_test.go | Log masked recipient and safe metadata, never message/OTP; propagate cancellation; username selection follows in NOTIF-03 factory | Capture slog output, assert no OTP/full phone/API key; concurrent dispatch; G(notifications) | Completed; race tests and vet passed |
     | NOTIF-03 | C3 Africa's Talking transport | NOTIF-01 | N/sms.go, N/sms_test.go | Context-bound HTTP client; encoded form and apiKey; validate per-recipient response; classify retryable/permanent failures | httptest request contract, provider rejection, timeout, malformed response, cancellation; G(notifications) | No false delivery success or secret-bearing errors; Pending |

## Activity

    $ git status --short && git diff --name-only && git diff --cached --name-only && git diff --check && git add backend/internal/billing/IMPLEMENTATION_PLAN.md backend/internal/notifications/sms.go backend/internal/notifications/sms_test.go && git diff --cached --name-only && git commit -m 'feat(notifications): add redacted sandbox SMS dispatcher'
     M Docs/DECISIONS.md
     M backend/internal/billing/IMPLEMENTATION_PLAN.md
     M backend/internal/notifications/sms.go
     M backend/internal/notifications/sms_test.go
    Docs/DECISIONS.md
    backend/internal/billing/IMPLEMENTATION_PLAN.md
    backend/internal/notifications/sms.go
    backend/internal/notifications/sms_test.go
    backend/internal/billing/IMPLEMENTATION_PLAN.md
    backend/internal/notifications/sms.go
    backend/internal/notifications/sms_test.go
    [feat/ian-backend-engineer-c 515b2f0] feat(notifications): add redacted sandbox SMS dispatcher
     3 files changed, 96 insertions(+), 1 deletion(-)
    ✓ • 375ms

## Assistant

`NOTIF-02` passed race tests and vet and is committed as `515b2f0`.

Next is `NOTIF-03`, the Africa’s Talking transport. I’ll test request formatting, recipient-level success and failure, retry classification, timeouts, cancellation, and redirect refusal using a local HTTP server. Sandbox configuration will select the local logger and make no HTTP requests.

## Activity

    file changes: Completed · 2 changes
    Add: /home/fian/frihk/Ikonex-Academy/backend/internal/notifications/sms_provider.go
    package notifications
    
    import (
    	"context"
    	"encoding/json"
    	"errors"
    	"io"
    	"log/slog"
    	"net/http"
    	"net/url"
    	"strings"
    	"time"
    )
    
    type SMSConfig struct {
    	Username string
    	APIKey   string
    	SenderID string
    }
    
    // SMSDeliveryError exposes retry policy without provider text, credentials,
    // message bodies, or recipient numbers. Retries belong to the queue worker.
    type SMSDeliveryError struct {
    	Code      string
    	Retryable bool
    }
    
    func (e *SMSDeliveryError) Error() string { return "SMS delivery failed: " + e.Code }
    
    type africaTalkingDispatcher struct {
    	config   SMSConfig
    	client   *http.Client
    	endpoint string
    }
    
    // NewSMSDispatcher selects the local logger for username "sandbox". The HTTP
    // client is copied so timeouts and redirect policy do not mutate a shared client.
    func NewSMSDispatcher(config SMSConfig, client *http.Client, logger *slog.Logger) (SMSDispatcher, error) {
    	config.Username = strings.TrimSpace(config.Username)
    	if config.Username == "sandbox" {
    		return NewSandboxSMSDispatcher(logger), nil
    	}
    	if config.Username == "" || strings.TrimSpace(config.APIKey) == "" {
    		return nil, errors.New("Africa's Talking username and API key are required")
    	}
    	httpClient := http.Client{Timeout: 15 * time.Second}
    	if client != nil {
    		httpClient = *client
    		if httpClient.Timeout <= 0 {
    			httpClient.Timeout = 15 * time.Second
    		}
    	}
    	httpClient.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
    	return &africaTalkingDispatcher{
    		config: config, client: &httpClient,
    		endpoint: "https://api.africastalking.com/version1/messaging",
    	}, nil
    }
    
    func (d *africaTalkingDispatcher) DispatchSMS(ctx context.Context, message SMSMessage) error {
    	if err := ctx.Err(); err != nil {
    		return err
    	}
    	message, err := message.normalized()
    	if err != nil {
    		return err
    	}
    	form := url.Values{"username": {d.config.Username}, "to": {message.To}, "message": {message.Body}}
    	if d.config.SenderID != "" {
    		form.Set("from", d.config.SenderID)
    	}
    	req, err := http.NewRequestWithContext(ctx, http.MethodPost, d.endpoint, strings.NewReader(form.Encode()))
    	if err != nil {
    		return &SMSDeliveryError{Code: "INVALID_REQUEST"}
    	}
    	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
    	req.Header.Set("Accept", "application/json")
    	req.Header.Set("apiKey", d.config.APIKey)
    	resp, err := d.client.Do(req)
    	if err != nil {
    		if ctx.Err() != nil {
    			return ctx.Err()
    		}
    		return &SMSDeliveryError{Code: "TRANSPORT_ERROR", Retryable: true}
    	}
    	defer resp.Body.Close()
    	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
    		return &SMSDeliveryError{Code: "HTTP_ERROR", Retryable: resp.StatusCode == 429 || resp.StatusCode == 408 || resp.StatusCode >= 500}
    	}
    	const maxResponseBytes = 64 << 10
    	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes+1))
    	if err != nil || len(body) > maxResponseBytes {
    		return &SMSDeliveryError{Code: "INVALID_RESPONSE"}
    	}
    	var response struct {
    		Data struct {
    			Recipients []struct {
    				Number     string `json:"number"`
    				StatusCode int    `json:"statusCode"`
    			} `json:"Recipients"`
    		} `json:"SMSMessageData"`
    	}
    	if err := json.Unmarshal(body, &response); err != nil || len(response.Data.Recipients) != 1 {
    		return &SMSDeliveryError{Code: "INVALID_RESPONSE"}
    	}
    	recipient := response.Data.Recipients[0]
    	phone, err := NormalizePhone(recipient.Number)
    	if err != nil || phone != message.To {
    		return &SMSDeliveryError{Code: "RECIPIENT_MISMATCH"}
    	}
    	switch recipient.StatusCode {
    	case 100, 101, 102: // Processed, sent, or queued by the provider.
    		return nil
    	default:
    		return &SMSDeliveryError{Code: "PROVIDER_REJECTED", Retryable: recipient.StatusCode >= 500 && recipient.StatusCode <= 502}
    	}
    }
    Add: /home/fian/frihk/Ikonex-Academy/backend/internal/notifications/sms_provider_test.go
    package notifications
    
    import (
    	"bytes"
    	"context"
    	"errors"
    	"fmt"
    	"io"
    	"log/slog"
    	"net/http"
    	"net/http/httptest"
    	"strings"
    	"sync/atomic"
    	"testing"
    	"time"
    )
    
    func providerDispatcher(t *testing.T, server *httptest.Server) *africaTalkingDispatcher {
    	t.Helper()
    	d, err := NewSMSDispatcher(SMSConfig{Username: "synthetic-user", APIKey: "synthetic-secret", SenderID: "IKONEX"}, server.Client(), nil)
    	if err != nil {
    		t.Fatal(err)
    	}
    	provider := d.(*africaTalkingDispatcher)
    	provider.endpoint = server.URL
    	return provider
    }
    
    func TestAfricaTalkingRequestAndResponses(t *testing.T) {
    	for _, tc := range []struct {
    		name, body string
    		status     int
    		wantCode   string
    		retry      bool
    	}{
    		{"sent", `{"SMSMessageData":{"Recipients":[{"number":"+254712345678","statusCode":101}]}}`, 201, "", false},
    		{"queued", `{"SMSMessageData":{"Recipients":[{"number":"254712345678","statusCode":102}]}}`, 201, "", false},
    		{"processed", `{"SMSMessageData":{"Recipients":[{"number":"254712345678","statusCode":100}]}}`, 201, "", false},
    		{"invalid phone", `{"SMSMessageData":{"Recipients":[{"number":"+254712345678","statusCode":403,"status":"private 654321"}]}}`, 201, "PROVIDER_REJECTED", false},
    		{"gateway failure", `{"SMSMessageData":{"Recipients":[{"number":"+254712345678","statusCode":501}]}}`, 201, "PROVIDER_REJECTED", true},
    		{"wrong recipient", `{"SMSMessageData":{"Recipients":[{"number":"+254712345679","statusCode":101}]}}`, 201, "RECIPIENT_MISMATCH", false},
    		{"no recipients", `{"SMSMessageData":{"Recipients":[]}}`, 201, "INVALID_RESPONSE", false},
    		{"multiple recipients", `{"SMSMessageData":{"Recipients":[{},{}]}}`, 201, "INVALID_RESPONSE", false},
    		{"malformed", `{"secret":"654321"`, 200, "INVALID_RESPONSE", false},
    		{"oversized", strings.Repeat("x", (64<<10)+1), 200, "INVALID_RESPONSE", false},
    		{"rate limited", "private 654321", 429, "HTTP_ERROR", true},
    		{"request timeout", "private 654321", 408, "HTTP_ERROR", true},
    		{"server failure", "private 654321", 503, "HTTP_ERROR", true},
    		{"unauthorized", "synthetic-secret", 401, "HTTP_ERROR", false},
    	} {
    		t.Run(tc.name, func(t *testing.T) {
    			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    				if r.Method != http.MethodPost || r.Header.Get("apiKey") != "synthetic-secret" || r.Header.Get("Accept") != "application/json" {
    					t.Error("incorrect provider request headers or method")
    				}
    				if err := r.ParseForm(); err != nil {
    					t.Error(err)
    				}
    				for key, want := range map[string]string{"username": "synthetic-user", "to": "+254712345678", "from": "IKONEX", "message": "Synthetic OTP 654321 & = +"} {
    					if r.Form.Get(key) != want {
    						t.Errorf("incorrect form field %s", key)
    					}
    				}
    				w.WriteHeader(tc.status)
    				_, _ = io.WriteString(w, tc.body)
    			}))
    			defer server.Close()
    			d := providerDispatcher(t, server)
    			err := d.DispatchSMS(context.Background(), SMSMessage{To: "0712345678", Body: "Synthetic OTP 654321 & = +", Purpose: AppLoginOTP})
    			if tc.wantCode == "" {
    				if err != nil {
    					t.Fatal(err)
    				}
    				return
    			}
    			var deliveryErr *SMSDeliveryError
    			if !errors.As(err, &deliveryErr) || deliveryErr.Code != tc.wantCode || deliveryErr.Retryable != tc.retry {
    				t.Fatalf("got %v; want %s retry=%v", err, tc.wantCode, tc.retry)
    			}
    			for _, secret := range []string{"654321", "synthetic-secret", "712345678"} {
    				if strings.Contains(err.Error(), secret) {
    					t.Fatal("provider error exposed private data")
    				}
    			}
    		})
    	}
    }
    
    func TestSMSFactorySandboxDoesNotUseHTTP(t *testing.T) {
    	var calls atomic.Int32
    	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { calls.Add(1) }))
    	defer server.Close()
    	var log bytes.Buffer
    	d, err := NewSMSDispatcher(SMSConfig{Username: "sandbox"}, server.Client(), slog.New(slog.NewJSONHandler(&log, nil)))
    	if err != nil {
    		t.Fatal(err)
    	}
    	if _, ok := d.(*sandboxSMSDispatcher); !ok {
    		t.Fatal("sandbox selected a provider client")
    	}
    	if err := d.DispatchSMS(context.Background(), SMSMessage{To: "0712345678", Body: "Synthetic OTP", Purpose: AppLoginOTP}); err != nil || calls.Load() != 0 || log.Len() == 0 {
    		t.Fatalf("sandbox did not log locally: %v", err)
    	}
    	for _, config := range []SMSConfig{{}, {Username: "synthetic-user"}, {APIKey: "secret"}} {
    		if _, err := NewSMSDispatcher(config, nil, nil); err == nil {
    			t.Fatal("incomplete provider configuration accepted")
    		}
    	}
    }
    
    func TestAfricaTalkingRejectsRedirects(t *testing.T) {
    	var redirected atomic.Int32
    	target := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { redirected.Add(1) }))
    	defer target.Close()
    	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    		http.Redirect(w, r, target.URL, http.StatusTemporaryRedirect)
    	}))
    	defer server.Close()
    	d := providerDispatcher(t, server)
    	err := d.DispatchSMS(context.Background(), SMSMessage{To: "0712345678", Body: "Synthetic OTP", Purpose: AppLoginOTP})
    	var deliveryErr *SMSDeliveryError
    	if !errors.As(err, &deliveryErr) || deliveryErr.Retryable || redirected.Load() != 0 {
    		t.Fatalf("redirect was followed or classified incorrectly: %v", err)
    	}
    }
    
    func TestAfricaTalkingCancellationAndTimeout(t *testing.T) {
    	var calls atomic.Int32
    	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    		calls.Add(1)
    		<-r.Context().Done()
    	}))
    	defer server.Close()
    	d := providerDispatcher(t, server)
    	msg := SMSMessage{To: "0712345678", Body: "Synthetic OTP", Purpose: AppLoginOTP}
    	ctx, cancel := context.WithCancel(context.Background())
    	cancel()
    	if err := d.DispatchSMS(ctx, msg); !errors.Is(err, context.Canceled) || calls.Load() != 0 {
    		t.Fatalf("canceled dispatch reached provider: %v", err)
    	}
    	msg.Purpose = "absence"
    	if err := d.DispatchSMS(context.Background(), msg); !errors.Is(err, ErrInvalidSMS) || calls.Load() != 0 {
    		t.Fatalf("invalid dispatch reached provider: %v", err)
    	}
    	msg.Purpose = AppLoginOTP
    	d.client.Timeout = 30 * time.Millisecond
    	var deliveryErr *SMSDeliveryError
    	if err := d.DispatchSMS(context.Background(), msg); !errors.As(err, &deliveryErr) || !deliveryErr.Retryable {
    		t.Fatalf("timeout retry classification: %v", err)
    	}
    }
    
    func TestSMSFactoryDoesNotMutateHTTPClient(t *testing.T) {
    	client := &http.Client{}
    	dispatcher, err := NewSMSDispatcher(SMSConfig{Username: "synthetic-user", APIKey: "synthetic-secret"}, client, nil)
    	if err != nil {
    		t.Fatal(err)
    	}
    	d := dispatcher.(*africaTalkingDispatcher)
    	if client.Timeout != 0 || client.CheckRedirect != nil || d.client == client || d.client.Timeout <= 0 {
    		t.Fatal("constructor mutated shared client or left timeout unset")
    	}
    	if got := fmt.Sprint(d.client.CheckRedirect(nil, nil)); got != http.ErrUseLastResponse.Error() {
    		t.Fatalf("redirect policy: %s", got)
    	}
    }

## Activity

    file changes: Completed · 2 changes
    Update { move_path: None }: /home/fian/frihk/Ikonex-Academy/backend/internal/billing/IMPLEMENTATION_PLAN.md
    @@ -57,3 +57,3 @@
     | NOTIF-02 | C3 deterministic sandbox SMS | NOTIF-01 | N/sms.go, N/sms_test.go | Log masked recipient and safe metadata, never message/OTP; propagate cancellation; username selection follows in NOTIF-03 factory | Capture slog output, assert no OTP/full phone/API key; concurrent dispatch; G(notifications) | Completed; race tests and vet passed |
    -| NOTIF-03 | C3 Africa's Talking transport | NOTIF-01 | N/sms.go, N/sms_test.go | Context-bound HTTP client; encoded form and apiKey; validate per-recipient response; classify retryable/permanent failures | httptest request contract, provider rejection, timeout, malformed response, cancellation; G(notifications) | No false delivery success or secret-bearing errors; Pending |
    +| NOTIF-03 | C3 Africa's Talking transport | NOTIF-01 | N/sms_provider.go, N/sms_provider_test.go | Select local sandbox by username; context-bound HTTP client; encoded form and apiKey; validate per-recipient response; classify retryable/permanent failures | httptest request contract, provider rejection, timeout, malformed response, cancellation; G(notifications) | In Progress |
     | INF-02 | C1–C3 transaction/auth/error contracts | INF-01, pgx | I/contracts/contracts.go, I/contracts/contracts_test.go | Small consumer contracts for pgx tenant transactions/audit/identity; APIError and standard envelopes; deny missing identity | Contract doubles, malformed identity, error envelope and safe error mapping; G(internalapi/...) | Compatible injection, no substitute auth; Blocked on module dependency |
    Update { move_path: None }: /home/fian/frihk/Ikonex-Academy/backend/internal/notifications/sms_provider_test.go
    @@ -128,7 +128,9 @@
     	var calls atomic.Int32
    +	release := make(chan struct{})
     	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
     		calls.Add(1)
    -		<-r.Context().Done()
    +		<-release
     	}))
     	defer server.Close()
    +	defer close(release)
     	d := providerDispatcher(t, server)

## Activity

    $ git diff --check && git diff --stat && git status --short && git diff --cached --name-only && git diff -- Docs/DECISIONS.md
     Docs/DECISIONS.md                               | 1 +
     backend/internal/billing/IMPLEMENTATION_PLAN.md | 2 +-
     2 files changed, 2 insertions(+), 1 deletion(-)
     M Docs/DECISIONS.md
     M backend/internal/billing/IMPLEMENTATION_PLAN.md
    ?? backend/internal/notifications/sms_provider.go
    ?? backend/internal/notifications/sms_provider_test.go
    diff --git a/Docs/DECISIONS.md b/Docs/DECISIONS.md
    index a9a94b6..ac82a09 100644
    --- a/Docs/DECISIONS.md
    +++ b/Docs/DECISIONS.md
    @@ -1,3 +1,4 @@
    +
     # DECISIONS.md
    
     A short, dated log of decisions worth remembering — so nobody (human or AI
    ✓ • 562ms

## Activity

    $ gofmt -w internal/notifications/sms_provider.go internal/notifications/sms_provider_test.go && go test -v -race ./internal/notifications && go vet ./internal/notifications
    === RUN   TestAfricaTalkingRequestAndResponses
    === RUN   TestAfricaTalkingRequestAndResponses/sent
    === RUN   TestAfricaTalkingRequestAndResponses/queued
    === RUN   TestAfricaTalkingRequestAndResponses/processed
    === RUN   TestAfricaTalkingRequestAndResponses/invalid_phone
    === RUN   TestAfricaTalkingRequestAndResponses/gateway_failure
    === RUN   TestAfricaTalkingRequestAndResponses/wrong_recipient
    === RUN   TestAfricaTalkingRequestAndResponses/no_recipients
    === RUN   TestAfricaTalkingRequestAndResponses/multiple_recipients
    === RUN   TestAfricaTalkingRequestAndResponses/malformed
    === RUN   TestAfricaTalkingRequestAndResponses/oversized
    === RUN   TestAfricaTalkingRequestAndResponses/rate_limited
    === RUN   TestAfricaTalkingRequestAndResponses/request_timeout
    === RUN   TestAfricaTalkingRequestAndResponses/server_failure
    === RUN   TestAfricaTalkingRequestAndResponses/unauthorized
    --- PASS: TestAfricaTalkingRequestAndResponses (0.08s)
        --- PASS: TestAfricaTalkingRequestAndResponses/sent (0.01s)
        --- PASS: TestAfricaTalkingRequestAndResponses/queued (0.00s)
        --- PASS: TestAfricaTalkingRequestAndResponses/processed (0.01s)
        --- PASS: TestAfricaTalkingRequestAndResponses/invalid_phone (0.00s)
        --- PASS: TestAfricaTalkingRequestAndResponses/gateway_failure (0.00s)
        --- PASS: TestAfricaTalkingRequestAndResponses/wrong_recipient (0.01s)
        --- PASS: TestAfricaTalkingRequestAndResponses/no_recipients (0.01s)
        --- PASS: TestAfricaTalkingRequestAndResponses/multiple_recipients (0.00s)
        --- PASS: TestAfricaTalkingRequestAndResponses/malformed (0.00s)
        --- PASS: TestAfricaTalkingRequestAndResponses/oversized (0.01s)
        --- PASS: TestAfricaTalkingRequestAndResponses/rate_limited (0.00s)
        --- PASS: TestAfricaTalkingRequestAndResponses/request_timeout (0.00s)
        --- PASS: TestAfricaTalkingRequestAndResponses/server_failure (0.01s)
        --- PASS: TestAfricaTalkingRequestAndResponses/unauthorized (0.01s)
    === RUN   TestSMSFactorySandboxDoesNotUseHTTP
    --- PASS: TestSMSFactorySandboxDoesNotUseHTTP (0.00s)
    === RUN   TestAfricaTalkingRejectsRedirects
    --- PASS: TestAfricaTalkingRejectsRedirects (0.01s)
    === RUN   TestAfricaTalkingCancellationAndTimeout
    --- PASS: TestAfricaTalkingCancellationAndTimeout (0.03s)
    === RUN   TestSMSFactoryDoesNotMutateHTTPClient
    --- PASS: TestSMSFactoryDoesNotMutateHTTPClient (0.00s)
    === RUN   TestNormalizePhone
    === RUN   TestNormalizePhone/0712345678
    === RUN   TestNormalizePhone/0112345678
    === RUN   TestNormalizePhone/254712345678
    === RUN   TestNormalizePhone/_+254_(712)_345-678_
    === RUN   TestNormalizePhone/+442079460000
    === RUN   TestNormalizePhone/#00
    === RUN   TestNormalizePhone/+
    === RUN   TestNormalizePhone/071234567
    === RUN   TestNormalizePhone/+2547123456789
    === RUN   TestNormalizePhone/+254012345678
    === RUN   TestNormalizePhone/+254212345678
    === RUN   TestNormalizePhone/+0712345678
    === RUN   TestNormalizePhone/0712345678_apiKey=secret
    === RUN   TestNormalizePhone/+254７12345678
    === RUN   TestNormalizePhone/+254+712345678
    === RUN   TestNormalizePhone/712345678
    === RUN   TestNormalizePhone/+1234567890123456
    --- PASS: TestNormalizePhone (0.02s)
        --- PASS: TestNormalizePhone/0712345678 (0.00s)
        --- PASS: TestNormalizePhone/0112345678 (0.00s)
        --- PASS: TestNormalizePhone/254712345678 (0.00s)
        --- PASS: TestNormalizePhone/_+254_(712)_345-678_ (0.00s)
        --- PASS: TestNormalizePhone/+442079460000 (0.00s)
        --- PASS: TestNormalizePhone/#00 (0.00s)
        --- PASS: TestNormalizePhone/+ (0.00s)
        --- PASS: TestNormalizePhone/071234567 (0.00s)
        --- PASS: TestNormalizePhone/+2547123456789 (0.00s)
        --- PASS: TestNormalizePhone/+254012345678 (0.00s)
        --- PASS: TestNormalizePhone/+254212345678 (0.00s)
        --- PASS: TestNormalizePhone/+0712345678 (0.00s)
        --- PASS: TestNormalizePhone/0712345678_apiKey=secret (0.00s)
        --- PASS: TestNormalizePhone/+254７12345678 (0.00s)
        --- PASS: TestNormalizePhone/+254+712345678 (0.00s)
        --- PASS: TestNormalizePhone/712345678 (0.00s)
        --- PASS: TestNormalizePhone/+1234567890123456 (0.00s)
    === RUN   TestSandboxSMSRedactsSensitiveData
    --- PASS: TestSandboxSMSRedactsSensitiveData (0.00s)
    === RUN   TestSandboxSMSConcurrentDispatch
    --- PASS: TestSandboxSMSConcurrentDispatch (0.00s)
    === RUN   TestTransactionalSMSValidation
    --- PASS: TestTransactionalSMSValidation (0.00s)
    PASS
    ok      github.com/Ashomondi/ikonex-academy/backend/internal/notifications    1.186s
    ✓ • 2m 12s

## Assistant

The Africa’s Talking transport tests passed under the race detector, including provider failures, malformed responses, timeouts, and redirects. The diff review also confirms that your existing `Docs/DECISIONS.md` edit is unchanged.

The next foundation task needs the documented `pgx.Tx` contract. That is currently blocked by the missing dependency and the explicit prohibition on editing `go.mod`/`go.sum`. I’m finishing the verification of the completed SMS work while the dependency decision remains pending.

## Activity

    $ npm run lint && npm run build
    
    > ikonex-academy-frontend@0.1.0 lint
    > next lint
    
    sh: 1: next: not found
    ✗ (127) • 1.87s

## Activity

    $ go test -v -race ./... && go build ./... && go vet ./... && gofmt -l internal/notifications/sms.go internal/notifications/sms_test.go internal/notifications/sms_provider.go internal/notifications/sms_provider_test.go
    ?       github.com/Ashomondi/ikonex-academy/backend/cmd/server    [no test files]
    ?       github.com/Ashomondi/ikonex-academy/backend/cmd/worker    [no test files]
    ?       github.com/Ashomondi/ikonex-academy/backend/internal/academics    [no test files]
    ?       github.com/Ashomondi/ikonex-academy/backend/internal/auth    [no test files]
    ?       github.com/Ashomondi/ikonex-academy/backend/internal/billing    [no test files]
    ?       github.com/Ashomondi/ikonex-academy/backend/internal/clubs    [no test files]
    ?       github.com/Ashomondi/ikonex-academy/backend/internal/config    [no test files]
    ?       github.com/Ashomondi/ikonex-academy/backend/internal/consent    [no test files]
    ?       github.com/Ashomondi/ikonex-academy/backend/internal/database    [no test files]
    ?       github.com/Ashomondi/ikonex-academy/backend/internal/internalapi    [no test files]
    ?       github.com/Ashomondi/ikonex-academy/backend/internal/jobs    [no test files]
    ?       github.com/Ashomondi/ikonex-academy/backend/internal/library    [no test files]
    ?       github.com/Ashomondi/ikonex-academy/backend/internal/middleware    [no test files]
    ?       github.com/Ashomondi/ikonex-academy/backend/internal/nurse    [no test files]
    ?       github.com/Ashomondi/ikonex-academy/backend/internal/onboarding    [no test files]
    === RUN   TestAfricaTalkingRequestAndResponses
    === RUN   TestAfricaTalkingRequestAndResponses/sent
    === RUN   TestAfricaTalkingRequestAndResponses/queued
    === RUN   TestAfricaTalkingRequestAndResponses/processed
    === RUN   TestAfricaTalkingRequestAndResponses/invalid_phone
    === RUN   TestAfricaTalkingRequestAndResponses/gateway_failure
    === RUN   TestAfricaTalkingRequestAndResponses/wrong_recipient
    === RUN   TestAfricaTalkingRequestAndResponses/no_recipients
    === RUN   TestAfricaTalkingRequestAndResponses/multiple_recipients
    === RUN   TestAfricaTalkingRequestAndResponses/malformed
    === RUN   TestAfricaTalkingRequestAndResponses/oversized
    === RUN   TestAfricaTalkingRequestAndResponses/rate_limited
    === RUN   TestAfricaTalkingRequestAndResponses/request_timeout
    === RUN   TestAfricaTalkingRequestAndResponses/server_failure
    === RUN   TestAfricaTalkingRequestAndResponses/unauthorized
    --- PASS: TestAfricaTalkingRequestAndResponses (0.08s)
        --- PASS: TestAfricaTalkingRequestAndResponses/sent (0.01s)
        --- PASS: TestAfricaTalkingRequestAndResponses/queued (0.00s)
        --- PASS: TestAfricaTalkingRequestAndResponses/processed (0.01s)
        --- PASS: TestAfricaTalkingRequestAndResponses/invalid_phone (0.00s)
        --- PASS: TestAfricaTalkingRequestAndResponses/gateway_failure (0.00s)
        --- PASS: TestAfricaTalkingRequestAndResponses/wrong_recipient (0.01s)
        --- PASS: TestAfricaTalkingRequestAndResponses/no_recipients (0.01s)
        --- PASS: TestAfricaTalkingRequestAndResponses/multiple_recipients (0.00s)
        --- PASS: TestAfricaTalkingRequestAndResponses/malformed (0.00s)
        --- PASS: TestAfricaTalkingRequestAndResponses/oversized (0.01s)
        --- PASS: TestAfricaTalkingRequestAndResponses/rate_limited (0.00s)
        --- PASS: TestAfricaTalkingRequestAndResponses/request_timeout (0.00s)
        --- PASS: TestAfricaTalkingRequestAndResponses/server_failure (0.01s)
        --- PASS: TestAfricaTalkingRequestAndResponses/unauthorized (0.01s)
    === RUN   TestSMSFactorySandboxDoesNotUseHTTP
    --- PASS: TestSMSFactorySandboxDoesNotUseHTTP (0.00s)
    === RUN   TestAfricaTalkingRejectsRedirects
    --- PASS: TestAfricaTalkingRejectsRedirects (0.01s)
    === RUN   TestAfricaTalkingCancellationAndTimeout
    --- PASS: TestAfricaTalkingCancellationAndTimeout (0.03s)
    === RUN   TestSMSFactoryDoesNotMutateHTTPClient
    --- PASS: TestSMSFactoryDoesNotMutateHTTPClient (0.00s)
    === RUN   TestNormalizePhone
    === RUN   TestNormalizePhone/0712345678
    === RUN   TestNormalizePhone/0112345678
    === RUN   TestNormalizePhone/254712345678
    === RUN   TestNormalizePhone/_+254_(712)_345-678_
    === RUN   TestNormalizePhone/+442079460000
    === RUN   TestNormalizePhone/#00
    === RUN   TestNormalizePhone/+
    === RUN   TestNormalizePhone/071234567
    === RUN   TestNormalizePhone/+2547123456789
    === RUN   TestNormalizePhone/+254012345678
    === RUN   TestNormalizePhone/+254212345678
    === RUN   TestNormalizePhone/+0712345678
    === RUN   TestNormalizePhone/0712345678_apiKey=secret
    === RUN   TestNormalizePhone/+254７12345678
    === RUN   TestNormalizePhone/+254+712345678
    === RUN   TestNormalizePhone/712345678
    === RUN   TestNormalizePhone/+1234567890123456
    --- PASS: TestNormalizePhone (0.02s)
        --- PASS: TestNormalizePhone/0712345678 (0.00s)
        --- PASS: TestNormalizePhone/0112345678 (0.00s)
        --- PASS: TestNormalizePhone/254712345678 (0.00s)
        --- PASS: TestNormalizePhone/_+254_(712)_345-678_ (0.00s)
        --- PASS: TestNormalizePhone/+442079460000 (0.00s)
        --- PASS: TestNormalizePhone/#00 (0.00s)
        --- PASS: TestNormalizePhone/+ (0.00s)
        --- PASS: TestNormalizePhone/071234567 (0.00s)
        --- PASS: TestNormalizePhone/+2547123456789 (0.00s)
        --- PASS: TestNormalizePhone/+254012345678 (0.00s)
        --- PASS: TestNormalizePhone/+254212345678 (0.00s)
        --- PASS: TestNormalizePhone/+0712345678 (0.00s)
        --- PASS: TestNormalizePhone/0712345678_apiKey=secret (0.00s)
        --- PASS: TestNormalizePhone/+254７12345678 (0.00s)
        --- PASS: TestNormalizePhone/+254+712345678 (0.00s)
        --- PASS: TestNormalizePhone/712345678 (0.00s)
        --- PASS: TestNormalizePhone/+1234567890123456 (0.00s)
    === RUN   TestSandboxSMSRedactsSensitiveData
    --- PASS: TestSandboxSMSRedactsSensitiveData (0.00s)
    === RUN   TestSandboxSMSConcurrentDispatch
    --- PASS: TestSandboxSMSConcurrentDispatch (0.00s)
    === RUN   TestTransactionalSMSValidation
    --- PASS: TestTransactionalSMSValidation (0.00s)
    PASS
    ok      github.com/Ashomondi/ikonex-academy/backend/internal/notifications    (cached)
    ✓ • 5.50s

## Activity

    $ make test
    Running Go Backend tests...
    cd backend && go test -v ./...
    ?       github.com/Ashomondi/ikonex-academy/backend/cmd/server    [no test files]
    ?       github.com/Ashomondi/ikonex-academy/backend/cmd/worker    [no test files]
    ?       github.com/Ashomondi/ikonex-academy/backend/internal/academics    [no test files]
    ?       github.com/Ashomondi/ikonex-academy/backend/internal/auth    [no test files]
    ?       github.com/Ashomondi/ikonex-academy/backend/internal/billing    [no test files]
    ?       github.com/Ashomondi/ikonex-academy/backend/internal/clubs    [no test files]
    ?       github.com/Ashomondi/ikonex-academy/backend/internal/config    [no test files]
    ?       github.com/Ashomondi/ikonex-academy/backend/internal/consent    [no test files]
    ?       github.com/Ashomondi/ikonex-academy/backend/internal/database    [no test files]
    ?       github.com/Ashomondi/ikonex-academy/backend/internal/internalapi    [no test files]
    ?       github.com/Ashomondi/ikonex-academy/backend/internal/jobs    [no test files]
    ?       github.com/Ashomondi/ikonex-academy/backend/internal/library    [no test files]
    ?       github.com/Ashomondi/ikonex-academy/backend/internal/middleware    [no test files]
    ?       github.com/Ashomondi/ikonex-academy/backend/internal/onboarding    [no test files]
    ?       github.com/Ashomondi/ikonex-academy/backend/internal/nurse    [no test files]
    === RUN   TestAfricaTalkingRequestAndResponses
    === RUN   TestAfricaTalkingRequestAndResponses/sent
    === RUN   TestAfricaTalkingRequestAndResponses/queued
    === RUN   TestAfricaTalkingRequestAndResponses/processed
    === RUN   TestAfricaTalkingRequestAndResponses/invalid_phone
    === RUN   TestAfricaTalkingRequestAndResponses/gateway_failure
    === RUN   TestAfricaTalkingRequestAndResponses/wrong_recipient
    === RUN   TestAfricaTalkingRequestAndResponses/no_recipients
    === RUN   TestAfricaTalkingRequestAndResponses/multiple_recipients
    === RUN   TestAfricaTalkingRequestAndResponses/malformed
    === RUN   TestAfricaTalkingRequestAndResponses/oversized
    === RUN   TestAfricaTalkingRequestAndResponses/rate_limited
    === RUN   TestAfricaTalkingRequestAndResponses/request_timeout
    === RUN   TestAfricaTalkingRequestAndResponses/server_failure
    === RUN   TestAfricaTalkingRequestAndResponses/unauthorized
    --- PASS: TestAfricaTalkingRequestAndResponses (0.15s)
        --- PASS: TestAfricaTalkingRequestAndResponses/sent (0.01s)
        --- PASS: TestAfricaTalkingRequestAndResponses/queued (0.01s)
        --- PASS: TestAfricaTalkingRequestAndResponses/processed (0.01s)
        --- PASS: TestAfricaTalkingRequestAndResponses/invalid_phone (0.02s)
        --- PASS: TestAfricaTalkingRequestAndResponses/gateway_failure (0.02s)
        --- PASS: TestAfricaTalkingRequestAndResponses/wrong_recipient (0.03s)
        --- PASS: TestAfricaTalkingRequestAndResponses/no_recipients (0.00s)
        --- PASS: TestAfricaTalkingRequestAndResponses/multiple_recipients (0.01s)
        --- PASS: TestAfricaTalkingRequestAndResponses/malformed (0.00s)
        --- PASS: TestAfricaTalkingRequestAndResponses/oversized (0.00s)
        --- PASS: TestAfricaTalkingRequestAndResponses/rate_limited (0.00s)
        --- PASS: TestAfricaTalkingRequestAndResponses/request_timeout (0.01s)
        --- PASS: TestAfricaTalkingRequestAndResponses/server_failure (0.01s)
        --- PASS: TestAfricaTalkingRequestAndResponses/unauthorized (0.00s)
    === RUN   TestSMSFactorySandboxDoesNotUseHTTP
    --- PASS: TestSMSFactorySandboxDoesNotUseHTTP (0.00s)
    === RUN   TestAfricaTalkingRejectsRedirects
    --- PASS: TestAfricaTalkingRejectsRedirects (0.00s)
    === RUN   TestAfricaTalkingCancellationAndTimeout
    --- PASS: TestAfricaTalkingCancellationAndTimeout (0.03s)
    === RUN   TestSMSFactoryDoesNotMutateHTTPClient
    --- PASS: TestSMSFactoryDoesNotMutateHTTPClient (0.00s)
    === RUN   TestNormalizePhone
    === RUN   TestNormalizePhone/0712345678
    === RUN   TestNormalizePhone/0112345678
    === RUN   TestNormalizePhone/254712345678
    === RUN   TestNormalizePhone/_+254_(712)_345-678_
    === RUN   TestNormalizePhone/+442079460000
    === RUN   TestNormalizePhone/#00
    === RUN   TestNormalizePhone/+
    === RUN   TestNormalizePhone/071234567
    === RUN   TestNormalizePhone/+2547123456789
    === RUN   TestNormalizePhone/+254012345678
    === RUN   TestNormalizePhone/+254212345678
    === RUN   TestNormalizePhone/+0712345678
    === RUN   TestNormalizePhone/0712345678_apiKey=secret
    === RUN   TestNormalizePhone/+254７12345678
    === RUN   TestNormalizePhone/+254+712345678
    === RUN   TestNormalizePhone/712345678
    === RUN   TestNormalizePhone/+1234567890123456
    --- PASS: TestNormalizePhone (0.01s)
        --- PASS: TestNormalizePhone/0712345678 (0.00s)
        --- PASS: TestNormalizePhone/0112345678 (0.00s)
        --- PASS: TestNormalizePhone/254712345678 (0.00s)
        --- PASS: TestNormalizePhone/_+254_(712)_345-678_ (0.00s)
        --- PASS: TestNormalizePhone/+442079460000 (0.00s)
        --- PASS: TestNormalizePhone/#00 (0.00s)
        --- PASS: TestNormalizePhone/+ (0.00s)
        --- PASS: TestNormalizePhone/071234567 (0.00s)
        --- PASS: TestNormalizePhone/+2547123456789 (0.00s)
        --- PASS: TestNormalizePhone/+254012345678 (0.00s)
        --- PASS: TestNormalizePhone/+254212345678 (0.00s)
        --- PASS: TestNormalizePhone/+0712345678 (0.00s)
        --- PASS: TestNormalizePhone/0712345678_apiKey=secret (0.00s)
        --- PASS: TestNormalizePhone/+254７12345678 (0.00s)
        --- PASS: TestNormalizePhone/+254+712345678 (0.00s)
        --- PASS: TestNormalizePhone/712345678 (0.00s)
        --- PASS: TestNormalizePhone/+1234567890123456 (0.00s)
    === RUN   TestSandboxSMSRedactsSensitiveData
    --- PASS: TestSandboxSMSRedactsSensitiveData (0.00s)
    === RUN   TestSandboxSMSConcurrentDispatch
    --- PASS: TestSandboxSMSConcurrentDispatch (0.00s)
    === RUN   TestTransactionalSMSValidation
    --- PASS: TestTransactionalSMSValidation (0.00s)
    PASS
    ok      github.com/Ashomondi/ikonex-academy/backend/internal/notifications    0.235s
    Running Python AI Service tests...
    cd ai_service && python3 -m pytest tests/ || true
    /usr/bin/python3: No module named pytest
    ✓ • 27.38s

## Activity

    $ npm run build
    
    > ikonex-academy-frontend@0.1.0 build
    > next build
    
    sh: 1: next: not found
    ✗ (127) • 244ms

## Activity

    $ rg -n '^func Test|^func New|^func Normalize|^type SMS' backend/internal/notifications && git diff --check
    backend/internal/notifications/sms_provider.go:15:type SMSConfig struct {
    backend/internal/notifications/sms_provider.go:23:type SMSDeliveryError struct {
    backend/internal/notifications/sms_provider.go:38:func NewSMSDispatcher(config SMSConfig, client *http.Client, logger *slog.Logger) (SMSDispatcher, error) {
    backend/internal/notifications/sms_provider_test.go:29:func TestAfricaTalkingRequestAndResponses(t *testing.T) {
    backend/internal/notifications/sms_provider_test.go:89:func TestSMSFactorySandboxDoesNotUseHTTP(t *testing.T) {
    backend/internal/notifications/sms_provider_test.go:111:func TestAfricaTalkingRejectsRedirects(t *testing.T) {
    backend/internal/notifications/sms_provider_test.go:127:func TestAfricaTalkingCancellationAndTimeout(t *testing.T) {
    backend/internal/notifications/sms_provider_test.go:155:func TestSMSFactoryDoesNotMutateHTTPClient(t *testing.T) {
    backend/internal/notifications/sms.go:11:type SMSPurpose string
    backend/internal/notifications/sms.go:26:type SMSMessage struct {
    backend/internal/notifications/sms.go:34:type SMSDispatcher interface {
    backend/internal/notifications/sms.go:40:func NormalizePhone(phone string) (string, error) {
    backend/internal/notifications/sms.go:97:func NewSandboxSMSDispatcher(logger *slog.Logger) SMSDispatcher {
    backend/internal/notifications/sms_test.go:13:func TestNormalizePhone(t *testing.T) {
    backend/internal/notifications/sms_test.go:39:func TestSandboxSMSRedactsSensitiveData(t *testing.T) {
    backend/internal/notifications/sms_test.go:79:func TestSandboxSMSConcurrentDispatch(t *testing.T) {
    backend/internal/notifications/sms_test.go:100:func TestTransactionalSMSValidation(t *testing.T) {
    ✓ • 18ms

## Assistant

Full backend race tests, build, vet, and formatting checks passed. The broader checks found existing setup gaps: Python tests could not run because `pytest` is missing, and frontend lint/build failed because `next` is unavailable. `make test` suppresses the Python failure, so its zero exit code does not mean the full repository passed.

I’ll record those results and commit the SMS transport. The next task remains blocked on the protected module-file dependency decision.

## Activity

    file changes: Completed · 2 changes
    Update { move_path: None }: /home/fian/frihk/Ikonex-Academy/backend/internal/billing/IMPLEMENTATION_PLAN.md
    @@ -57,3 +57,3 @@
     | NOTIF-02 | C3 deterministic sandbox SMS | NOTIF-01 | N/sms.go, N/sms_test.go | Log masked recipient and safe metadata, never message/OTP; propagate cancellation; username selection follows in NOTIF-03 factory | Capture slog output, assert no OTP/full phone/API key; concurrent dispatch; G(notifications) | Completed; race tests and vet passed |
    -| NOTIF-03 | C3 Africa's Talking transport | NOTIF-01 | N/sms_provider.go, N/sms_provider_test.go | Select local sandbox by username; context-bound HTTP client; encoded form and apiKey; validate per-recipient response; classify retryable/permanent failures | httptest request contract, provider rejection, timeout, malformed response, cancellation; G(notifications) | In Progress |
    +| NOTIF-03 | C3 Africa's Talking transport | NOTIF-01 | N/sms_provider.go, N/sms_provider_test.go | Select local sandbox by username; context-bound HTTP client; encoded form and apiKey; validate per-recipient response; classify retryable/permanent failures | httptest request contract, provider rejection, timeout, malformed response, cancellation; G(notifications) | Completed; package and full-backend race tests, build and vet passed |
     | INF-02 | C1–C3 transaction/auth/error contracts | INF-01, pgx | I/contracts/contracts.go, I/contracts/contracts_test.go | Small consumer contracts for pgx tenant transactions/audit/identity; APIError and standard envelopes; deny missing identity | Contract doubles, malformed identity, error envelope and safe error mapping; G(internalapi/...) | Compatible injection, no substitute auth; Blocked on module dependency |
    Add: /home/fian/frihk/Ikonex-Academy/backend/internal/billing/VERIFICATION.md
    # Developer 05 checkpoint report — 2026-11-09
    
    This is an incomplete implementation checkpoint, not Release 1 sign-off. Phase Zero, the detailed plan, and NOTIF-01 through NOTIF-03 are complete. The next task, INF-02, requires the missing pgx dependency in module files explicitly protected by the assignment. An exception for the exact dependency proposal in IMPLEMENTATION_PLAN.md has been requested and has not been assumed.
    
    ## Git baseline and ownership
    
    - Branch: `feat/ian-backend-engineer-c`.
    - Starting commit: `b2b2e3b6f15e60f25947f04c39b8f6b12b5a268f` on `main`, also the verified remote main tip.
    - No `release-1` or existing Ian branch was available. Created the requested branch without resetting/stashing/discarding anything.
    - Existing unrelated change: `Docs/DECISIONS.md` has a leading blank line. It remains unchanged and unstaged.
    - Tested commits preceding this report: `5925aa5` (SMS contract and plan), `515b2f0` (sandbox dispatcher). The commit containing this report adds the verified provider transport. Its exact hash is reported in the user-facing checkpoint response.
    - No push or merge performed.
    - `git status --short`, `git diff --name-only`, `git diff --cached --name-only`, and `git diff --check` checked before commits. Every staged implementation path belongs to Developer 05.
    
    ## Files created
    
    | File | Purpose |
    |---|---|
    | `backend/internal/billing/IMPLEMENTATION_PLAN.md` | Repository assessment, dependency map, detailed tasks, gates and acceptance tracker |
    | `backend/internal/billing/VERIFICATION.md` | This checkpoint's evidence and limitations |
    | `backend/internal/notifications/sms.go` | SMS consumer contract, transactional purposes, phone normalization and safe sandbox dispatcher |
    | `backend/internal/notifications/sms_test.go` | Normalization, purpose validation, sandbox redaction/determinism/cancellation/concurrency |
    | `backend/internal/notifications/sms_provider.go` | Factory and Africa's Talking HTTP transport with safe errors and retry classification |
    | `backend/internal/notifications/sms_provider_test.go` | Local HTTP provider contract and failure tests |
    
    No pre-existing tracked source files were modified. The original `notifications.go` remains a placeholder for the later in-app service. No module files, migrations, shared code, frontend, AI service, or external documentation were changed.
    
    ## Implemented and tested
    
    - `SMSDispatcher.DispatchSMS(context.Context, SMSMessage) error` is the consumer-facing dependency for auth/onboarding/WhatsApp and, later, the queue worker.
    - Supported purposes: app-login OTP, WhatsApp OTP, phone-change OTP, old-number phone-change notice. Other purposes, including absence and sickbay events, are rejected before delivery.
    - `NormalizePhone` accepts Kenyan local and country-prefixed numbers and international E.164 forms; normalizes common display separators; rejects malformed input without reflecting private input in errors.
    - `NewSMSDispatcher(SMSConfig, *http.Client, *slog.Logger)` selects local-only behavior when `Username == "sandbox"`; no key or network is required in that mode.
    - Sandbox logging contains only masked recipient suffix, purpose, and character count. It does not log body, OTP, full number, or credentials. Tests remove the logger timestamp to compare deterministic metadata.
    - Provider transport uses a context-bound form POST, API key header, normalized recipient, optional sender ID, and JSON response validation. It verifies a single matching recipient and recognized processed/sent/queued status before success.
    - A copied HTTP client receives a finite default timeout and refuses redirects, preventing credentials and SMS bodies from being forwarded to a redirect target.
    - `SMSDeliveryError` classifies transport failures, HTTP 408/429/5xx, and selected provider failures for retry without exposing provider text or private data. Invalid/ambiguous response bodies do not trigger automatic retry classification.
    
    ## Test evidence
    
    Nine top-level tests, including 18 phone cases and 14 provider-response cases:
    
    | Test | Evidence |
    |---|---|
    | TestNormalizePhone | Local/E.164 normalization, incorrect lengths/prefixes, Unicode digits, embedded private/injected text |
    | TestTransactionalSMSValidation | All four allowed purposes, rejection of event SMS and blank/invalid content, safe errors |
    | TestSandboxSMSRedactsSensitiveData | No full number/body/OTP, safe metadata, deterministic output, cancellation and invalid-purpose silence |
    | TestSandboxSMSConcurrentDispatch | 20 concurrent calls, no lost log events and no race findings |
    | TestAfricaTalkingRequestAndResponses | Exact method/headers/form, 100/101/102 accepted states, recipient failure/mismatch, malformed/oversized responses, retry classification |
    | TestSMSFactorySandboxDoesNotUseHTTP | Sandbox factory uses local logger; no provider transport; missing production config rejected |
    | TestAfricaTalkingRejectsRedirects | 307 redirect never reaches target and is not classified as retryable |
    | TestAfricaTalkingCancellationAndTimeout | Canceled/invalid requests never reach provider; transport timeout returns retryable safe error |
    | TestSMSFactoryDoesNotMutateHTTPClient | Original client remains unchanged; copied client has timeout and redirect policy |
    
    | Command | Result |
    |---|---|
    | Initial `cd backend && go test -v -race ./...` | Passed; every original package had no tests |
    | `gofmt -w` on each affected owned Go file | Applied before each task's verification |
    | `go test -v -race ./internal/notifications` after each task | Passed |
    | `go vet ./internal/notifications` after each task | Passed |
    | Final `cd backend && go test -v -race ./...` | Passed; notifications tests exercised, other packages remain scaffolds |
    | `cd backend && go build ./...` | Passed |
    | `cd backend && go vet ./...` | Passed |
    | `gofmt -l` on the four created Go files | Empty output; formatting clean |
    | `make test` | Go tests passed. Python invocation failed: `No module named pytest`. Makefile masks the failure with `|| true`; this is not an all-services pass |
    | `cd frontend && npm run lint` | Failed, exit 127: `next: not found` |
    | `cd frontend && npm run build` | Failed, exit 127: `next: not found` |
    | `git diff --check` | Passed |
    
    All SMS provider tests use local httptest servers. No live Africa's Talking, WhatsApp, Daraja, or AI-service end-to-end test has passed or been claimed. No race findings occurred in the implemented SMS code; financial/OTP-state concurrency remains unimplemented and untested.
    
    ## Database, APIs and acceptance status
    
    No database changes or HTTP API endpoints have been implemented at this checkpoint. The ten assigned tables, UP/DOWN migration, RLS tests, and UP/DOWN/UP verification remain pending. Full migration integration also lacks the other developers' 000002 and 000003 files.
    
    | Acceptance criterion | Status |
    |---|---|
    | AC-BE-C01 — WhatsApp identity/session security | Not implemented; SMS transport alone is not sufficient evidence |
    | AC-BE-C02 — cooldown and sender blocking | Not implemented |
    | AC-BE-C03 — charging and fee changes | Not implemented |
    | AC-BE-C04 — payments, credit, matching, phone lock | Not implemented |
    | AC-BE-C05 — internal AI boundary and audit | Not implemented |
    
    ## Blockers and integration work
    
    1. The required `WithTenantTx(ctx, schoolID, func(pgx.Tx) error)` contract cannot compile without pgx. Asynq is also absent. `backend/go.mod` currently declares only the module path and Go version. The assignment explicitly forbids editing it or go.sum. The minimal proposed direct requirements are pgx/v5 v5.7.2 and Asynq v0.25.1, plus generated indirect requirements/checksums. No dependency changes or alternate nested module have been made.
    2. Evans's real tenant transaction helper, authentication context, audit logger, cleanup integration, and migration 000002 are absent. Consumer interfaces and test doubles are planned, not replacements for those systems.
    3. Ema's academics summaries and migration 000003 are absent. Sos's AI client/parent assistant are placeholders. Server route mounting is reserved to Evans. These prevent eventual full integration, independently of unit-test results.
    4. The provider has not been exercised with live credentials. A transport timeout can mean delivery was accepted before the response was lost; the API does not provide a client idempotency guarantee here. The queue must handle deduplication where possible and must not claim exactly-once SMS delivery. Queue implementation is pending.
    5. HTTP source registration, in-app notifications, all billing/M-Pesa/WhatsApp state-machine/internal API functionality, and background workers remain unimplemented. Nothing in the successful scaffold build changes that status.
    
    The next implementation step is INF-02 after the protected dependency change is supplied or explicitly authorized, followed by transactional in-app notification creation and migration verification. Preserve the task-by-task gates in IMPLEMENTATION_PLAN.md.

## Activity

    $ git status --short && git diff --name-only && git diff --cached --name-only && git diff --check && git add backend/internal/billing/IMPLEMENTATION_PLAN.md backend/internal/billing/VERIFICATION.md backend/internal/notifications/sms_provider.go backend/internal/notifications/sms_provider_test.go && git diff --cached --name-only && git diff --cached --check && git commit -m 'feat(notifications): implement Africa Talking SMS transport'
     M Docs/DECISIONS.md
     M backend/internal/billing/IMPLEMENTATION_PLAN.md
    ?? backend/internal/billing/VERIFICATION.md
    ?? backend/internal/notifications/sms_provider.go
    ?? backend/internal/notifications/sms_provider_test.go
    Docs/DECISIONS.md
    backend/internal/billing/IMPLEMENTATION_PLAN.md
    backend/internal/billing/IMPLEMENTATION_PLAN.md
    backend/internal/billing/VERIFICATION.md
    backend/internal/notifications/sms_provider.go
    backend/internal/notifications/sms_provider_test.go
    [feat/ian-backend-engineer-c 0ef7d5e] feat(notifications): implement Africa Talking SMS transport
     4 files changed, 378 insertions(+), 1 deletion(-)
     create mode 100644 backend/internal/billing/VERIFICATION.md
     create mode 100644 backend/internal/notifications/sms_provider.go
     create mode 100644 backend/internal/notifications/sms_provider_test.go
    ✓ • 111ms

## Activity

    $ git status --short && git diff --name-only && git diff --cached --name-only && git diff b2b2e3b6f15e60f25947f04c39b8f6b12b5a268f HEAD --name-only && git rev-parse HEAD && git log -3 --oneline && git diff -- Docs/DECISIONS.md
     M Docs/DECISIONS.md
    Docs/DECISIONS.md
    backend/internal/billing/IMPLEMENTATION_PLAN.md
    backend/internal/billing/VERIFICATION.md
    backend/internal/notifications/sms.go
    backend/internal/notifications/sms_provider.go
    backend/internal/notifications/sms_provider_test.go
    backend/internal/notifications/sms_test.go
    0ef7d5e8ba65829c146dde8c1844e3c11efd351f
    0ef7d5e feat(notifications): implement Africa Talking SMS transport
    515b2f0 feat(notifications): add redacted sandbox SMS dispatcher
    5925aa5 feat(notifications): define transactional SMS contract
    diff --git a/Docs/DECISIONS.md b/Docs/DECISIONS.md
    index a9a94b6..ac82a09 100644
    --- a/Docs/DECISIONS.md
    +++ b/Docs/DECISIONS.md
    @@ -1,3 +1,4 @@
    +
     # DECISIONS.md
    
     A short, dated log of decisions worth remembering — so nobody (human or AI
    ✓ • 0ms

## Assistant

Phase Zero, the implementation plan, and three SMS tasks are complete. **The full assignment remains incomplete.**

- Branch: `feat/ian-backend-engineer-c`
- Starting commit: `b2b2e3b`
- Current commit: `0ef7d5e`
- Implemented: SMS interface, phone normalization, redacted sandbox dispatcher, and Africa’s Talking transport.
- Verified: nine test functions, backend race tests, build, vet, and formatting.
- External checks: Python lacks `pytest`; frontend lacks `next`.
- Ownership: all six committed files are within assigned paths. Your existing `Docs/DECISIONS.md` edit remains untouched. Nothing was pushed.

The [implementation plan](/home/fian/frihk/Ikonex-Academy/backend/internal/billing/IMPLEMENTATION_PLAN.md) and [verification report](/home/fian/frihk/Ikonex-Academy/backend/internal/billing/VERIFICATION.md) record requirements, evidence, and outstanding work. No migrations, billing endpoints, WhatsApp sessions, or internal AI endpoints are claimed complete.

**Pending decision:** may I add pgx `v5.7.2` and Asynq `v0.25.1`, including indirect dependencies and checksums, to `backend/go.mod` and `backend/go.sum`?

Your §4.1 ownership restriction explicitly prohibits those edits. The documented `pgx.Tx` contract and real Asynq worker require these currently absent dependencies.

## User

yes you can install for testisting. but you will have to update the whole group about the versions and resources used for testing, so make an md file to save these. then after testing and noting them down remove them from the prohibited documents, add this to you your memory

## Assistant

I’ll use the dependency exception only for testing, record the exact versions, tools, commands, and test resources in an owned Markdown file, then restore `backend/go.mod` and `backend/go.sum` to their original state. I’ll also record this rule in the implementation plan so it persists for future work.

## Activity

    $ cat backend/internal/billing/IMPLEMENTATION_PLAN.md backend/internal/billing/VERIFICATION.md
    # Developer 05 implementation plan
    
    ## Repository assessment — 2026-11-09
    
    - Branch: `feat/ian-backend-engineer-c`, created from `main`.
    - Starting commit: `b2b2e3b6f15e60f25947f04c39b8f6b12b5a268f`; remote main agrees. No local or remote `release-1` or existing Ian branch.
    - Pre-existing working-tree change: one leading blank line in `Docs/DECISIONS.md`. Preserve it; exclude it from every commit. Index initially empty.
    - Read both assignment documents in full, baseline UP/DOWN migrations, repository instructions, API/architecture/security/requirements documents, and relevant collaborator contracts.
    - Backend: Go 1.22.2, standard-library placeholder packages, empty server and worker entry points. No third-party dependencies or go.sum. No existing tests. Frontend and Python services are scaffolds.
    - Schema: only migration 000001 exists. Tenant policies use `current_school_id()` and transaction-local `app.school_id`. Users and schools are global. Fee structures are unique per school/year/grade/term; ledgers are per school/student/year/term. Payments have globally unique transaction references.
    - Schema gaps: fee items have neither school_id nor RLS; baseline foreign keys do not enforce matching school IDs; audit_log is not RLS protected. All owned queries must explicitly scope and validate related records. Never bypass RLS to route webhooks: inject trusted identifier-to-school routing, then verify configuration within that tenant.
    - Missing migration 000002 supplies school_profile, otp_requests, sender_rate_blocks, and users.chat_payment_paused_until. Missing 000003 supplies academics data. Do not create replacements for these migrations.
    - Documentation differences: old API_CONTRACTS.md leaves envelopes unspecified; the master plan defines them. Chat state is called both state/uppercase and status/lowercase: use the specified `state` column and uppercase state machine. StudentChargeHook/StudentBillingHook name the same admission capability. Fee variants are unresolved product policy; preserve the plan's default `all`, without inventing boarder classification. Term ledgers require explicit credit transfer, not repeated summing of rolled-forward balances.
    - Tooling: Go 1.22.2 and Docker 27.3.1 available; PostgreSQL/Redis executables absent, images not cached. Live provider credentials and deployed AI service have not been verified. No real SMS/payment requests are authorized by test execution.
    
    ## Dependency map
    
    | Dependency | Owner | Available | Required contract / approach |
    |---|---|---|---|
    | Tenant transactions | Evans | No | `WithTenantTx(ctx, schoolID, func(pgx.Tx) error) error`; injected consumer contract, test double; no edits to database package. |
    | Audit logger | Evans | No | Transaction-bound audit writer; injected contract; failures roll back financial changes. |
    | Authentication | Evans | No | Trusted user, membership, school, role and channel accessor; deny absent identity; no caller-controlled identity headers. |
    | SMSDispatcher | Ian | No | Context-aware transactional SMS message with purpose; local sandbox and Africa's Talking implementations. |
    | InAppNotifier | Ian | No | `CreateInAppNotification` participating in the caller's tenant transaction. |
    | Admission billing | Ian / Evans consumer | No | StudentChargeHook (StudentBillingHook alias if needed), caller transaction, tenant/student/year and actor. |
    | Term-date sync | Ian / Ema consumer | No | `OnTermStartDateChanged(ctx, tx, schoolID, termID, newStartDate, actorUserID, note) error`. |
    | Academic summaries | Ema | No | GetStudentAttendanceSummary / GetStudentMarksSummary consumer interface; mock only, no substitute academics system. |
    | Cleanup | Evans | No | `PurgeExpiredUnverifiedSubmissions(ctx)`; configuration retained by adapter; default seven days, scheduled daily. |
    | AI | Sos | Placeholder | HTTP parent-query; token plus raw base64url JSON + dot + hex HMAC-SHA256 of encoded payload, keyed by service token. |
    | Frontend | Ashy | Placeholder | Standard success/error envelopes, fixed two-decimal money strings, paginated lists, explicit offsets. |
    | pgx / Asynq | Evans module files | Absent | Native implementations cannot compile in the root module until dependencies are added by owner or explicitly authorized. |
    | Redis / PostgreSQL | Infrastructure | Docker available | Disposable synthetic services for migration, concurrency and worker verification. |
    
    ## Protected-file coordination proposal
    
    The minimal dependency change needed for the prescribed pgx transaction contract and real Asynq worker is the following direct requirements in `backend/go.mod`, plus their generated indirect requirements and `backend/go.sum`:
    
    ```go
    require (
        github.com/hibiken/asynq v0.25.1
        github.com/jackc/pgx/v5 v5.7.2
    )
    ```
    
    Both versions support this Go toolchain. This is a proposal, not authorization to edit protected files. Do not add nested modules, vendor libraries, replace Asynq, or hide unresolved imports behind build tags. Standard-library tasks can proceed independently. Migration 000002's global sender-block semantics and OTP school/session binding also need coordination; use boundary tests until the real schema exists.
    
    ## Execution and verification gates
    
    Paths below are relative to the repository root. `B`, `N`, `W`, `I`, `J` mean `backend/internal/billing`, `notifications`, `whatsapp`, `internalapi`, `jobs` respectively. `M` means `migrations/000004_r1_billing_whatsapp_notifications`. Tests live beside their implementation. Each row is one task: inspect, implement its ordered steps, gofmt only affected files, run verification, review the diff and ownership, then update status. Never advance a dependent task past a failed gate.
    
    `G(package)` means `cd backend && go test -v -race ./internal/<package>` and `go vet ./internal/<package>`; both must pass. `DB` means disposable PostgreSQL 16: baseline then owned UP, inspect schema/constraints/indexes, use a non-superuser role to test RLS, owned DOWN, verify baseline unchanged, UP again. Missing external integration is reported separately from a successful unit test. Commit tested coherent tasks only; no push or merge.
    
    | ID | Feature / objective | Dependencies | Exact files | Ordered implementation / expected behavior | Tests and verification | Acceptance / status |
    |---|---|---|---|---|---|---|
    | INF-01 | C1–C3 baseline and architecture | None | B/IMPLEMENTATION_PLAN.md | Read documents; inventory actual code/schema; preserve edits; inspect remote; create branch; map gaps | Git status/branches/HEAD, baseline race suite | Completed; baseline race suite passed (no pre-existing tests) |
    | NOTIF-01 | C3 SMS boundary and normalization | INF-01 | N/sms.go, N/sms_test.go | Define SMSDispatcher and allowed transactional purposes; normalize Kenyan/E.164 phones; validate without leaking content | Table tests for accepted/rejected numbers, purposes and missing content; G(notifications) | Completed; race tests and vet passed |
    | NOTIF-02 | C3 deterministic sandbox SMS | NOTIF-01 | N/sms.go, N/sms_test.go | Log masked recipient and safe metadata, never message/OTP; propagate cancellation; username selection follows in NOTIF-03 factory | Capture slog output, assert no OTP/full phone/API key; concurrent dispatch; G(notifications) | Completed; race tests and vet passed |
    | NOTIF-03 | C3 Africa's Talking transport | NOTIF-01 | N/sms_provider.go, N/sms_provider_test.go | Select local sandbox by username; context-bound HTTP client; encoded form and apiKey; validate per-recipient response; classify retryable/permanent failures | httptest request contract, provider rejection, timeout, malformed response, cancellation; G(notifications) | Completed; package and full-backend race tests, build and vet passed |
    | INF-02 | C1–C3 transaction/auth/error contracts | INF-01, pgx | I/contracts/contracts.go, I/contracts/contracts_test.go | Small consumer contracts for pgx tenant transactions/audit/identity; APIError and standard envelopes; deny missing identity | Contract doubles, malformed identity, error envelope and safe error mapping; G(internalapi/...) | Compatible injection, no substitute auth; Blocked on module dependency |
    | NOTIF-04 | C3 transaction-bound notification creation | INF-02 | N/notifications.go, N/notifications_test.go | Validate content/recipient; scoped insertion within supplied transaction; define InAppNotifier | Rollback, tenant/user checks; PostgreSQL insertion and RLS; G(notifications) | Usable by Ema and billing without separate commit; Pending |
    | NOTIF-05 | C3 notification HTTP routes | NOTIF-04, DB-03 | N/http.go, N/notifications_test.go | Register GET list / PATCH read; bounded pagination; recipient+school scope; preserve read_at on repeat | HTTP authorization, paging, ownership, read idempotency; G(notifications) | Both endpoints and envelopes verified; Pending |
    | DB-01 | C1–C3 migration design and UP | INF-01 | M.up.sql, B/migration_test.go | Add ten assigned tables; ledger credit/date; necessary financial constraints, indexes and tenant policies; enforce reference ownership | Review against baseline; DB schema and cross-tenant FK tests | No duplicated baseline/core tables; Pending |
    | DB-02 | C1–C3 reversible DOWN | DB-01 | M.down.sql, B/migration_test.go | Reverse owned changes in dependency order; no CASCADE into baseline; preserve unrelated records | DB UP/DOWN/UP and baseline schema/data comparison | Rollback verified; Pending |
    | DB-03 | C1–C3 RLS verification | DB-02 | B/migration_test.go, B/testdata/schema_checks.sql | Two synthetic schools; non-superuser sessions; no-tenant, wrong-tenant read/write and transaction-local context | DB policies, indexes, constraints, rollback | Actual PostgreSQL isolation evidence; Pending |
    | BILL-01 | C1 exact money and ledger arithmetic | INF-01 | B/money.go, B/money_test.go | Integer minor units; strict decimal parser/JSON strings/NUMERIC limits; compute nonnegative balance and credit | Precision, overflow, negative/zero, overpayment, serialization, conservation; G(billing) | No float arithmetic; Pending |
    | BILL-02 | C1 fee-structure validation | BILL-01, INF-02 | B/structures.go, B/structures_test.go | Validate team role, references, term/year agreement and items; check sums and duplicates | Invalid/cross-school/year/grade/items and duplicate request tests; G(billing) | Admin-only valid consistent structure; Pending |
    | BILL-03 | C1 fee periods and structure loading | BILL-02, DB-03 | B/billing.go, B/structures.go, B/structures_test.go | Transactionally persist itemized structure and term period with due_date=start_date | PostgreSQL reference/rollback/duplicate tests; G(billing) | No partial structure; Pending |
    | BILL-04 | C1 enrolled-student term charging | BILL-03 | B/charges.go, B/charges_test.go | Select applicable enrolled students; serialize student financial operations; insert unique charge and ledger entry | Concurrent repeat charge, inactive/other-grade exclusions, dates; G(billing) | One charge per student/period; Pending |
    | BILL-05 | C1 admission hook | BILL-04 | B/billing.go, B/charges_test.go | Export caller-transaction StudentChargeHook; load applicable structures; reuse charging logic | New admission, repeat hook, absent structures, rollback; G(billing) | Onboarding can inject without importing its code; Pending |
    | BILL-06 | C1 credit carry-forward | BILL-04 | B/ledgers.go, B/ledgers_test.go | Serialize student ledger changes; consume existing credit once; preserve term history; record credit_applied | Multi-term/year, partial/exact/excess credit and concurrency; G(billing) | Financial conservation and no double credit; Pending |
    | BILL-07 | C1 balance and ledger endpoints | BILL-06 | B/http.go, B/ledgers.go, B/ledgers_test.go | Guardian/team authorization; retrieve balances/items/history/notes; fixed money and date formats | Own/unrelated guardian, cross-tenant, item/entry accuracy; G(billing) | Two GET routes conform to envelope; Pending |
    | BILL-08 | C1 explained fee adjustments | BILL-07, NOTIF-04 | B/fee_changes.go, B/fee_changes_test.go | Require note; select affected charges; apply deltas, preserve entries; create note and audit in transaction | 422 code, positive/negative changes, rollback and history; G(billing) | Atomic adjustment with audit; Pending |
    | BILL-09 | C1 targeted fee notices | BILL-08 | B/fee_changes.go, B/fee_changes_test.go | Deduplicate affected guardians; create notifications inside same transaction | Shared guardian, unaffected guardian, notifier failure rollback; G(billing) | Only affected guardians receive one notice; Pending |
    | BILL-10 | C1 term-date synchronization | BILL-09 | B/fee_changes.go, B/fee_changes_test.go | Export exact date hook; shift periods/charges/ledger due dates; note/audit/notify | Same/date-shift, rollback, targeted notices; G(billing) | AC-22 hook verified; Pending |
    | MPESA-01 | C2 Daraja client and encrypted config | BILL-01 | B/daraja.go, B/daraja_test.go | School credential resolver; OAuth; safe HTTP timeout/errors; exact request encoding | httptest OAuth/STK/query payloads, secrets, provider failures; G(billing) | No hardcoded credentials or live calls; Pending |
    | MPESA-02 | C2 STK authorization and validation | MPESA-01, BILL-07 | B/mpesa_stk.go, B/mpesa_stk_test.go | Linked guardian/registered phone; positive amount; chat-only freeze; internal parent confirmation | Under/equal/over balance, 0.01, unauthorized, freeze boundary/web, missing confirmation; G(billing) | AC-23/26 enforced before provider call; Pending |
    | MPESA-03 | C2 durable STK initiation | MPESA-02 | B/mpesa_stk.go, B/mpesa_stk_test.go, M.up.sql, M.down.sql | Persist correlation before provider side effect; admission AccountReference; schedule 45s polling; recover queue failures | Provider failure, delayed callback, correlation and queue tests; G(billing), DB if schema changes | Pending STK survives process restart; Pending |
    | MPESA-04 | C2 idempotent callback credit | MPESA-03, BILL-06, NOTIF-04 | B/mpesa_stk.go, B/payments.go, B/payments_test.go | Authenticate callback route context; validate response against pending request; persist receipt/payment/entry/credit/notice atomically | Failed/malformed/duplicate/concurrent callback, wrong amount/receipt/school; G(billing) | Exactly one financial credit; Pending |
    | MPESA-05 | C2 direct Paybill and unmatched storage | MPESA-04 | B/mpesa_c2b.go, B/mpesa_c2b_test.go | Resolve receiving shortcode through trusted boundary; school-scoped normalized admission; persist match or full unmatched facts | Two schools/same admission, invalid/blank/ambiguous reference, duplicate receipt; G(billing) | No lost incoming payment; Pending |
    | MPESA-06 | C2 unmatched list and manual resolution | MPESA-05 | B/unmatched.go, B/unmatched_test.go | Team-only list; lock unresolved payment; validate target; credit once; resolve with actor/audit | Forbidden/cross-school/duplicate/concurrent resolution and overpayment; G(billing) | No fuzzy matching; Pending |
    | WA-01 | C3 guardian phone masking | NOTIF-01 | W/masking.go, W/masking_test.go | Normalize and mask; reject invalid values without echoing private input | Exact +254712345678 -> 07•• ••• 678; malformed cases; G(whatsapp) | No full guardian number returned; Pending |
    | WA-02 | C3 session persistence and bot routing | INF-02, DB-03 | W/whatsapp.go, W/repository.go, W/whatsapp_test.go | Trusted bot mapping then tenant verification; create/reuse sender-bound sessions; scoped admission/guardian lookup | Unknown admission zero SMS; two schools; masked multi/single selection; G(whatsapp) | No cross-school lookup or identity assumption; Pending |
    | WA-03 | C3 secure OTP and sender limits | WA-02, NOTIF-03, migration 000002 contract | W/otp.go, W/otp_test.go | Cryptographic 6-digit OTP/hash/expiry; bind guardian/session; serialize global sender and school/admission limits; persist sixth-request 24h block | 59/60s, 5/6 requests, 30min/24h boundaries, concurrent senders, selected recipient only; G(whatsapp) | Guardian account never blocked; Pending |
    | WA-04 | C3 verification and six-hour sessions | WA-03 | W/whatsapp.go, W/whatsapp_test.go | Atomic consume; bounded wrong-code attempts; verified_at/expires_at; deny at exact expiry; restart identity flow | Replay, wrong/expired code, exact six-hour boundary, ownership; G(whatsapp) | Fresh OTP after expiry; Pending |
    | WA-05 | C3 webhook and REST adapters | WA-04 | W/http.go, W/cloud.go, W/http_test.go | Verify Meta challenge and request signature; route events to same state machine; bind REST sender; bounded payloads and delivery retries | Invalid signature/session, duplicate events, REST 429, Cloud API HTTP tests; G(whatsapp) | One state machine across transports; Pending |
    | API-01 | C3 signed user context | INF-01 | I/context_signer.go, I/context_signer_test.go | Implement prescribed base64url/hex HMAC; constant-time comparisons; strict fields/channels/expiry; injectable clock | Fixed cross-language vector, tampering, malformed/expired/missing context/token; G(internalapi) | Verify signature before decoding/trusting payload; Pending |
    | API-02 | C3 internal authentication and authorization | API-01, INF-02, WA-04 | I/internalapi.go, I/internalapi_test.go | Authenticate both headers; recheck active membership/session; guardian-child tenant guard; inject academics interface | Cross-parent/school, revoked/expired session, bad membership; G(internalapi) | No AI trust for authorization; Pending |
    | API-03 | C3 five internal read endpoints | API-02, BILL-07 | I/reads.go, I/reads_test.go | School profile, own children, fees, attendance, marks; minimal WhatsApp facts and app link for marks | Route/data/envelope tests; academics mocks labeled; WhatsApp results denial; G(internalapi) | Backend enforces own-child and channel restrictions; Pending |
    | API-04 | C3 internal confirmed STK and audit | API-03, MPESA-03 | I/actions.go, I/actions_test.go | Force chat source/confirmation; call billing; persist safe action/read metadata; configurable lookup/start limiter | Missing confirmation/freeze, audit failures, secret metadata, rate limits; G(internalapi) | AC-28 and payment boundaries; Pending |
    | API-05 | C3 chat proxy and WhatsApp forwarding | API-04, WA-05 | I/client.go, I/client_test.go, W/whatsapp.go | Sign outbound identity; call AI parent-query with both headers; forward validated response; safe failure/fallback | httptest contract, context TTL<=session expiry, canceled/failed AI, channel limits; G(internalapi), G(whatsapp) | No AI end-to-end claim from mocks; Pending |
    | JOB-01 | C3 Asynq client and task payloads | Asynq dependency, NOTIF-03, MPESA-03 | J/jobs.go, J/jobs_test.go | Typed sms:transactional, mpesa:stk_poll, onboarding:purge_unverified tasks; validate/serialize; stable task IDs | Serialization, invalid payload, deduplication and 45s delay; G(jobs) | Actual Asynq, no substitute queue; Blocked on module dependency |
    | JOB-02 | C3 transactional SMS worker | JOB-01 | J/sms.go, J/sms_test.go | Context-bound dispatch; retry temporary errors; skip permanent/expired OTP; delivery deduplication | Redis-backed retry, permanent failure, cancellation, duplicate task; G(jobs) | No event SMS or unbounded OTP retries; Pending |
    | JOB-03 | C2 STK polling worker | JOB-01, MPESA-04 | J/mpesa.go, J/mpesa_test.go, B/mpesa_stk.go | Check pending; query provider; terminal failure/timeout; success requires verified receipt facts; same idempotent settlement | Callback/poll race, retry, pending, success without receipt; G(jobs), G(billing) | Query response never fabricates a receipt; Pending |
    | JOB-04 | C3 cleanup schedule and lifecycle | JOB-02, JOB-03 | J/worker.go, J/worker_test.go, backend/cmd/worker/main.go | Redis configuration; inject Evans cleanup; retention default 7; daily scheduler; signals/graceful shutdown | Test cleanup double, retention/env, handler registration/start/shutdown; G(jobs), go build ./cmd/worker | Missing cleanup explicitly reported, never silently skipped; Pending |
    | QA-01 | C1–C3 final verification and report | All above | B/VERIFICATION.md | Format owned Go; test/build/vet/race; DB cycle; inspect security/concurrency/contracts; map ACs; ownership audit; conventional commits | go test -v -race ./...; go build ./...; go vet ./...; make test; frontend lint/build; git status/diff/cached paths | Report evidence and blockers separately; Pending |
    
    ## Acceptance evidence tracker
    
    | Criterion | Required coverage | Evidence | Status |
    |---|---|---|---|
    | AC-BE-C01 | School routing, masked selected guardian, OTP, exact 6h expiry | WA-01–05, PostgreSQL/HTTP tests | Not implemented |
    | AC-BE-C02 | Sender/admission cooldown, sixth-request sender-only 24h block | WA-03 concurrent and boundary tests | Not implemented |
    | AC-BE-C03 | Idempotent term charge, credit, explained targeted fee/date changes | BILL-02–10, PostgreSQL tests | Not implemented |
    | AC-BE-C04 | STK rules, receipt idempotency, overpayment, school-scoped C2B/manual matching | MPESA-01–06, JOB-03 | Not implemented |
    | AC-BE-C05 | Token/HMAC/expiry, own-child permissions, AI audit | API-01–05 | Not implemented |
    
    ## Outstanding integration, not substitute implementations
    
    Evans must mount exported routes, supply authenticated identities/tenant transactions/audit/cleanup, and land core migration/dependencies. Ema supplies academic reads and invokes term-date synchronization. Sos supplies AI service and compatible signed-context verification; Ashy consumes documented owned endpoint shapes. Provider credentials, a reachable authenticated callback ingress, and provider sandbox end-to-end tests remain separate verification gates. Full integration of migrations 000001–000004 cannot be claimed before 000002/000003 exist.
    # Developer 05 checkpoint report — 2026-11-09
    
    This is an incomplete implementation checkpoint, not Release 1 sign-off. Phase Zero, the detailed plan, and NOTIF-01 through NOTIF-03 are complete. The next task, INF-02, requires the missing pgx dependency in module files explicitly protected by the assignment. An exception for the exact dependency proposal in IMPLEMENTATION_PLAN.md has been requested and has not been assumed.
    
    ## Git baseline and ownership
    
    - Branch: `feat/ian-backend-engineer-c`.
    - Starting commit: `b2b2e3b6f15e60f25947f04c39b8f6b12b5a268f` on `main`, also the verified remote main tip.
    - No `release-1` or existing Ian branch was available. Created the requested branch without resetting/stashing/discarding anything.
    - Existing unrelated change: `Docs/DECISIONS.md` has a leading blank line. It remains unchanged and unstaged.
    - Tested commits preceding this report: `5925aa5` (SMS contract and plan), `515b2f0` (sandbox dispatcher). The commit containing this report adds the verified provider transport. Its exact hash is reported in the user-facing checkpoint response.
    - No push or merge performed.
    - `git status --short`, `git diff --name-only`, `git diff --cached --name-only`, and `git diff --check` checked before commits. Every staged implementation path belongs to Developer 05.
    
    ## Files created
    
    | File | Purpose |
    |---|---|
    | `backend/internal/billing/IMPLEMENTATION_PLAN.md` | Repository assessment, dependency map, detailed tasks, gates and acceptance tracker |
    | `backend/internal/billing/VERIFICATION.md` | This checkpoint's evidence and limitations |
    | `backend/internal/notifications/sms.go` | SMS consumer contract, transactional purposes, phone normalization and safe sandbox dispatcher |
    | `backend/internal/notifications/sms_test.go` | Normalization, purpose validation, sandbox redaction/determinism/cancellation/concurrency |
    | `backend/internal/notifications/sms_provider.go` | Factory and Africa's Talking HTTP transport with safe errors and retry classification |
    | `backend/internal/notifications/sms_provider_test.go` | Local HTTP provider contract and failure tests |
    
    No pre-existing tracked source files were modified. The original `notifications.go` remains a placeholder for the later in-app service. No module files, migrations, shared code, frontend, AI service, or external documentation were changed.
    
    ## Implemented and tested
    
    - `SMSDispatcher.DispatchSMS(context.Context, SMSMessage) error` is the consumer-facing dependency for auth/onboarding/WhatsApp and, later, the queue worker.
    - Supported purposes: app-login OTP, WhatsApp OTP, phone-change OTP, old-number phone-change notice. Other purposes, including absence and sickbay events, are rejected before delivery.
    - `NormalizePhone` accepts Kenyan local and country-prefixed numbers and international E.164 forms; normalizes common display separators; rejects malformed input without reflecting private input in errors.
    - `NewSMSDispatcher(SMSConfig, *http.Client, *slog.Logger)` selects local-only behavior when `Username == "sandbox"`; no key or network is required in that mode.
    - Sandbox logging contains only masked recipient suffix, purpose, and character count. It does not log body, OTP, full number, or credentials. Tests remove the logger timestamp to compare deterministic metadata.
    - Provider transport uses a context-bound form POST, API key header, normalized recipient, optional sender ID, and JSON response validation. It verifies a single matching recipient and recognized processed/sent/queued status before success.
    - A copied HTTP client receives a finite default timeout and refuses redirects, preventing credentials and SMS bodies from being forwarded to a redirect target.
    - `SMSDeliveryError` classifies transport failures, HTTP 408/429/5xx, and selected provider failures for retry without exposing provider text or private data. Invalid/ambiguous response bodies do not trigger automatic retry classification.
    
    ## Test evidence
    
    Nine top-level tests, including 18 phone cases and 14 provider-response cases:
    
    | Test | Evidence |
    |---|---|
    | TestNormalizePhone | Local/E.164 normalization, incorrect lengths/prefixes, Unicode digits, embedded private/injected text |
    | TestTransactionalSMSValidation | All four allowed purposes, rejection of event SMS and blank/invalid content, safe errors |
    | TestSandboxSMSRedactsSensitiveData | No full number/body/OTP, safe metadata, deterministic output, cancellation and invalid-purpose silence |
    | TestSandboxSMSConcurrentDispatch | 20 concurrent calls, no lost log events and no race findings |
    | TestAfricaTalkingRequestAndResponses | Exact method/headers/form, 100/101/102 accepted states, recipient failure/mismatch, malformed/oversized responses, retry classification |
    | TestSMSFactorySandboxDoesNotUseHTTP | Sandbox factory uses local logger; no provider transport; missing production config rejected |
    | TestAfricaTalkingRejectsRedirects | 307 redirect never reaches target and is not classified as retryable |
    | TestAfricaTalkingCancellationAndTimeout | Canceled/invalid requests never reach provider; transport timeout returns retryable safe error |
    | TestSMSFactoryDoesNotMutateHTTPClient | Original client remains unchanged; copied client has timeout and redirect policy |
    
    | Command | Result |
    |---|---|
    | Initial `cd backend && go test -v -race ./...` | Passed; every original package had no tests |
    | `gofmt -w` on each affected owned Go file | Applied before each task's verification |
    | `go test -v -race ./internal/notifications` after each task | Passed |
    | `go vet ./internal/notifications` after each task | Passed |
    | Final `cd backend && go test -v -race ./...` | Passed; notifications tests exercised, other packages remain scaffolds |
    | `cd backend && go build ./...` | Passed |
    | `cd backend && go vet ./...` | Passed |
    | `gofmt -l` on the four created Go files | Empty output; formatting clean |
    | `make test` | Go tests passed. Python invocation failed: `No module named pytest`. Makefile masks the failure with `|| true`; this is not an all-services pass |
    | `cd frontend && npm run lint` | Failed, exit 127: `next: not found` |
    | `cd frontend && npm run build` | Failed, exit 127: `next: not found` |
    | `git diff --check` | Passed |
    
    All SMS provider tests use local httptest servers. No live Africa's Talking, WhatsApp, Daraja, or AI-service end-to-end test has passed or been claimed. No race findings occurred in the implemented SMS code; financial/OTP-state concurrency remains unimplemented and untested.
    
    ## Database, APIs and acceptance status
    
    No database changes or HTTP API endpoints have been implemented at this checkpoint. The ten assigned tables, UP/DOWN migration, RLS tests, and UP/DOWN/UP verification remain pending. Full migration integration also lacks the other developers' 000002 and 000003 files.
    
    | Acceptance criterion | Status |
    |---|---|
    | AC-BE-C01 — WhatsApp identity/session security | Not implemented; SMS transport alone is not sufficient evidence |
    | AC-BE-C02 — cooldown and sender blocking | Not implemented |
    | AC-BE-C03 — charging and fee changes | Not implemented |
    | AC-BE-C04 — payments, credit, matching, phone lock | Not implemented |
    | AC-BE-C05 — internal AI boundary and audit | Not implemented |
    
    ## Blockers and integration work
    
    1. The required `WithTenantTx(ctx, schoolID, func(pgx.Tx) error)` contract cannot compile without pgx. Asynq is also absent. `backend/go.mod` currently declares only the module path and Go version. The assignment explicitly forbids editing it or go.sum. The minimal proposed direct requirements are pgx/v5 v5.7.2 and Asynq v0.25.1, plus generated indirect requirements/checksums. No dependency changes or alternate nested module have been made.
    2. Evans's real tenant transaction helper, authentication context, audit logger, cleanup integration, and migration 000002 are absent. Consumer interfaces and test doubles are planned, not replacements for those systems.
    3. Ema's academics summaries and migration 000003 are absent. Sos's AI client/parent assistant are placeholders. Server route mounting is reserved to Evans. These prevent eventual full integration, independently of unit-test results.
    4. The provider has not been exercised with live credentials. A transport timeout can mean delivery was accepted before the response was lost; the API does not provide a client idempotency guarantee here. The queue must handle deduplication where possible and must not claim exactly-once SMS delivery. Queue implementation is pending.
    5. HTTP source registration, in-app notifications, all billing/M-Pesa/WhatsApp state-machine/internal API functionality, and background workers remain unimplemented. Nothing in the successful scaffold build changes that status.
    
    The next implementation step is INF-02 after the protected dependency change is supplied or explicitly authorized, followed by transactional in-app notification creation and migration verification. Preserve the task-by-task gates in IMPLEMENTATION_PLAN.md.
    ✓ • 50.97s

## Activity

    $ git status --short && git branch --show-current && git rev-parse HEAD && git diff --cached --name-only && cat backend/go.mod && rg --files -g AGENTS.md -g '*MEMORY*' -g '*memory*' -g 'go.sum' -g 'go.work*'
     M Docs/DECISIONS.md
    feat/ian-backend-engineer-c
    0ef7d5e8ba65829c146dde8c1844e3c11efd351f
    module github.com/Ashomondi/ikonex-academy/backend
    
    go 1.22.2
    AGENTS.md
    Docs/AGENTS.md
    ✓ • 7ms

## Activity

    file changes: Completed · 2 changes
    Update { move_path: None }: /home/fian/frihk/Ikonex-Academy/backend/internal/billing/IMPLEMENTATION_PLAN.md
    @@ -44,3 +44,9 @@
     
    -Both versions support this Go toolchain. This is a proposal, not authorization to edit protected files. Do not add nested modules, vendor libraries, replace Asynq, or hide unresolved imports behind build tags. Standard-library tasks can proceed independently. Migration 000002's global sender-block semantics and OTP school/session binding also need coordination; use boundary tests until the real schema exists.
    +The user authorized these dependencies for testing only on 2026-11-09, with two conditions: record all versions and test resources in an owned Markdown file for the group, and leave no dependency changes in protected files after testing. See TESTING_RESOURCES.md. Use a temporary alternate Go module manifest for test runs so the protected module files can remain unchanged throughout. This does not create a second source module or replace Asynq. Permanent module integration remains Evans's responsibility. Migration 000002's global sender-block semantics and OTP school/session binding also need coordination; use boundary tests until the real schema exists.
    +
    +## Persistent user instruction — testing dependencies
    +
    +The user explicitly said: "yes you can install for testisting. but you will have to update the whole group about the versions and resources used for testing, so make an md file to save these. then after testing and noting them down remove them from the prohibited documents, add this to you your memory".
    +
    +Apply this rule to all subsequent Developer 05 work: testing-only dependency installs are authorized; record exact versions, commands, resources, and results in TESTING_RESOURCES.md; preserve/restore protected files to their pre-test contents, never commit dependency changes to them, and verify their Git diff is empty before finishing. Keep the pre-existing Docs/DECISIONS.md edit intact. This file is the task's persistent repository memory. The requested group update is the Markdown handoff; no messaging destination has been provided.
     
    Add: /home/fian/frihk/Ikonex-Academy/backend/internal/billing/TESTING_RESOURCES.md
    # Developer 05 testing resources and team handoff
    
    ## Authorization and cleanup rule
    
    On 2026-11-09 the user authorized installing pgx and Asynq for testing only, required a Markdown record of versions/resources for the group, and required removing test dependency changes from protected documents afterward. This rule is also saved in IMPLEMENTATION_PLAN.md as persistent task memory.
    
    Use a temporary Go `-modfile` outside the source tree for these test dependencies. The application still has one source module. Preserve `backend/go.mod` byte-for-byte and preserve the absence of `backend/go.sum`; never commit testing dependencies to either protected path. Do not modify other developers' sources, shared configuration, or their migrations. Any temporary database/Redis instances must contain only synthetic test data.
    
    ## Starting state
    
    - Branch: `feat/ian-backend-engineer-c`.
    - Resumed implementation commit: `0ef7d5e8ba65829c146dde8c1844e3c11efd351f`.
    - Protected `backend/go.mod`: module `github.com/Ashomondi/ikonex-academy/backend`, Go `1.22.2`, no requirements.
    - Protected `backend/go.sum`: absent.
    - Existing unrelated edit: one leading blank line in `Docs/DECISIONS.md`; preserve unchanged.
    
    ## Direct test dependencies
    
    | Resource | Requested version | Purpose | Status |
    |---|---|---|---|
    | github.com/jackc/pgx/v5 | v5.7.2 | Compatible pgx.Tx contracts, PostgreSQL integration tests | Installation pending |
    | github.com/hibiken/asynq | v0.25.1 | Actual Redis-backed queue implementation and tests | Installation pending |
    | Go | go1.22.2 linux/amd64 | Build, vet, unit/integration/race tests | Verified installed |
    | Docker | 27.3.1 | Disposable PostgreSQL 16 and Redis 7 resources | Verified available |
    | PostgreSQL | 16, exact image digest to be recorded after pull | Migration UP/DOWN/UP, tenant RLS, financial transaction tests | Pending |
    | Redis | 7, exact image digest to be recorded after pull | Queue and worker integration tests | Pending |
    
    Record the resolved transitive module versions, exact image digests, container names and exposed loopback ports, commands, test results, and cleanup verification below as they become available. Never record live credentials, OTPs, full personal records, or service tokens.
    
    ## Integration responsibility
    
    Evans must add the recorded module requirements to the protected application manifest during integration. A successful test run with a temporary manifest is not evidence that the unchanged dependency-free application manifest can resolve newly imported pgx or Asynq packages. Report the exact manifest used with each result.

## Activity

    $ test_dir=$(mktemp -d /tmp/ikonex-dev05-testing.XXXXXX) && cp backend/go.mod "$test_dir/testing.mod" && printf '%s\n' "$test_dir"
    /tmp/ikonex-dev05-testing.xunwQ5
    ✓ • 5ms

## Assistant

The testing-only rule is now saved in the implementation plan and `TESTING_RESOURCES.md`. I’m using a temporary Go `-modfile` outside the repository, which lets us test with pgx and Asynq while keeping the protected files unchanged throughout. The handoff will identify the dependencies Evans needs to add during integration.

## Activity

    $ docker pull postgres:16-alpine && docker image inspect postgres:16-alpine --format '{{json .RepoDigests}}'
    16-alpine: Pulling from library/postgres
    881fc932939e: Pulling fs layer
    fd818ebd7c87: Pulling fs layer
    07e60219949d: Pulling fs layer
    ec3414c15456: Pulling fs layer
    a8b9fccd8d77: Pulling fs layer
    ec3414c15456: Waiting
    a8b9fccd8d77: Waiting
    850c0285d2f3: Pulling fs layer
    ac5e852d5d1d: Pulling fs layer
    55cd3be29a9d: Pulling fs layer
    850c0285d2f3: Waiting
    ac5e852d5d1d: Waiting
    108fda73eb54: Pulling fs layer
    45a91a95fdfe: Pulling fs layer
    9df8a060171e: Pulling fs layer
    55cd3be29a9d: Waiting
    108fda73eb54: Waiting
    45a91a95fdfe: Waiting
    9df8a060171e: Waiting
    fd818ebd7c87: Verifying Checksum
    fd818ebd7c87: Download complete
    881fc932939e: Download complete
    07e60219949d: Download complete
    881fc932939e: Pull complete
    fd818ebd7c87: Pull complete
    07e60219949d: Pull complete
    ec3414c15456: Verifying Checksum
    ec3414c15456: Download complete
    ec3414c15456: Pull complete
    a8b9fccd8d77: Verifying Checksum
    a8b9fccd8d77: Download complete
    a8b9fccd8d77: Pull complete
    ac5e852d5d1d: Verifying Checksum
    ac5e852d5d1d: Download complete
    55cd3be29a9d: Verifying Checksum
    55cd3be29a9d: Download complete
    45a91a95fdfe: Verifying Checksum
    45a91a95fdfe: Download complete
    108fda73eb54: Verifying Checksum
    108fda73eb54: Download complete
    9df8a060171e: Verifying Checksum
    9df8a060171e: Download complete
    850c0285d2f3: Download complete
    850c0285d2f3: Pull complete
    ac5e852d5d1d: Pull complete
    55cd3be29a9d: Pull complete
    108fda73eb54: Pull complete
    45a91a95fdfe: Pull complete
    9df8a060171e: Pull complete
    Digest: sha256:aa6cb33cf4c19c765cefcf48d96176258800870a9a1e930ad126200f9d825139
    Status: Downloaded newer image for postgres:16-alpine
    docker.io/library/postgres:16-alpine
    ["postgres@sha256:aa6cb33cf4c19c765cefcf48d96176258800870a9a1e930ad126200f9d825139"]
    ✓ • 59.36s

## Activity

    $ go get -modfile=/tmp/ikonex-dev05-testing.xunwQ5/testing.mod github.com/jackc/pgx/v5/pgxpool@v5.7.2 github.com/hibiken/asynq@v0.25.1 && go list -modfile=/tmp/ikonex-dev05-testing.xunwQ5/testing.mod -m all && go mod verify -modfile=/tmp/ikonex-dev05-testing.xunwQ5/testing.mod
    go: downloading github.com/hibiken/asynq v0.25.1
    go: downloading github.com/jackc/pgx/v5 v5.7.2
    go: downloading github.com/robfig/cron/v3 v3.0.1
    go: downloading github.com/redis/go-redis/v9 v9.7.0
    go: downloading golang.org/x/sys v0.27.0
    go: downloading golang.org/x/time v0.8.0
    go: downloading github.com/jackc/puddle/v2 v2.2.2
    go: downloading google.golang.org/protobuf v1.35.2
    go: downloading github.com/spf13/cast v1.7.0
    go: downloading github.com/jackc/pgpassfile v1.0.0
    go: downloading github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761
    go: downloading golang.org/x/crypto v0.31.0
    go: downloading golang.org/x/text v0.21.0
    go: downloading github.com/cespare/xxhash/v2 v2.2.0
    go: downloading github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f
    go: downloading golang.org/x/sync v0.10.0
    go: downloading golang.org/x/sys v0.28.0
    go: added github.com/cespare/xxhash/v2 v2.2.0
    go: added github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f
    go: added github.com/google/uuid v1.6.0
    go: added github.com/hibiken/asynq v0.25.1
    go: added github.com/jackc/pgpassfile v1.0.0
    go: added github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761
    go: added github.com/jackc/pgx/v5 v5.7.2
    go: added github.com/jackc/puddle/v2 v2.2.2
    go: added github.com/redis/go-redis/v9 v9.7.0
    go: added github.com/robfig/cron/v3 v3.0.1
    go: added github.com/spf13/cast v1.7.0
    go: added golang.org/x/crypto v0.31.0
    go: added golang.org/x/sync v0.10.0
    go: added golang.org/x/sys v0.28.0
    go: added golang.org/x/text v0.21.0
    go: added golang.org/x/time v0.8.0
    go: added google.golang.org/protobuf v1.35.2
    github.com/Ashomondi/ikonex-academy/backend
    github.com/bsm/ginkgo/v2 v2.12.0
    github.com/bsm/gomega v1.27.10
    github.com/cespare/xxhash/v2 v2.2.0
    github.com/davecgh/go-spew v1.1.1
    github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f
    github.com/frankban/quicktest v1.14.6
    github.com/golang/protobuf v1.5.0
    github.com/google/go-cmp v0.6.0
    github.com/google/uuid v1.6.0
    github.com/hibiken/asynq v0.25.1
    github.com/jackc/pgpassfile v1.0.0
    github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761
    github.com/jackc/pgx/v5 v5.7.2
    github.com/jackc/puddle/v2 v2.2.2
    github.com/kr/pretty v0.3.1
    github.com/kr/text v0.2.0
    github.com/pmezard/go-difflib v1.0.0
    github.com/redis/go-redis/v9 v9.7.0
    github.com/robfig/cron/v3 v3.0.1
    github.com/rogpeppe/go-internal v1.9.0
    github.com/spf13/cast v1.7.0
    github.com/stretchr/objx v0.1.0
    github.com/stretchr/testify v1.8.1
    go.uber.org/goleak v1.3.0
    golang.org/x/crypto v0.31.0
    golang.org/x/mod v0.17.0
    golang.org/x/net v0.21.0
    golang.org/x/sync v0.10.0
    golang.org/x/sys v0.28.0
    golang.org/x/term v0.27.0
    golang.org/x/text v0.21.0
    golang.org/x/time v0.8.0
    golang.org/x/tools v0.21.1-0.20240508182429-e35e4ccd0d2d
    golang.org/x/xerrors v0.0.0-20191204190536-9bdfabe68543
    google.golang.org/protobuf v1.35.2
    gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c
    gopkg.in/yaml.v3 v3.0.1
    all modules verified
    ✓ • 1m 01s

## Activity

    file changes: Completed · 1 changes
    Update { move_path: None }: /home/fian/frihk/Ikonex-Academy/backend/internal/billing/TESTING_RESOURCES.md
    @@ -20,7 +20,7 @@
     |---|---|---|---|
    -| github.com/jackc/pgx/v5 | v5.7.2 | Compatible pgx.Tx contracts, PostgreSQL integration tests | Installation pending |
    -| github.com/hibiken/asynq | v0.25.1 | Actual Redis-backed queue implementation and tests | Installation pending |
    +| github.com/jackc/pgx/v5 | v5.7.2 | Compatible pgx.Tx contracts, PostgreSQL integration tests | Installed; module checksums verified |
    +| github.com/hibiken/asynq | v0.25.1 | Actual Redis-backed queue implementation and tests | Installed; module checksums verified |
     | Go | go1.22.2 linux/amd64 | Build, vet, unit/integration/race tests | Verified installed |
     | Docker | 27.3.1 | Disposable PostgreSQL 16 and Redis 7 resources | Verified available |
    -| PostgreSQL | 16, exact image digest to be recorded after pull | Migration UP/DOWN/UP, tenant RLS, financial transaction tests | Pending |
    +| PostgreSQL | postgres:16-alpine, digest sha256:aa6cb33cf4c19c765cefcf48d96176258800870a9a1e930ad126200f9d825139 | Migration UP/DOWN/UP, tenant RLS, financial transaction tests | Image downloaded |
     | Redis | 7, exact image digest to be recorded after pull | Queue and worker integration tests | Pending |
    @@ -29,2 +29,54 @@
     
    +## Reproduction and resolved module graph
    +
    +The current temporary manifest is `/tmp/ikonex-dev05-testing.xunwQ5/testing.mod` (with its companion testing.sum). It was copied from the unchanged application manifest. From `backend/`:
    +
    +```sh
    +go get -modfile=/tmp/ikonex-dev05-testing.xunwQ5/testing.mod github.com/jackc/pgx/v5/pgxpool@v5.7.2 github.com/hibiken/asynq@v0.25.1
    +go list -modfile=/tmp/ikonex-dev05-testing.xunwQ5/testing.mod -m all
    +go mod verify -modfile=/tmp/ikonex-dev05-testing.xunwQ5/testing.mod
    +```
    +
    +Checksum verification returned `all modules verified`. The resolved module graph (including dependencies' own test/tool requirements, which are not all compiled into this application) is:
    +
    +```text
    +github.com/bsm/ginkgo/v2 v2.12.0
    +github.com/bsm/gomega v1.27.10
    +github.com/cespare/xxhash/v2 v2.2.0
    +github.com/davecgh/go-spew v1.1.1
    +github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f
    +github.com/frankban/quicktest v1.14.6
    +github.com/golang/protobuf v1.5.0
    +github.com/google/go-cmp v0.6.0
    +github.com/google/uuid v1.6.0
    +github.com/hibiken/asynq v0.25.1
    +github.com/jackc/pgpassfile v1.0.0
    +github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761
    +github.com/jackc/pgx/v5 v5.7.2
    +github.com/jackc/puddle/v2 v2.2.2
    +github.com/kr/pretty v0.3.1
    +github.com/kr/text v0.2.0
    +github.com/pmezard/go-difflib v1.0.0
    +github.com/redis/go-redis/v9 v9.7.0
    +github.com/robfig/cron/v3 v3.0.1
    +github.com/rogpeppe/go-internal v1.9.0
    +github.com/spf13/cast v1.7.0
    +github.com/stretchr/objx v0.1.0
    +github.com/stretchr/testify v1.8.1
    +go.uber.org/goleak v1.3.0
    +golang.org/x/crypto v0.31.0
    +golang.org/x/mod v0.17.0
    +golang.org/x/net v0.21.0
    +golang.org/x/sync v0.10.0
    +golang.org/x/sys v0.28.0
    +golang.org/x/term v0.27.0
    +golang.org/x/text v0.21.0
    +golang.org/x/time v0.8.0
    +golang.org/x/tools v0.21.1-0.20240508182429-e35e4ccd0d2d
    +golang.org/x/xerrors v0.0.0-20191204190536-9bdfabe68543
    +google.golang.org/protobuf v1.35.2
    +gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c
    +gopkg.in/yaml.v3 v3.0.1
    +```
    +
     ## Integration responsibility

## Activity

    $ go test -modfile=/tmp/ikonex-dev05-testing.xunwQ5/testing.mod -race ./...
    ?       github.com/Ashomondi/ikonex-academy/backend/cmd/server    [no test files]
    ?       github.com/Ashomondi/ikonex-academy/backend/cmd/worker    [no test files]
    ?       github.com/Ashomondi/ikonex-academy/backend/internal/academics    [no test files]
    ?       github.com/Ashomondi/ikonex-academy/backend/internal/auth    [no test files]
    ?       github.com/Ashomondi/ikonex-academy/backend/internal/billing    [no test files]
    ?       github.com/Ashomondi/ikonex-academy/backend/internal/clubs    [no test files]
    ?       github.com/Ashomondi/ikonex-academy/backend/internal/config    [no test files]
    ?       github.com/Ashomondi/ikonex-academy/backend/internal/consent    [no test files]
    ?       github.com/Ashomondi/ikonex-academy/backend/internal/database    [no test files]
    ?       github.com/Ashomondi/ikonex-academy/backend/internal/internalapi    [no test files]
    ?       github.com/Ashomondi/ikonex-academy/backend/internal/jobs    [no test files]
    ?       github.com/Ashomondi/ikonex-academy/backend/internal/library    [no test files]
    ?       github.com/Ashomondi/ikonex-academy/backend/internal/middleware    [no test files]
    ?       github.com/Ashomondi/ikonex-academy/backend/internal/nurse    [no test files]
    ?       github.com/Ashomondi/ikonex-academy/backend/internal/onboarding    [no test files]
    ok      github.com/Ashomondi/ikonex-academy/backend/internal/notifications    1.115s
    ✓ • 2.60s

## Activity

    $ docker run --detach --rm --name ikonex-dev05-postgres --publish 127.0.0.1::5432 --env POSTGRES_HOST_AUTH_METHOD=trust postgres@sha256:aa6cb33cf4c19c765cefcf48d96176258800870a9a1e930ad126200f9d825139
    02ab7aabd57d0e21dc645e0598d8487bf911c6bd8601ef46f053b8f58c779112
    ✓ • 1.23s

## Assistant

pgx `v5.7.2` and Asynq `v0.25.1` installed successfully, their checksums verified, and the current backend passed race tests with the temporary manifest. I’ve recorded the resolved module graph and PostgreSQL image digest.

I’m resuming `INF-02`: the transaction, audit, identity, and response contracts. These will let the owned modules compile and be tested without implementing Evans’s authentication or database package.

## Activity

    file changes: Completed · 2 changes
    Add: /home/fian/frihk/Ikonex-Academy/backend/internal/internalapi/contracts/contracts.go
    // Package contracts contains Developer 05's consumer boundaries. The application
    // supplies authenticated identities, tenant transactions and audit logging.
    package contracts
    
    import (
    	"context"
    	"encoding/json"
    	"errors"
    	"io"
    	"net/http"
    	"strconv"
    
    	"github.com/jackc/pgx/v5"
    	"github.com/jackc/pgx/v5/pgtype"
    )
    
    type TenantDB interface {
    	WithTenantTx(context.Context, string, func(pgx.Tx) error) error
    }
    
    // TenantTxFunc adapts the application's WithTenantTx without creating another
    // connection pool or taking ownership of transaction commit/rollback.
    type TenantTxFunc func(context.Context, string, func(pgx.Tx) error) error
    
    func (f TenantTxFunc) WithTenantTx(ctx context.Context, schoolID string, fn func(pgx.Tx) error) error {
    	if !ValidID(schoolID) {
    		return Fail(http.StatusBadRequest, "INVALID_SCHOOL", "A valid school is required.")
    	}
    	return f(ctx, schoolID, fn)
    }
    
    type AuditEvent struct {
    	SchoolID, UserID, Action, ResourceType, ResourceID string
    	Details                                         map[string]any
    }
    
    type AuditLogger interface {
    	Log(context.Context, pgx.Tx, AuditEvent) error
    }
    
    type Principal struct {
    	UserID       string `json:"user_id"`
    	SchoolID     string `json:"school_id"`
    	MembershipID string `json:"membership_id"`
    	Role         string `json:"role"`
    }
    
    // Identity must be supplied from trusted authentication middleware, never from
    // unchecked request headers or request JSON.
    type Identity func(*http.Request) (Principal, error)
    
    func RequireIdentity(r *http.Request, identity Identity) (Principal, error) {
    	if identity == nil {
    		return Principal{}, Fail(401, "UNAUTHENTICATED", "Authentication is required.")
    	}
    	p, err := identity(r)
    	if err != nil {
    		return Principal{}, Fail(401, "UNAUTHENTICATED", "Authentication is required.")
    	}
    	if !ValidID(p.UserID) || !ValidID(p.SchoolID) || !ValidID(p.MembershipID) || p.Role == "" {
    		return Principal{}, Fail(401, "UNAUTHENTICATED", "Authentication is required.")
    	}
    	return p, nil
    }
    
    func ValidID(id string) bool {
    	if len(id) != 36 || id == "00000000-0000-0000-0000-000000000000" {
    		return false
    	}
    	var value pgtype.UUID
    	return value.Scan(id) == nil && value.Valid
    }
    
    type APIError struct {
    	Status  int            `json:"-"`
    	Code    string         `json:"code"`
    	Message string         `json:"message"`
    	Details map[string]any `json:"details"`
    }
    
    func (e *APIError) Error() string { return e.Message }
    
    func Fail(status int, code, message string) *APIError {
    	return &APIError{Status: status, Code: code, Message: message, Details: map[string]any{}}
    }
    
    func WriteError(w http.ResponseWriter, err error) {
    	var apiError *APIError
    	if !errors.As(err, &apiError) {
    		apiError = Fail(500, "INTERNAL_ERROR", "The request could not be completed.")
    	}
    	write(w, apiError.Status, struct {
    		Status string    `json:"status"`
    		Error  *APIError `json:"error"`
    	}{"error", apiError})
    }
    
    type Page struct {
    	Page    int   `json:"page"`
    	PerPage int   `json:"per_page"`
    	Total   int64 `json:"total"`
    }
    
    func Success(w http.ResponseWriter, status int, data any, page *Page) {
    	write(w, status, struct {
    		Status string `json:"status"`
    		Data   any    `json:"data"`
    		Meta   *Page  `json:"meta,omitempty"`
    	}{"success", data, page})
    }
    
    func write(w http.ResponseWriter, status int, body any) {
    	encoded, err := json.Marshal(body)
    	if err != nil {
    		status = 500
    		encoded = []byte(`{"status":"error","error":{"code":"INTERNAL_ERROR","message":"The request could not be completed.","details":{}}}`)
    	}
    	w.Header().Set("Content-Type", "application/json")
    	w.Header().Set("Cache-Control", "no-store")
    	w.WriteHeader(status)
    	_, _ = w.Write(encoded)
    }
    
    func Decode(w http.ResponseWriter, r *http.Request, target any) error {
    	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
    	decoder := json.NewDecoder(r.Body)
    	decoder.DisallowUnknownFields()
    	if err := decoder.Decode(target); err != nil {
    		return Fail(400, "INVALID_JSON", "A valid JSON request is required.")
    	}
    	if err := decoder.Decode(new(any)); err != io.EOF {
    		return Fail(400, "INVALID_JSON", "Only one JSON value is allowed.")
    	}
    	return nil
    }
    
    func Pagination(r *http.Request) (Page, error) {
    	p := Page{Page: 1, PerPage: 25}
    	for key, field := range map[string]*int{"page": &p.Page, "per_page": &p.PerPage} {
    		if value := r.URL.Query().Get(key); value != "" {
    			n, err := strconv.Atoi(value)
    			if err != nil || n < 1 {
    				return Page{}, Fail(422, "INVALID_PAGINATION", "Pagination values must be positive integers.")
    			}
    			*field = n
    		}
    	}
    	if p.PerPage > 100 || p.Page > 1000000 {
    		return Page{}, Fail(422, "INVALID_PAGINATION", "Pagination values exceed the supported limit.")
    	}
    	return p, nil
    }
    Add: /home/fian/frihk/Ikonex-Academy/backend/internal/internalapi/contracts/contracts_test.go
    package contracts
    
    import (
    	"context"
    	"errors"
    	"net/http"
    	"net/http/httptest"
    	"strings"
    	"testing"
    
    	"github.com/jackc/pgx/v5"
    )
    
    const syntheticID = "10000000-0000-0000-0000-000000000001"
    
    func TestIdentityRejectsUntrustedOrIncompleteContext(t *testing.T) {
    	r := httptest.NewRequest("GET", "/", nil)
    	r.Header.Set("X-User-ID", syntheticID)
    	for _, identity := range []Identity{nil, func(*http.Request) (Principal, error) {
    		return Principal{UserID: syntheticID}, nil
    	}, func(*http.Request) (Principal, error) { return Principal{}, errors.New("private token") }} {
    		if _, err := RequireIdentity(r, identity); err == nil || strings.Contains(err.Error(), "private") {
    			t.Fatalf("invalid identity: %v", err)
    		}
    	}
    	p := Principal{syntheticID, syntheticID, syntheticID, "parent"}
    	if got, err := RequireIdentity(r, func(*http.Request) (Principal, error) { return p, nil }); err != nil || got != p {
    		t.Fatalf("trusted identity: %+v %v", got, err)
    	}
    }
    
    func TestTenantAdapterPropagatesTransactionFailure(t *testing.T) {
    	failure := errors.New("rollback")
    	called := 0
    	db := TenantTxFunc(func(ctx context.Context, school string, fn func(pgx.Tx) error) error {
    		called++
    		if school != syntheticID {
    			t.Fatal("tenant changed")
    		}
    		return fn(nil)
    	})
    	if err := db.WithTenantTx(context.Background(), "invalid", nil); err == nil || called != 0 {
    		t.Fatal("invalid tenant reached adapter")
    	}
    	if err := db.WithTenantTx(context.Background(), syntheticID, func(pgx.Tx) error { return failure }); !errors.Is(err, failure) || called != 1 {
    		t.Fatalf("transaction error lost: %v", err)
    	}
    }
    
    func TestResponseEnvelopesAndSafeErrors(t *testing.T) {
    	w := httptest.NewRecorder()
    	WriteError(w, errors.New("database password private"))
    	if w.Code != 500 || strings.Contains(w.Body.String(), "password") || !strings.Contains(w.Body.String(), `"details":{}`) {
    		t.Fatalf("unsafe error: %s", w.Body.String())
    	}
    	w = httptest.NewRecorder()
    	WriteError(w, Fail(403, "FORBIDDEN", "Access denied."))
    	if w.Code != 403 || !strings.Contains(w.Body.String(), `"code":"FORBIDDEN"`) {
    		t.Fatal(w.Body.String())
    	}
    	w = httptest.NewRecorder()
    	Success(w, 200, []string{}, &Page{1, 25, 0})
    	if w.Body.String() != `{"status":"success","data":[],"meta":{"page":1,"per_page":25,"total":0}}` || w.Header().Get("Cache-Control") != "no-store" {
    		t.Fatal(w.Body.String())
    	}
    }
    
    func TestDecodeAndPagination(t *testing.T) {
    	for _, body := range []string{`{"unexpected":true}`, `{}`, `{} {}`, strings.Repeat("x", (1<<20)+1)} {
    		var input struct{}
    		err := Decode(httptest.NewRecorder(), httptest.NewRequest("POST", "/", strings.NewReader(body)), &input)
    		if (err == nil) != (body == "{}") {
    			t.Fatalf("decode result %v", err)
    		}
    	}
    	for _, query := range []string{"page=0", "per_page=101", "page=1000001", "page=bad"} {
    		if _, err := Pagination(httptest.NewRequest("GET", "/?"+query, nil)); err == nil {
    			t.Fatalf("invalid pagination accepted: %s", query)
    		}
    	}
    	if p, err := Pagination(httptest.NewRequest("GET", "/", nil)); err != nil || p.Page != 1 || p.PerPage != 25 {
    		t.Fatalf("defaults: %+v %v", p, err)
    	}
    }

## Activity

    $ gofmt -w internal/internalapi/contracts/contracts.go internal/internalapi/contracts/contracts_test.go && go test -modfile=/tmp/ikonex-dev05-testing.xunwQ5/testing.mod -race ./internal/internalapi/contracts && go vet -modfile=/tmp/ikonex-dev05-testing.xunwQ5/testing.mod ./internal/internalapi/contracts
    ok      github.com/Ashomondi/ikonex-academy/backend/internal/internalapi/contracts    1.050s
    ✓ • 1m 12s

## Activity

    $ docker port ikonex-dev05-postgres 5432/tcp && docker exec ikonex-dev05-postgres psql -U postgres -Atc 'SELECT version();' && git diff -- backend/go.mod backend/go.sum
    127.0.0.1:32768
    PostgreSQL 16.15 on x86_64-pc-linux-musl, compiled by gcc (Alpine 15.2.0) 15.2.0, 64-bit
    ✓ • 214ms

## Assistant

`INF-02` passed race tests and vet. PostgreSQL `16.15` is running on `127.0.0.1:32768`, and the protected module files still have no diff.

I’m moving migration verification ahead of in-app notification persistence: that persistence task needs the new table for its database test gate. I’ll test migration `000004` against the existing baseline, including rollback and RLS under a non-superuser role.

## Activity

    file changes: Completed · 2 changes
    Update { move_path: None }: /home/fian/frihk/Ikonex-Academy/backend/internal/billing/IMPLEMENTATION_PLAN.md
    @@ -54,2 +54,4 @@
     
    +2026-11-09 execution adjustment: DB-01–03 precede NOTIF-04/05 because the actual notification persistence verification needs the new table. INF-02's temporary-manifest race tests and vet have passed. This breaks the original notification/schema test dependency cycle without weakening the database verification gate.
    +
     Paths below are relative to the repository root. `B`, `N`, `W`, `I`, `J` mean `backend/internal/billing`, `notifications`, `whatsapp`, `internalapi`, `jobs` respectively. `M` means `migrations/000004_r1_billing_whatsapp_notifications`. Tests live beside their implementation. Each row is one task: inspect, implement its ordered steps, gofmt only affected files, run verification, review the diff and ownership, then update status. Never advance a dependent task past a failed gate.
    @@ -64,3 +66,3 @@
     | NOTIF-03 | C3 Africa's Talking transport | NOTIF-01 | N/sms_provider.go, N/sms_provider_test.go | Select local sandbox by username; context-bound HTTP client; encoded form and apiKey; validate per-recipient response; classify retryable/permanent failures | httptest request contract, provider rejection, timeout, malformed response, cancellation; G(notifications) | Completed; package and full-backend race tests, build and vet passed |
    -| INF-02 | C1–C3 transaction/auth/error contracts | INF-01, pgx | I/contracts/contracts.go, I/contracts/contracts_test.go | Small consumer contracts for pgx tenant transactions/audit/identity; APIError and standard envelopes; deny missing identity | Contract doubles, malformed identity, error envelope and safe error mapping; G(internalapi/...) | Compatible injection, no substitute auth; Blocked on module dependency |
    +| INF-02 | C1–C3 transaction/auth/error contracts | INF-01, pgx | I/contracts/contracts.go, I/contracts/contracts_test.go | Small consumer contracts for pgx tenant transactions/audit/identity; APIError and standard envelopes; deny missing identity | Contract doubles, malformed identity, error envelope and safe error mapping; G(internalapi/...) | Completed; race tests and vet passed using temporary manifest |
     | NOTIF-04 | C3 transaction-bound notification creation | INF-02 | N/notifications.go, N/notifications_test.go | Validate content/recipient; scoped insertion within supplied transaction; define InAppNotifier | Rollback, tenant/user checks; PostgreSQL insertion and RLS; G(notifications) | Usable by Ema and billing without separate commit; Pending |
    Update { move_path: None }: /home/fian/frihk/Ikonex-Academy/backend/internal/billing/TESTING_RESOURCES.md
    @@ -27,2 +27,6 @@
     
    +PostgreSQL reports `16.15`, x86_64 Linux/musl, GCC 15.2.0. Test container: `ikonex-dev05-postgres`, published only on `127.0.0.1:32768`, with disposable trust authentication and no live data or mounted production volumes. Started from the digest recorded above. Stop/remove it after the test session. The test role will be non-superuser for application/RLS assertions; the bootstrap account is reserved for disposable schema setup and teardown.
    +
    +Initial dependency-enabled baseline: `go test -modfile=/tmp/ikonex-dev05-testing.xunwQ5/testing.mod -race ./...` passed. INF-02: the same manifest with `go test -race ./internal/internalapi/contracts` and `go vet ./internal/internalapi/contracts` passed. Protected module files have no diff.
    +
     Record the resolved transitive module versions, exact image digests, container names and exposed loopback ports, commands, test results, and cleanup verification below as they become available. Never record live credentials, OTPs, full personal records, or service tokens.

## Activity

    $ git status --short && git diff --name-only && git diff --cached --name-only && git diff --check && git add backend/internal/billing/IMPLEMENTATION_PLAN.md backend/internal/billing/TESTING_RESOURCES.md backend/internal/internalapi/contracts && git diff --cached --name-only && git commit -m 'feat(internalapi): define consumer integration contracts'
     M Docs/DECISIONS.md
     M backend/internal/billing/IMPLEMENTATION_PLAN.md
    ?? backend/internal/billing/TESTING_RESOURCES.md
    ?? backend/internal/internalapi/contracts/
    Docs/DECISIONS.md
    backend/internal/billing/IMPLEMENTATION_PLAN.md
    backend/internal/billing/IMPLEMENTATION_PLAN.md
    backend/internal/billing/TESTING_RESOURCES.md
    backend/internal/internalapi/contracts/contracts.go
    backend/internal/internalapi/contracts/contracts_test.go
    [feat/ian-backend-engineer-c 4db9d2d] feat(internalapi): define consumer integration contracts
     4 files changed, 334 insertions(+), 2 deletions(-)
     create mode 100644 backend/internal/billing/TESTING_RESOURCES.md
     create mode 100644 backend/internal/internalapi/contracts/contracts.go
     create mode 100644 backend/internal/internalapi/contracts/contracts_test.go
    ✓ • 354ms

## Activity

    file changes: Completed · 2 changes
    Add: /home/fian/frihk/Ikonex-Academy/migrations/000004_r1_billing_whatsapp_notifications.down.sql
    BEGIN;
    
    DROP TABLE ai_action_logs;
    DROP TABLE in_app_notifications;
    DROP TABLE chat_sessions;
    DROP TABLE school_whatsapp_bots;
    DROP TABLE mpesa_stk_requests;
    DROP TABLE unmatched_payments;
    DROP TABLE ledger_entries;
    DROP TABLE fee_change_note_students;
    DROP TABLE fee_change_notes;
    DROP TABLE fee_charges;
    DROP TABLE fee_periods;
    DROP TABLE school_mpesa_configs;
    
    ALTER TABLE fee_ledgers DROP CONSTRAINT r1_c_ledger_nonnegative,
        DROP COLUMN credit_balance, DROP COLUMN due_date;
    ALTER TABLE fee_structures DROP COLUMN boarding_status;
    ALTER TABLE payments DROP CONSTRAINT r1_c_payment_school_id;
    ALTER TABLE fee_structures DROP CONSTRAINT r1_c_structure_school_id;
    ALTER TABLE students DROP CONSTRAINT r1_c_student_school_id;
    ALTER TABLE terms DROP CONSTRAINT r1_c_term_school_id;
    ALTER TABLE terms DROP CONSTRAINT r1_c_term_school_year_id;
    ALTER TABLE academic_years DROP CONSTRAINT r1_c_year_school_id;
    
    COMMIT;
    Add: /home/fian/frihk/Ikonex-Academy/migrations/000004_r1_billing_whatsapp_notifications.up.sql
    BEGIN;
    
    -- Composite keys let new foreign keys enforce tenant ownership as well as RLS.
    ALTER TABLE academic_years ADD CONSTRAINT r1_c_year_school_id UNIQUE (school_id, id);
    ALTER TABLE terms ADD CONSTRAINT r1_c_term_school_year_id UNIQUE (school_id, academic_year_id, id);
    ALTER TABLE terms ADD CONSTRAINT r1_c_term_school_id UNIQUE (school_id, id);
    ALTER TABLE students ADD CONSTRAINT r1_c_student_school_id UNIQUE (school_id, id);
    ALTER TABLE fee_structures ADD CONSTRAINT r1_c_structure_school_id UNIQUE (school_id, id);
    ALTER TABLE payments ADD CONSTRAINT r1_c_payment_school_id UNIQUE (school_id, id);
    ALTER TABLE fee_structures ADD COLUMN boarding_status VARCHAR(20) NOT NULL DEFAULT 'all'
        CHECK (boarding_status IN ('all', 'boarder', 'day'));
    ALTER TABLE fee_ledgers ADD COLUMN credit_balance NUMERIC(12,2) NOT NULL DEFAULT 0.00,
        ADD COLUMN due_date DATE,
        ADD CONSTRAINT r1_c_ledger_nonnegative CHECK (current_balance >= 0 AND credit_balance >= 0);
    
    CREATE TABLE school_mpesa_configs (
        school_id UUID PRIMARY KEY REFERENCES schools(id) ON DELETE CASCADE,
        shortcode VARCHAR(30) UNIQUE NOT NULL,
        environment VARCHAR(20) NOT NULL DEFAULT 'sandbox' CHECK (environment IN ('sandbox', 'production')),
        passkey_encrypted TEXT,
        test_payment_verified BOOLEAN NOT NULL DEFAULT FALSE,
        updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
    );
    
    CREATE TABLE fee_periods (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
        school_id UUID NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
        academic_year_id UUID NOT NULL,
        term_id UUID,
        period_type VARCHAR(30) NOT NULL DEFAULT 'term' CHECK (btrim(period_type) <> ''),
        name VARCHAR(100) NOT NULL CHECK (btrim(name) <> ''),
        due_date DATE NOT NULL,
        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
        CHECK (period_type <> 'term' OR term_id IS NOT NULL),
        UNIQUE (school_id, id),
        UNIQUE (school_id, academic_year_id, term_id),
        FOREIGN KEY (school_id, academic_year_id) REFERENCES academic_years(school_id, id),
        FOREIGN KEY (school_id, academic_year_id, term_id) REFERENCES terms(school_id, academic_year_id, id)
    );
    
    CREATE TABLE fee_charges (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
        school_id UUID NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
        student_id UUID NOT NULL,
        fee_period_id UUID NOT NULL,
        fee_structure_id UUID NOT NULL,
        amount NUMERIC(12,2) NOT NULL CHECK (amount >= 0),
        due_date DATE NOT NULL,
        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
        UNIQUE (school_id, student_id, fee_period_id),
        FOREIGN KEY (school_id, student_id) REFERENCES students(school_id, id),
        FOREIGN KEY (school_id, fee_period_id) REFERENCES fee_periods(school_id, id),
        FOREIGN KEY (school_id, fee_structure_id) REFERENCES fee_structures(school_id, id)
    );
    CREATE INDEX fee_charges_period_idx ON fee_charges(school_id, fee_period_id);
    CREATE INDEX fee_charges_structure_idx ON fee_charges(school_id, fee_structure_id);
    
    CREATE TABLE fee_change_notes (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
        school_id UUID NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
        term_id UUID,
        fee_period_id UUID,
        change_type VARCHAR(50) NOT NULL CHECK (change_type IN ('amount', 'due_date')),
        amount_delta NUMERIC(12,2) NOT NULL DEFAULT 0,
        old_due_date DATE,
        new_due_date DATE,
        note TEXT NOT NULL CHECK (btrim(note) <> ''),
        created_by UUID NOT NULL REFERENCES users(id),
        affected_students_count INT NOT NULL CHECK (affected_students_count >= 0),
        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
        UNIQUE (school_id, id),
        FOREIGN KEY (school_id, term_id) REFERENCES terms(school_id, id),
        FOREIGN KEY (school_id, fee_period_id) REFERENCES fee_periods(school_id, id)
    );
    CREATE INDEX fee_change_notes_period_idx ON fee_change_notes(school_id, fee_period_id, created_at);
    
    -- A note may affect only some students in a period; preserve that history.
    CREATE TABLE fee_change_note_students (
        school_id UUID NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
        note_id UUID NOT NULL,
        student_id UUID NOT NULL,
        PRIMARY KEY (school_id, note_id, student_id),
        FOREIGN KEY (school_id, note_id) REFERENCES fee_change_notes(school_id, id) ON DELETE CASCADE,
        FOREIGN KEY (school_id, student_id) REFERENCES students(school_id, id)
    );
    CREATE INDEX fee_change_note_students_student_idx ON fee_change_note_students(school_id, student_id);
    
    CREATE TABLE ledger_entries (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
        school_id UUID NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
        student_id UUID NOT NULL,
        term_id UUID NOT NULL,
        entry_type VARCHAR(30) NOT NULL CHECK (entry_type IN ('charge', 'payment', 'credit_applied', 'adjustment')),
        amount NUMERIC(12,2) NOT NULL,
        reference VARCHAR(100) NOT NULL,
        description TEXT,
        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
        UNIQUE (school_id, student_id, term_id, entry_type, reference),
        FOREIGN KEY (school_id, student_id) REFERENCES students(school_id, id),
        FOREIGN KEY (school_id, term_id) REFERENCES terms(school_id, id)
    );
    CREATE INDEX ledger_entries_student_idx ON ledger_entries(school_id, student_id, created_at, id);
    
    CREATE TABLE unmatched_payments (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
        school_id UUID NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
        payment_id UUID NOT NULL,
        transaction_ref VARCHAR(100) NOT NULL,
        amount NUMERIC(12,2) NOT NULL CHECK (amount > 0),
        payer_phone VARCHAR(20),
        payer_name VARCHAR(255),
        raw_account_ref VARCHAR(100),
        status VARCHAR(30) NOT NULL DEFAULT 'unmatched' CHECK (status IN ('unmatched', 'resolved')),
        resolved_student_id UUID,
        resolved_by UUID REFERENCES users(id),
        resolution_note TEXT,
        resolved_at TIMESTAMPTZ,
        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
        UNIQUE (school_id, payment_id),
        UNIQUE (school_id, transaction_ref),
        CHECK ((status = 'unmatched' AND resolved_at IS NULL AND resolved_student_id IS NULL AND resolved_by IS NULL)
            OR (status = 'resolved' AND resolved_at IS NOT NULL AND resolved_student_id IS NOT NULL AND resolved_by IS NOT NULL)),
        FOREIGN KEY (school_id, payment_id) REFERENCES payments(school_id, id) ON DELETE CASCADE,
        FOREIGN KEY (school_id, resolved_student_id) REFERENCES students(school_id, id)
    );
    CREATE INDEX unmatched_payments_queue_idx ON unmatched_payments(school_id, created_at, id) WHERE status = 'unmatched';
    
    -- A durable attempt is distinct from a settled payment and its receipt number.
    CREATE TABLE mpesa_stk_requests (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
        school_id UUID NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
        student_id UUID NOT NULL,
        guardian_user_id UUID NOT NULL REFERENCES users(id),
        amount NUMERIC(12,2) NOT NULL CHECK (amount > 0),
        phone_number VARCHAR(20) NOT NULL,
        checkout_request_id VARCHAR(100) UNIQUE,
        merchant_request_id VARCHAR(100),
        status VARCHAR(30) NOT NULL DEFAULT 'initiating' CHECK (status IN ('initiating', 'pending', 'succeeded', 'failed', 'timeout', 'uncertain')),
        result_code INT,
        transaction_ref VARCHAR(100),
        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
        updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
        FOREIGN KEY (school_id, student_id) REFERENCES students(school_id, id)
    );
    CREATE INDEX mpesa_stk_requests_pending_idx ON mpesa_stk_requests(school_id, status, created_at);
    
    CREATE TABLE school_whatsapp_bots (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
        school_id UUID NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
        phone_number_id VARCHAR(100) UNIQUE NOT NULL,
        display_phone_number VARCHAR(30) UNIQUE NOT NULL,
        is_active BOOLEAN NOT NULL DEFAULT TRUE,
        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
    );
    CREATE INDEX school_whatsapp_bots_school_idx ON school_whatsapp_bots(school_id);
    
    CREATE TABLE chat_sessions (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
        school_id UUID NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
        channel VARCHAR(20) NOT NULL CHECK (channel IN ('whatsapp', 'in_app')),
        sender_wa_number VARCHAR(30),
        student_id UUID,
        guardian_user_id UUID REFERENCES users(id),
        state VARCHAR(40) NOT NULL DEFAULT 'AWAITING_ADMISSION_NUMBER'
            CHECK (state IN ('AWAITING_ADMISSION_NUMBER', 'AWAITING_GUARDIAN_SELECTION', 'AWAITING_OTP', 'VERIFIED')),
        otp_hash TEXT,
        otp_expires_at TIMESTAMPTZ,
        otp_attempts INT NOT NULL DEFAULT 0 CHECK (otp_attempts >= 0),
        access_token_hash TEXT,
        verified_at TIMESTAMPTZ,
        expires_at TIMESTAMPTZ,
        lookup_count INT NOT NULL DEFAULT 0 CHECK (lookup_count >= 0),
        payment_init_count INT NOT NULL DEFAULT 0 CHECK (payment_init_count >= 0),
        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
        updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
        CHECK (state <> 'VERIFIED' OR (guardian_user_id IS NOT NULL AND verified_at IS NOT NULL AND expires_at > verified_at)),
        UNIQUE (school_id, id),
        FOREIGN KEY (school_id, student_id) REFERENCES students(school_id, id)
    );
    CREATE UNIQUE INDEX chat_sessions_sender_idx ON chat_sessions(school_id, sender_wa_number) WHERE channel = 'whatsapp';
    CREATE INDEX chat_sessions_guardian_idx ON chat_sessions(school_id, guardian_user_id, expires_at);
    
    CREATE TABLE in_app_notifications (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
        school_id UUID NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
        recipient_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
        category VARCHAR(50) NOT NULL CHECK (btrim(category) <> ''),
        title VARCHAR(255) NOT NULL CHECK (btrim(title) <> ''),
        body TEXT NOT NULL CHECK (btrim(body) <> ''),
        metadata JSONB NOT NULL DEFAULT '{}' CHECK (jsonb_typeof(metadata) = 'object'),
        is_read BOOLEAN NOT NULL DEFAULT FALSE,
        read_at TIMESTAMPTZ,
        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
        CHECK (is_read = (read_at IS NOT NULL))
    );
    CREATE INDEX in_app_notifications_recipient_idx ON in_app_notifications(school_id, recipient_user_id, created_at DESC, id);
    
    CREATE TABLE ai_action_logs (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
        school_id UUID NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
        user_id UUID REFERENCES users(id),
        session_id UUID,
        channel VARCHAR(20) NOT NULL CHECK (channel IN ('whatsapp', 'in_app')),
        feature VARCHAR(50) NOT NULL,
        intent VARCHAR(50),
        endpoints_read JSONB NOT NULL DEFAULT '[]' CHECK (jsonb_typeof(endpoints_read) = 'array'),
        action_taken VARCHAR(100),
        metadata JSONB NOT NULL DEFAULT '{}' CHECK (jsonb_typeof(metadata) = 'object'),
        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
        FOREIGN KEY (school_id, session_id) REFERENCES chat_sessions(school_id, id)
    );
    CREATE INDEX ai_action_logs_session_idx ON ai_action_logs(school_id, session_id, created_at);
    
    DO $$
    DECLARE tbl TEXT;
    BEGIN
        FOREACH tbl IN ARRAY ARRAY['school_mpesa_configs', 'fee_periods', 'fee_charges', 'fee_change_notes',
            'fee_change_note_students', 'ledger_entries', 'unmatched_payments', 'mpesa_stk_requests',
            'school_whatsapp_bots', 'chat_sessions', 'in_app_notifications', 'ai_action_logs'] LOOP
            EXECUTE format('ALTER TABLE %I ENABLE ROW LEVEL SECURITY', tbl);
            EXECUTE format('ALTER TABLE %I FORCE ROW LEVEL SECURITY', tbl);
            EXECUTE format('CREATE POLICY %I ON %I FOR ALL USING (school_id = current_school_id()) WITH CHECK (school_id = current_school_id())', tbl || '_tenant_isolation', tbl);
        END LOOP;
    END $$;
    
    COMMIT;

## Assistant

The migration pair is written. It includes the ten required tables and two supporting tables: durable STK attempts and the students affected by each fee-change note. Composite foreign keys enforce matching school IDs, and all twelve tables enable and force RLS.

I’m now adding a disposable database test harness to verify application, rollback, schema preservation, and cross-tenant rejection against PostgreSQL.

## Activity

    file changes: Completed · 2 changes
    Add: /home/fian/frihk/Ikonex-Academy/backend/internal/billing/migration_test.go
    package billing_test
    
    import (
    	"context"
    	"errors"
    	"reflect"
    	"testing"
    
    	"github.com/Ashomondi/ikonex-academy/backend/internal/billing/testenv"
    	"github.com/jackc/pgx/v5"
    	"github.com/jackc/pgx/v5/pgconn"
    )
    
    func schemaSnapshot(t *testing.T, db *testenv.DB) []string {
    	t.Helper()
    	rows, err := db.Admin.Query(context.Background(), `
    		SELECT 'column:'||table_name||':'||column_name||':'||data_type||':'||is_nullable||':'||coalesce(column_default,'')
    		FROM information_schema.columns WHERE table_schema='public'
    		UNION ALL SELECT 'constraint:'||c.relname||':'||conname||':'||pg_get_constraintdef(k.oid)
    		FROM pg_constraint k JOIN pg_class c ON c.oid=k.conrelid JOIN pg_namespace n ON n.oid=c.relnamespace WHERE n.nspname='public'
    		UNION ALL SELECT 'index:'||indexname||':'||indexdef FROM pg_indexes WHERE schemaname='public'
    		ORDER BY 1`)
    	if err != nil {
    		t.Fatal(err)
    	}
    	defer rows.Close()
    	var snapshot []string
    	for rows.Next() {
    		var entry string
    		if err := rows.Scan(&entry); err != nil {
    			t.Fatal(err)
    		}
    		snapshot = append(snapshot, entry)
    	}
    	if err := rows.Err(); err != nil {
    		t.Fatal(err)
    	}
    	return snapshot
    }
    
    func TestMigrationUpDownUpPreservesBaseline(t *testing.T) {
    	db := testenv.New(t, false)
    	school := db.SeedSchool(t, "01")
    	baseline := schemaSnapshot(t, db)
    	db.Apply(t, "000004_r1_billing_whatsapp_notifications.up.sql")
    	db.Apply(t, "000004_r1_billing_whatsapp_notifications.down.sql")
    	if got := schemaSnapshot(t, db); !reflect.DeepEqual(got, baseline) {
    		t.Fatal("owned rollback changed baseline schema, constraints, or indexes")
    	}
    	var name string
    	if err := db.Admin.QueryRow(context.Background(), "SELECT full_name FROM students WHERE school_id=$1 AND id=$2", school.ID, school.StudentID).Scan(&name); err != nil || name != "Synthetic Student" {
    		t.Fatalf("rollback changed baseline data: %q %v", name, err)
    	}
    	db.Apply(t, "000004_r1_billing_whatsapp_notifications.up.sql")
    	for _, table := range []string{"school_mpesa_configs", "fee_periods", "fee_charges", "fee_change_notes", "fee_change_note_students", "ledger_entries", "unmatched_payments", "mpesa_stk_requests", "school_whatsapp_bots", "chat_sessions", "in_app_notifications", "ai_action_logs"} {
    		var enabled, forced, policy bool
    		err := db.Admin.QueryRow(context.Background(), `SELECT c.relrowsecurity,c.relforcerowsecurity,
    		EXISTS(SELECT 1 FROM pg_policy p WHERE p.polrelid=c.oid AND p.polcmd='*' AND p.polqual IS NOT NULL AND p.polwithcheck IS NOT NULL)
    		FROM pg_class c JOIN pg_namespace n ON n.oid=c.relnamespace WHERE n.nspname='public' AND c.relname=$1`, table).Scan(&enabled, &forced, &policy)
    		if err != nil || !enabled || !forced || !policy {
    			t.Fatalf("RLS missing on %s: %v", table, err)
    		}
    	}
    }
    
    func TestMigrationTenantIsolationAndReferences(t *testing.T) {
    	db := testenv.New(t, true)
    	a, b := db.SeedSchool(t, "01"), db.SeedSchool(t, "02")
    	ctx := context.Background()
    	if _, err := db.Admin.Exec(ctx, "INSERT INTO in_app_notifications(school_id,recipient_user_id,category,title,body) VALUES($1,$2,'test','Demo','Synthetic'),($3,$4,'test','Demo','Synthetic')", a.ID, a.UserID, b.ID, b.UserID); err != nil {
    		t.Fatal(err)
    	}
    	for _, tenant := range []string{"", a.ID, b.ID} {
    		err := db.WithTenantTx(ctx, tenant, func(tx pgx.Tx) error {
    			var count int
    			// Deliberately unscoped negative test: RLS itself must filter records.
    			if err := tx.QueryRow(ctx, "SELECT count(*) FROM in_app_notifications").Scan(&count); err != nil {
    				return err
    			}
    			want := 1
    			if tenant == "" {
    				want = 0
    			}
    			if count != want {
    				t.Fatalf("RLS returned %d rows, want %d", count, want)
    			}
    			var superuser, bypass bool
    			if err := tx.QueryRow(ctx, "SELECT rolsuper,rolbypassrls FROM pg_roles WHERE rolname=current_user").Scan(&superuser, &bypass); err != nil {
    				return err
    			}
    			if superuser || bypass {
    				t.Fatal("RLS tested with privileged role")
    			}
    			return nil
    		})
    		if err != nil {
    			t.Fatal(err)
    		}
    	}
    	for _, tc := range []struct {
    		sql, code string
    		args      []any
    	}{
    		{"INSERT INTO in_app_notifications(school_id,recipient_user_id,category,title,body) VALUES($1,$2,'test','Demo','Synthetic')", "42501", []any{b.ID, b.UserID}},
    		{"INSERT INTO fee_periods(school_id,academic_year_id,term_id,name,due_date) VALUES($1,$2,$3,'Term 1','2026-01-05')", "23503", []any{a.ID, b.YearID, b.TermID}},
    		{"INSERT INTO fee_change_notes(school_id,change_type,note,created_by,affected_students_count) VALUES($1,'amount',' ',$2,0)", "23514", []any{a.ID, a.UserID}},
    		{"INSERT INTO fee_ledgers(school_id,student_id,academic_year_id,term_id,current_balance) VALUES($1,$2,$3,$4,-1)", "23514", []any{a.ID, a.StudentID, a.YearID, a.TermID}},
    	} {
    		err := db.WithTenantTx(ctx, a.ID, func(tx pgx.Tx) error { _, err := tx.Exec(ctx, tc.sql, tc.args...); return err })
    		var pgErr *pgconn.PgError
    		if !errors.As(err, &pgErr) || pgErr.Code != tc.code {
    			t.Fatalf("expected PostgreSQL %s, got %v", tc.code, err)
    		}
    	}
    }
    Add: /home/fian/frihk/Ikonex-Academy/backend/internal/billing/testenv/database.go
    // Package testenv provides disposable synthetic PostgreSQL fixtures for owned
    // modules. It is never used by the application or worker binaries.
    package testenv
    
    import (
    	"context"
    	"crypto/rand"
    	"encoding/hex"
    	"os"
    	"path/filepath"
    	"runtime"
    	"testing"
    
    	"github.com/jackc/pgx/v5"
    	"github.com/jackc/pgx/v5/pgxpool"
    )
    
    type DB struct {
    	Admin *pgxpool.Pool
    	role  string
    }
    
    func New(t *testing.T, applyOwned bool) *DB {
    	t.Helper()
    	dsn := os.Getenv("DEV05_TEST_DATABASE_URL")
    	if dsn == "" {
    		t.Skip("DEV05_TEST_DATABASE_URL is required for real PostgreSQL verification")
    	}
    	ctx := context.Background()
    	root, err := pgxpool.New(ctx, dsn)
    	if err != nil {
    		t.Fatal(err)
    	}
    	var random [12]byte
    	if _, err := rand.Read(random[:]); err != nil {
    		t.Fatal(err)
    	}
    	name := "dev05_" + hex.EncodeToString(random[:])
    	ident := pgx.Identifier{name}.Sanitize()
    	if _, err := root.Exec(ctx, "CREATE DATABASE "+ident); err != nil {
    		root.Close()
    		t.Fatal(err)
    	}
    	t.Cleanup(func() {
    		if _, err := root.Exec(context.Background(), "DROP DATABASE "+ident+" WITH (FORCE)"); err != nil {
    			t.Error(err)
    		}
    		if _, err := root.Exec(context.Background(), "DROP ROLE IF EXISTS "+ident); err != nil {
    			t.Error(err)
    		}
    		root.Close()
    	})
    	if _, err := root.Exec(ctx, "CREATE ROLE "+ident+" NOLOGIN NOSUPERUSER NOBYPASSRLS"); err != nil {
    		t.Fatal(err)
    	}
    	config, err := pgxpool.ParseConfig(dsn)
    	if err != nil {
    		t.Fatal(err)
    	}
    	config.ConnConfig.Database = name
    	config.MaxConns = 8
    	pool, err := pgxpool.NewWithConfig(ctx, config)
    	if err != nil {
    		t.Fatal(err)
    	}
    	t.Cleanup(pool.Close)
    	db := &DB{Admin: pool, role: ident}
    	db.Apply(t, "000001_init_schema.up.sql")
    	if applyOwned {
    		db.Apply(t, "000004_r1_billing_whatsapp_notifications.up.sql")
    	}
    	db.Grant(t)
    	return db
    }
    
    func (d *DB) Apply(t *testing.T, filename string) {
    	t.Helper()
    	_, source, _, _ := runtime.Caller(0)
    	path := filepath.Join(filepath.Dir(source), "..", "..", "..", "..", "migrations", filename)
    	body, err := os.ReadFile(path)
    	if err != nil {
    		t.Fatal(err)
    	}
    	if _, err := d.Admin.Exec(context.Background(), string(body)); err != nil {
    		t.Fatalf("migration %s: %v", filename, err)
    	}
    }
    
    func (d *DB) Grant(t *testing.T) {
    	t.Helper()
    	if _, err := d.Admin.Exec(context.Background(), "GRANT USAGE ON SCHEMA public TO "+d.role+"; GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO "+d.role); err != nil {
    		t.Fatal(err)
    	}
    }
    
    func (d *DB) WithTenantTx(ctx context.Context, schoolID string, fn func(pgx.Tx) error) error {
    	tx, err := d.Admin.Begin(ctx)
    	if err != nil {
    		return err
    	}
    	defer tx.Rollback(context.Background())
    	if _, err = tx.Exec(ctx, "SET LOCAL ROLE "+d.role); err != nil {
    		return err
    	}
    	if _, err = tx.Exec(ctx, "SELECT set_config('app.school_id', $1, true)", schoolID); err != nil {
    		return err
    	}
    	if err := fn(tx); err != nil {
    		return err
    	}
    	return tx.Commit(ctx)
    }
    
    type School struct {
    	ID, UserID, MembershipID, YearID, TermID, GradeID, ClassID, StudentID string
    }
    
    func (d *DB) SeedSchool(t *testing.T, code string) School {
    	t.Helper()
    	ctx := context.Background()
    	var s School
    	queries := []struct {
    		sql  string
    		args []any
    		out  *string
    	}{
    		{"INSERT INTO schools(code,name) VALUES($1,$2) RETURNING id", []any{code, "Ikonex Demo School " + code}, &s.ID},
    		{"INSERT INTO users(phone_number,full_name) VALUES($1,'Synthetic Guardian') RETURNING id", []any{"+2547000000" + code}, &s.UserID},
    	}
    	for _, q := range queries {
    		if err := d.Admin.QueryRow(ctx, q.sql, q.args...).Scan(q.out); err != nil {
    			t.Fatal(err)
    		}
    	}
    	for _, q := range []struct {
    		sql  string
    		args func() []any
    		out  *string
    	}{
    		{"INSERT INTO memberships(school_id,user_id,role) VALUES($1,$2,'parent') RETURNING id", func() []any { return []any{s.ID, s.UserID} }, &s.MembershipID},
    		{"INSERT INTO academic_years(school_id,year,is_current) VALUES($1,2026,true) RETURNING id", func() []any { return []any{s.ID} }, &s.YearID},
    		{"INSERT INTO terms(school_id,academic_year_id,term_number,start_date,end_date,is_current) VALUES($1,$2,1,'2026-01-05','2026-04-10',true) RETURNING id", func() []any { return []any{s.ID, s.YearID} }, &s.TermID},
    		{"INSERT INTO grade_levels(school_id,number,phase,name) VALUES($1,7,'junior','Grade 7') RETURNING id", func() []any { return []any{s.ID} }, &s.GradeID},
    		{"INSERT INTO classes(school_id,grade_level_id,stream_name) VALUES($1,$2,'Demo') RETURNING id", func() []any { return []any{s.ID, s.GradeID} }, &s.ClassID},
    		{"INSERT INTO students(school_id,admission_number,full_name,dob,gender,current_class_id) VALUES($1,'DEMO-001','Synthetic Student','2013-01-01','female',$2) RETURNING id", func() []any { return []any{s.ID, s.ClassID} }, &s.StudentID},
    	} {
    		if err := d.Admin.QueryRow(ctx, q.sql, q.args()...).Scan(q.out); err != nil {
    			t.Fatal(err)
    		}
    	}
    	if _, err := d.Admin.Exec(ctx, "INSERT INTO student_guardians(school_id,student_id,user_id,relationship) VALUES($1,$2,$3,'guardian')", s.ID, s.StudentID, s.UserID); err != nil {
    		t.Fatal(err)
    	}
    	if _, err := d.Admin.Exec(ctx, "INSERT INTO enrolments(school_id,student_id,academic_year_id,grade_level_id,class_id) VALUES($1,$2,$3,$4,$5)", s.ID, s.StudentID, s.YearID, s.GradeID, s.ClassID); err != nil {
    		t.Fatal(err)
    	}
    	return s
    }

## Activity

    $ gofmt -w internal/billing/testenv/database.go internal/billing/migration_test.go && DEV05_TEST_DATABASE_URL=postgres://postgres@127.0.0.1:32768/postgres?sslmode=disable go test -modfile=/tmp/ikonex-dev05-testing.xunwQ5/testing.mod -v -race ./internal/billing -run TestMigration && go vet -modfile=/tmp/ikonex-dev05-testing.xunwQ5/testing.mod ./internal/billing/...
    go: could not create module cache: stat /home/bocal/go/pkg/mod: permission denied
    ✗ (1) • 10ms

## Activity

    $ gofmt -w internal/billing/testenv/database.go internal/billing/migration_test.go && DEV05_TEST_DATABASE_URL='postgres://postgres@127.0.0.1:32768/postgres?sslmode=disable' go test -modfile=/tmp/ikonex-dev05-testing.xunwQ5/testing.mod -v -race ./internal/billing -run TestMigration && go vet -modfile=/tmp/ikonex-dev05-testing.xunwQ5/testing.mod ./internal/billing/...
    === RUN   TestMigrationUpDownUpPreservesBaseline
    --- PASS: TestMigrationUpDownUpPreservesBaseline (5.95s)
    === RUN   TestMigrationTenantIsolationAndReferences
    --- PASS: TestMigrationTenantIsolationAndReferences (3.47s)
    PASS
    ok      github.com/Ashomondi/ikonex-academy/backend/internal/billing    10.530s
    ✓ • 23.36s

## Assistant

Migration verification passed: UP → DOWN → UP restored the baseline schema, constraints, indexes, and seeded student record. PostgreSQL also rejected cross-school writes and references, blank fee-change notes, and negative balances. The RLS tests ran under a role confirmed to have neither superuser nor `BYPASSRLS` privileges.

I’m recording that evidence and moving to transactional in-app notification creation.

## Activity

    file changes: Completed · 2 changes
    Update { move_path: None }: /home/fian/frihk/Ikonex-Academy/backend/internal/billing/IMPLEMENTATION_PLAN.md
    @@ -56,2 +56,4 @@
     
    +DB-01, DB-02, DB-03 completed: TestMigrationUpDownUpPreservesBaseline and TestMigrationTenantIsolationAndReferences passed against PostgreSQL 16.15 with race detection; owned billing/testenv vet passed. Baseline columns/constraints/indexes and synthetic student data survived DOWN unchanged; all twelve new tables have enabled/forced RLS and read/write policies; non-superuser isolation/reference/constraint checks passed. The extra owned tables persist STK attempts and the affected-student membership of fee-change notes. Full 000001–000004 integration still awaits the absent 000002/000003 migrations.
    +
     Paths below are relative to the repository root. `B`, `N`, `W`, `I`, `J` mean `backend/internal/billing`, `notifications`, `whatsapp`, `internalapi`, `jobs` respectively. `M` means `migrations/000004_r1_billing_whatsapp_notifications`. Tests live beside their implementation. Each row is one task: inspect, implement its ordered steps, gofmt only affected files, run verification, review the diff and ownership, then update status. Never advance a dependent task past a failed gate.
    Update { move_path: None }: /home/fian/frihk/Ikonex-Academy/backend/internal/billing/TESTING_RESOURCES.md
    @@ -87,2 +87,13 @@
     
    +## Migration verification
    +
    +```sh
    +DEV05_TEST_DATABASE_URL='postgres://postgres@127.0.0.1:32768/postgres?sslmode=disable' go test -modfile=/tmp/ikonex-dev05-testing.xunwQ5/testing.mod -v -race ./internal/billing -run TestMigration
    +go vet -modfile=/tmp/ikonex-dev05-testing.xunwQ5/testing.mod ./internal/billing/...
    +```
    +
    +Both migration tests passed (10.530 seconds including race instrumentation). Each test created its own randomly named synthetic database and non-superuser/NOBYPASSRLS role and removed both through test cleanup. UP/DOWN/UP, baseline schema/data preservation, RLS, cross-school reference rejection, missing-note checks and nonnegative balances were exercised. PostgreSQL/Redis versions remain separate from Go dependency versions. One initial invocation failed before testing because an alternate Bash login shell selected an inaccessible `/home/bocal/go/pkg/mod`; rerunning in the normal project shell succeeded without changing dependency versions or protected files.
    +
    +## Integration responsibility
    +
     Evans must add the recorded module requirements to the protected application manifest during integration. A successful test run with a temporary manifest is not evidence that the unchanged dependency-free application manifest can resolve newly imported pgx or Asynq packages. Report the exact manifest used with each result.

## Activity

    $ git status --short && git diff --name-only && git diff --cached --name-only && git diff --check && git add migrations/000004_r1_billing_whatsapp_notifications.up.sql migrations/000004_r1_billing_whatsapp_notifications.down.sql backend/internal/billing/migration_test.go backend/internal/billing/testenv/database.go backend/internal/billing/IMPLEMENTATION_PLAN.md backend/internal/billing/TESTING_RESOURCES.md && git diff --cached --name-only && git commit -m 'feat(billing): add tenant-isolated Release 1 migration'
     M Docs/DECISIONS.md
     M backend/internal/billing/IMPLEMENTATION_PLAN.md
     M backend/internal/billing/TESTING_RESOURCES.md
    ?? backend/internal/billing/migration_test.go
    ?? backend/internal/billing/testenv/
    ?? migrations/000004_r1_billing_whatsapp_notifications.down.sql
    ?? migrations/000004_r1_billing_whatsapp_notifications.up.sql
    Docs/DECISIONS.md
    backend/internal/billing/IMPLEMENTATION_PLAN.md
    backend/internal/billing/TESTING_RESOURCES.md
    backend/internal/billing/IMPLEMENTATION_PLAN.md
    backend/internal/billing/TESTING_RESOURCES.md
    backend/internal/billing/migration_test.go
    backend/internal/billing/testenv/database.go
    migrations/000004_r1_billing_whatsapp_notifications.down.sql
    migrations/000004_r1_billing_whatsapp_notifications.up.sql
    [feat/ian-backend-engineer-c 8988e8a] feat(billing): add tenant-isolated Release 1 migration
     6 files changed, 538 insertions(+)
     create mode 100644 backend/internal/billing/migration_test.go
     create mode 100644 backend/internal/billing/testenv/database.go
     create mode 100644 migrations/000004_r1_billing_whatsapp_notifications.down.sql
     create mode 100644 migrations/000004_r1_billing_whatsapp_notifications.up.sql
    ✓ • 121ms

## Activity

    file changes: Completed · 2 changes
    Update { move_path: None }: /home/fian/frihk/Ikonex-Academy/backend/internal/notifications/notifications.go
    @@ -2,3 +2,58 @@
     
    -// Service handles SMS dispatch (Africa's Talking), WhatsApp integration, and in-app notifications.
    -type Service struct{}
    +import (
    +	"context"
    +	"encoding/json"
    +	"strings"
    +	"unicode/utf8"
    +
    +	"github.com/Ashomondi/ikonex-academy/backend/internal/internalapi/contracts"
    +	"github.com/jackc/pgx/v5"
    +)
    +
    +type Service struct {
    +	DB contracts.TenantDB
    +}
    +
    +type CreateNotification struct {
    +	RecipientUserID string         `json:"recipient_user_id"`
    +	Category        string         `json:"category"`
    +	Title           string         `json:"title"`
    +	Body            string         `json:"body"`
    +	Metadata        map[string]any `json:"metadata"`
    +}
    +
    +// InAppNotifier participates in its caller's transaction: billing/calendar
    +// changes and their notices commit or roll back together. It never sends SMS.
    +type InAppNotifier interface {
    +	CreateInAppNotification(context.Context, pgx.Tx, string, CreateNotification) error
    +}
    +
    +func (s *Service) CreateInAppNotification(ctx context.Context, tx pgx.Tx, schoolID string, input CreateNotification) error {
    +	if !contracts.ValidID(schoolID) || !contracts.ValidID(input.RecipientUserID) ||
    +		strings.TrimSpace(input.Category) == "" || utf8.RuneCountInString(input.Category) > 50 ||
    +		strings.TrimSpace(input.Title) == "" || utf8.RuneCountInString(input.Title) > 255 ||
    +		strings.TrimSpace(input.Body) == "" {
    +		return contracts.Fail(422, "INVALID_NOTIFICATION", "A recipient, category, title and body are required.")
    +	}
    +	if input.Metadata == nil {
    +		input.Metadata = map[string]any{}
    +	}
    +	metadata, err := json.Marshal(input.Metadata)
    +	if err != nil {
    +		return contracts.Fail(422, "INVALID_NOTIFICATION", "Notification metadata must be valid JSON.")
    +	}
    +	var recipientExists bool
    +	err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM memberships m JOIN users u ON u.id=m.user_id
    +		WHERE m.school_id=$1 AND m.user_id=$2 AND m.status='active' AND u.is_active)`, schoolID, input.RecipientUserID).Scan(&recipientExists)
    +	if err != nil {
    +		return err
    +	}
    +	if !recipientExists {
    +		return contracts.Fail(422, "INVALID_NOTIFICATION_RECIPIENT", "The recipient is not active in this school.")
    +	}
    +	_, err = tx.Exec(ctx, `INSERT INTO in_app_notifications(school_id,recipient_user_id,category,title,body,metadata)
    +		VALUES($1,$2,$3,$4,$5,$6)`, schoolID, input.RecipientUserID, input.Category, input.Title, input.Body, metadata)
    +	return err
    +}
    +
    +var _ InAppNotifier = (*Service)(nil)
    Add: /home/fian/frihk/Ikonex-Academy/backend/internal/notifications/notifications_test.go
    package notifications_test
    
    import (
    	"context"
    	"errors"
    	"testing"
    
    	"github.com/Ashomondi/ikonex-academy/backend/internal/billing/testenv"
    	"github.com/Ashomondi/ikonex-academy/backend/internal/notifications"
    	"github.com/jackc/pgx/v5"
    )
    
    func TestCreateNotificationCommitsOrRollsBackWithCaller(t *testing.T) {
    	db := testenv.New(t, true)
    	a, b := db.SeedSchool(t, "01"), db.SeedSchool(t, "02")
    	s := &notifications.Service{DB: db}
    	ctx := context.Background()
    	input := notifications.CreateNotification{RecipientUserID: a.UserID, Category: "fee_change", Title: "Fee change", Body: "Synthetic fee notice", Metadata: map[string]any{"student_id": a.StudentID}}
    	if err := db.WithTenantTx(ctx, a.ID, func(tx pgx.Tx) error { return s.CreateInAppNotification(ctx, tx, a.ID, input) }); err != nil {
    		t.Fatal(err)
    	}
    	rollback := errors.New("caller transaction failed")
    	err := db.WithTenantTx(ctx, a.ID, func(tx pgx.Tx) error {
    		if err := s.CreateInAppNotification(ctx, tx, a.ID, input); err != nil {
    			return err
    		}
    		return rollback
    	})
    	if !errors.Is(err, rollback) {
    		t.Fatalf("caller rollback not propagated: %v", err)
    	}
    	var count int
    	if err := db.Admin.QueryRow(ctx, "SELECT count(*) FROM in_app_notifications WHERE school_id=$1 AND recipient_user_id=$2", a.ID, a.UserID).Scan(&count); err != nil || count != 1 {
    		t.Fatalf("partial or duplicate notice persisted: %d %v", count, err)
    	}
    	input.RecipientUserID = b.UserID
    	if err := db.WithTenantTx(ctx, a.ID, func(tx pgx.Tx) error { return s.CreateInAppNotification(ctx, tx, a.ID, input) }); err == nil {
    		t.Fatal("cross-school recipient accepted")
    	}
    	input.RecipientUserID = a.UserID
    	if err := db.WithTenantTx(ctx, b.ID, func(tx pgx.Tx) error { return s.CreateInAppNotification(ctx, tx, a.ID, input) }); err == nil {
    		t.Fatal("cross-school transaction accepted")
    	}
    	input.Title = " "
    	if err := s.CreateInAppNotification(ctx, nil, a.ID, input); err == nil {
    		t.Fatal("blank notification accepted")
    	}
    }

## Activity

    $ gofmt -w internal/notifications/notifications.go internal/notifications/notifications_test.go && DEV05_TEST_DATABASE_URL='postgres://postgres@127.0.0.1:32768/postgres?sslmode=disable' go test -modfile=/tmp/ikonex-dev05-testing.xunwQ5/testing.mod -v -race ./internal/notifications && go vet -modfile=/tmp/ikonex-dev05-testing.xunwQ5/testing.mod ./internal/notifications
    === RUN   TestAfricaTalkingRequestAndResponses
    === RUN   TestAfricaTalkingRequestAndResponses/sent
    === RUN   TestAfricaTalkingRequestAndResponses/queued
    === RUN   TestAfricaTalkingRequestAndResponses/processed
    === RUN   TestAfricaTalkingRequestAndResponses/invalid_phone
    === RUN   TestAfricaTalkingRequestAndResponses/gateway_failure
    === RUN   TestAfricaTalkingRequestAndResponses/wrong_recipient
    === RUN   TestAfricaTalkingRequestAndResponses/no_recipients
    === RUN   TestAfricaTalkingRequestAndResponses/multiple_recipients
    === RUN   TestAfricaTalkingRequestAndResponses/malformed
    === RUN   TestAfricaTalkingRequestAndResponses/oversized
    === RUN   TestAfricaTalkingRequestAndResponses/rate_limited
    === RUN   TestAfricaTalkingRequestAndResponses/request_timeout
    === RUN   TestAfricaTalkingRequestAndResponses/server_failure
    === RUN   TestAfricaTalkingRequestAndResponses/unauthorized
    --- PASS: TestAfricaTalkingRequestAndResponses (0.22s)
        --- PASS: TestAfricaTalkingRequestAndResponses/sent (0.02s)
        --- PASS: TestAfricaTalkingRequestAndResponses/queued (0.01s)
        --- PASS: TestAfricaTalkingRequestAndResponses/processed (0.01s)
        --- PASS: TestAfricaTalkingRequestAndResponses/invalid_phone (0.05s)
        --- PASS: TestAfricaTalkingRequestAndResponses/gateway_failure (0.01s)
        --- PASS: TestAfricaTalkingRequestAndResponses/wrong_recipient (0.02s)
        --- PASS: TestAfricaTalkingRequestAndResponses/no_recipients (0.01s)
        --- PASS: TestAfricaTalkingRequestAndResponses/multiple_recipients (0.02s)
        --- PASS: TestAfricaTalkingRequestAndResponses/malformed (0.01s)
        --- PASS: TestAfricaTalkingRequestAndResponses/oversized (0.02s)
        --- PASS: TestAfricaTalkingRequestAndResponses/rate_limited (0.01s)
        --- PASS: TestAfricaTalkingRequestAndResponses/request_timeout (0.01s)
        --- PASS: TestAfricaTalkingRequestAndResponses/server_failure (0.01s)
        --- PASS: TestAfricaTalkingRequestAndResponses/unauthorized (0.01s)
    === RUN   TestSMSFactorySandboxDoesNotUseHTTP
    --- PASS: TestSMSFactorySandboxDoesNotUseHTTP (0.00s)
    === RUN   TestAfricaTalkingRejectsRedirects
    --- PASS: TestAfricaTalkingRejectsRedirects (0.01s)
    === RUN   TestAfricaTalkingCancellationAndTimeout
    --- PASS: TestAfricaTalkingCancellationAndTimeout (0.03s)
    === RUN   TestSMSFactoryDoesNotMutateHTTPClient
    --- PASS: TestSMSFactoryDoesNotMutateHTTPClient (0.00s)
    === RUN   TestNormalizePhone
    === RUN   TestNormalizePhone/0712345678
    === RUN   TestNormalizePhone/0112345678
    === RUN   TestNormalizePhone/254712345678
    === RUN   TestNormalizePhone/_+254_(712)_345-678_
    === RUN   TestNormalizePhone/+442079460000
    === RUN   TestNormalizePhone/#00
    === RUN   TestNormalizePhone/+
    === RUN   TestNormalizePhone/071234567
    === RUN   TestNormalizePhone/+2547123456789
    === RUN   TestNormalizePhone/+254012345678
    === RUN   TestNormalizePhone/+254212345678
    === RUN   TestNormalizePhone/+0712345678
    === RUN   TestNormalizePhone/0712345678_apiKey=secret
    === RUN   TestNormalizePhone/+254７12345678
    === RUN   TestNormalizePhone/+254+712345678
    === RUN   TestNormalizePhone/712345678
    === RUN   TestNormalizePhone/+1234567890123456
    --- PASS: TestNormalizePhone (0.02s)
        --- PASS: TestNormalizePhone/0712345678 (0.00s)
        --- PASS: TestNormalizePhone/0112345678 (0.00s)
        --- PASS: TestNormalizePhone/254712345678 (0.00s)
        --- PASS: TestNormalizePhone/_+254_(712)_345-678_ (0.00s)
        --- PASS: TestNormalizePhone/+442079460000 (0.00s)
        --- PASS: TestNormalizePhone/#00 (0.00s)
        --- PASS: TestNormalizePhone/+ (0.00s)
        --- PASS: TestNormalizePhone/071234567 (0.00s)
        --- PASS: TestNormalizePhone/+2547123456789 (0.00s)
        --- PASS: TestNormalizePhone/+254012345678 (0.00s)
        --- PASS: TestNormalizePhone/+254212345678 (0.00s)
        --- PASS: TestNormalizePhone/+0712345678 (0.00s)
        --- PASS: TestNormalizePhone/0712345678_apiKey=secret (0.00s)
        --- PASS: TestNormalizePhone/+254７12345678 (0.00s)
        --- PASS: TestNormalizePhone/+254+712345678 (0.00s)
        --- PASS: TestNormalizePhone/712345678 (0.00s)
        --- PASS: TestNormalizePhone/+1234567890123456 (0.00s)
    === RUN   TestSandboxSMSRedactsSensitiveData
    --- PASS: TestSandboxSMSRedactsSensitiveData (0.00s)
    === RUN   TestSandboxSMSConcurrentDispatch
    --- PASS: TestSandboxSMSConcurrentDispatch (0.00s)
    === RUN   TestTransactionalSMSValidation
    --- PASS: TestTransactionalSMSValidation (0.00s)
    === RUN   TestCreateNotificationCommitsOrRollsBackWithCaller
    --- PASS: TestCreateNotificationCommitsOrRollsBackWithCaller (3.73s)
    PASS
    ok      github.com/Ashomondi/ikonex-academy/backend/internal/notifications    5.108s
    ✓ • 17.76s

## Assistant

Transactional notification creation passed its PostgreSQL tests: a notice commits with its caller, disappears when the caller rolls back, and rejects recipients or transactions from another school.

Next are the two notification routes: paginated listing and mark-as-read. Their tests will check user ownership, tenant isolation, and that repeated reads preserve the original `read_at` timestamp.

## Activity

    file changes: Completed · 2 changes
    Add: /home/fian/frihk/Ikonex-Academy/backend/internal/notifications/http.go
    package notifications
    
    import (
    	"context"
    	"encoding/json"
    	"net/http"
    	"time"
    
    	"github.com/Ashomondi/ikonex-academy/backend/internal/internalapi/contracts"
    	"github.com/jackc/pgx/v5"
    )
    
    type Notification struct {
    	ID        string          `json:"id"`
    	Category  string          `json:"category"`
    	Title     string          `json:"title"`
    	Body      string          `json:"body"`
    	Metadata  json.RawMessage `json:"metadata"`
    	IsRead    bool            `json:"is_read"`
    	ReadAt    *time.Time      `json:"read_at"`
    	CreatedAt time.Time       `json:"created_at"`
    }
    
    func (s *Service) List(ctx context.Context, p contracts.Principal, page *contracts.Page) ([]Notification, error) {
    	notices := []Notification{}
    	err := s.DB.WithTenantTx(ctx, p.SchoolID, func(tx pgx.Tx) error {
    		if err := tx.QueryRow(ctx, "SELECT count(*) FROM in_app_notifications WHERE school_id=$1 AND recipient_user_id=$2", p.SchoolID, p.UserID).Scan(&page.Total); err != nil {
    			return err
    		}
    		rows, err := tx.Query(ctx, `SELECT id,category,title,body,metadata,is_read,read_at,created_at
    			FROM in_app_notifications WHERE school_id=$1 AND recipient_user_id=$2
    			ORDER BY created_at DESC,id LIMIT $3 OFFSET $4`, p.SchoolID, p.UserID, page.PerPage, (page.Page-1)*page.PerPage)
    		if err != nil {
    			return err
    		}
    		defer rows.Close()
    		for rows.Next() {
    			var n Notification
    			if err := rows.Scan(&n.ID, &n.Category, &n.Title, &n.Body, &n.Metadata, &n.IsRead, &n.ReadAt, &n.CreatedAt); err != nil {
    				return err
    			}
    			n.CreatedAt = n.CreatedAt.UTC()
    			if n.ReadAt != nil {
    				readAt := n.ReadAt.UTC()
    				n.ReadAt = &readAt
    			}
    			notices = append(notices, n)
    		}
    		return rows.Err()
    	})
    	return notices, err
    }
    
    func (s *Service) MarkRead(ctx context.Context, p contracts.Principal, id string) error {
    	if !contracts.ValidID(id) {
    		return contracts.Fail(400, "INVALID_NOTIFICATION_ID", "A valid notification ID is required.")
    	}
    	return s.DB.WithTenantTx(ctx, p.SchoolID, func(tx pgx.Tx) error {
    		result, err := tx.Exec(ctx, `UPDATE in_app_notifications SET is_read=true,read_at=coalesce(read_at,now())
    			WHERE school_id=$1 AND recipient_user_id=$2 AND id=$3`, p.SchoolID, p.UserID, id)
    		if err != nil {
    			return err
    		}
    		if result.RowsAffected() != 1 {
    			return contracts.Fail(404, "NOTIFICATION_NOT_FOUND", "Notification not found.")
    		}
    		return nil
    	})
    }
    
    func RegisterRoutes(mux *http.ServeMux, service *Service, identity contracts.Identity) {
    	mux.HandleFunc("GET /api/v1/notifications", func(w http.ResponseWriter, r *http.Request) {
    		p, err := contracts.RequireIdentity(r, identity)
    		if err != nil {
    			contracts.WriteError(w, err)
    			return
    		}
    		page, err := contracts.Pagination(r)
    		if err != nil {
    			contracts.WriteError(w, err)
    			return
    		}
    		result, err := service.List(r.Context(), p, &page)
    		if err != nil {
    			contracts.WriteError(w, err)
    			return
    		}
    		contracts.Success(w, 200, result, &page)
    	})
    	mux.HandleFunc("PATCH /api/v1/notifications/{id}/read", func(w http.ResponseWriter, r *http.Request) {
    		p, err := contracts.RequireIdentity(r, identity)
    		if err == nil {
    			err = service.MarkRead(r.Context(), p, r.PathValue("id"))
    		}
    		if err != nil {
    			contracts.WriteError(w, err)
    			return
    		}
    		contracts.Success(w, 200, map[string]any{"id": r.PathValue("id"), "is_read": true}, nil)
    	})
    }
    Update { move_path: None }: /home/fian/frihk/Ikonex-Academy/backend/internal/notifications/notifications_test.go
    @@ -5,5 +5,10 @@
     	"errors"
    +	"encoding/json"
    +	"net/http"
    +	"net/http/httptest"
     	"testing"
    +	"time"
     
     	"github.com/Ashomondi/ikonex-academy/backend/internal/billing/testenv"
    +	"github.com/Ashomondi/ikonex-academy/backend/internal/internalapi/contracts"
     	"github.com/Ashomondi/ikonex-academy/backend/internal/notifications"
    @@ -48 +53,48 @@
     }
    +
    +func TestNotificationRoutesScopePaginationAndReadOwnership(t *testing.T) {
    +	db := testenv.New(t, true)
    +	a, b := db.SeedSchool(t, "01"), db.SeedSchool(t, "02")
    +	ctx := context.Background()
    +	var ownID, otherUserID, otherSchoolID string
    +	for _, row := range []struct {
    +		school, user string
    +		id *string
    +	}{{a.ID,a.UserID,&ownID},{a.ID,b.UserID,&otherUserID},{b.ID,b.UserID,&otherSchoolID}} {
    +		if err := db.Admin.QueryRow(ctx, "INSERT INTO in_app_notifications(school_id,recipient_user_id,category,title,body) VALUES($1,$2,'test','Synthetic','Synthetic') RETURNING id", row.school,row.user).Scan(row.id); err != nil {
    +			t.Fatal(err)
    +		}
    +	}
    +	s := &notifications.Service{DB:db}
    +	p := contracts.Principal{UserID:a.UserID,SchoolID:a.ID,MembershipID:a.MembershipID,Role:"parent"}
    +	mux := http.NewServeMux()
    +	notifications.RegisterRoutes(mux,s,func(*http.Request)(contracts.Principal,error){return p,nil})
    +	w := httptest.NewRecorder()
    +	mux.ServeHTTP(w,httptest.NewRequest("GET","/api/v1/notifications?per_page=1",nil))
    +	var result struct { Data []notifications.Notification `json:"data"`; Meta contracts.Page `json:"meta"` }
    +	if err := json.Unmarshal(w.Body.Bytes(),&result); err != nil || w.Code != 200 || len(result.Data)!=1 || result.Meta.Total!=1 || result.Data[0].ID!=ownID {
    +		t.Fatalf("list scoping/pagination: %d %s %v",w.Code,w.Body.String(),err)
    +	}
    +	for _, id := range []string{otherUserID,otherSchoolID} {
    +		w=httptest.NewRecorder()
    +		mux.ServeHTTP(w,httptest.NewRequest("PATCH","/api/v1/notifications/"+id+"/read",nil))
    +		if w.Code!=404 {t.Fatalf("another user's notice changed: %d",w.Code)}
    +	}
    +	var firstRead time.Time
    +	for i:=0;i<2;i++ {
    +		w=httptest.NewRecorder()
    +		mux.ServeHTTP(w,httptest.NewRequest("PATCH","/api/v1/notifications/"+ownID+"/read",nil))
    +		if w.Code!=200 {t.Fatal(w.Body.String())}
    +		var readAt time.Time
    +		if err:=db.Admin.QueryRow(ctx,"SELECT read_at FROM in_app_notifications WHERE school_id=$1 AND id=$2 AND is_read",a.ID,ownID).Scan(&readAt);err!=nil{t.Fatal(err)}
    +		if i==0 {firstRead=readAt} else if !readAt.Equal(firstRead){t.Fatal("repeat read replaced original timestamp")}
    +	}
    +	w=httptest.NewRecorder()
    +	mux.ServeHTTP(w,httptest.NewRequest("GET","/api/v1/notifications?page=2&per_page=1",nil))
    +	if err:=json.Unmarshal(w.Body.Bytes(),&result);err!=nil||len(result.Data)!=0||result.Meta.Total!=1{t.Fatal(w.Body.String())}
    +	unauthenticated:=http.NewServeMux()
    +	notifications.RegisterRoutes(unauthenticated,s,nil)
    +	w=httptest.NewRecorder()
    +	unauthenticated.ServeHTTP(w,httptest.NewRequest("GET","/api/v1/notifications",nil))
    +	if w.Code!=401{t.Fatal("unauthenticated list was allowed")}
    +}

## Activity

    $ gofmt -w internal/notifications/http.go internal/notifications/notifications_test.go && DEV05_TEST_DATABASE_URL='postgres://postgres@127.0.0.1:32768/postgres?sslmode=disable' go test -modfile=/tmp/ikonex-dev05-testing.xunwQ5/testing.mod -v -race ./internal/notifications && go vet -modfile=/tmp/ikonex-dev05-testing.xunwQ5/testing.mod ./internal/notifications
    === RUN   TestAfricaTalkingRequestAndResponses
    === RUN   TestAfricaTalkingRequestAndResponses/sent
    === RUN   TestAfricaTalkingRequestAndResponses/queued
    === RUN   TestAfricaTalkingRequestAndResponses/processed
    === RUN   TestAfricaTalkingRequestAndResponses/invalid_phone
    === RUN   TestAfricaTalkingRequestAndResponses/gateway_failure
    === RUN   TestAfricaTalkingRequestAndResponses/wrong_recipient
    === RUN   TestAfricaTalkingRequestAndResponses/no_recipients
    === RUN   TestAfricaTalkingRequestAndResponses/multiple_recipients
    === RUN   TestAfricaTalkingRequestAndResponses/malformed
    === RUN   TestAfricaTalkingRequestAndResponses/oversized
    === RUN   TestAfricaTalkingRequestAndResponses/rate_limited
    === RUN   TestAfricaTalkingRequestAndResponses/request_timeout
    === RUN   TestAfricaTalkingRequestAndResponses/server_failure
    === RUN   TestAfricaTalkingRequestAndResponses/unauthorized
    --- PASS: TestAfricaTalkingRequestAndResponses (0.24s)
        --- PASS: TestAfricaTalkingRequestAndResponses/sent (0.05s)
        --- PASS: TestAfricaTalkingRequestAndResponses/queued (0.02s)
        --- PASS: TestAfricaTalkingRequestAndResponses/processed (0.01s)
        --- PASS: TestAfricaTalkingRequestAndResponses/invalid_phone (0.01s)
        --- PASS: TestAfricaTalkingRequestAndResponses/gateway_failure (0.04s)
        --- PASS: TestAfricaTalkingRequestAndResponses/wrong_recipient (0.02s)
        --- PASS: TestAfricaTalkingRequestAndResponses/no_recipients (0.01s)
        --- PASS: TestAfricaTalkingRequestAndResponses/multiple_recipients (0.01s)
        --- PASS: TestAfricaTalkingRequestAndResponses/malformed (0.00s)
        --- PASS: TestAfricaTalkingRequestAndResponses/oversized (0.01s)
        --- PASS: TestAfricaTalkingRequestAndResponses/rate_limited (0.01s)
        --- PASS: TestAfricaTalkingRequestAndResponses/request_timeout (0.01s)
        --- PASS: TestAfricaTalkingRequestAndResponses/server_failure (0.00s)
        --- PASS: TestAfricaTalkingRequestAndResponses/unauthorized (0.00s)
    === RUN   TestSMSFactorySandboxDoesNotUseHTTP
    --- PASS: TestSMSFactorySandboxDoesNotUseHTTP (0.00s)
    === RUN   TestAfricaTalkingRejectsRedirects
    --- PASS: TestAfricaTalkingRejectsRedirects (0.01s)
    === RUN   TestAfricaTalkingCancellationAndTimeout
    --- PASS: TestAfricaTalkingCancellationAndTimeout (0.03s)
    === RUN   TestSMSFactoryDoesNotMutateHTTPClient
    --- PASS: TestSMSFactoryDoesNotMutateHTTPClient (0.00s)
    === RUN   TestNormalizePhone
    === RUN   TestNormalizePhone/0712345678
    === RUN   TestNormalizePhone/0112345678
    === RUN   TestNormalizePhone/254712345678
    === RUN   TestNormalizePhone/_+254_(712)_345-678_
    === RUN   TestNormalizePhone/+442079460000
    === RUN   TestNormalizePhone/#00
    === RUN   TestNormalizePhone/+
    === RUN   TestNormalizePhone/071234567
    === RUN   TestNormalizePhone/+2547123456789
    === RUN   TestNormalizePhone/+254012345678
    === RUN   TestNormalizePhone/+254212345678
    === RUN   TestNormalizePhone/+0712345678
    === RUN   TestNormalizePhone/0712345678_apiKey=secret
    === RUN   TestNormalizePhone/+254７12345678
    === RUN   TestNormalizePhone/+254+712345678
    === RUN   TestNormalizePhone/712345678
    === RUN   TestNormalizePhone/+1234567890123456
    --- PASS: TestNormalizePhone (0.06s)
        --- PASS: TestNormalizePhone/0712345678 (0.00s)
        --- PASS: TestNormalizePhone/0112345678 (0.00s)
        --- PASS: TestNormalizePhone/254712345678 (0.00s)
        --- PASS: TestNormalizePhone/_+254_(712)_345-678_ (0.00s)
        --- PASS: TestNormalizePhone/+442079460000 (0.00s)
        --- PASS: TestNormalizePhone/#00 (0.00s)
        --- PASS: TestNormalizePhone/+ (0.00s)
        --- PASS: TestNormalizePhone/071234567 (0.00s)
        --- PASS: TestNormalizePhone/+2547123456789 (0.00s)
        --- PASS: TestNormalizePhone/+254012345678 (0.00s)
        --- PASS: TestNormalizePhone/+254212345678 (0.00s)
        --- PASS: TestNormalizePhone/+0712345678 (0.00s)
        --- PASS: TestNormalizePhone/0712345678_apiKey=secret (0.00s)
        --- PASS: TestNormalizePhone/+254７12345678 (0.00s)
        --- PASS: TestNormalizePhone/+254+712345678 (0.00s)
        --- PASS: TestNormalizePhone/712345678 (0.00s)
        --- PASS: TestNormalizePhone/+1234567890123456 (0.00s)
    === RUN   TestSandboxSMSRedactsSensitiveData
    --- PASS: TestSandboxSMSRedactsSensitiveData (0.01s)
    === RUN   TestSandboxSMSConcurrentDispatch
    --- PASS: TestSandboxSMSConcurrentDispatch (0.02s)
    === RUN   TestTransactionalSMSValidation
    --- PASS: TestTransactionalSMSValidation (0.00s)
    === RUN   TestCreateNotificationCommitsOrRollsBackWithCaller
    --- PASS: TestCreateNotificationCommitsOrRollsBackWithCaller (4.00s)
    === RUN   TestNotificationRoutesScopePaginationAndReadOwnership
    --- PASS: TestNotificationRoutesScopePaginationAndReadOwnership (4.14s)
    PASS
    ok      github.com/Ashomondi/ikonex-academy/backend/internal/notifications    9.682s
    ✓ • 20.97s

## Assistant

Both notification routes passed their PostgreSQL and race tests. Listing is scoped to the authenticated user and school, other users’ notices cannot be marked read, pagination totals are correct, and repeated reads retain the first timestamp.

I’ll save the integration contract in the owned notification directory, then start the billing money type with exact decimal parsing and overflow tests.

## Activity

    file changes: Completed · 3 changes
    Update { move_path: None }: /home/fian/frihk/Ikonex-Academy/backend/internal/billing/IMPLEMENTATION_PLAN.md
    @@ -108,2 +108,4 @@
     
    +NOTIF-04 and NOTIF-05 completed: notification creation/rollback and HTTP ownership/pagination/read-idempotency tests passed against PostgreSQL 16.15, together with the complete SMS suite under the race detector; notification vet passed. Exact exported contracts and route shapes are documented in `../notifications/INTEGRATION.md` for Evans/Ema. The shared API documentation and server remain untouched.
    +
     | Criterion | Required coverage | Evidence | Status |
    Update { move_path: None }: /home/fian/frihk/Ikonex-Academy/backend/internal/billing/TESTING_RESOURCES.md
    @@ -85,4 +85,2 @@
     
    -## Integration responsibility
    -
     ## Migration verification
    @@ -98,2 +96,4 @@
     
    +NOTIF-04/05 verification: `DEV05_TEST_DATABASE_URL='postgres://postgres@127.0.0.1:32768/postgres?sslmode=disable' go test -modfile=/tmp/ikonex-dev05-testing.xunwQ5/testing.mod -v -race ./internal/notifications` passed (9.682 seconds), including real PostgreSQL transaction rollback, recipient/tenant isolation, HTTP pagination, and read timestamp idempotency. `go vet` with the same manifest passed. Test databases and roles were cleaned by the harness; protected files remained unchanged.
    +
     Evans must add the recorded module requirements to the protected application manifest during integration. A successful test run with a temporary manifest is not evidence that the unchanged dependency-free application manifest can resolve newly imported pgx or Asynq packages. Report the exact manifest used with each result.
    Add: /home/fian/frihk/Ikonex-Academy/backend/internal/notifications/INTEGRATION.md
    # Notifications integration contract
    
    The server owner mounts `notifications.RegisterRoutes(mux, service, identity)`. `identity` is a `contracts.Identity` supplied by trusted authentication middleware; caller headers must never be used directly as identity. `Service.DB` accepts the application's `WithTenantTx(ctx, schoolID, func(pgx.Tx) error) error` through `contracts.TenantDB`. No route is mounted in the shared server file by this change.
    
    `GET /api/v1/notifications?page=1&per_page=25` lists only the authenticated user's notices in the active school. Page size is limited to 100. Success has `status`, a `data` array, and `meta` with `page`, `per_page`, and `total`. Each notice contains `id`, `category`, `title`, `body`, `metadata`, `is_read`, nullable `read_at`, and UTC `created_at`.
    
    `PATCH /api/v1/notifications/{id}/read` accepts no required body and returns `{"status":"success","data":{"id":"...","is_read":true}}`. Repetition preserves the first `read_at`. Unknown notices, another user's notices, and another school's notices return 404 `NOTIFICATION_NOT_FOUND`; missing identity returns 401 `UNAUTHENTICATED`. Invalid pagination returns 422 `INVALID_PAGINATION`. Responses use the master plan's standard error envelope and `Cache-Control: no-store`.
    
    Billing and academics inject `InAppNotifier`:
    
    ```go
    CreateInAppNotification(ctx context.Context, tx pgx.Tx, schoolID string, input notifications.CreateNotification) error
    ```
    
    `CreateNotification` contains `RecipientUserID`, `Category`, `Title`, `Body`, and optional `Metadata`. Call it inside the caller's tenant transaction. It validates an active recipient membership in that school and participates in the caller's commit/rollback. It does not enqueue SMS or create its own transaction. Callers must deduplicate recipients for a domain event and avoid secrets in notice content/metadata.
    
    Auth, onboarding, WhatsApp and workers inject `SMSDispatcher.DispatchSMS(context.Context, SMSMessage) error`. A nil error means accepted by the dispatcher/provider, not proven handset delivery. The supported purposes are `app_login`, `whatsapp_session`, `phone_change`, and `phone_change_notice`. `NewSMSDispatcher` receives `SMSConfig` mapped by the application from Africa's Talking configuration. Username `sandbox` selects local redacted logging. `SMSDeliveryError.Retryable` is consumed by the later queue worker; ambiguous network failures cannot guarantee exactly-once SMS delivery.
    
    Verification uses the temporary dependencies recorded in `../billing/TESTING_RESOURCES.md`. Actual PostgreSQL tests cover rollback, tenant/user ownership, paging, and read idempotency. Local HTTP fixtures cover SMS provider behavior; no live SMS is claimed.

## Activity

    $ git status --short && git diff --name-only && git diff --cached --name-only && git diff --check && git add backend/internal/notifications/notifications.go backend/internal/notifications/notifications_test.go backend/internal/notifications/http.go backend/internal/notifications/INTEGRATION.md backend/internal/billing/IMPLEMENTATION_PLAN.md backend/internal/billing/TESTING_RESOURCES.md && git diff --cached --name-only && git commit -m 'feat(notifications): implement transactional in-app notifications'
     M Docs/DECISIONS.md
     M backend/internal/billing/IMPLEMENTATION_PLAN.md
     M backend/internal/billing/TESTING_RESOURCES.md
     M backend/internal/notifications/notifications.go
    ?? backend/internal/notifications/INTEGRATION.md
    ?? backend/internal/notifications/http.go
    ?? backend/internal/notifications/notifications_test.go
    Docs/DECISIONS.md
    backend/internal/billing/IMPLEMENTATION_PLAN.md
    backend/internal/billing/TESTING_RESOURCES.md
    backend/internal/notifications/notifications.go
    backend/internal/billing/IMPLEMENTATION_PLAN.md
    backend/internal/billing/TESTING_RESOURCES.md
    backend/internal/notifications/INTEGRATION.md
    backend/internal/notifications/http.go
    backend/internal/notifications/notifications.go
    backend/internal/notifications/notifications_test.go
    [feat/ian-backend-engineer-c 6c40594] feat(notifications): implement transactional in-app notifications
     6 files changed, 298 insertions(+), 4 deletions(-)
     create mode 100644 backend/internal/notifications/INTEGRATION.md
     create mode 100644 backend/internal/notifications/http.go
     create mode 100644 backend/internal/notifications/notifications_test.go
    ✓ • 892ms

## Activity

    file changes: Completed · 2 changes
    Add: /home/fian/frihk/Ikonex-Academy/backend/internal/billing/money.go
    package billing
    
    import (
    	"encoding/json"
    	"errors"
    	"fmt"
    	"math/big"
    	"strconv"
    	"strings"
    
    	"github.com/jackc/pgx/v5/pgtype"
    )
    
    // Money stores KES minor units with the exact range of NUMERIC(12,2).
    type Money int64
    
    const MaxMoney Money = 999999999999
    
    var ErrInvalidMoney = errors.New("amount must be an exact decimal within NUMERIC(12,2)")
    
    func ParseMoney(value string) (Money, error) {
    	if value == "" || len(value) > 14 {
    		return 0, ErrInvalidMoney
    	}
    	negative := value[0] == '-'
    	if negative {
    		value = value[1:]
    	}
    	parts := strings.Split(value, ".")
    	if len(parts) > 2 || parts[0] == "" {
    		return 0, ErrInvalidMoney
    	}
    	fraction := "00"
    	if len(parts) == 2 {
    		if len(parts[1]) < 1 || len(parts[1]) > 2 {
    			return 0, ErrInvalidMoney
    		}
    		fraction = parts[1] + strings.Repeat("0", 2-len(parts[1]))
    	}
    	digits := parts[0] + fraction
    	for _, digit := range digits {
    		if digit < '0' || digit > '9' {
    			return 0, ErrInvalidMoney
    		}
    	}
    	n, err := strconv.ParseInt(digits, 10, 64)
    	if err != nil || n > int64(MaxMoney) {
    		return 0, ErrInvalidMoney
    	}
    	if negative {
    		n = -n
    	}
    	return Money(n), nil
    }
    
    func (m Money) String() string {
    	n := int64(m)
    	sign := ""
    	if n < 0 {
    		sign, n = "-", -n
    	}
    	return fmt.Sprintf("%s%d.%02d", sign, n/100, n%100)
    }
    
    func (m Money) MarshalJSON() ([]byte, error) {
    	if m < -MaxMoney || m > MaxMoney {
    		return nil, ErrInvalidMoney
    	}
    	return json.Marshal(m.String())
    }
    
    func (m *Money) UnmarshalJSON(data []byte) error {
    	value := string(data)
    	if len(data) > 0 && data[0] == '"' {
    		if err := json.Unmarshal(data, &value); err != nil {
    			return ErrInvalidMoney
    		}
    	}
    	parsed, err := ParseMoney(value)
    	if err == nil {
    		*m = parsed
    	}
    	return err
    }
    
    func (m Money) NumericValue() (pgtype.Numeric, error) {
    	if m < -MaxMoney || m > MaxMoney {
    		return pgtype.Numeric{}, ErrInvalidMoney
    	}
    	return pgtype.Numeric{Int: big.NewInt(int64(m)), Exp: -2, Valid: true}, nil
    }
    
    func (m *Money) ScanNumeric(value pgtype.Numeric) error {
    	if !value.Valid || value.NaN || value.InfinityModifier != pgtype.Finite || value.Int == nil {
    		return ErrInvalidMoney
    	}
    	amount := new(big.Int).Set(value.Int)
    	exponent := int64(value.Exp) + 2
    	if amount.Sign() != 0 {
    		if exponent > 12 || exponent < -12 {
    			return ErrInvalidMoney
    		}
    		power := new(big.Int).Exp(big.NewInt(10), big.NewInt(abs(exponent)), nil)
    		if exponent >= 0 {
    			amount.Mul(amount, power)
    		} else {
    			var remainder big.Int
    			amount.QuoRem(amount, power, &remainder)
    			if remainder.Sign() != 0 {
    				return ErrInvalidMoney
    			}
    		}
    	}
    	if !amount.IsInt64() || amount.Int64() > int64(MaxMoney) || amount.Int64() < -int64(MaxMoney) {
    		return ErrInvalidMoney
    	}
    	*m = Money(amount.Int64())
    	return nil
    }
    
    func abs(n int64) int64 {
    	if n < 0 {
    		return -n
    	}
    	return n
    }
    
    func addMoney(a, b Money) (Money, error) {
    	if a < -MaxMoney || a > MaxMoney || b < -MaxMoney || b > MaxMoney || a+b < -MaxMoney || a+b > MaxMoney {
    		return 0, ErrInvalidMoney
    	}
    	return a + b, nil
    }
    
    func balances(opening, billed, paid Money) (current, credit Money, err error) {
    	if opening < 0 || billed < 0 || paid < 0 {
    		return 0, 0, ErrInvalidMoney
    	}
    	total, err := addMoney(opening, billed)
    	if err != nil {
    		return 0, 0, err
    	}
    	net, err := addMoney(total, -paid)
    	if err != nil {
    		return 0, 0, err
    	}
    	if net < 0 {
    		return 0, -net, nil
    	}
    	return net, 0, nil
    }
    Add: /home/fian/frihk/Ikonex-Academy/backend/internal/billing/money_test.go
    package billing
    
    import (
    	"context"
    	"encoding/json"
    	"math/big"
    	"testing"
    
    	"github.com/Ashomondi/ikonex-academy/backend/internal/billing/testenv"
    	"github.com/jackc/pgx/v5/pgtype"
    )
    
    func TestMoneyExactParsingAndJSON(t *testing.T) {
    	for _, tc := range []struct{ input, want string }{{"0.01", "0.01"}, {"1.2", "1.20"}, {"12500", "12500.00"}, {"-1.01", "-1.01"}, {"9999999999.99", "9999999999.99"}, {"-9999999999.99", "-9999999999.99"}, {"-0", "0.00"}} {
    		value, err := ParseMoney(tc.input)
    		if err != nil || value.String() != tc.want {
    			t.Fatalf("%s: %s %v", tc.input, value, err)
    		}
    		for _, input := range []string{tc.input, `"` + tc.input + `"`} {
    			var decoded Money
    			if err := json.Unmarshal([]byte(input), &decoded); err != nil || decoded != value {
    				t.Fatalf("JSON %s: %s %v", input, decoded, err)
    			}
    		}
    		encoded, err := json.Marshal(value)
    		if err != nil || string(encoded) != `"`+tc.want+`"` {
    			t.Fatalf("serialization: %s %v", encoded, err)
    		}
    	}
    	for _, input := range []string{"", "1.001", "1e2", "+1", " 1", "1.", ".1", "10000000000.00", "NaN", "-", "1..2", "１.００"} {
    		if _, err := ParseMoney(input); err == nil {
    			t.Fatalf("invalid money accepted: %q", input)
    		}
    	}
    	if _, err := addMoney(MaxMoney, 1); err == nil {
    		t.Fatal("overflow accepted")
    	}
    }
    
    func TestBalancesConserveMoney(t *testing.T) {
    	for _, tc := range []struct{ open, billed, paid, current, credit Money }{
    		{0, 10000, 5000, 5000, 0}, {0, 10000, 10000, 0, 0}, {0, 10000, 12500, 0, 2500}, {2000, 10000, 15000, 0, 3000}, {0, 0, 1, 0, 1},
    	} {
    		current, credit, err := balances(tc.open, tc.billed, tc.paid)
    		if err != nil || current != tc.current || credit != tc.credit || current-credit != tc.open+tc.billed-tc.paid {
    			t.Fatalf("balance %+v: %s %s %v", tc, current, credit, err)
    		}
    	}
    }
    
    func TestMoneyNumericRejectsRoundingAndSpecialValues(t *testing.T) {
    	for _, input := range []pgtype.Numeric{{}, {NaN: true, Valid: true}, {Int: big.NewInt(1001), Exp: -3, Valid: true}, {Int: big.NewInt(1), Exp: 20, Valid: true}} {
    		var money Money
    		if err := money.ScanNumeric(input); err == nil {
    			t.Fatal("inexact/invalid numeric accepted")
    		}
    	}
    }
    
    func TestMoneyPostgresRoundTrip(t *testing.T) {
    	db := testenv.New(t, false)
    	for _, expected := range []Money{0, 1, -1, 123456789, MaxMoney, -MaxMoney} {
    		var got Money
    		if err := db.Admin.QueryRow(context.Background(), "SELECT $1::numeric(12,2)", expected).Scan(&got); err != nil || got != expected {
    			t.Fatalf("NUMERIC round trip: want %s got %s: %v", expected, got, err)
    		}
    	}
    }

## Activity

    $ gofmt -w internal/billing/money.go internal/billing/money_test.go && DEV05_TEST_DATABASE_URL='postgres://postgres@127.0.0.1:32768/postgres?sslmode=disable' go test -modfile=/tmp/ikonex-dev05-testing.xunwQ5/testing.mod -v -race ./internal/billing -run 'TestMoney|TestBalances' && go vet -modfile=/tmp/ikonex-dev05-testing.xunwQ5/testing.mod ./internal/billing
    === RUN   TestMoneyExactParsingAndJSON
    --- PASS: TestMoneyExactParsingAndJSON (0.00s)
    === RUN   TestBalancesConserveMoney
    --- PASS: TestBalancesConserveMoney (0.00s)
    === RUN   TestMoneyNumericRejectsRoundingAndSpecialValues
    --- PASS: TestMoneyNumericRejectsRoundingAndSpecialValues (0.00s)
    === RUN   TestMoneyPostgresRoundTrip
    --- PASS: TestMoneyPostgresRoundTrip (2.36s)
    PASS
    ok      github.com/Ashomondi/ikonex-academy/backend/internal/billing    3.501s
    ✓ • 11.49s

## Assistant

The money type passed exact decimal, overflow, balance-conservation, and PostgreSQL round-trip tests, including one cent and the `NUMERIC(12,2)` limits.

The ledger design needs two additional counters: credit applied to a term and credit transferred out of a term. These preserve actual payment history without double-counting a transferred overpayment. I’ll add them to the owned migration and repeat its verification gate before implementing charging.

## Activity

    file changes: Completed · 5 changes
    Update { move_path: None }: /home/fian/frihk/Ikonex-Academy/backend/internal/billing/TESTING_RESOURCES.md
    @@ -96,2 +96,4 @@
     
    +BILL-01 passed with the temporary manifest: exact-money parsing/JSON, overflow, balance conservation, rejection of rounding/NaN, and PostgreSQL NUMERIC(12,2) round trips including ±9999999999.99. `go test -v -race ./internal/billing -run 'TestMoney|TestBalances'` used the local test database and passed in 3.501 seconds; billing vet passed. Per-term credit transfer will use separate `credit_applied`/`credit_transferred` counters, preserving actual `paid_amount` and preventing aggregate double-counting. The owned migration cycle is repeated after adding these columns.
    +
     NOTIF-04/05 verification: `DEV05_TEST_DATABASE_URL='postgres://postgres@127.0.0.1:32768/postgres?sslmode=disable' go test -modfile=/tmp/ikonex-dev05-testing.xunwQ5/testing.mod -v -race ./internal/notifications` passed (9.682 seconds), including real PostgreSQL transaction rollback, recipient/tenant isolation, HTTP pagination, and read timestamp idempotency. `go vet` with the same manifest passed. Test databases and roles were cleaned by the harness; protected files remained unchanged.
    Update { move_path: None }: /home/fian/frihk/Ikonex-Academy/backend/internal/internalapi/contracts/contracts.go
    @@ -69,2 +69,7 @@
     	}
    +	for _, index := range []int{8, 13, 18, 23} {
    +		if id[index] != '-' {
    +			return false
    +		}
    +	}
     	var value pgtype.UUID
    Update { move_path: None }: /home/fian/frihk/Ikonex-Academy/backend/internal/internalapi/contracts/contracts_test.go
    @@ -15,2 +15,13 @@
     
    +func TestIDRequiresCanonicalNonzeroUUID(t *testing.T) {
    +	for _, id := range []string{"", "00000000-0000-0000-0000-000000000000", "10000000x0000x0000x0000x000000000001", "10000000000000000000000000000001"} {
    +		if ValidID(id) {
    +			t.Fatalf("invalid ID accepted: %q", id)
    +		}
    +	}
    +	if !ValidID(syntheticID) {
    +		t.Fatal("canonical ID rejected")
    +	}
    +}
    +
     func TestIdentityRejectsUntrustedOrIncompleteContext(t *testing.T) {
    Update { move_path: None }: /home/fian/frihk/Ikonex-Academy/migrations/000004_r1_billing_whatsapp_notifications.down.sql
    @@ -16,3 +16,3 @@
     ALTER TABLE fee_ledgers DROP CONSTRAINT r1_c_ledger_nonnegative,
    -    DROP COLUMN credit_balance, DROP COLUMN due_date;
    +    DROP COLUMN credit_balance, DROP COLUMN credit_applied, DROP COLUMN credit_transferred, DROP COLUMN due_date;
     ALTER TABLE fee_structures DROP COLUMN boarding_status;
    Update { move_path: None }: /home/fian/frihk/Ikonex-Academy/migrations/000004_r1_billing_whatsapp_notifications.up.sql
    @@ -12,4 +12,6 @@
     ALTER TABLE fee_ledgers ADD COLUMN credit_balance NUMERIC(12,2) NOT NULL DEFAULT 0.00,
    +    ADD COLUMN credit_applied NUMERIC(12,2) NOT NULL DEFAULT 0.00,
    +    ADD COLUMN credit_transferred NUMERIC(12,2) NOT NULL DEFAULT 0.00,
         ADD COLUMN due_date DATE,
    -    ADD CONSTRAINT r1_c_ledger_nonnegative CHECK (current_balance >= 0 AND credit_balance >= 0);
    +    ADD CONSTRAINT r1_c_ledger_nonnegative CHECK (current_balance >= 0 AND credit_balance >= 0 AND credit_applied >= 0 AND credit_transferred >= 0);
     

## Activity

    $ gofmt -w internal/internalapi/contracts/contracts.go internal/internalapi/contracts/contracts_test.go && DEV05_TEST_DATABASE_URL='postgres://postgres@127.0.0.1:32768/postgres?sslmode=disable' go test -modfile=/tmp/ikonex-dev05-testing.xunwQ5/testing.mod -v -race ./internal/billing ./internal/internalapi/contracts && go vet -modfile=/tmp/ikonex-dev05-testing.xunwQ5/testing.mod ./internal/billing/... ./internal/internalapi/contracts
    === RUN   TestMoneyExactParsingAndJSON
    --- PASS: TestMoneyExactParsingAndJSON (0.00s)
    === RUN   TestBalancesConserveMoney
    --- PASS: TestBalancesConserveMoney (0.00s)
    === RUN   TestMoneyNumericRejectsRoundingAndSpecialValues
    --- PASS: TestMoneyNumericRejectsRoundingAndSpecialValues (0.00s)
    === RUN   TestMoneyPostgresRoundTrip
    --- PASS: TestMoneyPostgresRoundTrip (2.04s)
    === RUN   TestMigrationUpDownUpPreservesBaseline
    --- PASS: TestMigrationUpDownUpPreservesBaseline (5.98s)
    === RUN   TestMigrationTenantIsolationAndReferences
    --- PASS: TestMigrationTenantIsolationAndReferences (3.65s)
    PASS
    ok      github.com/Ashomondi/ikonex-academy/backend/internal/billing    12.791s
    === RUN   TestIDRequiresCanonicalNonzeroUUID
    --- PASS: TestIDRequiresCanonicalNonzeroUUID (0.00s)
    === RUN   TestIdentityRejectsUntrustedOrIncompleteContext
    --- PASS: TestIdentityRejectsUntrustedOrIncompleteContext (0.00s)
    === RUN   TestTenantAdapterPropagatesTransactionFailure
    --- PASS: TestTenantAdapterPropagatesTransactionFailure (0.00s)
    === RUN   TestResponseEnvelopesAndSafeErrors
    --- PASS: TestResponseEnvelopesAndSafeErrors (0.00s)
    === RUN   TestDecodeAndPagination
    --- PASS: TestDecodeAndPagination (0.02s)
    PASS
    ok      github.com/Ashomondi/ikonex-academy/backend/internal/internalapi/contracts    1.132s
    ✓ • 19.74s

## User

continue

## Assistant

Continuing. The migration recheck is running now. Once it passes, I’ll implement fee-structure validation and loading, followed by idempotent student charging, using the temporary dependency manifest and recording each test gate.

## User

continue

## User

continue

## Assistant

Continuing with fee-structure validation and loading. I’ll first check the working tree after the interruption, then implement and test the next billing task before moving to student charging.

## Activity

    $ git status --short && git branch --show-current && git rev-parse HEAD && git diff --cached --name-only && git diff --stat && git diff -- backend/go.mod backend/go.sum && cat backend/internal/billing/billing.go && sed -n '1,240p' backend/internal/billing/money.go
     M Docs/DECISIONS.md
     M backend/internal/billing/TESTING_RESOURCES.md
     M backend/internal/internalapi/contracts/contracts.go
     M backend/internal/internalapi/contracts/contracts_test.go
     M migrations/000004_r1_billing_whatsapp_notifications.down.sql
     M migrations/000004_r1_billing_whatsapp_notifications.up.sql
    ?? backend/internal/billing/money.go
    ?? backend/internal/billing/money_test.go
    feat/ian-backend-engineer-c
    6c4059466a3da2ee0177d9350c4bb0927d2625b3
     Docs/DECISIONS.md                                            |  1 +
     backend/internal/billing/TESTING_RESOURCES.md                |  2 ++
     backend/internal/internalapi/contracts/contracts.go          |  5 +++++
     backend/internal/internalapi/contracts/contracts_test.go     | 11 +++++++++++
     migrations/000004_r1_billing_whatsapp_notifications.down.sql |  2 +-
     migrations/000004_r1_billing_whatsapp_notifications.up.sql   |  4 +++-
     6 files changed, 23 insertions(+), 2 deletions(-)
    package billing
    
    // Service handles fee structures, student ledgers, M-Pesa STK push payments, and reconciliation.
    type Service struct{}
    package billing
    
    import (
        "encoding/json"
        "errors"
        "fmt"
        "math/big"
        "strconv"
        "strings"
    
        "github.com/jackc/pgx/v5/pgtype"
    )
    
    // Money stores KES minor units with the exact range of NUMERIC(12,2).
    type Money int64
    
    const MaxMoney Money = 999999999999
    
    var ErrInvalidMoney = errors.New("amount must be an exact decimal within NUMERIC(12,2)")
    
    func ParseMoney(value string) (Money, error) {
        if value == "" || len(value) > 14 {
            return 0, ErrInvalidMoney
        }
        negative := value[0] == '-'
        if negative {
            value = value[1:]
        }
        parts := strings.Split(value, ".")
        if len(parts) > 2 || parts[0] == "" {
            return 0, ErrInvalidMoney
        }
        fraction := "00"
        if len(parts) == 2 {
            if len(parts[1]) < 1 || len(parts[1]) > 2 {
                return 0, ErrInvalidMoney
            }
            fraction = parts[1] + strings.Repeat("0", 2-len(parts[1]))
        }
        digits := parts[0] + fraction
        for _, digit := range digits {
            if digit < '0' || digit > '9' {
                return 0, ErrInvalidMoney
            }
        }
        n, err := strconv.ParseInt(digits, 10, 64)
        if err != nil || n > int64(MaxMoney) {
            return 0, ErrInvalidMoney
        }
        if negative {
            n = -n
        }
        return Money(n), nil
    }
    
    func (m Money) String() string {
        n := int64(m)
        sign := ""
        if n < 0 {
            sign, n = "-", -n
        }
        return fmt.Sprintf("%s%d.%02d", sign, n/100, n%100)
    }
    
    func (m Money) MarshalJSON() ([]byte, error) {
        if m < -MaxMoney || m > MaxMoney {
            return nil, ErrInvalidMoney
        }
        return json.Marshal(m.String())
    }
    
    func (m *Money) UnmarshalJSON(data []byte) error {
        value := string(data)
        if len(data) > 0 && data[0] == '"' {
            if err := json.Unmarshal(data, &value); err != nil {
                return ErrInvalidMoney
            }
        }
        parsed, err := ParseMoney(value)
        if err == nil {
            *m = parsed
        }
        return err
    }
    
    func (m Money) NumericValue() (pgtype.Numeric, error) {
        if m < -MaxMoney || m > MaxMoney {
            return pgtype.Numeric{}, ErrInvalidMoney
        }
        return pgtype.Numeric{Int: big.NewInt(int64(m)), Exp: -2, Valid: true}, nil
    }
    
    func (m *Money) ScanNumeric(value pgtype.Numeric) error {
        if !value.Valid || value.NaN || value.InfinityModifier != pgtype.Finite || value.Int == nil {
            return ErrInvalidMoney
        }
        amount := new(big.Int).Set(value.Int)
        exponent := int64(value.Exp) + 2
        if amount.Sign() != 0 {
            if exponent > 12 || exponent < -12 {
                return ErrInvalidMoney
            }
            power := new(big.Int).Exp(big.NewInt(10), big.NewInt(abs(exponent)), nil)
            if exponent >= 0 {
                amount.Mul(amount, power)
            } else {
                var remainder big.Int
                amount.QuoRem(amount, power, &remainder)
                if remainder.Sign() != 0 {
                    return ErrInvalidMoney
                }
            }
        }
        if !amount.IsInt64() || amount.Int64() > int64(MaxMoney) || amount.Int64() < -int64(MaxMoney) {
            return ErrInvalidMoney
        }
        *m = Money(amount.Int64())
        return nil
    }
    
    func abs(n int64) int64 {
        if n < 0 {
            return -n
        }
        return n
    }
    
    func addMoney(a, b Money) (Money, error) {
        if a < -MaxMoney || a > MaxMoney || b < -MaxMoney || b > MaxMoney || a+b < -MaxMoney || a+b > MaxMoney {
            return 0, ErrInvalidMoney
        }
        return a + b, nil
    }
    
    func balances(opening, billed, paid Money) (current, credit Money, err error) {
        if opening < 0 || billed < 0 || paid < 0 {
            return 0, 0, ErrInvalidMoney
        }
        total, err := addMoney(opening, billed)
        if err != nil {
            return 0, 0, err
        }
        net, err := addMoney(total, -paid)
        if err != nil {
            return 0, 0, err
        }
        if net < 0 {
            return 0, -net, nil
        }
        return net, 0, nil
    }
    ✓ • 1.05s

## Activity

    $ test -f /tmp/ikonex-dev05-testing.xunwQ5/testing.mod && docker port ikonex-dev05-postgres 5432/tcp && tail -70 backend/internal/billing/TESTING_RESOURCES.md
    127.0.0.1:32768
    Record the resolved transitive module versions, exact image digests, container names and exposed loopback ports, commands, test results, and cleanup verification below as they become available. Never record live credentials, OTPs, full personal records, or service tokens.
    
    ## Reproduction and resolved module graph
    
    The current temporary manifest is `/tmp/ikonex-dev05-testing.xunwQ5/testing.mod` (with its companion testing.sum). It was copied from the unchanged application manifest. From `backend/`:
    
    ```sh
    go get -modfile=/tmp/ikonex-dev05-testing.xunwQ5/testing.mod github.com/jackc/pgx/v5/pgxpool@v5.7.2 github.com/hibiken/asynq@v0.25.1
    go list -modfile=/tmp/ikonex-dev05-testing.xunwQ5/testing.mod -m all
    go mod verify -modfile=/tmp/ikonex-dev05-testing.xunwQ5/testing.mod
    ```
    
    Checksum verification returned `all modules verified`. The resolved module graph (including dependencies' own test/tool requirements, which are not all compiled into this application) is:
    
    ```text
    github.com/bsm/ginkgo/v2 v2.12.0
    github.com/bsm/gomega v1.27.10
    github.com/cespare/xxhash/v2 v2.2.0
    github.com/davecgh/go-spew v1.1.1
    github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f
    github.com/frankban/quicktest v1.14.6
    github.com/golang/protobuf v1.5.0
    github.com/google/go-cmp v0.6.0
    github.com/google/uuid v1.6.0
    github.com/hibiken/asynq v0.25.1
    github.com/jackc/pgpassfile v1.0.0
    github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761
    github.com/jackc/pgx/v5 v5.7.2
    github.com/jackc/puddle/v2 v2.2.2
    github.com/kr/pretty v0.3.1
    github.com/kr/text v0.2.0
    github.com/pmezard/go-difflib v1.0.0
    github.com/redis/go-redis/v9 v9.7.0
    github.com/robfig/cron/v3 v3.0.1
    github.com/rogpeppe/go-internal v1.9.0
    github.com/spf13/cast v1.7.0
    github.com/stretchr/objx v0.1.0
    github.com/stretchr/testify v1.8.1
    go.uber.org/goleak v1.3.0
    golang.org/x/crypto v0.31.0
    golang.org/x/mod v0.17.0
    golang.org/x/net v0.21.0
    golang.org/x/sync v0.10.0
    golang.org/x/sys v0.28.0
    golang.org/x/term v0.27.0
    golang.org/x/text v0.21.0
    golang.org/x/time v0.8.0
    golang.org/x/tools v0.21.1-0.20240508182429-e35e4ccd0d2d
    golang.org/x/xerrors v0.0.0-20191204190536-9bdfabe68543
    google.golang.org/protobuf v1.35.2
    gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c
    gopkg.in/yaml.v3 v3.0.1
    ```
    
    ## Migration verification
    
    ```sh
    DEV05_TEST_DATABASE_URL='postgres://postgres@127.0.0.1:32768/postgres?sslmode=disable' go test -modfile=/tmp/ikonex-dev05-testing.xunwQ5/testing.mod -v -race ./internal/billing -run TestMigration
    go vet -modfile=/tmp/ikonex-dev05-testing.xunwQ5/testing.mod ./internal/billing/...
    ```
    
    Both migration tests passed (10.530 seconds including race instrumentation). Each test created its own randomly named synthetic database and non-superuser/NOBYPASSRLS role and removed both through test cleanup. UP/DOWN/UP, baseline schema/data preservation, RLS, cross-school reference rejection, missing-note checks and nonnegative balances were exercised. PostgreSQL/Redis versions remain separate from Go dependency versions. One initial invocation failed before testing because an alternate Bash login shell selected an inaccessible `/home/bocal/go/pkg/mod`; rerunning in the normal project shell succeeded without changing dependency versions or protected files.
    
    ## Integration responsibility
    
    BILL-01 passed with the temporary manifest: exact-money parsing/JSON, overflow, balance conservation, rejection of rounding/NaN, and PostgreSQL NUMERIC(12,2) round trips including ±9999999999.99. `go test -v -race ./internal/billing -run 'TestMoney|TestBalances'` used the local test database and passed in 3.501 seconds; billing vet passed. Per-term credit transfer will use separate `credit_applied`/`credit_transferred` counters, preserving actual `paid_amount` and preventing aggregate double-counting. The owned migration cycle is repeated after adding these columns.
    
    NOTIF-04/05 verification: `DEV05_TEST_DATABASE_URL='postgres://postgres@127.0.0.1:32768/postgres?sslmode=disable' go test -modfile=/tmp/ikonex-dev05-testing.xunwQ5/testing.mod -v -race ./internal/notifications` passed (9.682 seconds), including real PostgreSQL transaction rollback, recipient/tenant isolation, HTTP pagination, and read timestamp idempotency. `go vet` with the same manifest passed. Test databases and roles were cleaned by the harness; protected files remained unchanged.
    
    Evans must add the recorded module requirements to the protected application manifest during integration. A successful test run with a temporary manifest is not evidence that the unchanged dependency-free application manifest can resolve newly imported pgx or Asynq packages. Report the exact manifest used with each result.
    ✓ • 67ms
