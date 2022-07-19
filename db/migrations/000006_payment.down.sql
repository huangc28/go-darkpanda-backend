BEGIN;

ALTER TABLE IF EXISTS payments
	DROP CONSTRAINT IF EXISTS block_list_user_id,
	DROP CONSTRAINT IF EXISTS block_list_blocked_user_id,
	DROP CONSTRAINT payment_service_id_fkey;

DROP TABLE IF EXISTS payments;

COMMIT;
