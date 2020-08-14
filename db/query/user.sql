-- name: GetUserByUsername :one
SELECT * FROM users
WHERE username = $1 LIMIT 1;

-- name: CreateUser :one
INSERT INTO users (
	username,
	phone_verified,
	auth_sms_code,
	gender,
	premium_type,
	premium_expiry_date
) VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

