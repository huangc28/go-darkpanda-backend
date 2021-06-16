BEGIN;

CREATE TABLE service_rating (
	id BIGSERIAL PRIMARY KEY,

	rater_id INT REFERENCES users(id),
	ratee_id INT REFERENCES users(id),
	service_id INT REFERENCES services(id) ON DELETE CASCADE,
	rating INT DEFAULT 0,

	created_at timestamp NOT NULL DEFAULT NOW(),
	updated_at timestamp NULL DEFAULT current_timestamp,
	deleted_at timestamp
);

COMMIT;
