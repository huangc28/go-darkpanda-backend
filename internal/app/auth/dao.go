package auth

import (
	"context"
	"database/sql"

	"github.com/go-redis/redis/v8"
)

type UserCheckerDAOer interface {
	CheckUsernameExists(ctx context.Context, username string) (bool, error)
	CheckReferCodeExists(ctx context.Context, referCode string) (bool, error)
}

type UserCheckerDAO struct {
	db *sql.DB
}

func NewUserCheckerDAO(db *sql.DB) *UserCheckerDAO {
	return &UserCheckerDAO{
		db: db,
	}
}

func (dao *UserCheckerDAO) CheckUsernameExists(ctx context.Context, username string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE username = $1) AS "exists"`
	var exists bool

	if err := dao.db.QueryRow(query, username).Scan(&exists); err != nil {
		return false, err
	}

	return exists, nil
}

func (dao *UserCheckerDAO) CheckReferCodeExists(ctx context.Context, referCode string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM user_refcodes WHERE ref_code = $1) AS "exists"`
	var exists bool

	if err := dao.db.QueryRow(query, referCode).Scan(&exists); err != nil {
		return false, err
	}

	return exists, nil
}

const (
	INVALIDATE_TOKEN_REDIS_KEY = "invalidate_token"
)

type AuthDAOer interface {
	RevokeJwt(jwt string) error
}

type AuthDAO struct {
	redis *redis.Client
}

func NewAuthDao(redis *redis.Client) *AuthDAO {
	return &AuthDAO{
		redis: redis,
	}
}

func (dao *AuthDAO) RevokeJwt(ctx context.Context, jwt string) error {
	if err := dao.redis.SAdd(ctx, INVALIDATE_TOKEN_REDIS_KEY, jwt).Err(); err != nil {
		return err
	}

	return nil
}
