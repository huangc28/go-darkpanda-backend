BEGIN;

ALTER TABLE service_inquiries
DROP COLUMN expired_at;

COMMIT;
