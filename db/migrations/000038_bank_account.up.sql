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
