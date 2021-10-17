package coin

import (
	"database/sql"
	"errors"
	"fmt"
	"math"

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

func (dao *UserBalanceDAO) CreateOrTopUpBalance(params contracts.CreateOrTopUpBalanceParams) (*models.UserBalance, error) {
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
		params.UserID,
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

func (dao *UserBalanceDAO) deductCostFromBalance(userID int, cost float64) (*models.UserBalance, error) {
	query := `
UPDATE user_balance
SET balance = balance - $1
WHERE user_id = $2
RETURNING *;
`
	var m models.UserBalance

	if err := dao.db.QueryRowx(
		query,
		math.Round(cost*100)/100,
		userID,
	).StructScan(&m); err != nil {
		return nil, err
	}

	return &m, nil
}

func (dao *UserBalanceDAO) DeductUserPackageCostFromBalance(userID int, pkg *models.CoinPackage) (*models.UserBalance, error) {
	cDeci, err := decimal.NewFromString(pkg.Cost.String)

	if err != nil {
		return nil, err
	}

	cF, _ := cDeci.Float64()

	return dao.deductCostFromBalance(userID, cF)
}

func (s *UserBalanceDAO) DeductMachingFee(userID int, matchingFee decimal.Decimal) (*models.UserBalance, error) {
	mf, _ := matchingFee.Float64()

	return s.deductCostFromBalance(userID, mf)
}

func (s *UserBalanceDAO) HasEnoughBalanceToChargePackage(userID int, pkg *models.CoinPackage) error {
	ub, err := s.GetCoinBalanceByUserId(userID)

	if err != nil {
		return err
	}

	balanceDeci, err := decimal.NewFromString(ub.Balance)

	if err != nil {
		return err
	}

	pkgCost, err := decimal.NewFromString(pkg.Cost.String)

	if err != nil {
		return err
	}

	hasEnough := balanceDeci.GreaterThan(pkgCost) || balanceDeci.Equal(pkgCost)

	if !hasEnough {
		return errors.New("insufficient fund")
	}

	return nil
}

func (s *UserBalanceDAO) HasEnoughBalanceToCharge(userID int, cost decimal.Decimal) error {
	ub, err := s.GetCoinBalanceByUserId(userID)

	if err != nil {
		// If user does not have a balance record, we create a new balance record.
		// It's impossible for user to have enough blance to purchase any product.
		if err == sql.ErrNoRows {
			_, err = s.CreateOrTopUpBalance(contracts.CreateOrTopUpBalanceParams{
				UserID:      userID,
				TopupAmount: 0,
			})

			return fmt.Errorf("insufficient fund: %v", err.Error())
		}

		return err
	}

	balanceDeci, err := decimal.NewFromString(ub.Balance)

	if err != nil {
		return err
	}

	hasEnough := balanceDeci.GreaterThan(cost) || balanceDeci.Equal(cost)

	if !hasEnough {
		return errors.New("insufficient fund")
	}

	return nil

}

func (dao *UserBalanceDAO) AddBalance(userID int, amount decimal.Decimal) error {
	query := `
UPDATE user_balance
SET balance = balance + $1
WHERE user_id = $2
RETURNING *;
	`
	amountF, _ := amount.Float64()

	if _, err := dao.db.Exec(query, amountF, userID); err != nil {
		return err
	}

	return nil
}
