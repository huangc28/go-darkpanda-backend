BEGIN;

ALTER TABLE service_options
DROP COLUMN IF EXISTS service_options_type;


DROP TYPE IF EXISTS service_options_type;

COMMIT;