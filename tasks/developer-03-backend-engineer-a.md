# Ikonex Academy – Release 1 Task Assignment
## Developer 3: Backend Engineer A (**Evans** — @eojuma)

- **Developer Name:** Evans (@eojuma)
- **GitHub Username:** `@eojuma`
- **Role:** Backend Engineer A (Go — Platform Foundation, Multi-Tenant RLS, Auth & Badges, School Setup Wizard, Staff & Parent QR Onboarding, Verification & Student Admission)
- **Primary Directory Ownership:** `backend/cmd/server/`, `backend/internal/config/`, `backend/internal/database/`, `backend/internal/middleware/`, `backend/internal/auth/`, `backend/internal/onboarding/`, `backend/internal/consent/`, and `migrations/000001_init_schema.*` + `migrations/000002_r1_core_onboarding.*`
- **Release Scope:** Release 1 (Core Multi-Tenant Foundation, SMS OTP Login, Scoped Badge Engine, Audit Trail, Setup Wizard, Staff/Parent QR Registration, In-Person Document Verification & Auto-Purge, Verifier Delegation, Dual-Route Student Admission, and Guardian 2FA Phone Change)

---

## 1. Role and Objective

As **Backend Engineer A (Evans — @eojuma)**, you own the foundational spine of the Go modular monolith and the entire **Identity, Permissions, and Onboarding** lifecycle. Notably, your scope directly covers **all three Release 1 Pass Conditions (`Docs/Ikonex.md` C53)**:
1. Onboarding works end-to-end (school and principal creation, setup wizard, staff & parent QR registration, in-person verification, and delegation).
2. Parents and staff can log in reliably via Phone Number + SMS OTP.
3. Roles and scoped/time-bound Badges (`class_teacher`, `delegated_verifier`, etc.) work reliably and stress-free across every school tenant.

Additionally, as the owner of `backend/internal/database`, `backend/internal/middleware`, and `backend/cmd/server/main.go`, you provide the shared **PostgreSQL Row-Level Security (RLS) transaction helper**, **authentication/badge middleware**, **standard JSON response helper**, and **audit logger** that Backend Engineers B (**Ema** — @ejao-000) and C (**Ian** — @frihk) plug their domain modules into.

---

## 2. Assigned Features

You own **3 cohesive, end-to-end backend features**:

| Feature ID | Feature Name | Owned Go Packages | Why These Belong Together |
|---|---|---|---|
| **BE-A1** | **Multi-Tenant RLS Database Layer, Phone SMS OTP Auth, Scoped Badge Permission Engine & Audit Trail** | `internal/config`, `internal/database`, `internal/middleware`, `internal/auth`, `internal/consent`, `cmd/server` | Authentication, tenant context (`SET LOCAL app.school_id`), badge evaluation, and audit logging form the request pipeline executed on every HTTP call before domain handlers run. |
| **BE-A2** | **Ikonex School Provisioning, Principal Setup Wizard & Staff QR Onboarding** | `internal/onboarding` (`school_wizard.go`, `staff_onboarding.go`) | Covers school bootstrap from Ikonex team creation through the Principal's initial setup wizard and staff QR recruitment/in-person document verification. |
| **BE-A3** | **Parent 12h QR Registration, Verifier Delegation, Dual-Route Student Admission & Guardian 2FA Phone Change** | `internal/onboarding` (`parent_qr.go`, `verification.go`, `students.go`, `guardian_phone.go`) | Manages parent and student entry into the school, sharing QR token validation, pending submission isolation, in-person verification, document auto-deletion, and guardian identity management. |

---

## 3. Detailed Feature Specifications

### Feature BE-A1: Multi-Tenant RLS Database Layer, Phone SMS OTP Auth, Scoped Badge Engine & Audit Trail
1. **PostgreSQL Row-Level Security (RLS) Enforcement (`Docs/ARCHITECTURE.md`, `AC-27`):**
   - Every tenant-scoped database query must run inside a transaction that executes `SELECT set_config('app.school_id', $1, true)` (equivalent to `SET LOCAL app.school_id = '<uuid>'`).
   - If `app.school_id` is unset, PostgreSQL's `FORCE ROW LEVEL SECURITY` policies return zero rows (`AC-27`).
   - Provide `db.WithTenantTx(ctx, schoolID, func(tx pgx.Tx) error)` in `backend/internal/database` so Ema (@ejao-000) and Ian (@frihk) never have to write raw session-variable boilerplate.
