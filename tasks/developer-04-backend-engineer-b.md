# Ikonex Academy – Release 1 Task Assignment
## Developer 4: Backend Engineer B (**Ema** — @ejao-000)

- **Developer Name:** Ema (@ejao-000)
- **GitHub Username:** `@ejao-000`
- **Role:** Backend Engineer B (Go — CBC Academics, Dean of Studies Timetables & Calendar, Daily Attendance, Marks Entry, Hard-Gated Results Release/Reopen State Machine & Performance Graph)
- **Primary Directory Ownership:** `backend/internal/academics/` and `migrations/000003_r1_academics.*`
- **Release Scope:** Release 1 (Teacher Class/Subject Assignments, Timetables, School Calendar & Term-Date Fee Sync, Versioned CBC Grading Scales, Daily Roll-Call Attendance, Anonymised Class Means, Subject Marks Entry & Reassignment, Hard-Gated Results Release/Reopen/Re-Release State Machine, AI Report Comment Drafting Proxy, and Student CBC Performance History Graph)

---

## 1. Role and Objective

As **Backend Engineer B (Ema — @ejao-000)**, you own the entire **CBC Academics & Classroom Operations** domain (`backend/internal/academics/`) for Ikonex Academy Release 1.

Your core responsibility is to implement the academic workflows defined in `Docs/Ikonex.md` Section 6.5, `Docs/ARCHITECTURE.md`, and `Docs/DECISIONS.md` with strict adherence to Ikonex's pedagogical and data-privacy rules:
1. **Dean of Studies Authority (`C59, C60, C61`):** The Dean of Studies assigns teachers to classes and subjects, owns all teacher and student timetables, and manages the school calendar—including editing term dates after initial wizard setup (which automatically shifts term fee due dates and requires a mandatory fee-change note).
2. **One Daily Class-Teacher Roll-Call (`C35, C36`):** Only the teacher holding the `class_teacher` badge scoped to a specific `class_id` can record the daily attendance roll-call for that class (`attendance_daily`), triggering an in-app notification to guardians of absent students.
3. **Strict Privacy Across Classes (`Docs/Ikonex.md` Section 4):** A class teacher sees full rosters for their own class; any teacher viewing another class via `/api/v1/classes/{id}/stats` receives **only class and subject mean scores/levels, never individual student names**.
4. **Hard-Gated Results Release & Snapshot-Safe Reopen State Machine (`C30–C34`):**
   - Subject teachers enter marks only for their assigned `(class_id, subject_id)` (unless reassigned by the Dean/Principal/Deputy).
   - The class teacher **cannot** release results until **every subject has marks for every enrolled student in the class** (hard gate, zero override in Release 1).
   - When a released class is **reopened** for mark correction, parents **continue seeing the previously released snapshot** until the class teacher **re-releases** the results with a mandatory **teacher's explanation note** per corrected mark, notifying **only** the parents of affected students.
5. **Versioned CBC Grading Scales & Non-Ranking Performance Graph (`C42, C43`):** A term's CBC levels (`EE`, `ME`, `AE`, `BE`) are computed using the grading scale version in force when that term's results were released. Past terms never change when a scale is edited later, and the student performance graph **never includes class rank or position**.

---

## 2. Assigned Features

You own **3 cohesive, end-to-end backend features** inside `backend/internal/academics/`:

| Feature ID | Feature Name | Owned Files in `internal/academics/` | Why These Belong Together |
|---|---|---|---|
| **BE-B1** | **Dean of Studies Management: Teacher Assignments, Timetables, School Calendar & Versioned Grading Scales** | `academics.go`, `assignments.go`, `timetables_calendar.go`, `grading_scale.go` | Defines the structural academic configuration (who teaches what, when lessons/terms happen, and how scores map to CBC levels) that attendance and marks depend on. |
| **BE-B2** | **Daily Roll-Call Attendance, Scoped Class Roster & Anonymised Class/Subject Means** | `attendance.go`, `roster_stats.go` | Both operate on daily classroom management, enforcing the boundary between a `class_teacher`'s full student roster view and other teachers' anonymised mean-only view. |
| **BE-B3** | **Subject Marks Entry, Marks Reassignment, Hard-Gated Results Release/Reopen State Machine, AI Comment Proxy & CBC Performance Graph** | `marks.go`, `results_release.go`, `report_comments.go`, `performance_graph.go` | Represents the complete assessment lifecycle from raw mark entry to teacher comment drafting, hard-gated release, reopen corrections, and multi-term parent visualization. |

