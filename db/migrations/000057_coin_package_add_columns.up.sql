BEGIN;

	ALTER TABLE coin_packages
	ADD COLUMN name varchar(255);

COMMIT;
