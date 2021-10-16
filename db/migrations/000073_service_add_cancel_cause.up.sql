BEGIN;
    CREATE TYPE cancel_cause AS ENUM (
        'none',
        'girl_cancel_before_appointment_time',
        'girl_cancel_after_appointment_time',
        'guy_cancel_before_appointment_time',
        'guy_cancel_after_appointment_time'
    );
    
    ALTER TABLE services
    ADD COLUMN cancel_cause cancel_cause DEFAULT 'none';
    
    COMMENT ON COLUMN services.cancel_cause IS 'cause states the intention of cancelling a service.';
COMMIT;