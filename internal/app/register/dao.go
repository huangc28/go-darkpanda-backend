package register

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	cintrnal "github.com/golobby/container/pkg/container"
	"github.com/huangc28/go-darkpanda-backend/db"
	"github.com/huangc28/go-darkpanda-backend/internal/app/contracts"
	"github.com/huangc28/go-darkpanda-backend/internal/app/models"
)

type RegisterDAO struct {
	db db.Conn
}

func NewRegisterDAO(db db.Conn) *RegisterDAO {
	return &RegisterDAO{
		db: db,
	}
}

func RegisterDaoServiceProvider(c cintrnal.Container) func() error {
	return func() error {
		c.Transient(func() contracts.Registerar {
			return NewRegisterDAO(db.GetDB())
		})

		return nil
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

func (dao *RegisterDAO) GetReferralCodeByReferralCode(refCode string) (models.UserRefcode, error) {
	query := `
SELECT id, invitor_id, invitee_id, ref_code, ref_code_type, created_at, updated_at, deleted_at FROM user_refcodes
WHERE ref_code = $1 LIMIT 1
`
	var refCodeM models.UserRefcode

	if err := dao.
		db.
		QueryRowx(query, refCode).
		StructScan(&refCodeM); err != nil {
		return refCodeM, err
	}

	return refCodeM, nil
}

const (
	RegisterMobileVerifyCodeKey      = "register_mobile_verify_code:%s"
	RegisterMobileVerifyCodeFieldKey = "verify_code"
	RegisterMobileNumberFieldKey     = "mobile"
)

const SmsWhiteListKey = "sms_white_list"

// We don't send real sms message to users in this list.
// We want to save some money. Users in this list are
// developers.
func (dao *RegisterDAO) CheckUserInSMSWhiteList(ctx context.Context, p contracts.CheckUserInSMSWhiteListParams) (bool, error) {
	l, err := p.RedisClient.LRange(ctx, SmsWhiteListKey, 0, -1).Result()

	if err != nil {
		return false, err
	}

	for _, uuid := range l {
		if uuid == p.UserUuid {
			return true, nil
		}
	}

	return false, nil
}

type CreateRegisterMobileVerifyCodeParams struct {
	RedisCli   *redis.Client
	UserUuid   string
	VerifyCode string
	Mobile     string
}

func CreateRegisterMobileVerifyCode(ctx context.Context, p CreateRegisterMobileVerifyCodeParams) error {
	if p.RedisCli == nil {
		return errors.New("redis client is not provided")
	}

	pipe := p.RedisCli.TxPipeline()
	defer pipe.Close()

	p.RedisCli.HSet(
		ctx,
		fmt.Sprintf(RegisterMobileVerifyCodeKey, p.UserUuid),
		RegisterMobileVerifyCodeFieldKey,
		p.VerifyCode,
		RegisterMobileNumberFieldKey,
		p.Mobile,
	)

	p.RedisCli.Expire(
		ctx,
		fmt.Sprintf(RegisterMobileVerifyCodeKey, p.UserUuid),
		5*time.Minute,
	)

	if _, err := pipe.Exec(ctx); err != nil {
		return err

	}

	return nil
}

type RegisterMobileVerifyCode struct {
	VerifyCode string
	Mobile     string
}

type GetRegisterMobileVerifyCodeParams struct {
	RedisCli *redis.Client
	UserUuid string
}

func GetRegisterMobileVerifyCode(ctx context.Context, p GetRegisterMobileVerifyCodeParams) (*RegisterMobileVerifyCode, error) {
	val, err := p.RedisCli.HGetAll(
		ctx,
		fmt.Sprintf(RegisterMobileVerifyCodeKey, p.UserUuid),
	).Result()

	if err != nil {
		return nil, err

	}

	return parseRegisterMobileVerifyCode(val), nil
}

func parseRegisterMobileVerifyCode(res map[string]string) *RegisterMobileVerifyCode {
	m := &RegisterMobileVerifyCode{}

	if v, ok := res[RegisterMobileVerifyCodeFieldKey]; ok {
		m.VerifyCode = v
	}

	if v, ok := res[RegisterMobileNumberFieldKey]; ok {
		m.Mobile = v
	}

	return m
}
