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
	InquiryUuid     sql.NullString `json:"inquiry_uuid"`
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

	// Female user requesting service list, we retrieve customer info.
	// vice versa, male user requesting service list, we retrieve service
	// provider info.
	joinTargetPersonClause := ""
	if gender == models.GenderFemale {
		whereClause += "services.service_provider_id = $1"

		joinTargetPersonClause += "services.customer_id = users.id"
	} else {
		whereClause += "services.customer_id = $1"

		joinTargetPersonClause += "services.service_provider_id = users.id"
	}

	query := fmt.Sprintf(
		`
SELECT
	services.uuid as service_uuid,
	services.service_status,
	services.appointment_time,
	users.username,
	users.uuid as user_uuid,
	users.avatar_url,
	chatrooms.channel_uuid,
	service_inquiries.uuid as inquiry_uuid
FROM services
INNER JOIN users
	ON %s
INNER JOIN service_inquiries
	ON services.inquiry_id = service_inquiries.id
INNER JOIN chatrooms
	ON services.inquiry_id = chatrooms.inquiry_id
WHERE
	%s
AND %s
ORDER BY services.created_at DESC
LIMIT $2
OFFSET $3;
	`,
		joinTargetPersonClause,
		whereClause,
		sCondStr,
	)

	rows, err := dao.DB.Queryx(
		query,
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
	query := `
UPDATE services SET
	price = COALESCE($1, price),
	uuid = uuid,
	duration = COALESCE($2, duration),
	appointment_time = COALESCE($3, appointment_time),
	service_status = COALESCE($4, service_status),
	service_type = COALESCE($5, service_type),
	start_time = COALESCE($6, start_time),
	end_time = COALESCE($7, end_time)
WHERE id = $8
RETURNING *;
	`
	service := models.Service{}

	if err := dao.DB.QueryRowx(
		query,
		params.Price,
		params.Duration,
		params.Appointment,
		params.ServiceStatus,
		params.ServiceType,
		params.StartTime,
		params.EndTime,
		params.ID,
	).StructScan(&service); err != nil {
		return (*models.Service)(nil), err
	}

	return &service, nil
}

func (dao *ServiceDAO) CreateServiceQRCode(params contracts.CreateServiceQRCodeParams) (*models.ServiceQrcode, error) {
	query := `
INSERT INTO service_qrcode (
	uuid,
	url,
	service_id
) VALUES ($1, $2, $3)
RETURNING *;
`
	var m models.ServiceQrcode

	if err := dao.DB.QueryRowx(
		query,
		params.Uuid,
		params.Url,
		params.ServiceId,
	).StructScan(&m); err != nil {
		return nil, err
	}

	return &m, nil
}

func (dao *ServiceDAO) GetServiceByQrcodeUuid(qrCodeUuid string) (*models.Service, error) {
	query := `
SELECT
	services.*
FROM
	services
INNER JOIN service_qrcode
	ON service_qrcode.service_id = services.id
WHERE
	service_qrcode.uuid = $1;
`

	var m models.Service

	if err := dao.DB.QueryRowx(query, qrCodeUuid).StructScan(&m); err != nil {
		return nil, err
	}

	return &m, nil
}

// ScanExpiredServices scan services with service status `to_be_fulfilled`.
// If current time is later than the service end_time, we set the service status to be `expired`
func (dao *ServiceDAO) ScanExpiredServices() ([]*models.Service, error) {
	return dao.ScanAndUpdateServiceStatusIfNeeded(
		ScanAndUpdateServiceStatusIfNeededParams{
			ScanStatus:     string(models.ServiceStatusToBeFulfilled),
			UpdateToStatus: string(models.ServiceStatusExpired),
		},
	)
}

// ScanCompletedServices scan those services with status `fulfilling`. If current time
// is greater than `end_time`, update the status to `completed`
func (dao *ServiceDAO) ScanCompletedServices() ([]*models.Service, error) {
	return dao.ScanAndUpdateServiceStatusIfNeeded(
		ScanAndUpdateServiceStatusIfNeededParams{
			ScanStatus:     string(models.ServiceStatusFulfilling),
			UpdateToStatus: string(models.ServiceStatusCompleted),
		},
	)
}

type ScanAndUpdateServiceStatusIfNeededParams struct {
	ScanStatus     string
	UpdateToStatus string
}

func (dao *ServiceDAO) ScanAndUpdateServiceStatusIfNeeded(params ScanAndUpdateServiceStatusIfNeededParams) ([]*models.Service, error) {
	query := `
WITH found_services AS (
	SELECT
		id,
		uuid
	FROM
		services
	WHERE
		service_status = $1 AND
		now() >= end_time
), updated AS (
	UPDATE
		services
	SET
		service_status = $2
	FROM
		 found_services
	WHERE
		found_services.id = services.id
)
SELECT id, uuid FROM found_services;
`
	rows, err := dao.DB.Queryx(
		query,
		params.ScanStatus,
		params.UpdateToStatus,
	)

	defer rows.Close()

	if err != nil {
		return nil, err
	}

	srvs := make([]*models.Service, 0)

	for rows.Next() {
		srv := models.Service{}

		if err := rows.StructScan(&srv); err != nil {
			return nil, err
		}

		srvs = append(srvs, &srv)
	}

	return srvs, nil
}

func (dao *ServiceDAO) GetQrcodeByServiceUuid(srvUuid string) (*models.ServiceQrcode, error) {
	query := `
SELECT
	service_qrcode.*
FROM
	service_qrcode
INNER JOIN services
	ON service_qrcode.service_id = services.id
	AND services.uuid = $1;

`

	var m models.ServiceQrcode

	if err := dao.DB.QueryRowx(query, srvUuid).StructScan(&m); err != nil {
		return nil, err
	}

	return &m, nil
}

func (dao *ServiceDAO) GetServiceNames() ([]*models.ServiceName, error) {
	query := `
SELECT
	service_names.*
FROM service_names;
`

	rows, err := dao.DB.Queryx(query)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	srvNames := make([]*models.ServiceName, 0)

	for rows.Next() {
		srvName := models.ServiceName{}

		if err := rows.StructScan(&srvName); err != nil {
			return nil, err
		}

		srvNames = append(srvNames, &srvName)
	}

	return srvNames, nil
}

func (dao *ServiceDAO) GetServiceByUuid(srvUuid string, fields ...string) (*models.Service, error) {
	if len(fields) == 0 {
		fields = append(fields, "*")
	}

	fieldsStr := strings.TrimSuffix(strings.Join(fields, ","), ",")

	baseQuery := `
SELECT %s
FROM services
WHERE uuid = $1;
`
	query := fmt.Sprintf(baseQuery, fieldsStr)

	var m models.Service

	if err := dao.DB.QueryRowx(query, srvUuid).StructScan(&m); err != nil {
		return nil, err
	}

	return &m, nil
}
