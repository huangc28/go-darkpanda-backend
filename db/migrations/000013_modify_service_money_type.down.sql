BEGIN;

ALTER TABLE services
DROP COLUMN IF EXISTS budget;

COMMIT;
