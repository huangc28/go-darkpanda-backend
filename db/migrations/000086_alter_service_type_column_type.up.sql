BEGIN;

ALTER TABLE services
ALTER COLUMN service_type TYPE text; 

COMMIT;