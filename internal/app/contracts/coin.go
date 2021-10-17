package contracts

import (
	"github.com/huangc28/go-darkpanda-backend/db"
	"github.com/huangc28/go-darkpanda-backend/internal/app/models"
)

type OrderCoinParams struct {
	BuyerID     int
	PackageId   int
	Quantity    int
	Cost        float64
	OrderStatus models.OrderStatus
}

type UpdateOrderCoinStatusParams struct {
	ID          int                `json:"id"`
	OrderStatus models.OrderStatus `json:"order_status"`
}

type CoinDAOer interface {
	WithTx(tx db.Conn) CoinDAOer
	OrderCoin(params OrderCoinParams) (*models.CoinOrder, error)
}

type CoinPackageDAOer interface {
	GetMatchingFee() (*models.CoinPackage, error)
	GetMatchingFeeRate() (*models.CoinPackage, error)
}