### Explicitly Out of Scope for Release 1 (Do Not Implement Yet)
Per `Docs/Ikonex.md` Section 3 & Appendix C:
- **Syllabus Coverage & AI Strand Mapping (`POST /api/v1/syllabus/coverage`):** Placed in **Release 2** (`Docs/Ikonex.md` Appendix C, `REQUIREMENTS.md` Section D). Keep stubbed (`501 Not Implemented`).
- **Biometric Attendance Capture:** Release 2 (`C35`). However, design `attendance_daily` with a `source VARCHAR(30) DEFAULT 'manual_roll_call'` column so Release 2 biometric attendance writes into the exact same daily table (`Docs/Ikonex.md` 6.5).
- **Nurse (`internal/nurse`), Library (`internal/library`), and Clubs (`internal/clubs`):** Release 2 modules. Do not implement in Release 1.

---

## 3. Detailed Feature Specifications

### Feature BE-B1: Dean of Studies Management (Assignments, Timetables, Calendar & Grading Scales)
1. **Teacher Class & Subject Assignments (`C59`):**
   - `POST /api/v1/academics/assignments`: Restricted to `dean_of_studies` (and `principal`/`deputy` per `O16` configurable default).
   - Supports two assignment types:
     - **Class Teacher Assignment:** Grants the `class_teacher` badge in `membership_badges` with `scope_id = class_id` for the current academic year (revoking any previous class teacher badge on that same class so each class has one active class teacher).
     - **Subject Teacher Assignment:** Creates/updates a row in `subject_teacher_assignments` linking `(school_id, academic_year_id, class_id, subject_id, teacher_membership_id)`.
2. **Timetables & School Calendar (`C59, C60, C61, AC-22`):**
   - `GET /api/v1/timetables` and `PUT /api/v1/timetables`: Manages weekly timetable slots (`timetable_slots`) per class and teacher. Updating a timetable sends an in-app notification via Ian's (@frihk) `notifications.Service` to affected teachers and parents of students in that class.
   - `GET /api/v1/calendar` and `PUT /api/v1/calendar`: Manages school events (`calendar_events`) and allows the `dean_of_studies` to update term dates (`terms.start_date`, `terms.end_date`) after the setup wizard (`C60`).
   - **Term Start-Date Change Rule (`C61, AC-22`):**
     - If `PUT /api/v1/calendar` modifies `terms.start_date`, the request **must** include a non-empty `fee_change_note` string. If `fee_change_note` is missing or blank, reject with `422 Unprocessable Entity` (`TERM_DATE_CHANGE_NOTE_REQUIRED`).
     - Within the transaction, call Ian's (@frihk) `billing.OnTermStartDateChanged(ctx, tx, schoolID, termID, newStartDate, deanUserID, feeChangeNote)` hook, which shifts the term's fee due dates, records a `fee_change_notes` row, and notifies affected parents (`AC-22`).
     - Log the term-date change to `audit_log` (`AC-29`).
3. **Versioned CBC Grading Scale (`C43, AC-19`):**
   - `GET /api/v1/academics/grading-scales` & `PUT /api/v1/academics/grading-scales`:
   - Stores score bands for `EE` (e.g., `80–100`), `ME` (`60–79`), `AE` (`40–59`), `BE` (`0–39`) in `grading_scale_versions`.
   - Editing the grading scale creates a new version that applies **only to terms whose results have not yet been released** (`C43`). Once a term's results are released for a class, `result_releases.grading_scale_version_id` is permanently locked to the scale version active at release time (`AC-19`).

