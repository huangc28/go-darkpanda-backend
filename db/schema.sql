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
BEGIN;

CREATE TYPE verify_status AS ENUM (
	'pending',
	'verifying',
	'verified',
	'verify_failed'
);

CREATE TABLE bank_accounts(
	id SERIAL PRIMARY KEY,
	user_id INT REFERENCES users (id) NOT NULL,

	bank_name VARCHAR(255) NOT NULL,
	branch VARCHAR(255) NOT NULL,
	account_number VARCHAR(255) NOT NULL,
	verify_status verify_status  NOT NULL DEFAULT 'pending',

	created_at timestamp NOT NULL DEFAULT NOW(),
	updated_at timestamp NULL DEFAULT current_timestamp,
	deleted_at timestamp
);

COMMIT;
BEGIN;

CREATE TYPE order_status AS ENUM (
	'init',
	'ordering',
	'success',
	'failed'
);

CREATE TABLE coin_orders(
	id SERIAL PRIMARY KEY,
	buyer_id INT REFERENCES users(id) NOT NULL,
	amount numeric(12, 2) NOT NULL,
	cost numeric(12, 2) NOT NULL,
	order_status order_status NOT NULL DEFAULT 'init',

	created_at timestamp NOT NULL DEFAULT NOW(),
	updated_at timestamp NULL DEFAULT current_timestamp,
	deleted_at timestamp
);

COMMENT ON COLUMN coin_orders.amount IS 'amount of DP coins to buy';
COMMENT ON COLUMN coin_orders.cost IS 'cost to buy, currency in TWD';

COMMIT;
BEGIN;

CREATE TABLE block_list(
	id SERIAL PRIMARY KEY,
	user_id INT REFERENCES users (id) NOT NULL,
	blocked_user_id INT REFERENCES users (id) NOT NULL,

	created_at timestamp NOT NULL DEFAULT NOW(),
	updated_at timestamp NULL DEFAULT current_timestamp,
	deleted_at timestamp
);


COMMIT;
BEGIN;

ALTER TABLE services
ADD COLUMN address VARCHAR(500);

COMMIT;
BEGIN;

CREATE TABLE coin_packages (
	id BIGSERIAL PRIMARY KEY,
	db_coins INT,
	cost INT,
	currency varchar(10) DEFAULT 'TWD'
);

COMMIT;
BEGIN;

CREATE TABLE user_balance (
	id BIGSERIAL PRIMARY KEY,
	user_id INT REFERENCES users (id) NOT NULL,
	balance numeric(12, 2) NOT NULL DEFAULT 0,

	created_at timestamp NOT NULL DEFAULT NOW(),
	updated_at timestamp NULL DEFAULT current_timestamp,
	deleted_at timestamp
);

COMMENT ON COLUMN user_balance.balance IS 'use for update when reading the column';

COMMIT;
BEGIN;

ALTER TABLE coin_orders
DROP COLUMN amount,
ADD COLUMN package_id INT REFERENCES coin_packages(id),
ADD COLUMN quantity INT NOT NULL DEFAULT 1;


COMMIT;
BEGIN;

ALTER TABLE coin_orders
ADD COLUMN rec_trade_id VARCHAR(255),
ADD COLUMN raw text;

COMMIT;
ALTER TABLE user_balance ADD CONSTRAINT uniq_user_balance_user_id UNIQUE (user_id);
BEGIN;

ALTER TABLE users
DROP COLUMN auth_sms_code;

COMMIT;
ALTER TABLE services ALTER COLUMN service_status TYPE varchar(255);

DROP TYPE IF EXISTS service_status CASCADE;

CREATE TYPE service_status AS ENUM  (
	'unpaid',
	'payment_failed',
	'to_be_fulfilled',
	'canceled',
	'expired',
	'fulfilling',
	'completed'
);

ALTER TABLE services ALTER COLUMN service_status TYPE service_status USING service_status::text::service_status;

BEGIN;

CREATE TABLE service_qrcode (
	id BIGSERIAL PRIMARY KEY,
	service_id INT REFERENCES services (id) NOT NULL,
	uuid VARCHAR(255) UNIQUE,
	url VARCHAR(255)
);

COMMIT;
BEGIN;

ALTER TABLE services
ADD COLUMN start_time timestamp,
ADD COLUMN end_time timestamp;

COMMIT;
BEGIN;

CREATE TABLE service_names (
	id BIGSERIAL PRIMARY KEY,
	service_name service_type,

	created_at timestamp NOT NULL DEFAULT NOW(),
	updated_at timestamp NULL DEFAULT current_timestamp,
	deleted_at timestamp
);

COMMIT;
DROP TABLE IF EXISTS user_ratings;
BEGIN;

