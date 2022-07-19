BEGIN;

ALTER TABLE users

DROP COLUMN nationality,
DROP COLUMN region,
DROP COLUMN age,
DROP COLUMN height,
DROP COLUMN weight,
DROP COLUMN habbits,
DROP COLUMN description,
DROP COLUMN breast_size,

DROP CONSTRAINT IF EXISTS breast_size_regex;

COMMIT;
