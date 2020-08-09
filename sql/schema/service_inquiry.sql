-- service_inquiries.sql

CREATE TYPE inquiry_status AS ENUM (
	'inquiring',
	'canceled',
	'expired',
	'booked'
);

CREATE TABLE IF NOT EXISTS service_inquiries (
	id BIGSERIAL PRIMARY KEY,
	inquirer_id INT REFERENCES users (id)  ON DELETE CASCADE,
	budget FLOAT NOT NULL,
	service_type service_type NOT NULL,
	inquiry_status inquiry_status NOT NULL,

	created_at timestamp NOT NULL DEFAULT NOW(),
	updated_at timestamp NULL DEFAULT current_timestamp,
	deleted_at timestamp,

	CONSTRAINT inquirer_id_fk PRIMARY KEY (inquirer_id)
)
