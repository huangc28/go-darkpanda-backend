BEGIN;

CREATE TYPE inquiry_type AS ENUM (
    'direct', 
    'random'
);

ALTER TABLE service_inquiries
ADD COLUMN inquiry_type inquiry_type DEFAULT 'random';

COMMIT;