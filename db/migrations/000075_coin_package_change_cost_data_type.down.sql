BEGIN;

ALTER TABLE coin_packages
ALTER COLUMN cost TYPE int;

COMMIT;