BEGIN;

ALTER TABLE service_inquiries
DROP COLUMN address;

COMMIT;
