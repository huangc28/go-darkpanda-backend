BEGIN;

CREATE TABLE IF NOT exists service_options (
    id BIGSERIAL PRIMARY KEY,
    name text NOT NULL, 
    description text, 
    price numeric(12, 2),    
    
	created_at timestamp NOT NULL DEFAULT NOW(),
	updated_at timestamp NULL DEFAULT current_timestamp,
	deleted_at timestamp
);

COMMIT;