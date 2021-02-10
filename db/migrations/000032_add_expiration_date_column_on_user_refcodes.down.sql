BEGIN;

ALTER TABLE user_refcodes
DROP COLUMN expired_at;

COMMIT;