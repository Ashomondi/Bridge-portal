BEGIN;

DROP INDEX IF EXISTS mpesa_stk_requests_unscheduled_poll_idx;
DROP INDEX IF EXISTS mpesa_stk_requests_idempotency_idx;

ALTER TABLE mpesa_stk_requests
    DROP CONSTRAINT IF EXISTS r1_mpesa_stk_idempotency_pair,
    DROP CONSTRAINT IF EXISTS r1_mpesa_stk_callback_token_hash_key,
    DROP COLUMN IF EXISTS poll_scheduled,
    DROP COLUMN IF EXISTS callback_token_hash,
    DROP COLUMN IF EXISTS request_hash,
    DROP COLUMN IF EXISTS idempotency_key;

ALTER TABLE users
    DROP COLUMN IF EXISTS chat_payment_paused_until;

COMMIT;
