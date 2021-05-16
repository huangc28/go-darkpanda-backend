BEGIN;

ALTER TABLE coin_orders
DROP COLUMN rec_trade_id,
DROP COLUMN raw;

COMMIT;

