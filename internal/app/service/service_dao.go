package service

import (
	"database/sql"
	"fmt"

	"github.com/huangc28/go-darkpanda-backend/db"
	"github.com/huangc28/go-darkpanda-backend/internal/app/contracts"
	"github.com/huangc28/go-darkpanda-backend/internal/app/models"
	"github.com/jmoiron/sqlx"
	"github.com/prometheus/common/log"

	cintrnal "github.com/golobby/container/pkg/container"
)

type ServiceDAO struct {
	DB db.Conn
}

func NewServiceDAO(db *sqlx.DB) *ServiceDAO {
	return &ServiceDAO{
		DB: db,
	}
}

func ServiceDAOServiceProvider(c cintrnal.Container) func() error {
	return func() error {
		c.Transient(func() contracts.ServiceDAOer {
			return NewServiceDAO(db.GetDB())
		})

		return nil
	}
}

func (dao *ServiceDAO) WithTx(tx db.Conn) contracts.ServiceDAOer {
	dao.DB = tx

	return dao
}

// GetIncomingServicesByProviderId Gets list of services by service provider ID.
//   - service uuid
//   - service status
//   - appointment time
//   - customer name
//   - customer uuid
//   - customer avatar
//   - chatroom channel uuid to subscribe to chatroom
type IncomingServiceResult struct {
	ServiceUuid     sql.NullString `json:"service_uuid"`
	ServiceStatus   sql.NullString `json:"service_status"`
	AppointmentTime sql.NullTime   `json:"appointment_time"`
	Username        sql.NullString `json:"username"`
	UserUuid        sql.NullString `json:"user_uuid"`
	AvatarUrl       sql.NullString `json:"avatar_url"`
	ChannelUuid     sql.NullString `json:"channel_uuid"`
}

func (dao *ServiceDAO) GetIncomingServicesByProviderId(providerID int) ([]IncomingServiceResult, error) {
	sql := `
SELECT
	services.uuid as service_uuid,
	services.service_status,
	services.appointment_time,
	users.username,
	users.uuid as user_uuid,
	users.avatar_url,
	chatrooms.channel_uuid
FROM services
INNER JOIN users
	ON services.customer_id = users.id
JOIN chatrooms
	ON services.inquiry_id = chatrooms.inquiry_id
WHERE
	service_provider_id = $1
AND
	(
		service_status = $2 OR
		service_status = $3
	);
	`

	rows, err := dao.DB.Queryx(
		sql,
		providerID,
		models.ServiceStatusUnpaid,
		models.ServiceStatusToBeFulfilled,
	)

	if err != nil {
		log.Errorf("Failed to get incoming services")

		return nil, err
	}

	defer rows.Close()

	srvs := make([]IncomingServiceResult, 0)

	for rows.Next() {
		srv := IncomingServiceResult{}
		if err := rows.StructScan(&srv); err != nil {
			return nil, err
		}

		srvs = append(srvs, srv)
	}

	return srvs, nil
}

func (dao *ServiceDAO) GetUserHistoricalServicesByUuid(uuid string, perPage int, offset int) ([]models.Service, error) {
	query := `
SELECT
	services.uuid,
	price,
	service_type,
	service_status,
	services.created_at
FROM services
INNER JOIN users ON users.id = customer_id
WHERE users.uuid = $1
ORDER BY created_at DESC
LIMIT $2
OFFSET $3;
	`
	rows, err := dao.DB.Query(query, uuid, perPage, offset)
	defer rows.Close()

	if err != nil {
		return nil, err
	}

	services := make([]models.Service, 0)

	for rows.Next() {
		service := models.Service{}

		rows.Scan(
			&service.Uuid,
			&service.Price,
			&service.ServiceType,
			&service.ServiceStatus,
			&service.CreatedAt,
		)

		services = append(services, service)
	}

	return services, nil
}

func (dao *ServiceDAO) GetServiceByInquiryUUID(uuid string, fields ...string) (*models.Service, error) {
	if len(fields) == 0 {
		fields = append(fields, "*")
	}

	baseQuery := `
SELECT %s
FROM services
LEFT JOIN service_inquiries ON service_inquiries.id = services.inquiry_id
WHERE service_inquiries.uuid = $1;
`

	query := fmt.Sprintf(baseQuery, db.ComposeFieldsSQLString(fields...))
	service := models.Service{}

	if err := dao.DB.QueryRowx(query, uuid).StructScan(&service); err != nil {
		return (*models.Service)(nil), err
	}

	return &service, nil
}

func (dao *ServiceDAO) UpdateServiceByID(params contracts.UpdateServiceByIDParams) (*models.Service, error) {
	sql := `
UPDATE services SET
	price = COALESCE($1, price),
	uuid = uuid,
	duration = COALESCE($2, duration),
	appointment_time = COALESCE($3, appointment_time),
	service_status = COALESCE($4, service_status),
	service_type = COALESCE($5, service_type)
WHERE id = $6
RETURNING
	uuid,
	price,
	duration,
	appointment_time,
	service_type;
	`
	service := models.Service{}

	if err := dao.DB.QueryRowx(
		sql,
		params.Price,
		params.Duration,
		params.Appointment,
		params.ServiceStatus,
		params.ServiceType,
		params.ID,
	).StructScan(&service); err != nil {
		return (*models.Service)(nil), err
	}

	return &service, nil
}
