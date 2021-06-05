BEGIN;

CREATE TABLE service_types (
	id BIGSERIAL PRIMARY KEY,
	service_name service_type,

	created_at timestamp NOT NULL DEFAULT NOW(),
	updated_at timestamp NULL DEFAULT current_timestamp,
	deleted_at timestamp
);

COMMIT;
