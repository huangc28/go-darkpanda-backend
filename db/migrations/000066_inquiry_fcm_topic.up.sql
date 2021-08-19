BEGIN;

ALTER TABLE service_inquiries
ADD COLUMN fcm_topic VARCHAR;

COMMIT;