BEGIN;

ALTER TABLE user_refcodes
ADD COLUMN expired_at timestamp
DEFAULT NOW() + interval '3 days';


COMMENT ON COLUMN user_refcodes.expired_at IS 'Time that this referral code will be invalid.';

COMMIT;
