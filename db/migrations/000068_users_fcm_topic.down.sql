BEGIN;

ALTER TABLE users 
DROP COLUMN IF EXISTS fcm_topic; 

COMMIT;