### Feature BE-B2: Daily Class Roll-Call Attendance, Scoped Class Roster & Anonymised Stats
1. **Daily Roll-Call (`C35, C36, AC-14`):**
   - `POST /api/v1/classes/{id}/attendance`:
   - **Authorization:** Caller must hold an active `class_teacher` badge in `membership_badges` with `scope_id == {id}` (the target `class_id`). A teacher not holding the `class_teacher` badge for `{id}` is rejected with `403 Forbidden` (`AC-14`).
   - Upserts one row per student per `attendance_date` in `attendance_daily` (`status IN ('present', 'absent', 'late')`, `source = 'manual_roll_call'`).
   - For every student marked `'absent'`, enqueues an **in-app notification** (`C36, C58`) to all guardians linked to that student in `student_guardians`. (No SMS is sent for absences in Release 1).
2. **Class Roster vs. Anonymised Class Statistics (`Docs/Ikonex.md` Section 4):**
   - `GET /api/v1/classes/{id}/roster`: Requires `class.students.view` scoped to `{id}` (held by the class teacher of `{id}`, subject teachers assigned to `{id}`, Dean, Principal, Deputy). Returns the full list of enrolled students (`student_id`, `admission_number`, `full_name`).
   - `GET /api/v1/classes/{id}/stats`: Callable by **any teacher** in the school (`class.stats.view`). Returns **only aggregate statistics** (`class_mean_score`, `total_students`, and per-subject `{ "subject_name", "mean_score", "predominant_cbc_level", "level_distribution": {"EE": 5, "ME": 18, "AE": 7, "BE": 2} }`). **Never returns student names, admission numbers, or individual rows.**

### Feature BE-B3: Subject Marks Entry, Hard-Gated Results Release/Reopen State Machine & Performance Graph
1. **Subject Marks Entry & Reassignment (`C30, C32, AC-15, AC-17`):**
   - `PUT /api/v1/classes/{id}/subjects/{sid}/marks`:
   - **Authorization (`AC-15`):** Checks `subject_teacher_assignments` to verify the caller is the active assigned teacher for `(class_id = {id}, subject_id = {sid})` in the target term's academic year. If not assigned, returns `403 Forbidden`.
   - **State Check:** Allowed only when the class's `result_releases` state for that term is `'draft'` or `'reopened'`. If `'released'` or `'rereleased'`, rejects with `409 Conflict` (`RESULTS_LOCKED_RELEASED`).
   - Saves raw numeric `score` (`0.00–100.00`) in `student_marks` and derives the provisional `cbc_level` (`EE`, `ME`, `AE`, `BE`).
   - `POST /api/v1/classes/{id}/marks-assignments` (`C32, AC-17`): Callable by `dean_of_studies`, `principal`, or `deputy`. Reassigns `(class_id, subject_id)` to a new `teacher_membership_id`, immediately revoking the previous teacher's write access (`AC-17`), and writes to `audit_log` (`AC-29`).
2. **Hard-Gated Results Release (`C31, AC-16`):**
   - `POST /api/v1/classes/{id}/results/release`: Callable **only** by the `class_teacher` of `{id}`.
   - **Hard Gate (`AC-16`):** Queries all active enrolled students in `class_id` and all active subjects for that class/grade level. If **any** `(student_id, subject_id)` pair is missing a mark in `student_marks` for the target `term_id`, rejects with `422 Unprocessable Entity`:
     ```json
     {
       "status": "error",
       "error": {
         "code": "RESULTS_INCOMPLETE_MARKS",
         "message": "Cannot release results until every subject has marks for every student.",
         "details": {
           "missing_count": 3,
           "missing_subjects": [
             { "subject_id": "uuid", "subject_name": "Mathematics", "missing_students_count": 3 }
           ]
         }
       }
     }
     ```
   - When 100% complete:
     - Locks `grading_scale_version_id` on `result_releases`.
     - Copies the current marks and report comments into an immutable JSONB/table snapshot (`released_marks_snapshots`) so parents always read from the last released snapshot.
     - Sets `result_releases.status = 'released'`, logs to `audit_log`, and sends in-app notifications to all guardians of students in the class.
