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
