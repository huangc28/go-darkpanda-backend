package contracts

import (
	"github.com/huangc28/go-darkpanda-backend/db"
	"github.com/huangc28/go-darkpanda-backend/internal/app/models"
)

type OrderCoinParams struct {
	BuyerID     int32              `json:"buyer_id"`
	Amount      float32            `json:"amount"`
	Cost        float32            `json:"cost"`
	OrderStatus models.OrderStatus `json:"order_status"`
}

type CoinDAOer interface {
	WithTx(tx db.Conn) CoinDAOer
}
