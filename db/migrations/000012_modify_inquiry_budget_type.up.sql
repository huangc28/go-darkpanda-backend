BEGIN;

ALTER TABLE service_inquiries
ALTER COLUMN budget TYPE numeric(12, 2);

COMMIT;
