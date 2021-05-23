ALTER TABLE services ALTER COLUMN service_status TYPE varchar(255);

DROP TYPE IF EXISTS service_status CASCADE;

CREATE TYPE service_status AS ENUM  (
	'unpaid',
	'payment_failed',
	'to_be_fulfilled',
	'canceled',
	'expired',
	'fulfilling',
	'completed'
);

ALTER TABLE services ALTER COLUMN service_status TYPE service_status USING service_status::text::service_status;

