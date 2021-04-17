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
	inquirer_id INT REFERENCES users (id) ON DELETE CASCADE,
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
BEGIN;

ALTER TABLE service_inquiries
ADD COLUMN uuid VARCHAR(40) UNIQUE NOT NULL;

COMMIT;
BEGIN;

ALTER TABLE services
ALTER COLUMN price DROP NOT NULL;

ALTER TABLE services
ALTER COLUMN duration DROP NOT NULL;

ALTER TABLE services
ALTER COLUMN appointment_time DROP NOT NULL;

ALTER TABLE services
ALTER COLUMN lng DROP NOT NULL;

ALTER TABLE services
ALTER COLUMN lat DROP NOT NULL;

COMMIT;
BEGIN;

ALTER TABLE service_inquiries
ALTER COLUMN budget TYPE numeric(12, 2);

COMMIT;
BEGIN;

ALTER TABLE services
ADD COLUMN budget numeric(12, 2);

ALTER TABLE services
ALTER COLUMN price TYPE numeric(12, 2);

COMMIT;
BEGIN;

ALTER TABLE services
ADD COLUMN inquiry_id INT NOT NULL;

ALTER TABLE services
   ADD CONSTRAINT fk_inquiry_id
   FOREIGN KEY (inquiry_id)
   REFERENCES service_inquiries(id);

COMMIT;
BEGIN;

DROP TYPE IF EXISTS service_status;

CREATE TYPE service_status AS ENUM  (
	'unpaid',
	'to_be_fulfilled',
	'canceled',
	'failed_due_to_both',
	'girl_waiting',
	'fufilling',
	'failed_due_to_girl',
	'failed_due_to_man',
	'completed'
);

ALTER TABLE services
ADD COLUMN service_status service_status NOT NULL DEFAULT 'unpaid';

COMMIT;
ALTER TYPE inquiry_status ADD VALUE 'chatting';
ALTER TYPE inquiry_status ADD VALUE 'wait_for_inquirer_approve';
BEGIN;

ALTER TABLE service_inquiries
ADD COLUMN price numeric(12, 2),
ADD COLUMN duration int,
ADD COLUMN appointment_time timestamp,
ADD COLUMN lng numeric(17, 8),
ADD COLUMN lat numeric(17, 8);

COMMIT;

BEGIN;

CREATE TABLE images (
	id BIGSERIAL PRIMARY KEY,
	user_id INT REFERENCES users (id) NOT NULL,
	url VARCHAR(255) NOT NULL,

	created_at timestamp NOT NULL DEFAULT NOW(),
	updated_at timestamp NULL DEFAULT current_timestamp,
	deleted_at timestamp
);

COMMIT;
BEGIN;

ALTER TABLE users
ADD COLUMN avatar_url varchar(255);

COMMIT;
BEGIN;

ALTER TABLE users

ADD COLUMN nationality varchar(255),
ADD COLUMN region varchar(255),
ADD COLUMN age INT,
ADD COLUMN height numeric(5, 2),
ADD COLUMN weight numeric(5, 2),
ADD COLUMN habbits varchar(40),
ADD COLUMN description varchar(255),
ADD COLUMN breast_size varchar(40),
ADD CONSTRAINT breast_size_regex CHECK (breast_size ~ '^[a-zA-Z]$');

COMMIT;
BEGIN;

ALTER TABLE users
ALTER COLUMN uuid TYPE varchar(60);

COMMIT;
BEGIN;

ALTER TABLE users
ALTER COLUMN phone_verified SET NOT NULL;

COMMIT;
BEGIN;
	ALTER TABLE users
	ADD COLUMN mobile varchar(20);
COMMIT;
BEGIN;

ALTER TABLE IF EXISTS payment
RENAME TO payments;

COMMIT;
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
BEGIN;

CREATE TABLE IF NOT EXISTS lobby_users (
	id BIGSERIAL PRIMARY KEY,
	channel_uuid VARCHAR(255) NOT NULL,
	inquiry_id INT REFERENCES service_inquiries(id) NOT NULL,

	created_at timestamp NOT NULL DEFAULT NOW(),
	updated_at timestamp NULL DEFAULT current_timestamp,
	deleted_at timestamp
);

COMMIT;
BEGIN;

ALTER TABLE service_inquiries
ADD COLUMN expired_at timestamp;

COMMENT ON COLUMN service_inquiries.expired_at IS 'Time that this inquiry will be invalid.';

COMMIT;
BEGIN;

ALTER TABLE service_inquiries 
ADD COLUMN picker_id INT REFERENCES users(id) NULL;

COMMIT;ALTER TYPE service_status ADD VALUE 'negotiating';BEGIN;

ALTER TABLE user_refcodes
ADD COLUMN expired_at timestamp
DEFAULT NOW() + interval '3 days';


COMMENT ON COLUMN user_refcodes.expired_at IS 'Time that this referral code will be invalid.';

COMMIT;
-- Add default timestamp value in minute: https://stackoverflow.com/questions/21745125/add-minutes-to-current-timestamp-in-postgresql
BEGIN;

CREATE TYPE lobby_status AS ENUM (
	'waiting',
	'pause',
	'expired',
	'left'
);

ALTER TABLE lobby_users
ADD COLUMN expired_at timestamp NOT NULL DEFAULT current_timestamp + (26 * interval '1 minute'),
ADD COLUMN lobby_status lobby_status NOT NULL DEFAULT 'waiting';

COMMIT;
ALTER TYPE lobby_status ADD VALUE 'asking';
    ALTER TYPE inquiry_status ADD VALUE 'asking';
CREATE TYPE chatroom_type AS ENUM(
	'inquiry_chat',
	'service_chat'
);

BEGIN;

ALTER TABLE chatrooms
ADD COLUMN chatroom_type chatroom_type NOT NULL DEFAULT 'inquiry_chat';

COMMIT;
BEGIN;

ALTER TABLE service_inquiries
ADD COLUMN address varchar(255);

COMMIT;
