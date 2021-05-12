package coin

import (
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
	var coin models.CoinOrder

	query := `
		INSERT INTO coin_orders(
			buyer_id,
			amount,
			cost,
			order_status
		) VALUES ($1, $2, $3, $4)
		RETURNING *;
	`

	if err := dao.
		db.QueryRowx(
		query,
		params.BuyerID,
		params.Amount,
		params.Cost,
		params.OrderStatus,
	).StructScan(&coin); err != nil {
		return nil, err
	}
	// if _, err := dao.db.Exec(
	// 	query,
	// 	params.BuyerID,
	// 	params.Amount,
	// 	params.Cost,
	// 	params.OrderStatus,
	// ); err != nil {
	// 	return nil,err
	// }

	return &coin, nil
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
