BEGIN;

CREATE TABLE IF NOT EXISTS user_service_options (
    id BIGSERIAL PRIMARY KEY,
    users_id INT,
    service_option_id INT,

	created_at timestamp NOT NULL DEFAULT NOW(),
	updated_at timestamp NULL DEFAULT current_timestamp,
	deleted_at timestamp,

    CONSTRAINT fk_user_id
    FOREIGN KEY (users_id)
    REFERENCES users(id),
    
    CONSTRAINT fk_service_option_id
    FOREIGN KEY (service_option_id)
    REFERENCES service_options(id)
);

COMMIT;