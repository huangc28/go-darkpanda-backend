BEGIN;

ALTER TABLE coin_orders
ADD COLUMN rec_trade_id VARCHAR(255),
ADD COLUMN raw text;

COMMIT;
