BEGIN;

CREATE TABLE IF NOT EXISTS chatrooms (
	id BIGSERIAL PRIMARY KEY,
	inquiry_id INT REFERENCES service_inquiries (id) NOT NULL,
	channel_uuid VARCHAR(255),
	message_count INT,
	enabled BOOLEAN DEFAULT true,

	created_at timestamp NOT NULL DEFAULT NOW(),
	expired_at timestamp NOT NULL DEFAULT NOW(),
	updated_at timestamp NULL DEFAULT current_timestamp,
	deleted_at timestamp
);

COMMIT;
