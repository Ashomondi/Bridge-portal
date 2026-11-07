-- ==============================================================================
-- 000003_r1_academics.up.sql
-- Ikonex Academy - Release 1 CBC Academics & Classroom Operations
-- Owner: Ema (@ejao-000), Backend Engineer B
-- ==============================================================================

-- ------------------------------------------------------------------------------
-- Subjects (CBC learning areas per grade level)
-- ------------------------------------------------------------------------------
CREATE TABLE subjects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id UUID NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
    grade_level_id UUID NOT NULL REFERENCES grade_levels(id) ON DELETE CASCADE,
    code VARCHAR(50) NOT NULL,
    name VARCHAR(255) NOT NULL,
    is_compulsory BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (school_id, grade_level_id, code)
);

-- ------------------------------------------------------------------------------
-- Teacher subject assignments (C59)
-- ------------------------------------------------------------------------------
CREATE TABLE subject_teacher_assignments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id UUID NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
    academic_year_id UUID NOT NULL REFERENCES academic_years(id) ON DELETE CASCADE,
    class_id UUID NOT NULL REFERENCES classes(id) ON DELETE CASCADE,
    subject_id UUID NOT NULL REFERENCES subjects(id) ON DELETE CASCADE,
    teacher_membership_id UUID NOT NULL REFERENCES memberships(id) ON DELETE CASCADE,
    assigned_by UUID REFERENCES users(id) ON DELETE SET NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (school_id, academic_year_id, class_id, subject_id)
);

CREATE INDEX idx_subject_teacher_assignments_teacher
    ON subject_teacher_assignments (school_id, teacher_membership_id, class_id);

-- ------------------------------------------------------------------------------
-- Versioned CBC grading scales (C43, AC-19)
-- ------------------------------------------------------------------------------
CREATE TABLE grading_scale_versions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id UUID NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
    version_number INT NOT NULL,
    ee_min NUMERIC(5,2) NOT NULL,
    me_min NUMERIC(5,2) NOT NULL,
    ae_min NUMERIC(5,2) NOT NULL,
    be_min NUMERIC(5,2) NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (school_id, version_number),
    CHECK (ee_min > me_min AND me_min > ae_min AND ae_min > be_min),
    CHECK (be_min >= 0 AND ee_min <= 100)
);

CREATE INDEX idx_grading_scale_versions_active
    ON grading_scale_versions (school_id) WHERE is_active = TRUE;

-- ------------------------------------------------------------------------------
-- Timetables (C59)
-- ------------------------------------------------------------------------------
CREATE TABLE timetable_slots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id UUID NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
    academic_year_id UUID NOT NULL REFERENCES academic_years(id) ON DELETE CASCADE,
    term_id UUID NOT NULL REFERENCES terms(id) ON DELETE CASCADE,
    class_id UUID NOT NULL REFERENCES classes(id) ON DELETE CASCADE,
    subject_id UUID NOT NULL REFERENCES subjects(id) ON DELETE CASCADE,
    teacher_membership_id UUID NOT NULL REFERENCES memberships(id) ON DELETE CASCADE,
    day_of_week INT NOT NULL CHECK (day_of_week BETWEEN 1 AND 7),
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    room VARCHAR(50),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (end_time > start_time)
);

CREATE INDEX idx_timetable_slots_class_term
    ON timetable_slots (school_id, term_id, class_id, day_of_week, start_time);
CREATE INDEX idx_timetable_slots_teacher
    ON timetable_slots (school_id, teacher_membership_id, term_id);

-- ------------------------------------------------------------------------------
-- School calendar (C59, C60)
-- ------------------------------------------------------------------------------
CREATE TABLE calendar_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id UUID NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
    academic_year_id UUID REFERENCES academic_years(id) ON DELETE CASCADE,
    term_id UUID REFERENCES terms(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    event_type VARCHAR(50),
    start_at TIMESTAMPTZ NOT NULL,
    end_at TIMESTAMPTZ NOT NULL,
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (end_at >= start_at)
);

CREATE INDEX idx_calendar_events_term
    ON calendar_events (school_id, term_id, start_at);

-- ------------------------------------------------------------------------------
-- Daily roll-call attendance (C35, C36)
-- The source column lets Release 2 biometric capture write into this table.
-- ------------------------------------------------------------------------------
CREATE TABLE attendance_daily (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id UUID NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
    student_id UUID NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    class_id UUID NOT NULL REFERENCES classes(id) ON DELETE CASCADE,
    term_id UUID NOT NULL REFERENCES terms(id) ON DELETE CASCADE,
    attendance_date DATE NOT NULL,
    status VARCHAR(20) NOT NULL CHECK (status IN ('present', 'absent', 'late')),
    source VARCHAR(30) NOT NULL DEFAULT 'manual_roll_call',
    recorded_by UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (school_id, student_id, attendance_date)
);

CREATE INDEX idx_attendance_daily_class_date
    ON attendance_daily (school_id, class_id, attendance_date);
