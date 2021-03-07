package service

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/huangc28/go-darkpanda-backend/db"
	"github.com/huangc28/go-darkpanda-backend/internal/app/contracts"
	"github.com/huangc28/go-darkpanda-backend/internal/app/models"
	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"

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

type ServiceResult struct {
	ServiceUuid     sql.NullString `json:"service_uuid"`
	ServiceStatus   sql.NullString `json:"service_status"`
	AppointmentTime sql.NullTime   `json:"appointment_time"`
	Username        sql.NullString `json:"username"`
	UserUuid        sql.NullString `json:"user_uuid"`
	AvatarUrl       sql.NullString `json:"avatar_url"`
	ChannelUuid     sql.NullString `json:"channel_uuid"`
}

// GetServicesByStatus gets services of given status
//   - service uuid
//   - service status
//   - appointment time
//   - customer name
//   - customer uuid
//   - customer avatar
//   - chatroom channel uuid to subscribe to chatroom
func (dao *ServiceDAO) GetServicesByStatus(providerID int, gender models.Gender, offset, perPage int, slist ...models.ServiceStatus) ([]ServiceResult, error) {
	// Default to have 10 records perpage
	if perPage == 0 {
		perPage = 10
	}

	// Start formatting status condition string
	sCondStr := ""

	for _, s := range slist {
		sCondStr += fmt.Sprintf(
			"services.service_status = '%s' OR ",
			s.ToString(),
		)
	}

	sCondStr = fmt.Sprintf("(%s)", strings.TrimSuffix(sCondStr, " OR "))

	// If gender equals female, the column to match should be `service_provider_id`
	// If gender equals male, the columns to match should be `customer_id`
	whereClause := ""

	if gender == models.GenderFemale {
		whereClause += "services.service_provider_id = $1"
	} else {
		whereClause += "services.customer_id = $1"
	}

	sql := fmt.Sprintf(
		`
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
LEFT JOIN chatrooms
	ON services.inquiry_id = chatrooms.inquiry_id
WHERE
	%s
AND %s
ORDER BY services.created_at DESC
LIMIT $2
OFFSET $3;
	`,
		whereClause,
		sCondStr,
	)

	rows, err := dao.DB.Queryx(
		sql,
		providerID,
		perPage,
		offset,
	)

	if err != nil {
		log.Errorf("Failed to get service list")

		return nil, err
	}

	defer rows.Close()
	srvs := make([]ServiceResult, 0)
	for rows.Next() {
		srv := ServiceResult{}
		if err := rows.StructScan(&srv); err != nil {
			return nil, err
		}

		srvs = append(srvs, srv)
	}

	return srvs, nil
}

type GetServicesParams struct {
	UserID  int
	Offset  int
	PerPage int
}

// GetIncomingServicesByProviderId Gets list of services of following service status:
//   - unpaid
//   - to_be_fulfilled
func (dao *ServiceDAO) GetIncomingServicesByProviderId(p GetServicesParams) ([]ServiceResult, error) {
	return dao.GetServicesByStatus(
		p.UserID,
		models.GenderFemale,
		p.Offset,
		p.PerPage,
		models.ServiceStatusUnpaid,
		models.ServiceStatusToBeFulfilled,
	)
}

// GetOverduedServicesByProviderId Get list of services of following service status:
//   - canceled
//   - completed
//   - failed_due_to_due
//   - failed_due_to_girl
//   - failed_due_to_man
func (dao *ServiceDAO) GetOverduedServicesByProviderId(p GetServicesParams) ([]ServiceResult, error) {
	return dao.GetServicesByStatus(
		p.UserID,
		models.GenderFemale,
		p.Offset,
		p.PerPage,
		models.ServiceStatusCanceled,
		models.ServiceStatusCompleted,
		models.ServiceStatusFailedDueToBoth,
		models.ServiceStatusFailedDueToGirl,
		models.ServiceStatusFailedDueToMan,
	)
}

func (dao *ServiceDAO) GetIncomingServicesByCustomerId(p GetServicesParams) ([]ServiceResult, error) {
	return dao.GetServicesByStatus(
		p.UserID,
		models.GenderMale,
		p.Offset,
		p.PerPage,
		models.ServiceStatusUnpaid,
		models.ServiceStatusToBeFulfilled,
	)
}

func (dao *ServiceDAO) GetOverduedServicesByCustomerId(p GetServicesParams) ([]ServiceResult, error) {
	return dao.GetServicesByStatus(
		p.UserID,
		models.GenderMale,
		p.Offset,
		p.PerPage,
		models.ServiceStatusCanceled,
		models.ServiceStatusCompleted,
		models.ServiceStatusFailedDueToBoth,
		models.ServiceStatusFailedDueToGirl,
		models.ServiceStatusFailedDueToMan,
	)
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
