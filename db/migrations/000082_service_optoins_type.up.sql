BEGIN;

CREATE TYPE service_options_type AS ENUM ('default', 'custom');

ALTER TABLE service_options 
ADD COLUMN service_options_type service_options_type DEFAULT 'default';

COMMIT;