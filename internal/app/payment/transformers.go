package payment

import (
	"github.com/huangc28/go-darkpanda-backend/internal/app/models"
	"github.com/shopspring/decimal"
)

func TrfCreatePayment(bal *models.UserBalance, user *models.User) (interface{}, error) {
	balDeci, err := decimal.NewFromString(bal.Balance)

	if err != nil {
		return nil, err
	}

	balFloat, _ := balDeci.Float64()

	return struct {
		Uuid    string  `json:"uuid"`
		Balance float64 `json:"balance"`
	}{
		Uuid:    user.Uuid,
		Balance: balFloat,
	}, nil
}
