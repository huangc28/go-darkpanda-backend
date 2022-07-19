BEGIN;

ALTER TABLE IF EXISTS service_inquiries
DROP COLUMN IF EXISTS inquiry_type;

DROP TYPE inquiry_type;

COMMIT;
