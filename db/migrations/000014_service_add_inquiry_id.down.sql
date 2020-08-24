BEGIN;

ALTER TABLE services
DROP COLUMN inquiry_id;

COMMIT;
