BEGIN;

CREATE TABLE images (
	id BIGSERIAL PRIMARY KEY,
	user_id INT REFERENCES users (id) NOT NULL,
	url VARCHAR(255) NOT NULL,

	created_at timestamp NOT NULL DEFAULT NOW(),
	updated_at timestamp NULL DEFAULT current_timestamp,
	deleted_at timestamp
);

COMMIT;
