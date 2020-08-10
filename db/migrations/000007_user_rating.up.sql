-- user_ratings.sql

BEGIN;

CREATE TABLE IF NOT EXISTS user_ratings (
	id BIGSERIAL PRIMARY KEY,
	from_user_id INT REFERENCES users(id),
	to_user_id INT REFERENCES users(id),
	rating INT,
	comments VARCHAR,

	created_at timestamp NOT NULL DEFAULT NOW(),
	updated_at timestamp NULL DEFAULT current_timestamp,
	deleted_at timestamp
);

COMMIT;
