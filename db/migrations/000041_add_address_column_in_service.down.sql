BEGIN;

ALTER TABLE services
ALTER COLUMN address DROP NOT NULL;

COMMIT;
