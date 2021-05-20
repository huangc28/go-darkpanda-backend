package coin

import (
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