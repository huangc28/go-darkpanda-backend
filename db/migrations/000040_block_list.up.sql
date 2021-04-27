BEGIN;

CREATE TABLE block_list(
	id SERIAL PRIMARY KEY,
	user_id INT REFERENCES users (id) NOT NULL,
	blocked_user_id INT REFERENCES users (id) NOT NULL,

	created_at timestamp NOT NULL DEFAULT NOW(),
	updated_at timestamp NULL DEFAULT current_timestamp,
	deleted_at timestamp
);


COMMIT;
