BEGIN;

ALTER TABLE IF EXISTS block_list
	DROP CONSTRAINT IF EXISTS block_list_user_id_fkey;

DROP TABLE IF EXISTS block_list;

COMMIT;
