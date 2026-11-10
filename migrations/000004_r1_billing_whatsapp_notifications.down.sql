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

DROP INDEX fee_structure_items_structure_idx;
ALTER TABLE fee_ledgers DROP CONSTRAINT r1_c_ledger_nonnegative,
    DROP COLUMN credit_balance, DROP COLUMN credit_applied, DROP COLUMN credit_transferred, DROP COLUMN due_date;
ALTER TABLE payments DROP CONSTRAINT r1_c_payment_school_id;
ALTER TABLE fee_structures DROP CONSTRAINT r1_c_structure_variant,
    DROP CONSTRAINT r1_c_structure_school_id,
    DROP COLUMN boarding_status,
    ADD CONSTRAINT fee_structures_school_id_academic_year_id_grade_level_id_te_key UNIQUE (school_id, academic_year_id, grade_level_id, term_id);
ALTER TABLE students DROP CONSTRAINT r1_c_student_school_id;
ALTER TABLE terms DROP CONSTRAINT r1_c_term_school_id;
ALTER TABLE terms DROP CONSTRAINT r1_c_term_school_year_id;
ALTER TABLE academic_years DROP CONSTRAINT r1_c_year_school_id;

COMMIT;
