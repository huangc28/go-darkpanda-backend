BEGIN;

CREATE TABLE coin_packages (
	id BIGSERIAL PRIMARY KEY,
	db_coins INT,
	cost INT,
	currency varchar(10) DEFAULT 'TWD'
);

COMMIT;
