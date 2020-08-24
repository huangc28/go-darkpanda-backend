BEGIN;

ALTER TABLE services
ADD COLUMN budget numeric(12, 2);

ALTER TABLE services
ALTER COLUMN price TYPE numeric(12, 2);

COMMIT;
