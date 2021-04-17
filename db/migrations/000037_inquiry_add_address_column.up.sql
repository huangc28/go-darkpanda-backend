BEGIN;

ALTER TABLE service_inquiries
ADD COLUMN address varchar(255);

COMMIT;