CREATE TABLE service_rating (
	id BIGSERIAL PRIMARY KEY,

	rater_id INT REFERENCES users(id),
	ratee_id INT REFERENCES users(id),
	service_id INT REFERENCES services(id) ON DELETE CASCADE,
	rating INT DEFAULT 0,

	created_at timestamp NOT NULL DEFAULT NOW(),
	updated_at timestamp NULL DEFAULT current_timestamp,
	deleted_at timestamp
);

COMMIT;
BEGIN;

ALTER TABLE service_rating
ADD COLUMN comments text;

COMMIT;
BEGIN;

ALTER TABLE service_rating
RENAME TO service_ratings;

COMMIT;
BEGIN;

	ALTER TABLE service_ratings
	DROP COLUMN ratee_id;

COMMIT;
BEGIN;

	ALTER TABLE coin_packages
	ADD COLUMN name varchar(255);

COMMIT;
BEGIN;
	ALTER TABLE payments
	DROP COLUMN IF EXISTS payee_id,
	DROP COLUMN IF EXISTS rec_trade_id;
COMMIT;
BEGIN;
	ALTER TABLE services
	DROP COLUMN IF EXISTS lng,
	DROP COLUMN IF EXISTS lat,
	DROP COLUMN IF EXISTS girl_ready,
	DROP COLUMN IF EXISTS man_ready;
COMMIT;
BEGIN;

    ALTER TABLE services 
    ALTER COLUMN uuid TYPE VARCHAR;

COMMIT;BEGIN;
    ALTER TABLE chatrooms 
    DROP COLUMN IF EXISTS "enabled";
COMMIT;BEGIN; 
    ALTER TABLE services   
    ADD COLUMN canceller_id INT REFERENCES users(id) NULL;  
COMMIT;BEGIN;
    ALTER TABLE users 
    DROP COLUMN IF EXISTS phone_verify_code;

COMMIT;CREATE UNIQUE INDEX idx_userid_blocked_userid ON block_list(user_id, blocked_user_id);ALTER TYPE service_status ADD VALUE 'negotiating';BEGIN;

ALTER TABLE service_inquiries
ADD COLUMN fcm_topic VARCHAR;

COMMIT;BEGIN;
	ALTER TABLE service_inquiries
	DROP COLUMN IF EXISTS price;
COMMIT;BEGIN; 

ALTER TABLE users 
ADD COLUMN IF NOT EXISTS fcm_topic VARCHAR(128); 

COMMIT;BEGIN;

ALTER TABLE service_inquiries 
DROP COLUMN IF EXISTS fcm_topic;

COMMIT;BEGIN;

ALTER TABLE service_ratings 
ADD COLUMN ratee_id INT REFERENCES users(id);

COMMIT;BEGIN;

CREATE TYPE inquiry_type AS ENUM (
    'direct', 
    'random'
);

ALTER TABLE service_inquiries
ADD COLUMN inquiry_type inquiry_type DEFAULT 'random';

COMMIT;BEGIN;

ALTER TABLE payments
ADD COLUMN  refunded boolean default false;

COMMIT;BEGIN;
    CREATE TYPE cancel_cause AS ENUM (
        'none',
        'girl_cancel_before_appointment_time',
        'girl_cancel_after_appointment_time',
        'guy_cancel_before_appointment_time',
        'guy_cancel_after_appointment_time'
    );
    
    ALTER TABLE services
    ADD COLUMN cancel_cause cancel_cause DEFAULT 'none';
    
    COMMENT ON COLUMN services.cancel_cause IS 'cause states the intention of cancelling a service.';
COMMIT;ALTER TYPE cancel_cause ADD VALUE 'payment_failed';BEGIN; 

ALTER TABLE coin_packages
ALTER COLUMN cost TYPE NUMERIC(12, 2);

COMMIT;BEGIN;

ALTER TABLE services
ADD COLUMN IF NOT EXISTS matching_fee numeric(12, 2) default 0;

COMMIT;BEGIN;

CREATE TABLE IF NOT exists service_options (
    id BIGSERIAL PRIMARY KEY,
    name text NOT NULL, 
    description text, 
    price numeric(12, 2),    
    
	created_at timestamp NOT NULL DEFAULT NOW(),
	updated_at timestamp NULL DEFAULT current_timestamp,
	deleted_at timestamp
);

COMMIT;BEGIN;

CREATE TABLE IF NOT EXISTS user_service_options (
    id BIGSERIAL PRIMARY KEY,
    users_id INT,
    service_option_id INT,

	created_at timestamp NOT NULL DEFAULT NOW(),
	updated_at timestamp NULL DEFAULT current_timestamp,
	deleted_at timestamp,

    CONSTRAINT fk_user_id
    FOREIGN KEY (users_id)
    REFERENCES users(id),
    
    CONSTRAINT fk_service_option_id
    FOREIGN KEY (service_option_id)
    REFERENCES service_options(id)
);

