-- ==============================================================================
-- 000002_r1_core_onboarding.up.sql
-- Ikonex Academy - Release 1 Core Foundation & Onboarding
-- Owner: Evans (@eojuma), Backend Engineer A
-- ==============================================================================

-- ------------------------------------------------------------------------------
-- Self-scoped membership reads for login (C18-C20)
-- A user must be able to discover their own school memberships before a
-- school context exists. This permissive policy does not grant access to any
-- school's student or academic data; those still require app.school_id.
-- ------------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION current_user_id() RETURNS uuid AS $$
    SELECT NULLIF(current_setting('app.user_id', true), '')::uuid;
$$ LANGUAGE sql STABLE;

DROP POLICY IF EXISTS memberships_self_read ON memberships;
CREATE POLICY memberships_self_read ON memberships
    FOR SELECT USING (user_id = current_user_id());

-- ------------------------------------------------------------------------------
-- School profile (contact details are required for a school to be usable, C57)
-- ------------------------------------------------------------------------------
CREATE TABLE school_profile (
    school_id UUID PRIMARY KEY REFERENCES schools(id) ON DELETE CASCADE,
    contact_phone VARCHAR(20) NOT NULL,
    contact_email VARCHAR(255) NOT NULL,
    office_hours VARCHAR(255),
    physical_address TEXT,
    student_onboarding_route VARCHAR(30) NOT NULL DEFAULT 'both'
        CHECK (student_onboarding_route IN ('manual', 'qr', 'both')),
    wizard_completed BOOLEAN NOT NULL DEFAULT FALSE,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ------------------------------------------------------------------------------
-- Phone + SMS OTP requests (C18-C20)
-- Global identity table: a phone number is not tenant-scoped until a membership
-- exists. OTPs are stored only as SHA-256 hashes.
-- ------------------------------------------------------------------------------
CREATE TABLE otp_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    phone_number VARCHAR(20) NOT NULL,
    admission_number VARCHAR(100),
    sender_identifier VARCHAR(50) NOT NULL,
    otp_hash TEXT NOT NULL,
    purpose VARCHAR(30) NOT NULL
        CHECK (purpose IN ('app_login', 'whatsapp_session', 'phone_change')),
    expires_at TIMESTAMPTZ NOT NULL,
    consumed_at TIMESTAMPTZ,
    failed_attempts INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_otp_requests_phone_created
    ON otp_requests (phone_number, purpose, created_at DESC);
CREATE INDEX idx_otp_requests_sender_created
    ON otp_requests (sender_identifier, created_at DESC);

-- ------------------------------------------------------------------------------
-- Sender rate blocks (C18-C20, AC-10)
-- A sender (phone or WhatsApp number) exceeding the request threshold is
-- blocked for 24 hours. Shared by app login and chat OTP flows.
-- ------------------------------------------------------------------------------
CREATE TABLE sender_rate_blocks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sender_identifier VARCHAR(50) UNIQUE NOT NULL,
    blocked_until TIMESTAMPTZ NOT NULL,
    reason VARCHAR(100) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ------------------------------------------------------------------------------
-- Parent 12-hour multi-use QR codes (C26, C28, AC-03)
-- ------------------------------------------------------------------------------
CREATE TABLE parent_qr_codes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id UUID NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
    qr_token VARCHAR(255) UNIQUE NOT NULL,
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    notes TEXT,
    valid_from TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    valid_until TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,
    revoked_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_parent_qr_codes_valid ON parent_qr_codes (school_id, valid_until);

-- ------------------------------------------------------------------------------
-- Pending registration submissions (C28, C63, C64, AC-04, AC-06, AC-07)
-- A pending submission grants no access to any student data. Uploaded document
-- bytes are stored here only until in-person verification deletes them.
-- ------------------------------------------------------------------------------
CREATE TABLE registration_submissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id UUID NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
    submission_type VARCHAR(20) NOT NULL CHECK (submission_type IN ('parent', 'staff')),
    parent_qr_id UUID REFERENCES parent_qr_codes(id) ON DELETE SET NULL,
    staff_invite_id UUID REFERENCES staff_invites(id) ON DELETE SET NULL,
    applicant_phone VARCHAR(20) NOT NULL,
    applicant_name VARCHAR(255) NOT NULL,
    national_id VARCHAR(50),
    payload JSONB NOT NULL DEFAULT '{}'::jsonb,
    document_type VARCHAR(100),
    document_blob BYTEA,
    document_deleted_at TIMESTAMPTZ,
    matched_student_id UUID REFERENCES students(id) ON DELETE SET NULL,
    is_duplicate_flag BOOLEAN NOT NULL DEFAULT FALSE,
    status VARCHAR(20) NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'approved', 'rejected')),
    expires_for_purge_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_registration_submissions_school_status
    ON registration_submissions (school_id, status, created_at DESC);
CREATE INDEX idx_registration_submissions_purge
    ON registration_submissions (status, expires_for_purge_at);
CREATE INDEX idx_registration_submissions_matched_student
    ON registration_submissions (matched_student_id)
    WHERE matched_student_id IS NOT NULL;

-- ------------------------------------------------------------------------------
-- Permanent verification records (AC-06)
-- Kept after the uploaded document is deleted so the in-person check is auditable.
-- ------------------------------------------------------------------------------
CREATE TABLE verification_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id UUID NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
    submission_id UUID REFERENCES registration_submissions(id) ON DELETE SET NULL,
    verified_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    student_id UUID REFERENCES students(id) ON DELETE SET NULL,
    verifier_user_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    document_type VARCHAR(100) NOT NULL,
    notes TEXT,
    verified_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_verification_records_school
    ON verification_records (school_id, verified_at DESC);
CREATE INDEX idx_verification_records_student
    ON verification_records (school_id, student_id);

-- ------------------------------------------------------------------------------
-- Guardian phone-change security state (C25, C27, AC-26)
-- ------------------------------------------------------------------------------
ALTER TABLE users ADD COLUMN IF NOT EXISTS chat_payment_paused_until TIMESTAMPTZ;
ALTER TABLE users ADD COLUMN IF NOT EXISTS phone_changed_at TIMESTAMPTZ;

-- ------------------------------------------------------------------------------
-- Consent purposes: allow WhatsApp consent in Release 1 (C22)
-- ------------------------------------------------------------------------------
ALTER TABLE consents DROP CONSTRAINT IF EXISTS consents_purpose_check;
ALTER TABLE consents ADD CONSTRAINT consents_purpose_check
    CHECK (purpose IN ('biometric_attendance', 'photo', 'sms', 'whatsapp'));

-- One current consent row per child, guardian and purpose. History of changes
-- stays in audit_log; the consent record itself is updated in place.
CREATE UNIQUE INDEX consents_school_student_guardian_purpose_key
    ON consents (school_id, student_id, guardian_user_id, purpose);

-- ------------------------------------------------------------------------------
-- Row-Level Security on every new tenant-owned table (AC-27)
-- ------------------------------------------------------------------------------
DO $$
DECLARE
    tbl TEXT;
    tenant_tables TEXT[] := ARRAY[
        'school_profile', 'parent_qr_codes', 'registration_submissions',
        'verification_records'
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
