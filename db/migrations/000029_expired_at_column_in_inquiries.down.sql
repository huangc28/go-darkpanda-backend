BEGIN;

ALTER TABLE IF EXISTS service_inquiries
DROP COLUMN expired_at;

COMMIT;
