BEGIN; 

ALTER TABLE coin_packages
ALTER COLUMN cost TYPE NUMERIC(12, 2);

COMMIT;