3. **Reopen & Re-Release with Mandatory Teacher Note (`C33, C34, A9, AC-18`):**
   - `POST /api/v1/classes/{id}/results/reopen`: Callable by the `class_teacher` of `{id}`. Transitions `result_releases.status` from `'released'` (or `'rereleased'`) to `'reopened'` and logs to `audit_log`.
   - **Parent Visibility Rule (`AC-18`):** While `status == 'reopened'`, all parent-facing endpoints (`GET /api/v1/students/{id}/results`, `GET /api/v1/students/{id}/performance-history`) continue serving the **last released snapshot** from `released_marks_snapshots`—never the in-progress reopened edits.
   - `POST /api/v1/classes/{id}/results/re-release`: Callable by the `class_teacher` of `{id}`.
     - Compares current `student_marks` against the previous `released_marks_snapshots` to detect every `(student_id, subject_id)` whose mark or CBC level changed.
     - **Mandatory Teacher Note Rule (`C34, AC-18`):** Every changed `(student_id, subject_id)` must have a corresponding non-empty `teacher_note` in the request body explaining why the mark was corrected. If any modified mark lacks a `teacher_note`, reject with `422 Unprocessable Entity` (`RERELEASE_NOTE_REQUIRED`).
     - Saves the new snapshot into `released_marks_snapshots` (marking changed subjects with `is_amended = true` and `amendment_note = teacher_note`), sets `status = 'rereleased'`, logs to `audit_log`, and sends in-app notifications **strictly to guardians of the affected students** (never unaffected parents, `AC-18`).
4. **Report Card Comments & AI Drafter Proxy (`C5, C30`):**
   - `POST /api/v1/report-comments/draft`: Authenticated endpoint for teachers. Gathers the student's current term marks, prior term trend, and attendance summary (or `teacher_reason` for a re-release note), signs `X-User-Context`, and calls Sos's `POST /api/v1/ai/report-comments/draft` endpoint, returning the draft to the teacher UI.
   - `PUT /api/v1/classes/{id}/report-comments`: Allows the `class_teacher` to save/edit their approved comments in `report_comments` before release/re-release.
5. **Student CBC Performance History Graph (`C42, C43, AC-19`):**
   - `GET /api/v1/students/{id}/performance-history`: Scoped to linked guardians of `{id}` or teachers of `{id}`.
   - Reads across all `enrolments` and `released_marks_snapshots` for `{id}` from the day the student joined the school (`Docs/ARCHITECTURE.md`).
   - Returns per-term, per-subject CBC performance levels (`EE`, `ME`, `AE`, `BE`, plus numeric level value `4, 3, 2, 1` for graphing) computed using each term's locked `grading_scale_version_id`. **Never computes or returns class rank or student position (`AC-19`).**

---

## 4. Technical Implementation

