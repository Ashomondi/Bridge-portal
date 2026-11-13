BEGIN;

-- Composite keys let new foreign keys enforce tenant ownership as well as RLS.
ALTER TABLE academic_years ADD CONSTRAINT r1_c_year_school_id UNIQUE (school_id, id);
ALTER TABLE terms ADD CONSTRAINT r1_c_term_school_year_id UNIQUE (school_id, academic_year_id, id);
ALTER TABLE terms ADD CONSTRAINT r1_c_term_school_id UNIQUE (school_id, id);
ALTER TABLE students ADD CONSTRAINT r1_c_student_school_id UNIQUE (school_id, id);
ALTER TABLE fee_structures DROP CONSTRAINT fee_structures_school_id_academic_year_id_grade_level_id_te_key,
    ADD COLUMN boarding_status VARCHAR(20) NOT NULL DEFAULT 'all'
        CHECK (boarding_status IN ('all', 'boarder', 'day')),
    ADD CONSTRAINT r1_c_structure_school_id UNIQUE (school_id, id),
    ADD CONSTRAINT r1_c_structure_variant UNIQUE (school_id, academic_year_id, grade_level_id, term_id, boarding_status);
CREATE INDEX fee_structure_items_structure_idx ON fee_structure_items(fee_structure_id, lower(title), id);
ALTER TABLE payments ADD CONSTRAINT r1_c_payment_school_id UNIQUE (school_id, id);
ALTER TABLE fee_ledgers ADD COLUMN credit_balance NUMERIC(12,2) NOT NULL DEFAULT 0.00,
    ADD COLUMN credit_applied NUMERIC(12,2) NOT NULL DEFAULT 0.00,
    ADD COLUMN credit_transferred NUMERIC(12,2) NOT NULL DEFAULT 0.00,
    ADD COLUMN due_date DATE;

-- Preserve legacy overpayments before enforcing nonnegative receivables.
-- Keep data changes tenant-scoped even when run by the migration owner.
DO $$
DECLARE tenant UUID;
BEGIN
    FOR tenant IN SELECT id FROM schools LOOP
        PERFORM set_config('app.school_id', tenant::text, true);
        UPDATE fee_ledgers SET credit_balance = -current_balance, current_balance = 0
        WHERE school_id = tenant AND current_balance < 0;
    END LOOP;
END $$;
ALTER TABLE fee_ledgers ADD CONSTRAINT r1_c_ledger_nonnegative
    CHECK (current_balance >= 0 AND credit_balance >= 0 AND credit_applied >= 0 AND credit_transferred >= 0);

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
    CHECK (state <> 'VERIFIED' OR (guardian_user_id IS NOT NULL AND verified_at IS NOT NULL AND expires_at IS NOT NULL AND expires_at > verified_at)),
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
