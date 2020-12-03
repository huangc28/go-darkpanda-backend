package contracts

import "github.com/huangc28/go-darkpanda-backend/internal/app/models"

type PaymentDAOer interface {
	GetPaymentsByUuid(uuid string) ([]models.PaymentInfo, error)
}
