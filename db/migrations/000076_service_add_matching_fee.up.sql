BEGIN;

ALTER TABLE services
ADD COLUMN IF NOT EXISTS matching_fee numeric(12, 2) default 0;

COMMIT;