COMMIT;BEGIN;

CREATE OR REPLACE FUNCTION trigger_set_timestamp()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

COMMIT;CREATE TRIGGER set_timestamp
BEFORE UPDATE ON services
FOR EACH ROW
EXECUTE PROCEDURE trigger_set_timestamp();--- bank_accounts
CREATE TRIGGER bank_accounts_updated_at_set_timestamp
BEFORE UPDATE ON bank_accounts
FOR EACH ROW
EXECUTE PROCEDURE trigger_set_timestamp();

--- block_list
CREATE TRIGGER block_list_updated_at_set_timestamp
BEFORE UPDATE ON block_list
FOR EACH ROW
EXECUTE PROCEDURE trigger_set_timestamp();

--- chatroom_users
CREATE TRIGGER chatroom_users_updated_at_set_timestamp
BEFORE UPDATE ON chatroom_users
FOR EACH ROW
EXECUTE PROCEDURE trigger_set_timestamp();

--- chatrooms
CREATE TRIGGER chatrooms_updated_at_set_timestamp
BEFORE UPDATE ON chatrooms
FOR EACH ROW
EXECUTE PROCEDURE trigger_set_timestamp();

--- coin_orders 
CREATE TRIGGER coin_orders_updated_at_set_timestamp
BEFORE UPDATE ON coin_orders
FOR EACH ROW
EXECUTE PROCEDURE trigger_set_timestamp();

--- images 
CREATE TRIGGER images_updated_at_set_timestamp
BEFORE UPDATE ON images
FOR EACH ROW
EXECUTE PROCEDURE trigger_set_timestamp();

--- payments 
CREATE TRIGGER payments_updated_at_set_timestamp
BEFORE UPDATE ON payments
FOR EACH ROW
EXECUTE PROCEDURE trigger_set_timestamp();

--- service_inquiries 
CREATE TRIGGER service_inquiries_updated_at_set_timestamp
BEFORE UPDATE ON service_inquiries 
FOR EACH ROW
EXECUTE PROCEDURE trigger_set_timestamp();

--- service_options 
CREATE TRIGGER service_options_updated_at_set_timestamp
BEFORE UPDATE ON service_options 
FOR EACH ROW
EXECUTE PROCEDURE trigger_set_timestamp();

--- service_qrcode 
CREATE TRIGGER service_qrcode_updated_at_set_timestamp
BEFORE UPDATE ON service_qrcode 
FOR EACH ROW
EXECUTE PROCEDURE trigger_set_timestamp();

--- service_ratings 
CREATE TRIGGER service_ratings_updated_at_set_timestamp
BEFORE UPDATE ON service_ratings 
FOR EACH ROW
EXECUTE PROCEDURE trigger_set_timestamp();

--- services 
CREATE TRIGGER services_updated_at_set_timestamp
BEFORE UPDATE ON services 
FOR EACH ROW
EXECUTE PROCEDURE trigger_set_timestamp();

--- user_balance 
CREATE TRIGGER user_balance_updated_at_set_timestamp
BEFORE UPDATE ON user_balance 
FOR EACH ROW
EXECUTE PROCEDURE trigger_set_timestamp();

--- user_refcodes 
CREATE TRIGGER user_refcodes_updated_at_set_timestamp
BEFORE UPDATE ON user_refcodes 
FOR EACH ROW
EXECUTE PROCEDURE trigger_set_timestamp();

--- user_service_options 
CREATE TRIGGER user_service_options_updated_at_set_timestamp
BEFORE UPDATE ON user_service_options 
FOR EACH ROW
EXECUTE PROCEDURE trigger_set_timestamp();

--- users
CREATE TRIGGER users_updated_at_set_timestamp
BEFORE UPDATE ON users
FOR EACH ROW
EXECUTE PROCEDURE trigger_set_timestamp();BEGIN;

CREATE TYPE service_options_type AS ENUM ('default', 'custom');

ALTER TABLE service_options 
ADD COLUMN service_options_type service_options_type DEFAULT 'default';

COMMIT;BEGIN;

DROP TABLE IF EXISTS service_names;

COMMIT;BEGIN;

DROP TABLE IF EXISTS lobby_users;

COMMIT;BEGIN;

ALTER TABLE service_inquiries
DROP COLUMN service_type,
ADD COLUMN expect_service_type text;

COMMIT;BEGIN;

ALTER TABLE services
ALTER COLUMN service_type TYPE text; 

COMMIT;BEGIN;

ALTER TABLE service_options 
ADD COLUMN duration INT;

COMMIT;BEGIN;

ALTER TABLE service_inquiries
ADD COLUMN currency TEXT;

COMMIT;

BEGIN;

ALTER TABLE services
ADD COLUMN currency TEXT;

COMMIT;

