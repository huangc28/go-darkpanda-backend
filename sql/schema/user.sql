-- user.sql

CREATE TYPE gender AS ENUM (
	'male',
	'female'
);

CREATE TYPE premium_kind AS ENUM (
	'normal',
	'paid'
);

CREATE TABLE IF NOT EXISTS users (
	id BIGSERIAL PRIMARY KEY,
	username VARCHAR UNIQUE NOT NULL,
	phone_verified BOOL DEFAULT false,
	auth_sms_code INT NULL,
	gender gender NOT NULL,
	premium_kind premium_kind DEFAULT 'normal',
	premium_expiry_date timestamp NULL,

	created_at timestamp NOT NULL DEFAULT NOW(),
	updated_at timestamp NULL DEFAULT current_timestamp,
	deleted_at timestamp
);
