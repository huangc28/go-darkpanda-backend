-- name: GetUserByUsername :one
SELECT * FROM users
WHERE username = $1 LIMIT 1;

-- name: CreateUser :one
INSERT INTO users (
	username,
	uuid,
	phone_verified,
	auth_sms_code,
	gender,
	premium_type,
	premium_expiry_date
) VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetUserByUuid :one
SELECT * FROM users
WHERE uuid = $1 LIMIT 1;

-- name: UpdateVerifyCodeById :exec
UPDATE users SET phone_verify_code = $1
WHERE id = $2;
