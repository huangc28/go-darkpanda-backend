// Code generated by sqlc. DO NOT EDIT.
// source: user.sql

package models

import (
	"context"
	"database/sql"
)

const createUser = `-- name: CreateUser :one
INSERT INTO users (
	username,
	uuid,
	phone_verified,
	gender,
	premium_type,
	premium_expiry_date,
	avatar_url,
	mobile,
	fcm_topic,
	description
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING id, username, phone_verified, gender, premium_type, premium_expiry_date, created_at, updated_at, deleted_at, uuid, avatar_url, nationality, region, age, height, weight, habbits, description, breast_size, mobile, fcm_topic
`

type CreateUserParams struct {
	Username          string         `json:"username"`
	Uuid              string         `json:"uuid"`
	PhoneVerified     bool           `json:"phone_verified"`
	Gender            Gender         `json:"gender"`
	PremiumType       PremiumType    `json:"premium_type"`
	PremiumExpiryDate sql.NullTime   `json:"premium_expiry_date"`
	AvatarUrl         sql.NullString `json:"avatar_url"`
	Mobile            sql.NullString `json:"mobile"`
	FcmTopic          sql.NullString `json:"fcm_topic"`
	Description       sql.NullString `json:"description"`
}

func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) (User, error) {
	row := q.queryRow(ctx, q.createUserStmt, createUser,
		arg.Username,
		arg.Uuid,
		arg.PhoneVerified,
		arg.Gender,
		arg.PremiumType,
		arg.PremiumExpiryDate,
		arg.AvatarUrl,
		arg.Mobile,
		arg.FcmTopic,
		arg.Description,
	)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Username,
		&i.PhoneVerified,
		&i.Gender,
		&i.PremiumType,
		&i.PremiumExpiryDate,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.DeletedAt,
		&i.Uuid,
		&i.AvatarUrl,
		&i.Nationality,
		&i.Region,
		&i.Age,
		&i.Height,
		&i.Weight,
		&i.Habbits,
		&i.Description,
		&i.BreastSize,
		&i.Mobile,
		&i.FcmTopic,
	)
	return i, err
}

const getUserByID = `-- name: GetUserByID :one
SELECT id, username, phone_verified, gender, premium_type, premium_expiry_date, created_at, updated_at, deleted_at, uuid, avatar_url, nationality, region, age, height, weight, habbits, description, breast_size, mobile, fcm_topic FROM users
WHERE id = $1 LIMIT 1
`

func (q *Queries) GetUserByID(ctx context.Context, id int64) (User, error) {
	row := q.queryRow(ctx, q.getUserByIDStmt, getUserByID, id)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Username,
		&i.PhoneVerified,
		&i.Gender,
		&i.PremiumType,
		&i.PremiumExpiryDate,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.DeletedAt,
		&i.Uuid,
		&i.AvatarUrl,
		&i.Nationality,
		&i.Region,
		&i.Age,
		&i.Height,
		&i.Weight,
		&i.Habbits,
		&i.Description,
		&i.BreastSize,
		&i.Mobile,
		&i.FcmTopic,
	)
	return i, err
}

const getUserByUsername = `-- name: GetUserByUsername :one
SELECT id, username, phone_verified, gender, premium_type, premium_expiry_date, created_at, updated_at, deleted_at, uuid, avatar_url, nationality, region, age, height, weight, habbits, description, breast_size, mobile, fcm_topic FROM users
WHERE username = $1 LIMIT 1
`

func (q *Queries) GetUserByUsername(ctx context.Context, username string) (User, error) {
	row := q.queryRow(ctx, q.getUserByUsernameStmt, getUserByUsername, username)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Username,
		&i.PhoneVerified,
		&i.Gender,
		&i.PremiumType,
		&i.PremiumExpiryDate,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.DeletedAt,
		&i.Uuid,
		&i.AvatarUrl,
		&i.Nationality,
		&i.Region,
		&i.Age,
		&i.Height,
		&i.Weight,
		&i.Habbits,
		&i.Description,
		&i.BreastSize,
		&i.Mobile,
		&i.FcmTopic,
	)
	return i, err
}

