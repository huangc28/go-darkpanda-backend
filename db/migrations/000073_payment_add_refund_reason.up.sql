BEGIN;
    CREATE TYPE cause AS ENUM (
        'none',
        'girl_cancel_before_appointment_time',
        'girl_cancel_after_appointment_time',
        'guy_cancel_before_appointment_time',
        'guy_cancel_after_appointment_time'
    );
    
    ALTER TABLE payments
    ADD COLUMN cause cause DEFAULT 'none';
    
    COMMENT ON COLUMN payments.cause IS 'cause states the intention of cancelling a payment.';
COMMIT;