BEGIN;
	ALTER TABLE users
	ADD COLUMN mobile varchar(20);
COMMIT;
