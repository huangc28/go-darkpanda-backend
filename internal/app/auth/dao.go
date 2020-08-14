package auth

import (
	"context"
	"database/sql"
)

type AuthDAOer interface {
	CheckUsernameExists(ctx context.Context, username string) (bool, error)
}

type AuthDAO struct {
	db *sql.DB
}

func NewAuthDao(db *sql.DB) *AuthDAO {
	return &AuthDAO{
		db: db,
	}
}

func (dao *AuthDAO) CheckUsernameExists(ctx context.Context, username string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE username = $1) AS "exists"`
	var exists bool

	if err := dao.db.QueryRow(query, username).Scan(&exists); err != nil {
		return false, err
	}

	return exists, nil
}

func (dao *AuthDAO) CheckReferCodeExists(ctx context.Context, referCode string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM user_refcodes WHERE ref_code = $1) AS "exists"`
	var exists bool

	if err := dao.db.QueryRow(query, referCode).Scan(&exists); err != nil {
		return false, err
	}

	return exists, nil
}
