BEGIN;

DROP TABLE IF EXISTS services;

CREATE TABLE IF NOT EXISTS services (
	id BIGSERIAL PRIMARY KEY,
	uuid UUID DEFAULT uuid_generate_v4(),
	customer_id INT REFERENCES users (id) ON DELETE CASCADE,
	service_provider_id INT REFERENCES users(id)  ON DELETE CASCADE,
	price FLOAT NOT NULL,
	duration INT NOT NULL,
	appointment_time timestamp NOT NULL,
	lng NUMERIC(17, 8) NOT NULL,
	lat NUMERIC(17, 8) NOT NULL,
	service_type service_type NOT NULL,
	girl_ready BOOLEAN DEFAULT false,
	man_ready BOOLEAN DEFAULT false,

	created_at timestamp NOT NULL DEFAULT NOW(),
	updated_at timestamp NULL DEFAULT current_timestamp,
	deleted_at timestamp
);

COMMIT;