2. **Phone + SMS OTP Authentication (`Docs/Ikonex.md` 6.4):**
   - Users are global in `users` (one login per human phone number); school roles live in `memberships` (`DECISIONS.md`).
   - `POST /api/v1/auth/otp/request`: Validates Kenyan phone number (`+254...`), enforces rate limits (`1 OTP request per minute` per phone number; `> 5 requests in 30 minutes` blocks the requesting phone/IP for 24 hours—see `Docs/Ikonex.md` C18–C20, O20), generates a 6-digit cryptographic OTP, stores its SHA-256 hash in `otp_requests` (5-minute expiry), and enqueues a transactional SMS via Ian's (@frihk) `notifications.SMSDispatcher` interface.
   - `POST /api/v1/auth/otp/verify`: Verifies the OTP, issues a signed JWT (`HS256` with `JWT_SECRET`, expiry `JWT_EXPIRY_HOURS`), and returns the user profile along with all active school `memberships` and their active `badges`.
3. **Roles vs. Scoped Badges Permission Engine (`Docs/Ikonex.md` Section 4, `C53`):**
   - Never hard-code `"class_teacher"` or `"delegated_verifier"` as a column on `classes` or `memberships`.
   - Middleware `RequirePermission(permCode string, scopeExtractor ...ScopeFunc)` checks:
     1. Role-based default permissions for the caller's active `membership.role` in `X-School-ID`.
     2. Active rows in `membership_badges` joined to `badges` and `badge_permissions` where `(valid_from IS NULL OR valid_from <= NOW()) AND (valid_until IS NULL OR valid_until > NOW())` and optional `scope_id` matches the target resource (e.g., `class_id`).
   - Rejects unauthorized requests with `403 Forbidden` **before** reaching handler logic (`Docs/API_CONTRACTS.md`).
4. **Centralized Audit Trail (`Docs/Ikonex.md` Section 5 Rule 6, `AC-29`):**
   - Expose `audit.Log(ctx, tx, AuditEvent{...})` writing to `audit_log` for: QR creation/revocation, registration verification/rejection, badge delegation, marks reassignment, results release/reopen/re-release, fee changes, term-date changes, and guardian phone number changes.
5. **Consent Centre Schema (`Docs/Ikonex.md` 6.10):**
   - Update `consents.purpose` constraint in `000002_r1_core_onboarding.up.sql` to allow `('biometric_attendance', 'photo', 'sms', 'whatsapp')` (`C22`) and provide `GET/POST /api/v1/consents` for synthetic testing in Release 1.

### Feature BE-A2: Ikonex School Provisioning, Principal Setup Wizard & Staff QR Onboarding
1. **Ikonex Team School & Principal Creation (`Docs/Ikonex.md` 6.1, `C37`, `C57`, `AC-01`):**
   - `POST /api/v1/admin/schools`: Creates `schools` row, `school_profile` row (with required contact phone, email, office hours per `C57`), global `users` row for the Principal, `memberships` row (`role = 'principal'`), and seeds default school `badges` (`class_teacher`, `delegated_verifier`) and `badge_permissions`.
2. **Principal Setup Wizard (`Docs/Ikonex.md` 6.1, `C38`, `C43`, `C60`):**
   - `POST /api/v1/schools/setup-wizard`: Accessible to `principal` (and `deputy`). Atomically creates:
     - Updated `school_profile` contact details (`C57`).
     - `academic_years` and `terms` (Terms 1–3 with `start_date` and `end_date`; sets `wizard_completed = true` on `school_profile`).
     - `grade_levels` (constrained to CBC Grades `7..12`, `phase IN ('junior', 'senior')`).
     - `classes` (streams linked to `grade_level_id`) and `dormitories`.
     - Initial `grading_scale_versions` row (`EE`, `ME`, `AE`, `BE` score thresholds) locked to the active term (`C43`).
3. **Staff QR Onboarding & In-Person Document Verification (`Docs/Ikonex.md` 6.2, `C62`, `C63`, `AC-02`, `AC-06`):**
   - `POST /api/v1/staff-invites`: Principal/Deputy generates a Staff QR code (`staff_qr_codes` / `staff_invites`), default valid window `08:00–18:00` (adjustable by Principal/Deputy), revocable at any time (`POST /api/v1/staff-invites/{id}/revoke`).
   - Scanning an expired or revoked Staff QR is refused with `410 Gone` (`AC-02`).

