BEGIN;

	ALTER TABLE service_ratings
	DROP COLUMN ratee_id;

COMMIT;
