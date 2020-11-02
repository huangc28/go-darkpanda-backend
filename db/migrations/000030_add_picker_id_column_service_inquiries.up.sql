BEGIN;

ALTER TABLE service_inquiries 
ADD COLUMN picker_id INT REFERENCES users(id) NULL;

COMMIT;