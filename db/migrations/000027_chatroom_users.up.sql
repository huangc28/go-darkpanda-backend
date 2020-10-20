BEGIN;

CREATE TABLE IF NOT EXISTS chatroom_users (
	id BIGSERIAL PRIMARY KEY,
	chatroom_id INT REFERENCES chatrooms (id) NOT NULL,
	user_id INT REFERENCES users (id) NOT NULL,

	created_at timestamp NOT NULL DEFAULT NOW(),
	updated_at timestamp NULL DEFAULT current_timestamp,
	deleted_at timestamp
);

COMMIT;
