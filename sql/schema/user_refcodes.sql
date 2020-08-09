-- user_refcodes.sql
CREATE TYPE ref_code_type AS ENUM (
	'invitor',
	'manager'
);

CREATE TABLE IF NOT EXISTS user_refcodes (
	id BIGSERIAL PRIMARY KEY,
	invitor_id INT REFERENCES users(id) NOT NULL,
	invitee_id INT REFERENCES users(id),
	ref_code VARCHAR NOT NULL,
	ref_code_type ref_code_type NOT NULL,

	created_at timestamp NOT NULL DEFAULT NOW(),
	updated_at timestamp NULL DEFAULT current_timestamp,
	deleted_at timestamp,

	CONSTRAINT invitor_id_fk PRIMARY KEY(invitor_id)
)

