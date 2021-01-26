package register

import (
	"context"

	"github.com/huangc28/go-darkpanda-backend/db"
)

type RegisterDAO struct {
	db db.Conn
}

func NewRegisterDAO(db db.Conn) *RegisterDAO {
	return &RegisterDAO{
		db: db,
	}
}

func (dao *RegisterDAO) CheckUsernameExists(ctx context.Context, username string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE username = $1) AS "exists"`
	var exists bool

	if err := dao.db.QueryRow(query, username).Scan(&exists); err != nil {
		return false, err
	}

	return exists, nil
}

func (dao *RegisterDAO) CheckReferCodeExists(ctx context.Context, referCode string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM user_refcodes WHERE ref_code = $1) AS "exists"`
	var exists bool

	if err := dao.db.QueryRow(query, referCode).Scan(&exists); err != nil {
		return false, err
	}

	return exists, nil
}
