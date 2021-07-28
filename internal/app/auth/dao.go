package auth

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/golobby/container/pkg/container"
	"github.com/huangc28/go-darkpanda-backend/db"
	"github.com/huangc28/go-darkpanda-backend/internal/app/contracts"
	"github.com/huangc28/go-darkpanda-backend/internal/app/models"
	"github.com/jmoiron/sqlx"
)

type UserDAO interface {
	GetUserByUuid(uuid string, fields ...string) (*models.User, error)
}

type UserCheckerDAOer interface {
	CheckUsernameExists(ctx context.Context, username string) (bool, error)
	CheckReferCodeExists(ctx context.Context, referCode string) (bool, error)
}

type UserCheckerDAO struct {
	db *sqlx.DB
}

func NewUserCheckerDAO(db *sqlx.DB) *UserCheckerDAO {
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

func (dao *UserCheckerDAO) GetUserByUsername(ctx context.Context, username string) error {
	query := `
SELECT id FROM users WHERE username = $1 LIMIT 1;
`
	user := &models.User{}

	if err := dao.db.QueryRow(query, username).Scan(&user.ID); err != nil {
		return err
	}

	return nil
}

const (
	INVALIDATE_TOKEN_REDIS_KEY = "invalid_tokens"
)

type AuthDAOer interface {
	RevokeJwt(jwt string) error
	GetLoginRecord(ctx context.Context, userUuid string) (*LoginAuthenticator, error)
	UpdateLoginRecord(ctx context.Context, userUuid string, record LoginAuthenticator) error
	CreateLoginVerifyCode(ctx context.Context, loginVerifyCode, userUuid string) error
}

type AuthDAO struct {
	redis *redis.Client
}

func AuthDaoerServiceProvider(c container.Container) func() error {
	return func() error {
		c.Transient(func() contracts.AuthDaoer {
			return NewAuthDao(db.GetRedis())
		})

		return nil
	}
}

func NewAuthDao(redis *redis.Client) *AuthDAO {
	return &AuthDAO{
		redis: redis,
	}
}

// We store the invalid jwt tokens in the following struct in redis.
//
// invalid_tokens: {
//	{token_1}: expired_date_1
//	{token_2}: expired_date_2
//	{token_3}: expired_date_3
// }
func (dao *AuthDAO) RevokeJwt(ctx context.Context, jwt string, expTs int64) error {

	if err := dao.redis.HSet(
		ctx,
		INVALIDATE_TOKEN_REDIS_KEY,
		jwt,
		expTs,
	).Err(); err != nil {
		return err
	}

	return nil
}

// CreateLoginVerifyCode store the code to redis, set TTL to 300 seconds.
// Key: login_authenticator_uuid:{{ USER_UUID }}
// Structure:
//   {
//      verify_code: {{ VERIFY_CODE }}
//      num_retried: 0,
//   }
//
// `num_retried` records the number of times the user has
const (
	LoginAuthenticatorHashKey = "login_authenticator_uuid:%s"
	LoginVerifyCodeFieldKey   = "verify_code"
	LoginNumRetriedFieldKey   = "num_retried"
)

type LoginAuthenticator struct {
	VerifyCode string
	NumRetried int
}

func ParseLoginAuthenticatorFromMap(data map[string]string) (*LoginAuthenticator, error) {
	s := &LoginAuthenticator{}

	if val, ok := data["verify_code"]; ok {
		s.VerifyCode = val
	}

	if val, ok := data["num_retried"]; ok {

		num, err := strconv.Atoi(val)

		if err != nil {
			return nil, err
		}

		s.NumRetried = num
	}

	return s, nil
}

func (dao *AuthDAO) GetLoginRecord(ctx context.Context, userUuid string) (*LoginAuthenticator, error) {
	exists, err := dao.redis.Exists(ctx, fmt.Sprintf(LoginAuthenticatorHashKey, userUuid)).Result()

	if err != nil {
		return nil, err
	}

	if exists == 0 {
		return nil, redis.Nil
	}

	val, err := dao.redis.HGetAll(ctx, fmt.Sprintf(LoginAuthenticatorHashKey, userUuid)).Result()

	if err != nil {
		return nil, err
	}

	return ParseLoginAuthenticatorFromMap(val)
}

func (dao *AuthDAO) UpdateLoginRecord(ctx context.Context, userUuid string, record LoginAuthenticator) error {
	pipe := dao.redis.TxPipeline()
	defer pipe.Close()

	pipe.HMSet(
		ctx,
		fmt.Sprintf(LoginAuthenticatorHashKey, userUuid),
		LoginVerifyCodeFieldKey,
		record.VerifyCode,
		LoginNumRetriedFieldKey,
		record.NumRetried,
	)

	pipe.Expire(
		ctx,
		fmt.Sprintf(LoginAuthenticatorHashKey, userUuid),
		5*time.Minute,
	)

	if _, err := pipe.Exec(ctx); err != nil {
		return err
	}

	return nil
}

func (dao *AuthDAO) CreateLoginVerifyCode(ctx context.Context, loginVerifyCode, userUuid string) (*LoginAuthenticator, error) {
	pipe := dao.redis.TxPipeline()
	defer pipe.Close()

	pipe.HSet(
		ctx,
		fmt.Sprintf(LoginAuthenticatorHashKey, userUuid),
		LoginVerifyCodeFieldKey,
		loginVerifyCode,
		LoginNumRetriedFieldKey,
		0,
	)

	pipe.Expire(
		ctx,
		fmt.Sprintf(LoginAuthenticatorHashKey, userUuid),
		5*time.Minute,
	)

	if _, err := pipe.Exec(ctx); err != nil {
		return nil, err
	}

	val, err := dao.redis.HGetAll(ctx, fmt.Sprintf(LoginAuthenticatorHashKey, userUuid)).Result()

	if err != nil {
		return nil, err
	}

	return ParseLoginAuthenticatorFromMap(val)
}

func (dao *AuthDAO) IsTokenInvalid(ctx context.Context, jwtToken string) (bool, error) {
	exists, err := dao.redis.HExists(
		ctx,
		INVALIDATE_TOKEN_REDIS_KEY,
		jwtToken,
	).Result()

	if err != nil {
		return false, err
	}

	return exists, nil
}
