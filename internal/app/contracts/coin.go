package contracts

import (
	"github.com/huangc28/go-darkpanda-backend/db"
	"github.com/huangc28/go-darkpanda-backend/internal/app/models"
)

type OrderCoinParams struct {
	BuyerID     int                `json:"buyer_id"`
	Amount      int                `json:"amount"`
	Cost        int                `json:"cost"`
	OrderStatus models.OrderStatus `json:"order_status"`
}

type UpdateOrderCoinStatusParams struct {
	ID          int                `json:"id"`
	OrderStatus models.OrderStatus `json:"order_status"`
}

type CoinDAOer interface {
	WithTx(tx db.Conn) CoinDAOer
}