### Feature BE-A3: Parent 12h QR Registration, Verifier Delegation, Student Admission & Guardian 2FA Phone Change
1. **Parent 12-Hour Multi-Use QR Code (`Docs/Ikonex.md` 6.3, `C26`, `C28`, `AC-03`):**
   - `POST /api/v1/parent-qr`: Principal, Deputy, or Dean of Studies generates a `parent_qr_codes` record with `valid_until = NOW() + INTERVAL '12 hours'`. One QR serves many parents.
   - `POST /api/v1/parent-qr/{id}/revoke`: Immediately revokes the QR code and invalidates any delegated verifier badges tied to that QR (`AC-03`, `AC-05`).
2. **Pending Submissions & In-Person Verification (`C28`, `C63`, `C64`, `AC-04`, `AC-06`, `AC-07`):**
   - `POST /api/v1/registrations`: Public endpoint accepting a valid `qr_token` (staff or parent), applicant details, and uploaded verification document metadata/file. Creates a row in `registration_submissions` with `status = 'pending'`.
   - **Strict Isolation (`AC-04`):** A pending submission creates **no** active `memberships` or `student_guardians` row and grants **zero** access to any child's data.
   - `POST /api/v1/verifier-grants`: Principal, Deputy, or Dean delegates the `delegated_verifier` badge (`membership_badges`) to an active **staff** member (`role != 'parent'`), with `valid_until` set to the active Parent QR's `valid_until` (`<= 12 hours`, `AC-05`). Attempting to grant to a parent returns `422 Unprocessable Entity`.
   - `POST /api/v1/registrations/{id}/verify`: Callable by Principal, Deputy, Dean, or a staff member holding an active `delegated_verifier` badge.
     - On `action = "approve"`: Creates/links `users`, `memberships`, `students`, `enrolments`, and `student_guardians` (and triggers Ian's (@frihk) `billing.EnsureStudentTermCharges` hook so newly admitted students receive term fee charges), **immediately deletes the uploaded document file/blob** (`document_storage_path = NULL, document_deleted_at = NOW()`), writes a permanent row to `verification_records` (`verifier_user_id`, `verified_at`, `document_type`), and logs to `audit_log` (`AC-06`).
     - On `action = "reject"`: Marks `status = 'rejected'`, deletes uploaded document, and schedules automatic purge of the rejected/unverified submission after `UNVERIFIED_SUBMISSION_RETENTION_DAYS` (default `7` days via env var, addressing `O17` / `AC-07`).
3. **Dual-Route Student Admission & Re-Enrolment (`Docs/Ikonex.md` 6.3, `C39–C41`, `A12`):**
   - `POST /api/v1/students/admit`: Used by Class Teachers (first load manual route) or Dean of Studies to admit a student with `admission_number`, `full_name`, `dob`, `gender`, `class_id`, `dormitory_id`, and guardian phone number.
   - If a parent later registers via the Opening-Day QR route with an `admission_number` that already exists in `students`, `registration_submissions.matched_student_id` is populated and `is_duplicate_flag = true` is surfaced to the verifier (`A12`).
   - `POST /api/v1/students/re-enrol`: Carries existing active students forward into a new academic year's `enrolments` automatically (`C39`).
4. **Guardian 2FA Registered Phone Number Change (`Docs/Ikonex.md` 6.3, `C25`, `C27`, `AC-26`):**
   - `POST /api/v1/guardians/me/phone-change/request-otp`: Sends a 6-digit OTP to the **new** phone number (`A5`).
   - `POST /api/v1/guardians/me/phone-change/confirm`: Verifies the OTP, updates `users.phone_number`, sets `chat_payment_paused_until = NOW() + INTERVAL '24 hours'` on the user/guardian record, sends a transactional SMS alert to the **old** phone number, invalidates existing WhatsApp sessions, and writes an `audit_log` entry (`AC-26`).

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
│   ├── config/
│   │   └── config.go                    # Env config loader (DB, Redis, JWT, Daraja, AT, AI token)
│   ├── database/
│   │   ├── database.go                  # pgxpool connection, WithTenantTx(ctx, schoolID, fn)
│   │   └── audit.go                     # Shared AuditLogger writing to audit_log
│   ├── middleware/
│   │   ├── middleware.go                # RequireAuth, RequireTenant, RequirePermission, RequireInternalService
│   │   └── response.go                  # Shared RespondJSON / RespondError helpers
│   ├── auth/
│   │   ├── auth.go                      # OTP request/verify, rate limiter, JWT issue/parse, memberships
│   │   └── auth_test.go
│   ├── consent/
│   │   ├── consent.go                   # GET/POST /api/v1/consents (biometrics, photo, sms, whatsapp)
│   │   └── consent_test.go
│   └── onboarding/
│       ├── onboarding.go                # Router registration & shared service struct
│       ├── school_wizard.go             # Team school creation & Principal 5-step setup wizard
│       ├── staff_onboarding.go          # Staff QR invites (8am-6pm), revocation
│       ├── parent_qr.go                 # 12h Parent QR generation, revocation, verifier badge delegation
│       ├── verification.go              # Public QR submission, in-person verification, doc deletion
│       ├── students.go                  # Student admission, duplicate flagging, re-enrolment
│       ├── guardian_phone.go            # 2FA phone change, old-number SMS alert, 24h payment pause
│       └── onboarding_test.go
```

### 4.3 Key API Contracts Owned by Evans

| Method & Endpoint | Auth / Permission | Request Body | Response (`200`/`201`) |
|---|---|---|---|
| `POST /api/v1/auth/otp/request` | Public | `{ "phone_number": "+254712345678" }` | `{ "status": "success", "data": { "message": "OTP sent", "cooldown_seconds": 60 } }` |
| `POST /api/v1/auth/otp/verify` | Public | `{ "phone_number": "+254712345678", "otp": "123456" }` | `{ "status": "success", "data": { "token": "<jwt>", "user": {...}, "memberships": [...] } }` |
| `POST /api/v1/admin/schools` | Ikonex Team | `{ "code": "IKX-001", "name": "Alliance High", "county": "Kiambu", "principal_name": "...", "principal_phone": "+2547...", "contact_phone": "...", "contact_email": "..." }` | `{ "status": "success", "data": { "school_id": "uuid", "principal_user_id": "uuid" } }` |
| `POST /api/v1/schools/setup-wizard` | `principal`, `deputy` | `{ "contact": {...}, "academic_year": 2026, "terms": [...], "grade_levels": [7,8,9,10], "classes": [...], "dormitories": [...], "grading_scale": {...} }` | `{ "status": "success", "data": { "wizard_completed": true } }` |
| `POST /api/v1/staff-invites` | `staff.invites.create` (`principal`, `deputy`) | `{ "phone_number": "+2547...", "role": "teacher", "valid_from": "2026-11-05T08:00:00Z", "valid_to": "2026-11-05T18:00:00Z" }` | `{ "status": "success", "data": { "invite_id": "uuid", "qr_token": "..." } }` |
| `POST /api/v1/parent-qr` | `parents.qr.manage` (`principal`, `deputy`, `dean_of_studies`) | `{ "notes": "Term 1 Opening Day" }` | `{ "status": "success", "data": { "qr_id": "uuid", "qr_token": "...", "valid_until": "... (+12h)" } }` |
| `POST /api/v1/parent-qr/{id}/revoke` | `parents.qr.manage` | `{}` | `{ "status": "success", "data": { "revoked": true } }` |
| `POST /api/v1/verifier-grants` | `verifier.grant` (`principal`, `deputy`, `dean_of_studies`) | `{ "staff_membership_id": "uuid", "parent_qr_id": "uuid" }` | `{ "status": "success", "data": { "badge_grant_id": "uuid", "valid_until": "..." } }` |
| `POST /api/v1/registrations` | Public (requires active `qr_token`) | `{ "qr_token": "...", "applicant_name": "...", "applicant_phone": "+2547...", "national_id": "...", "student": {...}, "document_type": "national_id", "document_base64": "..." }` | `{ "status": "success", "data": { "submission_id": "uuid", "status": "pending", "is_duplicate_flag": false } }` |
| `POST /api/v1/registrations/{id}/verify` | `registrations.verify` (or active `delegated_verifier` badge) | `{ "action": "approve", "document_type_verified": "national_id", "notes": "Verified ID in person" }` | `{ "status": "success", "data": { "status": "approved", "document_deleted": true, "verification_record_id": "uuid" } }` |
| `POST /api/v1/students/admit` | `students.admit` (`dean_of_studies`, `class_teacher`) | `{ "admission_number": "ADM-1001", "full_name": "...", "dob": "2011-05-14", "gender": "F", "class_id": "uuid", "dormitory_id": "uuid", "guardian_name": "...", "guardian_phone": "+2547...", "relationship": "mother" }` | `{ "status": "success", "data": { "student_id": "uuid", "enrolment_id": "uuid" } }` |
| `POST /api/v1/guardians/me/phone-change/request-otp` | Authenticated `parent` | `{ "new_phone_number": "+254799887766" }` | `{ "status": "success", "data": { "otp_sent_to": "+254799887766" } }` |
| `POST /api/v1/guardians/me/phone-change/confirm` | Authenticated `parent` | `{ "new_phone_number": "+254799887766", "otp": "654321" }` | `{ "status": "success", "data": { "phone_updated": true, "old_number_notified": true, "chat_payment_paused_until": "..." } }` |

---

## 5. Exact Deliverables

1. **Database RLS Isolation (`AC-27`):** Queries run via `WithTenantTx` return only rows for `app.school_id`; queries without `app.school_id` set return zero rows.
2. **SMS OTP Login & Rate Limiting (`AC-01`, `AC-10`):** Users can request and verify OTPs, receive JWTs with their school memberships and badges, and are rate-limited to 1 OTP/minute and blocked for 24 hours on the 6th request within 30 minutes.
3. **Scoped & Time-Bound Badge Middleware (`C53`, `AC-05`, `AC-14`):** Middleware accurately validates whether a caller holds a required role permission or an active, non-expired `membership_badges` grant (optionally scoped to a `class_id`).
4. **School Provisioning & Setup Wizard (`AC-01`):** Ikonex team can create a school + principal, and the principal can complete the setup wizard creating academic years, terms, CBC grades 7–12, classes, dormitories, initial grading scale, and required school contact details.
5. **Staff & 12h Parent QR Codes (`AC-02`, `AC-03`):** Staff QR codes are rejected outside their valid window or after revocation; Parent QR codes serve multiple parents for 12 hours and are rejected after 12 hours or revocation.
6. **Pending Registration Isolation, Delegation & Document Deletion (`AC-04`, `AC-05`, `AC-06`, `AC-07`):**
   - Pending submissions cannot access any student data (`AC-04`).
   - Delegated verifiers must be staff (`role != 'parent'`) and their badge expires automatically when the Parent QR expires or is revoked (`AC-05`).
   - Approving a registration creates the active user/student records, deletes the uploaded document bytes immediately, and keeps the `verification_records` row (`AC-06`).
   - Rejected or never-verified submissions and their files are purged after the retention window (`AC-07`).
7. **Guardian 2FA Phone Change (`AC-26`):** Verifies OTP sent to the new number, updates the phone number, sends an SMS notice to the old number, and sets `chat_payment_paused_until = NOW() + 24h`.

---

## 6. Acceptance Criteria

- [ ] **AC-BE-A01 (aligns with AC-27):** Automated integration test proves that querying `students`, `memberships`, or `registration_submissions` without `app.school_id` returns 0 rows, and School A cannot read School B's rows.
- [ ] **AC-BE-A02 (aligns with AC-01):** Creating a school and principal, logging in via OTP, and submitting `POST /api/v1/schools/setup-wizard` populates all foundational tables in a single transaction.
- [ ] **AC-BE-A03 (aligns with AC-02, AC-03):** `POST /api/v1/registrations` returns `410 Gone` if the Staff QR is outside its time window/revoked or if the Parent QR is older than 12 hours/revoked.
- [ ] **AC-BE-A04 (aligns with AC-04):** A user with only a `pending` submission in `registration_submissions` receives `403 Forbidden` on all parent/staff endpoints.
- [ ] **AC-BE-A05 (aligns with AC-05):** `POST /api/v1/verifier-grants` rejects parent memberships (`422`), allows staff memberships, and `POST /api/v1/registrations/{id}/verify` rejects a delegated verifier once the linked Parent QR expires or is revoked.
- [ ] **AC-BE-A06 (aligns with AC-06):** After `POST /api/v1/registrations/{id}/verify` (`action="approve"`), `registration_submissions.document_blob` is `NULL`, `document_deleted_at` is non-null, and a row exists in `verification_records`.
- [ ] **AC-BE-A07 (aligns with AC-26, AC-29):** Completing a 2FA phone change sets `chat_payment_paused_until` to 24h in the future, enqueues an SMS to the old phone number, and logs every sensitive action in `audit_log`.

---

## 7. Testing Requirements

Write automated Go tests (`go test -v ./internal/...`) covering:
1. **RLS Isolation Tests (`internal/database/database_test.go`):** Cross-tenant read/write rejection and unset `app.school_id` empty result test (`AC-27`).
2. **Auth & OTP Rate Limit Tests (`internal/auth/auth_test.go`):** 1-minute cooldown refusal (`429`), 6th request in 30 minutes triggering a 24h sender block (`AC-10`), and JWT claims validation.
3. **Badge Permission Middleware Tests (`internal/middleware/middleware_test.go`):** Scoped `class_teacher` badge allowed on own `class_id` and rejected (`403`) on another `class_id`; expired `delegated_verifier` badge rejected (`AC-05`).
4. **Onboarding & Verification Lifecycle Tests (`internal/onboarding/onboarding_test.go`):**
   - Staff QR window & revocation (`AC-02`).
   - Parent 12h QR multi-parent usage, expiry, and revocation (`AC-03`).
   - Pending isolation (`AC-04`), approval document deletion & `verification_records` persistence (`AC-06`), duplicate admission number flagging (`A12`), and 2FA phone change 24h payment lock (`AC-26`).

---

## 8. Integration Responsibilities

- **Provides to Developer 4 (Ema — @ejao-000) & Developer 5 (Ian — @frihk):**
  - `database.DB.WithTenantTx(ctx, schoolID, fn)` and `database.AuditLogger`.
  - `middleware.RequireAuth`, `middleware.RequirePermission`, `middleware.RequireInternalService`, and `middleware.RespondJSON` / `middleware.RespondError`.
  - `cmd/server/main.go` route mounting: Define clean `academics.RegisterRoutes(mux, deps)`, `billing.RegisterRoutes(mux, deps)`, `whatsapp.RegisterRoutes(mux, deps)`, `notifications.RegisterRoutes(mux, deps)`, and `internalapi.RegisterRoutes(mux, deps)` calls so Ema (@ejao-000) and Ian (@frihk) register their routes inside their own packages.
- **Consumes from Developer 5 (Ian — @frihk):**
  - Calls Ian's (@frihk) `notifications.SMSDispatcher` interface to send login OTPs and the old-number phone-change SMS alert.
  - Calls Ian's (@frihk) `billing.StudentChargeHook` interface when a student is admitted or verified so their initial term fee charges are created automatically.
  - Provides a clean cleanup query/method `onboarding.PurgeExpiredUnverifiedSubmissions(ctx)` that Ian's (@frihk) Asynq background worker invokes daily (`AC-07`).

---

## 9. File Ownership and Conflict Prevention

| Category | Paths | Rule |
|---|---|---|
| **Exclusively Owned by Evans (@eojuma, Dev 3)** | `backend/cmd/server/main.go`, `backend/internal/config/**`, `backend/internal/database/**`, `backend/internal/middleware/**`, `backend/internal/auth/**`, `backend/internal/onboarding/**`, `backend/internal/consent/**`, `migrations/000001_init_schema.*`, `migrations/000002_r1_core_onboarding.*` | Full ownership. Deliver `config`, `database`, and `middleware` in Milestone 1 so Ema (@ejao-000) and Ian (@frihk) can build against them immediately. |
| **Shared Files (Coordinate via PR)** | `backend/go.mod`, `backend/go.sum`, `Docs/API_CONTRACTS.md` | Coordinate before adding third-party Go dependencies. |
| **Strictly Do Not Modify** | `backend/internal/academics/**` (Ema — @ejao-000), `backend/internal/billing/**`, `backend/internal/whatsapp/**`, `backend/internal/notifications/**`, `backend/internal/jobs/**`, `backend/internal/internalapi/**`, `backend/cmd/worker/**` (Ian — @frihk), `ai_service/**` (Sos — @sospeter-57), `frontend/**` (Ashy — @Ashomondi) | Owned by other developers. |

---

## 10. Definition of Done

- [ ] `migrations/000002_r1_core_onboarding.up.sql` and `.down.sql` apply and roll back cleanly on PostgreSQL 16 with RLS forced on all new tenant tables.
- [ ] `database`, `middleware`, `auth`, `consent`, and `onboarding` packages compile and pass `go test -v ./...`.
- [ ] All endpoints in Section 4.3 return the standard JSON envelope and enforce badge/role permissions before handler execution.
- [ ] Every sensitive onboarding, QR, delegation, verification, and phone-change event is recorded in `audit_log`.
- [ ] Uploaded verification documents are provably deleted upon approval while `verification_records` remain intact.

