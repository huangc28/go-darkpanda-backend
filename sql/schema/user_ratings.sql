-- user_ratings.sql

CREATE TABLE IF NOT EXISTS user_ratings  (
	id BIGSERIAL PRIMARY KEY,
	from_user_id INT REFERENCES users(id),
	to_user_id INT REFERENCES users(id),
	rating INT,
	comments VARCHAR,

	created_at timestamp NOT NULL DEFAULT NOW(),
	updated_at timestamp NULL DEFAULT current_timestamp,
	deleted_at timestamp,

	CONSTRAINT user_ratings_fk PRIMARY KEY(from_user_id, to_user_id)
);
