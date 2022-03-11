BEGIN;

ALTER TABLE service_inquiries
DROP COLUMN IF EXISTS currency;

COMMIT;
