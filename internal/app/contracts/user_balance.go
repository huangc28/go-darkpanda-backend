package contracts

import (
	"github.com/huangc28/go-darkpanda-backend/db"
	"github.com/huangc28/go-darkpanda-backend/internal/app/models"
)

type UserBalancer interface {
	WithTx(tx db.Conn) UserBalancer
	GetCoinBalanceByUserId(userId int) (*models.UserBalance, error)
	DeductUserBalance(userId int, pkg *models.CoinPackage) (*models.UserBalance, error)
	HasEnoughBalanceToCharge(userId int, pkg *models.CoinPackage) error
}
