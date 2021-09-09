BEGIN;

ALTER TABLE service_ratings 
ADD COLUMN ratee_id INT REFERENCES users(id);

COMMIT;