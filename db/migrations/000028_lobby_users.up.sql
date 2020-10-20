BEGIN;

CREATE TABLE IF NOT EXISTS lobby_users (
	id BIGSERIAL PRIMARY KEY,
	channel_uuid VARCHAR(255) NOT NULL,
	inquiry_id INT REFERENCES service_inquiries(id) NOT NULL,

	created_at timestamp NOT NULL DEFAULT NOW(),
	expired_at timestamp NOT NULL DEFAULT NOW(),
	updated_at timestamp NULL DEFAULT current_timestamp,
	deleted_at timestamp
);

COMMIT;
