BEGIN;

ALTER TABLE service_inquiries
DROP COLUMN IF EXISTS inquiry_type;

DROP TYPE inquiry_type;

COMMIT;