### 4.1 Database Migration Owned by Ema (`migrations/000003_r1_academics.up.sql`)
Create `000003_r1_academics.up.sql` and `.down.sql` with all Release 1 Academic tables and RLS policies:
- `subjects`: `id`, `school_id`, `grade_level_id`, `code`, `name`, `is_compulsory BOOLEAN DEFAULT TRUE`, `UNIQUE(school_id, grade_level_id, code)`.
- `subject_teacher_assignments`: `id`, `school_id`, `academic_year_id`, `class_id`, `subject_id`, `teacher_membership_id`, `assigned_by`, `updated_at`, `UNIQUE(school_id, academic_year_id, class_id, subject_id)`.
- `grading_scale_versions`: `id`, `school_id`, `version_number INT NOT NULL`, `ee_min NUMERIC(5,2) NOT NULL`, `me_min NUMERIC(5,2) NOT NULL`, `ae_min NUMERIC(5,2) NOT NULL`, `be_min NUMERIC(5,2) NOT NULL DEFAULT 0`, `is_active BOOLEAN DEFAULT TRUE`, `created_by`, `created_at`.
- `timetable_slots`: `id`, `school_id`, `academic_year_id`, `term_id`, `class_id`, `subject_id`, `teacher_membership_id`, `day_of_week INT CHECK (day_of_week BETWEEN 1 AND 7)`, `start_time TIME NOT NULL`, `end_time TIME NOT NULL`, `room VARCHAR(50)`.
- `calendar_events`: `id`, `school_id`, `academic_year_id`, `term_id`, `title`, `description`, `event_type`, `start_at TIMESTAMPTZ`, `end_at TIMESTAMPTZ`, `created_by`, `created_at`.
- `attendance_daily`: `id`, `school_id`, `student_id`, `class_id`, `term_id`, `attendance_date DATE NOT NULL`, `status VARCHAR(20) NOT NULL CHECK (status IN ('present', 'absent', 'late'))`, `source VARCHAR(30) NOT NULL DEFAULT 'manual_roll_call'`, `recorded_by UUID NOT NULL`, `updated_at`, `UNIQUE(school_id, student_id, attendance_date)`.
- `student_marks`: `id`, `school_id`, `student_id`, `class_id`, `term_id`, `subject_id`, `score NUMERIC(5,2) NOT NULL CHECK (score BETWEEN 0 AND 100)`, `entered_by UUID NOT NULL`, `updated_at`, `UNIQUE(school_id, student_id, term_id, subject_id)`.
- `report_comments`: `id`, `school_id`, `student_id`, `class_id`, `term_id`, `comment_text TEXT NOT NULL`, `ai_draft_used BOOLEAN DEFAULT FALSE`, `approved_by UUID NOT NULL`, `updated_at`, `UNIQUE(school_id, student_id, term_id)`.
- `result_releases`: `id`, `school_id`, `class_id`, `term_id`, `status VARCHAR(20) NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'released', 'reopened', 'rereleased'))`, `grading_scale_version_id UUID REFERENCES grading_scale_versions(id)`, `released_by UUID`, `released_at TIMESTAMPTZ`, `reopened_by UUID`, `reopened_at TIMESTAMPTZ`, `rereleased_by UUID`, `rereleased_at TIMESTAMPTZ`, `UNIQUE(school_id, class_id, term_id)`.
- `released_marks_snapshots`: `id`, `school_id`, `result_release_id UUID NOT NULL REFERENCES result_releases(id) ON DELETE CASCADE`, `student_id UUID NOT NULL`, `term_id UUID NOT NULL`, `subject_id UUID NOT NULL`, `score NUMERIC(5,2) NOT NULL`, `cbc_level VARCHAR(10) NOT NULL CHECK (cbc_level IN ('EE', 'ME', 'AE', 'BE'))`, `is_amended BOOLEAN NOT NULL DEFAULT FALSE`, `amendment_note TEXT`, `snapshot_version INT NOT NULL DEFAULT 1`, `published_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`, `UNIQUE(school_id, student_id, term_id, subject_id)`.
- Apply `ENABLE ROW LEVEL SECURITY`, `FORCE ROW LEVEL SECURITY`, and `tenant_isolation` policies on all 10 tables above.

### 4.2 Owned Go Files & Package Structure (`backend/internal/academics/`)
```text
backend/internal/academics/
├── academics.go               # Service struct, dependencies, RegisterRoutes(mux, deps)
├── assignments.go             # Teacher subject assignments & class_teacher badge assignment
├── timetables_calendar.go     # Timetables, calendar events, and term start-date change hook
├── grading_scale.go           # Versioned CBC grading scales & score-to-level resolver
├── attendance.go              # Daily class roll-call & guardian in-app absence notifications
├── roster_stats.go            # Scoped class roster & anonymised other-class mean statistics
├── marks.go                   # Subject teacher marks entry & Dean marks reassignment
├── results_release.go         # Hard-gated release, reopen & re-release snapshot state machine
├── report_comments.go         # Teacher report comments & proxy to Sos's AI drafter
├── performance_graph.go       # Multi-term CBC level history graph & parent released results view
└── academics_test.go          # Comprehensive unit & integration tests
```

### 4.3 Key API Contracts Owned by Ema

