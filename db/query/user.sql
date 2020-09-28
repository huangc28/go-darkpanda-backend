-- name: GetUserByUsername :one
SELECT * FROM users
WHERE username = $1 LIMIT 1;

-- name: CreateUser :one
INSERT INTO users (
	username,
	uuid,
	phone_verify_code,
	phone_verified,
	auth_sms_code,
	gender,
	premium_type,
	premium_expiry_date,
	avatar_url,
	mobile
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING *;

-- name: GetUserByUuid :one
SELECT * FROM users
WHERE uuid = $1 LIMIT 1;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = $1 LIMIT 1;

-- name: GetUserIDByUuid :one
SELECT id FROM users
WHERE uuid = $1 LIMIT 1;

-- name: UpdateVerifyCodeById :exec
UPDATE users SET phone_verify_code = $1
WHERE id = $2;

-- name: GetUserByVerifyCode :one
SELECT * FROM users
WHERE phone_verify_code = $1 LIMIT 1;

-- name: UpdateVerifyStatusById :exec
UPDATE users
SET phone_verified = $2,
    mobile = $3
WHERE id = $1;

-- name: PatchUserInfoByUuid :one
UPDATE users
SET avatar_url = $1, nationality = $2, region = $3, age = $4, height = $5, weight = $6, description = $7, breast_size = $8
WHERE uuid = $9
RETURNING *;

