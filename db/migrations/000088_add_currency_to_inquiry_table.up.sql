BEGIN;

ALTER TABLE service_inquiries
ADD COLUMN currency TEXT;

COMMIT;