CREATE INDEX idx_attendance_daily_student
    ON attendance_daily (school_id, student_id, attendance_date);

-- ------------------------------------------------------------------------------
-- Raw subject marks (C30)
-- ------------------------------------------------------------------------------
CREATE TABLE student_marks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id UUID NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
    student_id UUID NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    class_id UUID NOT NULL REFERENCES classes(id) ON DELETE CASCADE,
    term_id UUID NOT NULL REFERENCES terms(id) ON DELETE CASCADE,
    subject_id UUID NOT NULL REFERENCES subjects(id) ON DELETE CASCADE,
    score NUMERIC(5,2) NOT NULL CHECK (score BETWEEN 0 AND 100),
    entered_by UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (school_id, student_id, term_id, subject_id)
);

CREATE INDEX idx_student_marks_class_term
    ON student_marks (school_id, class_id, term_id, subject_id);
CREATE INDEX idx_student_marks_student_term
    ON student_marks (school_id, student_id, term_id);

-- ------------------------------------------------------------------------------
-- Approved report comments (C30, C5)
-- ------------------------------------------------------------------------------
CREATE TABLE report_comments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id UUID NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
    student_id UUID NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    class_id UUID NOT NULL REFERENCES classes(id) ON DELETE CASCADE,
    term_id UUID NOT NULL REFERENCES terms(id) ON DELETE CASCADE,
    comment_text TEXT NOT NULL,
    ai_draft_used BOOLEAN NOT NULL DEFAULT FALSE,
    approved_by UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (school_id, student_id, term_id)
);

-- ------------------------------------------------------------------------------
-- Results release state machine (C31, C33, C34)
-- ------------------------------------------------------------------------------
CREATE TABLE result_releases (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id UUID NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
    class_id UUID NOT NULL REFERENCES classes(id) ON DELETE CASCADE,
    term_id UUID NOT NULL REFERENCES terms(id) ON DELETE CASCADE,
    status VARCHAR(20) NOT NULL DEFAULT 'draft'
        CHECK (status IN ('draft', 'released', 'reopened', 'rereleased')),
    grading_scale_version_id UUID REFERENCES grading_scale_versions(id) ON DELETE RESTRICT,
    released_by UUID REFERENCES users(id) ON DELETE SET NULL,
    released_at TIMESTAMPTZ,
    reopened_by UUID REFERENCES users(id) ON DELETE SET NULL,
    reopened_at TIMESTAMPTZ,
    rereleased_by UUID REFERENCES users(id) ON DELETE SET NULL,
    rereleased_at TIMESTAMPTZ,
    UNIQUE (school_id, class_id, term_id)
);

CREATE INDEX idx_result_releases_status
    ON result_releases (school_id, status, term_id);

-- ------------------------------------------------------------------------------
-- Immutable parent-facing released snapshots (AC-18)
-- The report_comment is copied here so a reopen cannot change what parents see.
-- ------------------------------------------------------------------------------
CREATE TABLE released_marks_snapshots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id UUID NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
    result_release_id UUID NOT NULL REFERENCES result_releases(id) ON DELETE CASCADE,
    student_id UUID NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    term_id UUID NOT NULL REFERENCES terms(id) ON DELETE CASCADE,
    subject_id UUID NOT NULL REFERENCES subjects(id) ON DELETE CASCADE,
    score NUMERIC(5,2) NOT NULL,
    cbc_level VARCHAR(10) NOT NULL CHECK (cbc_level IN ('EE', 'ME', 'AE', 'BE')),
    is_amended BOOLEAN NOT NULL DEFAULT FALSE,
    amendment_note TEXT,
    snapshot_version INT NOT NULL DEFAULT 1,
    report_comment TEXT,
    published_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (school_id, student_id, term_id, subject_id)
);

CREATE INDEX idx_released_snapshots_release
    ON released_marks_snapshots (school_id, result_release_id);
CREATE INDEX idx_released_snapshots_student_term
    ON released_marks_snapshots (school_id, student_id, term_id);

-- ------------------------------------------------------------------------------
-- Row-Level Security on every tenant-owned academic table (AC-27)
-- ------------------------------------------------------------------------------
DO $$
DECLARE
    tbl TEXT;
    tenant_tables TEXT[] := ARRAY[
        'subjects', 'subject_teacher_assignments', 'grading_scale_versions',
        'timetable_slots', 'calendar_events', 'attendance_daily',
        'student_marks', 'report_comments', 'result_releases',
        'released_marks_snapshots'
    ];
BEGIN
    FOREACH tbl IN ARRAY tenant_tables LOOP
        EXECUTE format('ALTER TABLE %I ENABLE ROW LEVEL SECURITY;', tbl);
        EXECUTE format('ALTER TABLE %I FORCE ROW LEVEL SECURITY;', tbl);
        EXECUTE format(
            'CREATE POLICY %I_tenant_isolation ON %I FOR ALL USING (school_id = current_school_id()) WITH CHECK (school_id = current_school_id());',
            tbl, tbl
        );
    END LOOP;
END $$;
