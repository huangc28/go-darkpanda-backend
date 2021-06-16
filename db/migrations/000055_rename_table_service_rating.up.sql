BEGIN;

ALTER TABLE service_rating
RENAME TO service_ratings;

COMMIT;
