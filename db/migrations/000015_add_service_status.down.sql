BEGIN;

DROP TYPE IF EXISTS service_status;

ALTER TABLE services
DROP COLUMN IF EXISTS service_status;

COMMIT;
