BEGIN;

ALTER TABLE services
ADD COLUMN start_time timestamp,
ADD COLUMN end_time timestamp;

COMMIT;