const getUserByUuid = `-- name: GetUserByUuid :one
SELECT id, username, phone_verified, gender, premium_type, premium_expiry_date, created_at, updated_at, deleted_at, uuid, avatar_url, nationality, region, age, height, weight, habbits, description, breast_size, mobile, fcm_topic FROM users
WHERE uuid = $1 LIMIT 1
`

func (q *Queries) GetUserByUuid(ctx context.Context, uuid string) (User, error) {
	row := q.queryRow(ctx, q.getUserByUuidStmt, getUserByUuid, uuid)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Username,
		&i.PhoneVerified,
		&i.Gender,
		&i.PremiumType,
		&i.PremiumExpiryDate,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.DeletedAt,
		&i.Uuid,
		&i.AvatarUrl,
		&i.Nationality,
		&i.Region,
		&i.Age,
		&i.Height,
		&i.Weight,
		&i.Habbits,
		&i.Description,
		&i.BreastSize,
		&i.Mobile,
		&i.FcmTopic,
	)
	return i, err
}

const getUserIDByUuid = `-- name: GetUserIDByUuid :one
SELECT id FROM users
WHERE uuid = $1 LIMIT 1
`

func (q *Queries) GetUserIDByUuid(ctx context.Context, uuid string) (int64, error) {
	row := q.queryRow(ctx, q.getUserIDByUuidStmt, getUserIDByUuid, uuid)
	var id int64
	err := row.Scan(&id)
	return id, err
}

const patchUserInfoByUuid = `-- name: PatchUserInfoByUuid :one
UPDATE users
SET avatar_url = $1, nationality = $2, region = $3, age = $4, height = $5, weight = $6, description = $7, breast_size = $8
WHERE uuid = $9
RETURNING id, username, phone_verified, gender, premium_type, premium_expiry_date, created_at, updated_at, deleted_at, uuid, avatar_url, nationality, region, age, height, weight, habbits, description, breast_size, mobile, fcm_topic
`

type PatchUserInfoByUuidParams struct {
	AvatarUrl   sql.NullString `json:"avatar_url"`
	Nationality sql.NullString `json:"nationality"`
	Region      sql.NullString `json:"region"`
	Age         sql.NullInt32  `json:"age"`
	Height      sql.NullString `json:"height"`
	Weight      sql.NullString `json:"weight"`
	Description sql.NullString `json:"description"`
	BreastSize  sql.NullString `json:"breast_size"`
	Uuid        string         `json:"uuid"`
}

func (q *Queries) PatchUserInfoByUuid(ctx context.Context, arg PatchUserInfoByUuidParams) (User, error) {
	row := q.queryRow(ctx, q.patchUserInfoByUuidStmt, patchUserInfoByUuid,
		arg.AvatarUrl,
		arg.Nationality,
		arg.Region,
		arg.Age,
		arg.Height,
		arg.Weight,
		arg.Description,
		arg.BreastSize,
		arg.Uuid,
	)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Username,
		&i.PhoneVerified,
		&i.Gender,
		&i.PremiumType,
		&i.PremiumExpiryDate,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.DeletedAt,
		&i.Uuid,
		&i.AvatarUrl,
		&i.Nationality,
		&i.Region,
		&i.Age,
		&i.Height,
		&i.Weight,
		&i.Habbits,
		&i.Description,
		&i.BreastSize,
		&i.Mobile,
		&i.FcmTopic,
	)
	return i, err
}

const updateVerifyStatusById = `-- name: UpdateVerifyStatusById :exec
UPDATE users
SET phone_verified = $2,
    mobile = $3
WHERE id = $1
`

type UpdateVerifyStatusByIdParams struct {
	ID            int64          `json:"id"`
	PhoneVerified bool           `json:"phone_verified"`
	Mobile        sql.NullString `json:"mobile"`
}

func (q *Queries) UpdateVerifyStatusById(ctx context.Context, arg UpdateVerifyStatusByIdParams) error {
	_, err := q.exec(ctx, q.updateVerifyStatusByIdStmt, updateVerifyStatusById, arg.ID, arg.PhoneVerified, arg.Mobile)
	return err
}
