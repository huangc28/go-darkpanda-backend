
BEGIN;

ALTER TABLE user_service_options
    DROP CONSTRAINT fk_user_id,
    DROP CONSTRAINT fk_service_option_id;

DROP TABLE IF EXISTS user_service_options;

COMMIT;