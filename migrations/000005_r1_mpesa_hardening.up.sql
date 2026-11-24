BEGIN;

-- Hardening for Safaricom Daraja STK Push:
-- 1. Scoped idempotency key + request payload hash to prevent duplicate STK prompts on retry.
-- 2. Per-request SHA-256 callback token hash for authenticated, server-bound STK callbacks.
-- 3. Durable poll_scheduled flag so failed poll-queue publications can be safely recovered.
--
-- users.chat_payment_paused_until is owned by 000002_r1_core_onboarding, which
-- always runs earlier, so this migration must not add or drop it.
ALTER TABLE mpesa_stk_requests
    ADD COLUMN idempotency_key VARCHAR(128),
    ADD COLUMN request_hash VARCHAR(64),
    ADD COLUMN callback_token_hash VARCHAR(64),
    ADD COLUMN poll_scheduled BOOLEAN NOT NULL DEFAULT FALSE,
    ADD CONSTRAINT r1_mpesa_stk_callback_token_hash_key UNIQUE (callback_token_hash),
    ADD CONSTRAINT r1_mpesa_stk_idempotency_pair CHECK (
        (idempotency_key IS NULL AND request_hash IS NULL)
        OR (idempotency_key IS NOT NULL AND btrim(idempotency_key) <> ''
            AND request_hash IS NOT NULL AND btrim(request_hash) <> '')
    );

CREATE UNIQUE INDEX mpesa_stk_requests_idempotency_idx
    ON mpesa_stk_requests (school_id, guardian_user_id, idempotency_key)
    WHERE idempotency_key IS NOT NULL;

CREATE INDEX mpesa_stk_requests_unscheduled_poll_idx
    ON mpesa_stk_requests (school_id, created_at, id)
    WHERE status = 'pending' AND poll_scheduled = FALSE;

COMMIT;
