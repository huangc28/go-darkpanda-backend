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
