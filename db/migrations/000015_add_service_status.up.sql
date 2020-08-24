BEGIN;

DROP TYPE IF EXISTS service_status;

CREATE TYPE service_status AS ENUM  (
	'unpaid',
	'to_be_fulfilled',
	'canceled',
	'failed_due_to_both',
	'girl_waiting',
	'fufilling',
	'failed_due_to_girl',
	'failed_due_to_man',
	'completed'
);

ALTER TABLE services
ADD COLUMN service_status service_status NOT NULL DEFAULT 'unpaid';

COMMIT;
