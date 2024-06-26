package payment

import (
	"github.com/golobby/container/pkg/container"
	"github.com/huangc28/go-darkpanda-backend/db"
	"github.com/huangc28/go-darkpanda-backend/internal/app/contracts"
	"github.com/huangc28/go-darkpanda-backend/internal/app/models"
)

type PaymentDAO struct {
	DB db.Conn
}

func NewPaymentDAO(DB db.Conn) *PaymentDAO {
	return &PaymentDAO{
		DB: DB,
	}
}

func PaymentDAOServiceProvider(c container.Container) func() error {
	return func() error {
		c.Transient(func() contracts.PaymentDAOer {
			return NewPaymentDAO(db.GetDB())
		})

		return nil
	}
}

func (dao *PaymentDAO) WithTx(tx db.Conn) contracts.PaymentDAOer {
	dao.DB = tx

	return dao
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
	payments.price,
	payments.rec_trade_id,
	payments.service_id,
	payments.payer_id
 FROM payments
 INNER JOIN users ON users.id = payments.payer_id
 WHERE users.uuid = $1
) AS payment
INNER JOIN services ON services.id = payment.service_id
INNER JOIN users AS payer ON payer.id = payment.payer_id;
	`

	paymentInfos := make([]models.PaymentInfo, 0)

	rows, err := dao.DB.Query(query, uuid)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		paymentInfo := models.PaymentInfo{
			Service: models.Service{},
			Payer:   models.User{},
		}

		err = rows.Scan(
			&paymentInfo.ID,
			&paymentInfo.Price,

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

func (dao *PaymentDAO) GetPaymentByServiceUuid(srvUuid string) (*models.ServicePaymentDetail, error) {
	query := `
	SELECT
		-- Retrieve payment info
		payments.id AS payment_id,
	 	payments.price,
		(
			CASE WHEN payments.refunded IS NULL
			THEN
				false
			ELSE
				payments.refunded::BOOLEAN
			END
		) AS refunded,

	 	-- Retrieve service info
		services.service_type,
	 	services.address,
	 	services.appointment_time,
	 	services.duration,
		services.cancel_cause,
		services.currency,

	  	-- Retrieve picker info
	  	pickers.uuid AS picker_uuid,
	  	pickers.username AS picker_username,
	  	pickers.avatar_url AS picker_avatar_url
	FROM
		services
	LEFT JOIN payments ON payments.service_id = services.id
	INNER JOIN users AS pickers ON pickers.id = services.service_provider_id
	WHERE
		services.uuid = $1
`

	var m models.ServicePaymentDetail

	if err := dao.DB.QueryRowx(
		query,
		srvUuid,
	).StructScan(&m); err != nil {
		return nil, err
	}

	return &m, nil
}

func (dao *PaymentDAO) SetRefunded(paymentID int) error {
	query := `
UPDATE payments
SET refunded = true
WHERE id = $1
RETURNING *;
	`
	if _, err := dao.DB.Exec(query, paymentID); err != nil {
		return err
	}

	return nil
}
