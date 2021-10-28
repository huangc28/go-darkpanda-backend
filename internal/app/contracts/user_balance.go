package contracts

import (
	"github.com/huangc28/go-darkpanda-backend/db"
	"github.com/huangc28/go-darkpanda-backend/internal/app/models"
	"github.com/shopspring/decimal"
)

type CreateOrTopUpBalanceParams struct {
	UserID      int
	TopupAmount int
}
type UserBalancer interface {
	WithTx(tx db.Conn) UserBalancer
	GetCoinBalanceByUserId(userId int) (*models.UserBalance, error)
	DeductUserPackageCostFromBalance(userId int, pkg *models.CoinPackage) (*models.UserBalance, error)
	DeductMachingFee(userID int, matchingFee decimal.Decimal) (*models.UserBalance, error)
	HasEnoughBalanceToChargePackage(userId int, pkg *models.CoinPackage) error
	HasEnoughBalanceToCharge(userID int, cost decimal.Decimal) error
	CreateOrTopUpBalance(params CreateOrTopUpBalanceParams) (*models.UserBalance, error)
	AddBalance(userID int, amount decimal.Decimal) error
}
