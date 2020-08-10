BEGIN;

DROP TABLE IF EXISTS users;
DROP TYPE IF EXISTS gender;
DROP TYPE IF EXISTS premium_type;

CREATE TYPE gender AS ENUM (
	'male',
	'female'
);

CREATE TYPE premium_type AS ENUM (
	'normal',
	'paid'
);

CREATE TABLE users (
	id BIGSERIAL PRIMARY KEY,

	username VARCHAR UNIQUE NOT NULL,
	phone_verified BOOL DEFAULT false,
	auth_sms_code INT NULL,
	gender gender NOT NULL,
	premium_type premium_type DEFAULT 'normal',
	premium_expiry_date timestamp NULL,

	created_at timestamp NOT NULL DEFAULT NOW(),
	updated_at timestamp NULL DEFAULT current_timestamp,
	deleted_at timestamp
);

COMMIT;
