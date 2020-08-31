BEGIN;

ALTER TABLE users

ADD COLUMN nationality varchar(255),
ADD COLUMN region varchar(255),
ADD COLUMN age INT,
ADD COLUMN height numeric(5, 2),
ADD COLUMN weight numeric(5, 2),
ADD COLUMN habbits varchar(40),
ADD COLUMN description varchar(255),
ADD COLUMN breast_size varchar(40),
ADD CONSTRAINT breast_size_regex CHECK (breast_size ~ '^[a-zA-Z]$');

COMMIT;