| Method & Endpoint | Auth / Permission | Request Body | Response (`200`/`201`) |
|---|---|---|---|
| `POST /api/v1/academics/assignments` | `dean_of_studies`, `principal`, `deputy` | `{ "class_id": "uuid", "class_teacher_membership_id": "uuid", "subject_assignments": [{"subject_id": "uuid", "teacher_membership_id": "uuid"}] }` | `{ "status": "success", "data": { "assigned": true } }` |
| `GET /api/v1/classes/{id}/roster` | `class.students.view` (scoped to `{id}`) | — | `{ "status": "success", "data": { "class_id": "uuid", "students": [{"student_id": "...", "admission_number": "...", "full_name": "..."}] } }` |
| `GET /api/v1/classes/{id}/stats` | `class.stats.view` (any teacher) | `?term_id=uuid` | `{ "status": "success", "data": { "class_id": "uuid", "class_mean": 68.4, "subject_means": [{"subject_name": "Math", "mean_score": 71.2, "predominant_cbc_level": "ME"}] } }` *(Zero student names)* |
| `POST /api/v1/classes/{id}/attendance` | `class_teacher` badge scoped to `{id}` | `{ "date": "2026-11-05", "records": [{"student_id": "uuid", "status": "present"}, {"student_id": "uuid", "status": "absent"}] }` | `{ "status": "success", "data": { "recorded_count": 35, "absent_notifications_queued": 1 } }` |
| `PUT /api/v1/classes/{id}/subjects/{sid}/marks` | Assigned subject teacher for `({id}, {sid})` | `{ "term_id": "uuid", "marks": [{"student_id": "uuid", "score": 84.5}] }` | `{ "status": "success", "data": { "saved_count": 35 } }` |
| `POST /api/v1/classes/{id}/marks-assignments` | `marks.reassign` (`dean_of_studies`, `principal`, `deputy`) | `{ "subject_id": "uuid", "new_teacher_membership_id": "uuid", "reason": "Medical leave" }` | `{ "status": "success", "data": { "reassigned": true } }` |
| `POST /api/v1/classes/{id}/results/release` | `class_teacher` badge scoped to `{id}` | `{ "term_id": "uuid" }` | `200 OK` when 100% marks entered; `422 RESULTS_INCOMPLETE_MARKS` if any mark is missing |
| `POST /api/v1/classes/{id}/results/reopen` | `class_teacher` badge scoped to `{id}` | `{ "term_id": "uuid", "reason": "Correcting Math mark" }` | `{ "status": "success", "data": { "status": "reopened" } }` |
| `POST /api/v1/classes/{id}/results/re-release` | `class_teacher` badge scoped to `{id}` | `{ "term_id": "uuid", "correction_notes": [{"student_id": "uuid", "subject_id": "uuid", "teacher_note": "Paper 2 score addition fixed"}] }` | `{ "status": "success", "data": { "status": "rereleased", "affected_students_notified": 1 } }` |
| `POST /api/v1/report-comments/draft` | `teacher`, `dean_of_studies`, `principal` | `{ "mode": "term_comment", "student_id": "uuid", "term_id": "uuid", "teacher_reason": null }` | `{ "status": "success", "data": { "draft_comment": "...", "requires_teacher_approval": true } }` |
| `GET/PUT /api/v1/timetables` | `dean_of_studies` (`PUT`), Authenticated (`GET`) | `{ "term_id": "uuid", "slots": [...] }` | `{ "status": "success", "data": { "slots": [...] } }` |
| `GET/PUT /api/v1/calendar` | `dean_of_studies` (`PUT`), Authenticated (`GET`) | `{ "term_updates": [{"term_id": "uuid", "start_date": "2026-05-11", "end_date": "2026-08-07", "fee_change_note": "Ministry delayed Term 2 opening by 1 week"}], "events": [...] }` | `{ "status": "success", "data": { "updated": true, "fee_due_dates_shifted": true } }` |
| `GET /api/v1/students/{id}/performance-history` | Guardian of `{id}` or Teacher of `{id}` | — | `{ "status": "success", "data": { "student_id": "uuid", "terms": [{"term_id": "...", "term_name": "Term 1 2026", "subjects": [{"subject_name": "Math", "cbc_level": "EE", "level_value": 4, "is_amended": false, "amendment_note": null}]}] } }` |

---

## 5. Exact Deliverables

