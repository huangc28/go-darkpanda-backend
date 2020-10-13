package payment

import (
	"database/sql"

	"github.com/huangc28/go-darkpanda-backend/internal/models"
)

type PaymentDAO struct {
	DB *sql.DB
}

func (dao *PaymentDAO) Ping() error {
	return dao.DB.Ping()
}

func (dao *PaymentDAO) GetPaymentsByUuid(uuid string) ([]models.PaymentInfo, error) {
	// Retrieve list of payments first.
	// Retrieve payee info and service info via keys from payment.
	query := `
SELECT
  -- Retrieve payment info
  payment.id,
  payment.price,
  payment.rec_trade_id,

  -- Retrieve service info
  services.uuid,
  services.service_type,
  services.price,

  -- Retrieve payer info
  payer.uuid,
  payer.username,
  payer.avatar_url
FROM (
 SELECT
	payments.id,
	payments.payee_id,
	payments.payer_id,
	payments.service_id,
	payments.price,
	payments.rec_trade_id
 FROM payments
 INNER JOIN users ON users.id = payments.payer_id
 WHERE users.uuid = $1
) AS payment
INNER JOIN services ON services.id = payment.service_id
INNER JOIN users AS payer ON payer.id = payment.payer_id;
	`

	paymentInfos := make([]models.PaymentInfo, 0)

	rows, err := dao.DB.Query(query, uuid)
	defer rows.Close()

	if err != nil {
		return nil, err
	}

	for rows.Next() {
		paymentInfo := models.PaymentInfo{
			Service: models.Service{},
			Payer:   models.User{},
		}

		err = rows.Scan(
			&paymentInfo.ID,
			&paymentInfo.Price,
			&paymentInfo.RecTradeID,

			&paymentInfo.Service.Uuid,
			&paymentInfo.Service.ServiceType,
			&paymentInfo.Service.Price,

			&paymentInfo.Payer.Uuid,
			&paymentInfo.Payer.Username,
			&paymentInfo.Payer.AvatarUrl,
		)

		if err != nil {
			return nil, err
		}

		paymentInfos = append(paymentInfos, paymentInfo)
	}

	return paymentInfos, nil
}
