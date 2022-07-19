BEGIN;

ALTER TABLE IF EXISTS service_inquiries
DROP COLUMN address;

COMMIT;