1. **Class-Teacher Scoped Roll-Call (`AC-14`):** A teacher with the `class_teacher` badge for Class A can submit daily attendance for Class A (creating in-app absence notifications for parents of absent students) and is rejected (`403`) if attempting to submit attendance for Class B.
2. **Subject-Teacher Scoped Marks Entry & Reassignment (`AC-15`, `AC-17`):** A subject teacher can enter marks only for their assigned `(class_id, subject_id)`. Once the Dean reassigns that `(class_id, subject_id)` to a second teacher, the second teacher can enter marks and the first teacher is rejected with `403`.
3. **Hard-Gated Results Release (`AC-16`):** Calling `POST /api/v1/classes/{id}/results/release` when even 1 student is missing a mark in 1 subject fails with `422 RESULTS_INCOMPLETE_MARKS` listing the incomplete subject(s). When all students have marks for all subjects, release succeeds and creates the immutable snapshot.
4. **Snapshot-Preserving Reopen & Targeted Re-Release (`AC-18`):**
   - Reopening a released class leaves the previously released results visible to parents.
   - Re-releasing after changing a mark requires a `teacher_note` for every changed mark, marks the amended subject (`is_amended: true`), shows the `amendment_note` to the parent, and sends an in-app notification **only** to parents of students whose marks changed.
5. **Versioned Grading Scale & Non-Ranking Performance Graph (`AC-19`):** Editing a school's grading scale creates a new `grading_scale_versions` row that applies only to unreleased/future terms; already-released terms retain their original CBC levels, and `GET /api/v1/students/{id}/performance-history` never returns class rank or position.
6. **Term Start-Date Fee Sync (`AC-22`):** Updating a term's `start_date` via `PUT /api/v1/calendar` requires a `fee_change_note`, shifts the term's fee due date via Ian's (@frihk) Billing hook, and notifies parents.

---

## 6. Acceptance Criteria

- [ ] **AC-BE-B01 (aligns with AC-14):** `POST /api/v1/classes/{id}/attendance` returns `200` for the class's `class_teacher` and `403` for any other teacher.
- [ ] **AC-BE-B02 (aligns with AC-15, AC-17):** `PUT /api/v1/classes/{id}/subjects/{sid}/marks` enforces `subject_teacher_assignments` and immediately respects reassignments via `POST /api/v1/classes/{id}/marks-assignments`.
- [ ] **AC-BE-B03 (aligns with AC-16):** `POST /api/v1/classes/{id}/results/release` returns `422` if `count(entered_marks) < count(active_students) * count(class_subjects)`.
- [ ] **AC-BE-B04 (aligns with AC-18):** During `status == 'reopened'`, `GET /api/v1/students/{id}/performance-history` returns the pre-reopen mark. After `re-release` with a required `teacher_note`, it returns the new mark with `is_amended: true` and notifies only the affected student's guardian(s).
- [ ] **AC-BE-B05 (aligns with AC-19):** Editing the grading scale in Term 2 does not alter the CBC levels returned for Term 1, and neither `/stats` nor `/performance-history` ever exposes student rankings or individual names to unauthorized teachers.
- [ ] **AC-BE-B06 (aligns with AC-22):** Moving a term `start_date` in `PUT /api/v1/calendar` without `fee_change_note` fails with `422`; with `fee_change_note`, it updates the term fee due date and triggers the fee-change notification.

---

## 7. Testing Requirements

Write automated Go tests in `backend/internal/academics/academics_test.go`:
1. **Attendance Authorization & Notification Test (`AC-14`):** Verifies class teacher succeeds, non-class teacher gets `403`, and absent students trigger in-app notifications.
2. **Marks Assignment, Hard Gate & Reassignment Test (`AC-15`, `AC-16`, `AC-17`):**
   - Tests unassigned teacher rejection (`403`).
   - Tests release rejection (`422`) when 1 student is missing 1 subject mark.
   - Tests Dean reassignment revoking old teacher access and granting new teacher access (`AC-17`).
   - Tests successful release once the missing mark is entered.
