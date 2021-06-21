package coin

import (
	"errors"

	cintrnal "github.com/golobby/container/pkg/container"
	"github.com/huangc28/go-darkpanda-backend/db"
	"github.com/huangc28/go-darkpanda-backend/internal/app/contracts"
	"github.com/huangc28/go-darkpanda-backend/internal/app/models"
	"github.com/shopspring/decimal"
)

type UserBalanceDAO struct {
	db db.Conn
}

func NewUserBalanceDAO(db db.Conn) *UserBalanceDAO {
	return &UserBalanceDAO{
		db: db,
	}
}

func UserBalanceDAOServiceProvider(c cintrnal.Container) func() error {
	return func() error {
		c.Transient(func() contracts.UserBalancer {
			return NewUserBalanceDAO(db.GetDB())
		})

		return nil
	}
}

func (dao *UserBalanceDAO) WithTx(tx db.Conn) contracts.UserBalancer {
	dao.db = tx

	return dao
}

type CreateOrTopUpBalanceParams struct {
	UserId      int
	TopupAmount float64
}

func (dao *UserBalanceDAO) CreateOrTopUpBalance(params CreateOrTopUpBalanceParams) (*models.UserBalance, error) {
	query := `
INSERT INTO user_balance (user_id, balance)
VALUES ($1, $2)
ON CONFLICT (user_id)
DO UPDATE SET balance = user_balance.balance + $3
RETURNING *;
	`

	userBalance := models.UserBalance{}

	if err := dao.db.QueryRowx(
		query,
		params.UserId,
		params.TopupAmount,
		params.TopupAmount,
	).StructScan(&userBalance); err != nil {
		return nil, err
	}

	return &userBalance, nil
}

func (dao *UserBalanceDAO) GetCoinBalanceByUserId(userId int) (*models.UserBalance, error) {
	query := `
SELECT * FROM user_balance WHERE user_id = $1;
`
	userBal := models.UserBalance{}

	if err := dao.db.QueryRowx(
		query,
		userId,
	).StructScan(&userBal); err != nil {
		return (*models.UserBalance)(nil), err
	}

	return &userBal, nil
}

func (dao *UserBalanceDAO) DeductUserBalance(userId int, pkg *models.CoinPackage) (*models.UserBalance, error) {
	query := `
UPDATE user_balance
SET balance = balance - $1
WHERE user_id = $2
RETURNING *;
`

	var m models.UserBalance

	if err := dao.db.QueryRowx(query, pkg.Cost.Int32, userId).StructScan(&m); err != nil {
		return nil, err
	}

	return &m, nil
}

func (s *UserBalanceDAO) HasEnoughBalanceToCharge(userId int, pkg *models.CoinPackage) error {
	ub, err := s.GetCoinBalanceByUserId(userId)

	if err != nil {
		return err
	}

	balanceDeci, err := decimal.NewFromString(ub.Balance)

	if err != nil {
		return err
	}

	pkgCost := decimal.NewFromInt32(pkg.Cost.Int32)

	hasEnough := balanceDeci.GreaterThan(pkgCost) || balanceDeci.Equal(pkgCost)

	if !hasEnough {
		return errors.New("insufficient fund")
	}

	return nil
}
