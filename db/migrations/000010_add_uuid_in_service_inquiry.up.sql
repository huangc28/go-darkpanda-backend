BEGIN;

ALTER TABLE service_inquiries
ADD COLUMN uuid VARCHAR(40) UNIQUE NOT NULL;

COMMIT;
