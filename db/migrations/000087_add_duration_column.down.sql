BEGIN;

ALTER TABLE service_options 
DROP COLUMN IF EXISTS duration INT;

COMMIT;