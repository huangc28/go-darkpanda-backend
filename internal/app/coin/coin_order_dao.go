package coin

import (
	"errors"

	"github.com/golobby/container/pkg/container"
	"github.com/huangc28/go-darkpanda-backend/db"
	"github.com/huangc28/go-darkpanda-backend/internal/app/contracts"
	"github.com/huangc28/go-darkpanda-backend/internal/app/models"
)

type CoinDAO struct {
	db db.Conn
}

func NewCoinDAO(db db.Conn) *CoinDAO {
	return &CoinDAO{
		db: db,
	}
}

func CoinDAOServiceProvider(c container.Container) func() error {
	return func() error {
		c.Transient(func() contracts.CoinDAOer {
			return NewCoinDAO(db.GetDB())
		})

		return nil
	}
}

func (dao *CoinDAO) WithTx(tx db.Conn) contracts.CoinDAOer {
	dao.db = tx

	return dao
}

func (dao *CoinDAO) OrderCoin(params contracts.OrderCoinParams) (*models.CoinOrder, error) {
	query := `
INSERT INTO coin_orders(
	buyer_id,
	cost,
	order_status,
	package_id,
	quantity
) VALUES ($1, $2, $3, $4, $5)
RETURNING *;
	`

	coinOrderModel := models.CoinOrder{}

	if err := dao.
		db.QueryRowx(
		query,
		params.BuyerID,
		params.Cost,
		params.OrderStatus,
		params.PackageId,
		params.Quantity,
	).StructScan(&coinOrderModel); err != nil {
		return nil, err
	}

	return &coinOrderModel, nil
}

func (dao *CoinDAO) UpdateOrderCoinStatus(params contracts.UpdateOrderCoinStatusParams) error {
	query := `
		UPDATE coin_orders
		SET order_status = COALESCE($1, order_status)
		WHERE id=$2;
	`

	if _, err := dao.db.Exec(
		query,
		params.OrderStatus,
		params.ID,
	); err != nil {
		return err
	}

	return nil
}

type UpdateOrderCoinByIdParam struct {
	OrderStatus models.OrderStatus
	RecTradeId  string
	Raw         string
	Id          int
}

func (dao *CoinDAO) UpdateOrderCoinById(params UpdateOrderCoinByIdParam) (*models.CoinOrder, error) {
	if params.Id == 0 {
		return nil, errors.New("Failed to update order coin, Id not provided.")
	}

	query := `
UPDATE
	coin_orders
SET
	order_status = COALESCE($1, order_status),
	rec_trade_id = COALESCE($2, rec_trade_id),
	raw = COALESCE($3, raw)
WHERE
	id = $4
RETURNING *;
	`
	coinOrderModel := models.CoinOrder{}

	if err := dao.db.QueryRowx(
		query,
		params.OrderStatus,
		params.RecTradeId,
		params.Raw,
		params.Id,
	).StructScan(&coinOrderModel); err != nil {
		return nil, err
	}

	return &coinOrderModel, nil
}
