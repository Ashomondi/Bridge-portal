-- ==============================================================================
-- 000004_r1_billing_whatsapp_notifications.down.sql
-- Rollback for Release 1 Billing, Daraja M-Pesa, WhatsApp Bot, Notifications,
-- and AI Bridge Schema
-- ==============================================================================

DROP TABLE IF EXISTS ai_action_logs CASCADE;
DROP TABLE IF EXISTS in_app_notifications CASCADE;
DROP TABLE IF EXISTS chat_otp_requests CASCADE;
DROP TABLE IF EXISTS sender_rate_blocks CASCADE;
DROP TABLE IF EXISTS chat_sessions CASCADE;
DROP TABLE IF EXISTS school_whatsapp_bots CASCADE;
DROP TABLE IF EXISTS unmatched_payments CASCADE;
DROP TABLE IF EXISTS ledger_entries CASCADE;
DROP TABLE IF EXISTS fee_change_notes CASCADE;
DROP TABLE IF EXISTS fee_charges CASCADE;
DROP TABLE IF EXISTS fee_periods CASCADE;
DROP TABLE IF EXISTS school_mpesa_configs CASCADE;

ALTER TABLE fee_ledgers DROP COLUMN IF EXISTS credit_balance;
ALTER TABLE fee_ledgers DROP COLUMN IF EXISTS due_date;
ALTER TABLE users DROP COLUMN IF EXISTS chat_payment_paused_until;
