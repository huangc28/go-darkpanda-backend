package payment

import (
	"github.com/golobby/container/pkg/container"
	"github.com/huangc28/go-darkpanda-backend/db"
	"github.com/huangc28/go-darkpanda-backend/internal/app/contracts"
	"github.com/huangc28/go-darkpanda-backend/internal/app/models"
	"github.com/jmoiron/sqlx"
)

type PaymentDAO struct {
	DB *sqlx.DB
}

func NewPaymentDAO(DB *sqlx.DB) *PaymentDAO {
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
  payments.price,
  payments.rec_trade_id,

  -- Retrieve service info
  services.address,
  services.start_time,
  services.duration,

  -- Retrieve picker info
  pickers.uuid AS picker_uuid,
  pickers.username AS picker_username,
  pickers.avatar_url AS picker_avatar_url
FROM payments
INNER JOIN services
	ON services.id = payments.service_id
	AND services.uuid =  $1
INNER JOIN users AS pickers ON pickers.id = services.service_provider_id;
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
