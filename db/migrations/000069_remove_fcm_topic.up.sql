BEGIN;

ALTER TABLE service_inquiries 
DROP COLUMN IF EXISTS fcm_topic;

COMMIT;