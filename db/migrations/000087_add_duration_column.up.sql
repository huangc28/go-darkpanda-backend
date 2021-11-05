BEGIN;

ALTER TABLE service_options 
ADD COLUMN duration INT;

COMMIT;