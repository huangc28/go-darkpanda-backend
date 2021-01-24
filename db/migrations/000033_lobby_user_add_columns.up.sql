-- Add default timestamp value in minute: https://stackoverflow.com/questions/21745125/add-minutes-to-current-timestamp-in-postgresql
BEGIN;

DROP TYPE IF EXISTS lobby_status;
CREATE TYPE lobby_status AS ENUM (
	'waiting',
	'pause',
	'expired',
	'left'
);

ALTER TABLE lobby_users
ADD COLUMN expired_at timestamp NOT NULL DEFAULT current_timestamp + (26 * interval '1 minute'),
ADD COLUMN lobby_status lobby_status NOT NULL DEFAULT 'waiting';

COMMIT;
