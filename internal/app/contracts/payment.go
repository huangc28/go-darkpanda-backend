package contracts

import (
	"github.com/huangc28/go-darkpanda-backend/db"
	"github.com/huangc28/go-darkpanda-backend/internal/app/models"
)

type PaymentDAOer interface {
	WithTx(tx db.Conn) PaymentDAOer
	GetPaymentsByUuid(uuid string) ([]models.PaymentInfo, error)
	GetPaymentByServiceUuid(srvUuid string) (*models.ServicePaymentDetail, error)
	SetRefunded(paymentID int) error
}
