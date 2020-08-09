// Code generated by sqlc. DO NOT EDIT.
// source: user.sql

package db

import (
	"context"
)

const getAuthor = `-- name: GetAuthor :one
SELECT id, username, phone_verified, auth_sms_code, gender, premium_kind, premium_expiry_date, created_at, updated_at, deleted_at FROM users
WHERE id = $1 LIMIT 1
`

func (q *Queries) GetAuthor(ctx context.Context, id int64) (User, error) {
	row := q.queryRow(ctx, q.getAuthorStmt, getAuthor, id)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Username,
		&i.PhoneVerified,
		&i.AuthSmsCode,
		&i.Gender,
		&i.PremiumKind,
		&i.PremiumExpiryDate,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.DeletedAt,
	)
	return i, err
}
