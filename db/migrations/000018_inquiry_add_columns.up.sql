BEGIN;

ALTER TABLE service_inquiries
ADD COLUMN price numeric(12, 2),
ADD COLUMN duration int,
ADD COLUMN appointment_time timestamp,
ADD COLUMN lng numeric(17, 8),
ADD COLUMN lat numeric(17, 8);

COMMIT;

