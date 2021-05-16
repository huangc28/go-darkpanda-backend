package coin

import (
	"log"

	"github.com/huangc28/go-darkpanda-backend/db"
	"github.com/huangc28/go-darkpanda-backend/internal/app/models"
)

type UserBalanceDAO struct {
	db db.Conn
}

func NewUserBalanceDAO(db db.Conn) *UserBalanceDAO {
	return &UserBalanceDAO{
		db: db,
	}
}

type CreateOrTopUpBalanceParams struct {
	UserId      int
	TopupAmount float64
}

func (dao *UserBalanceDAO) CreateOrTopUpBalance(params CreateOrTopUpBalanceParams) (*models.UserBalance, error) {
	log.Printf("DEBUG spot 1** %v", params.UserId)

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
