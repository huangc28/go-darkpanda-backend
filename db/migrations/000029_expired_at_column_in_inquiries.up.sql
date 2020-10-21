BEGIN;

ALTER TABLE service_inquiries
ADD COLUMN expired_at timestamp;

COMMENT ON COLUMN service_inquiries.expired_at IS 'Time that this inquiry will be invalid.';

COMMIT;
