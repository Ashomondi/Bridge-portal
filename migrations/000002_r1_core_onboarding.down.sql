-- ==============================================================================
-- 000002_r1_core_onboarding.down.sql
-- Rollback Release 1 Core Foundation & Onboarding
-- ==============================================================================

DROP INDEX IF EXISTS consents_school_student_guardian_purpose_key;

ALTER TABLE consents DROP CONSTRAINT IF EXISTS consents_purpose_check;
ALTER TABLE consents ADD CONSTRAINT consents_purpose_check
    CHECK (purpose IN ('biometric_attendance', 'photo', 'sms'));

ALTER TABLE users DROP COLUMN IF EXISTS phone_changed_at;
ALTER TABLE users DROP COLUMN IF EXISTS chat_payment_paused_until;

DROP TABLE IF EXISTS verification_records CASCADE;
DROP TABLE IF EXISTS registration_submissions CASCADE;
DROP TABLE IF EXISTS parent_qr_codes CASCADE;
DROP TABLE IF EXISTS sender_rate_blocks CASCADE;
DROP TABLE IF EXISTS otp_requests CASCADE;
DROP TABLE IF EXISTS school_profile CASCADE;

DROP POLICY IF EXISTS memberships_self_read ON memberships;
DROP FUNCTION IF EXISTS current_user_id();
