BEGIN;

CREATE TABLE user_balance (
	id BIGSERIAL PRIMARY KEY,
	balance INT,

	created_at timestamp NOT NULL DEFAULT NOW(),
	updated_at timestamp NULL DEFAULT current_timestamp,
	deleted_at timestamp
);

COMMENT ON COLUMN user_balance.balance IS 'use for update when reading the column';

COMMIT;