3. **Reopen & Re-Release Snapshot Isolation Test (`AC-18`):**
   - Releases Term 1 with Student A having Math = `70 (ME)`.
   - Reopens Term 1 and updates Student A's Math to `85 (EE)`.
   - Asserts parent query while `reopened` still returns `70 (ME)`.
   - Attempts `re-release` without `teacher_note` -> asserts `422`.
   - Calls `re-release` with `teacher_note` -> asserts parent query now returns `85 (EE)` with `is_amended = true`, and asserts notification was queued only for Student A's guardian (not Student B's guardian).
4. **Grading Scale Versioning & Anonymised Stats Test (`AC-19`):**
   - Releases Term 1 with `EE >= 80`. Updates scale to `EE >= 85`. Asserts Term 1 student with `82` remains `EE`.
   - Asserts `GET /api/v1/classes/{id}/stats` JSON payload contains zero student names or IDs.

---

## 8. Integration Responsibilities

- **Depends on Developer 3 (Evans — @eojuma – Backend Engineer A):**
  - Uses `database.WithTenantTx`, `database.AuditLogger`, `middleware.RequireAuth`, and `middleware.RequirePermission` (including scoped `class_teacher` badge check).
- **Integrates with Developer 5 (Ian — @frihk – Backend Engineer C):**
  - Calls Ian's (@frihk) `notifications.Service.CreateInAppNotification(ctx, tx, ...)` for absence alerts, timetable/calendar changes, results release, and targeted re-release alerts.
  - Calls Ian's (@frihk) `billing.Service.OnTermStartDateChanged(ctx, tx, ...)` when the Dean changes a term's start date (`C61, AC-22`).
  - Exposes internal read methods (`GetStudentAttendanceSummary`, `GetStudentMarksSummary`) on `academics.Service` so Ian's (@frihk) `internalapi` package can serve `/api/v1/internal/students/{id}/attendance-summary` and `/api/v1/internal/students/{id}/marks-summary` to Sos's (@sospeter-57) AI service.
- **Integrates with Developer 1 (Sos — @sospeter-57 – AI Engineer):**
  - `POST /api/v1/report-comments/draft` calls Sos's (@sospeter-57) `POST /api/v1/ai/report-comments/draft` with `X-Internal-Service-Token` and signed `X-User-Context`.

---

## 9. File Ownership and Conflict Prevention

| Category | Paths | Rule |
|---|---|---|
| **Exclusively Owned by Ema (@ejao-000, Dev 4)** | `backend/internal/academics/**`, `migrations/000003_r1_academics.up.sql`, `migrations/000003_r1_academics.down.sql` | Full read/write ownership. No other developer edits `internal/academics/` or `000003_r1_academics.*`. |
| **Shared Files (Coordinate via PR)** | `backend/cmd/server/main.go` (owned by Evans — @eojuma; Ema exports `academics.RegisterRoutes`), `Docs/API_CONTRACTS.md` | Do not edit `main.go` directly without coordinating with Evans (@eojuma). |
| **Strictly Do Not Modify** | `backend/internal/auth/**`, `backend/internal/onboarding/**`, `backend/internal/middleware/**` (Evans — @eojuma), `backend/internal/billing/**`, `backend/internal/whatsapp/**`, `backend/internal/notifications/**`, `backend/internal/internalapi/**` (Ian — @frihk), `ai_service/**` (Sos — @sospeter-57), `frontend/**` (Ashy — @Ashomondi) | Owned by other developers. |

---

## 10. Definition of Done

- [ ] `migrations/000003_r1_academics.up.sql` and `.down.sql` run cleanly with `FORCE ROW LEVEL SECURITY` on all 10 academic tables.
- [ ] All endpoints in Section 4.3 are implemented and wired via `academics.RegisterRoutes`.
- [ ] Hard-gated release (`AC-16`), snapshot-preserving reopen & targeted re-release with required teacher note (`AC-18`), versioned grading scales (`AC-19`), and term-date fee due-date sync (`AC-22`) are verified by automated tests.
- [ ] All sensitive academic actions (marks reassignment, results release, reopen, re-release, term-date changes) are written to `audit_log` (`AC-29`).
- [ ] `go test -v ./internal/academics/...` passes with zero failures.

