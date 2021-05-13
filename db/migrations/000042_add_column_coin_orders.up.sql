BEGIN;

ALTER TABLE coin_orders
DROP COLUMN amount,
ADD COLUMN package_id INT REFERECES coin_packages(id),
ADD COLUMN quantity INT NOT NULL DEFAULT 1;


COMMIT;
