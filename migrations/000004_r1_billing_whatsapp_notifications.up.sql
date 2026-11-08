-- ==============================================================================
-- 000004_r1_billing_whatsapp_notifications.up.sql
-- Ikonex Academy - Release 1 Billing, Daraja M-Pesa, WhatsApp Bot,
-- Notifications, and AI Bridge Schema with Row-Level Security (RLS)
-- ==============================================================================

-- ------------------------------------------------------------------------------
-- 1. School M-Pesa Configurations
-- Each school has its own Safaricom Daraja Paybill/Till shortcode configuration.
-- ------------------------------------------------------------------------------
CREATE TABLE school_mpesa_configs (
    school_id UUID PRIMARY KEY REFERENCES schools(id) ON DELETE CASCADE,
    shortcode VARCHAR(30) UNIQUE NOT NULL,
    environment VARCHAR(20) NOT NULL DEFAULT 'sandbox',
    passkey_encrypted TEXT,
    consumer_key_encrypted TEXT,
    consumer_secret_encrypted TEXT,
    test_payment_verified BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ------------------------------------------------------------------------------
-- 2. Generic Billing Periods (C46, Section 9)
-- Generic billing periods (period_type = 'term' in Release 1) tied to term_id
-- and due_date = terms.start_date so future custom periods need no migration.
-- ------------------------------------------------------------------------------
CREATE TABLE fee_periods (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id UUID NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
    academic_year_id UUID NOT NULL REFERENCES academic_years(id) ON DELETE RESTRICT,
    term_id UUID NOT NULL REFERENCES terms(id) ON DELETE RESTRICT,
    period_type VARCHAR(30) NOT NULL DEFAULT 'term',
    name VARCHAR(100) NOT NULL,
    due_date DATE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (school_id, academic_year_id, term_id)
);

CREATE INDEX idx_fee_periods_school_term ON fee_periods(school_id, term_id);

-- ------------------------------------------------------------------------------
-- 3. Fee Charges per Student and Period (AC-20)
-- Stores individual student billing rows with explicit due_date.
-- ------------------------------------------------------------------------------
CREATE TABLE fee_charges (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id UUID NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
    student_id UUID NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    fee_period_id UUID NOT NULL REFERENCES fee_periods(id) ON DELETE RESTRICT,
    fee_structure_id UUID NOT NULL REFERENCES fee_structures(id) ON DELETE RESTRICT,
    amount NUMERIC(12, 2) NOT NULL,
    due_date DATE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (school_id, student_id, fee_period_id)
);

CREATE INDEX idx_fee_charges_student ON fee_charges(school_id, student_id);

-- ------------------------------------------------------------------------------
-- 4. Mandatory Mid-Year Fee Change & Term Date Notes (C45, C61, AC-21, AC-22)
-- Explains reason for any fee change or due date shift.
-- ------------------------------------------------------------------------------
CREATE TABLE fee_change_notes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id UUID NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
    term_id UUID REFERENCES terms(id) ON DELETE SET NULL,
    fee_period_id UUID REFERENCES fee_periods(id) ON DELETE SET NULL,
    change_type VARCHAR(50) NOT NULL, -- 'mid_year_adjustment', 'term_date_shift', 'structure_update'
    amount_delta NUMERIC(12, 2) NOT NULL DEFAULT 0.00,
    old_due_date DATE,
    new_due_date DATE,
    note TEXT NOT NULL,
    created_by UUID NOT NULL REFERENCES users(id),
    affected_students_count INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_fee_change_notes_school_term ON fee_change_notes(school_id, term_id);

-- ------------------------------------------------------------------------------
-- 5. Ledger Entries (Double-Entry Financial Accounting Ledger)
-- Tracks charges, payments, credit applications, and adjustments.
-- ------------------------------------------------------------------------------
CREATE TABLE ledger_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id UUID NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
    student_id UUID NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    term_id UUID REFERENCES terms(id) ON DELETE SET NULL,
    entry_type VARCHAR(30) NOT NULL CHECK (entry_type IN ('charge', 'payment', 'credit_applied', 'adjustment')),
    amount NUMERIC(12, 2) NOT NULL,
    reference VARCHAR(100),
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_ledger_entries_student ON ledger_entries(school_id, student_id);

-- ------------------------------------------------------------------------------
-- 6. Alter Existing Tables: fee_ledgers & users
-- ------------------------------------------------------------------------------
ALTER TABLE fee_ledgers ADD COLUMN IF NOT EXISTS credit_balance NUMERIC(12, 2) NOT NULL DEFAULT 0.00;
ALTER TABLE fee_ledgers ADD COLUMN IF NOT EXISTS due_date DATE;

-- 24-hour chat payment security pause on guardian phone change (C27, AC-26)
ALTER TABLE users ADD COLUMN IF NOT EXISTS chat_payment_paused_until TIMESTAMPTZ;

-- ------------------------------------------------------------------------------
-- 7. Unmatched Payments Queue (C50, AC-25)
-- Holds C2B Paybill payments with invalid/missing admission numbers.
-- ------------------------------------------------------------------------------
CREATE TABLE unmatched_payments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id UUID NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
    payment_id UUID NOT NULL REFERENCES payments(id) ON DELETE CASCADE,
    transaction_ref VARCHAR(100) NOT NULL,
    amount NUMERIC(12, 2) NOT NULL,
    payer_phone VARCHAR(20),
    payer_name VARCHAR(255),
    raw_account_ref VARCHAR(100),
    status VARCHAR(30) NOT NULL DEFAULT 'unmatched', -- 'unmatched', 'resolved'
    resolved_student_id UUID REFERENCES students(id),
    resolved_by UUID REFERENCES users(id),
    resolution_note TEXT,
    resolved_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_unmatched_payments_school_status ON unmatched_payments(school_id, status);

-- ------------------------------------------------------------------------------
-- 8. School WhatsApp Bots (C16, AC-12)
-- Dedicated WhatsApp Business number mapping per school.
-- ------------------------------------------------------------------------------
CREATE TABLE school_whatsapp_bots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id UUID NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
    phone_number_id VARCHAR(100) UNIQUE NOT NULL,
    display_phone_number VARCHAR(30) UNIQUE NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_school_whatsapp_bots_phone ON school_whatsapp_bots(display_phone_number);

-- ------------------------------------------------------------------------------
-- 9. Chat Sessions & 6-Hour OTP Identity State Machine (C12–C20, AC-08–AC-12)
-- ------------------------------------------------------------------------------
CREATE TABLE chat_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id UUID NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
    channel VARCHAR(20) NOT NULL CHECK (channel IN ('whatsapp', 'in_app')),
    sender_wa_number VARCHAR(30),
    student_id UUID REFERENCES students(id),
    guardian_user_id UUID REFERENCES users(id),
    state VARCHAR(40) NOT NULL DEFAULT 'AWAITING_ADMISSION_NUMBER',
    verified_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ,
    lookup_count INT NOT NULL DEFAULT 0,
    payment_init_count INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_chat_sessions_sender ON chat_sessions(school_id, sender_wa_number);

-- ------------------------------------------------------------------------------
-- 10. WhatsApp Sender Rate Limiting & 24h Blocks (C18–C20, AC-10)
-- ------------------------------------------------------------------------------
CREATE TABLE sender_rate_blocks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id UUID NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
    sender_wa_number VARCHAR(30) NOT NULL,
    blocked_until TIMESTAMPTZ NOT NULL,
    reason VARCHAR(100) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_sender_rate_blocks_lookup ON sender_rate_blocks(sender_wa_number, blocked_until);

CREATE TABLE chat_otp_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id UUID NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
    sender_wa_number VARCHAR(30) NOT NULL,
    admission_number VARCHAR(100) NOT NULL,
    guardian_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    otp_code VARCHAR(20) NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    attempts INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_chat_otp_requests_recent ON chat_otp_requests(school_id, sender_wa_number, created_at);
CREATE INDEX idx_chat_otp_requests_student ON chat_otp_requests(school_id, admission_number, created_at);

-- ------------------------------------------------------------------------------
-- 11. In-App Notifications (C58)
-- ------------------------------------------------------------------------------
CREATE TABLE in_app_notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id UUID NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
    recipient_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    category VARCHAR(50) NOT NULL, -- 'fee_change', 'payment_receipt', 'term_date_shift', 'system'
    title VARCHAR(255) NOT NULL,
    body TEXT NOT NULL,
    metadata JSONB,
    is_read BOOLEAN NOT NULL DEFAULT FALSE,
    read_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_in_app_notifications_user ON in_app_notifications(recipient_user_id, is_read, created_at);

-- ------------------------------------------------------------------------------
-- 12. AI Action Logs (C21, AC-28, SECURITY.md)
-- Immutable audit log of every call from the AI service.
-- ------------------------------------------------------------------------------
CREATE TABLE ai_action_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id UUID NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    session_id UUID,
    channel VARCHAR(20) NOT NULL,
    feature VARCHAR(50) NOT NULL,
    intent VARCHAR(50),
    endpoints_read JSONB,
    action_taken VARCHAR(100),
    metadata JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_ai_action_logs_school ON ai_action_logs(school_id, created_at);

-- ------------------------------------------------------------------------------
-- Row-Level Security on every tenant-owned table (AC-27)
-- ------------------------------------------------------------------------------
DO $$
DECLARE
    tbl TEXT;
    tenant_tables TEXT[] := ARRAY[
        'school_mpesa_configs',
        'fee_periods',
        'fee_charges',
        'fee_change_notes',
        'ledger_entries',
        'unmatched_payments',
        'school_whatsapp_bots',
        'chat_sessions',
        'chat_otp_requests',
        'sender_rate_blocks',
        'in_app_notifications',
        'ai_action_logs'
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
