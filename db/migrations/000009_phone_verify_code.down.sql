BEGIN;

ALTER TABLE users
DROP COLUMN IF EXISTS phone_verify_code;

COMMIT;
