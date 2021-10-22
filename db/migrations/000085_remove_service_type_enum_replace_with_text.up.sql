BEGIN;

ALTER TABLE service_inquiries
DROP COLUMN service_type,
ADD COLUMN expect_service_type text;

COMMIT;