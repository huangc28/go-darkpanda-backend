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
-- service_inquiries.sql
BEGIN;

DROP TABLE IF EXISTS service_inquiries;
DROP TYPE IF EXISTS inquiry_status;
DROP TYPE IF EXISTS service_type;

CREATE TYPE inquiry_status AS ENUM (
	'inquiring',
	'canceled',
	'expired',
	'booked'
);

CREATE TYPE service_type AS ENUM (
	'sex',
	'diner',
	'movie',
	'shopping',
	'chat'
);

CREATE TABLE IF NOT EXISTS service_inquiries (
	id BIGSERIAL PRIMARY KEY,
	inquirer_id INT REFERENCES users (id)  ON DELETE CASCADE,
	budget FLOAT NOT NULL,
	service_type service_type NOT NULL,
	inquiry_status inquiry_status NOT NULL,

	created_at timestamp NOT NULL DEFAULT NOW(),
	updated_at timestamp NULL DEFAULT current_timestamp,
	deleted_at timestamp
);

COMMIT;
BEGIN;

CREATE extension IF NOT EXISTS "uuid-ossp";

COMMIT;
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
BEGIN;

DROP TYPE IF EXISTS ref_code_type;

CREATE TYPE ref_code_type AS ENUM (
	'invitor',
	'manager'
);

CREATE TABLE IF NOT EXISTS user_refcodes (
	id BIGSERIAL PRIMARY KEY,
	invitor_id INT REFERENCES users(id) NOT NULL,
	invitee_id INT REFERENCES users(id),
	ref_code VARCHAR NOT NULL,
	ref_code_type ref_code_type NOT NULL,

	created_at timestamp NOT NULL DEFAULT NOW(),
	updated_at timestamp NULL DEFAULT current_timestamp,
	deleted_at timestamp
);


COMMIT;
BEGIN;

CREATE TABLE payment (
	id BIGSERIAL PRIMARY KEY,
	payer_id INT REFERENCES users(id) NOT NULL,
	payee_id INT REFERENCES users(id) NOT NULL,
	service_id INT REFERENCES services(id)  NOT NULL,
	price DECIMAL(12, 2) NOT NULL,
	rec_trade_id VARCHAR,

	created_at timestamp NOT NULL DEFAULT NOW(),
	updated_at timestamp NULL DEFAULT current_timestamp,
	deleted_at timestamp

);

COMMIT;
-- user_ratings.sql

BEGIN;

CREATE TABLE IF NOT EXISTS user_ratings (
	id BIGSERIAL PRIMARY KEY,
	from_user_id INT REFERENCES users(id),
	to_user_id INT REFERENCES users(id),
	rating INT,
	comments VARCHAR,

	created_at timestamp NOT NULL DEFAULT NOW(),
	updated_at timestamp NULL DEFAULT current_timestamp,
	deleted_at timestamp
);

COMMIT;
BEGIN;

ALTER TABLE users
ADD COLUMN uuid VARCHAR(20) UNIQUE NOT NULL;

COMMIT;
BEGIN;

ALTER TABLE users
ADD COLUMN phone_verify_code VARCHAR(20);

COMMIT;
