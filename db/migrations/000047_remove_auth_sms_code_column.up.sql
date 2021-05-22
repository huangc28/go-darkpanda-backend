BEGIN;

ALTER TABLE users
DROP COLUMN auth_sms_code;

COMMIT;
