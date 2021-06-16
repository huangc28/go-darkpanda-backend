BEGIN;

ALTER TABLE service_rating
ADD COLUMN comments text;

COMMIT;
