package service

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/huangc28/go-darkpanda-backend/db"
	"github.com/huangc28/go-darkpanda-backend/internal/app/contracts"
	"github.com/huangc28/go-darkpanda-backend/internal/app/models"
	"github.com/jmoiron/sqlx"

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
	CancelCause     string         `json:"cancel_cause"`
	Username        sql.NullString `json:"username"`
	UserUuid        sql.NullString `json:"user_uuid"`
	AvatarUrl       sql.NullString `json:"avatar_url"`
	ChannelUuid     sql.NullString `json:"channel_uuid"`
	InquiryUuid     sql.NullString `json:"inquiry_uuid"`
	CreatedAt       sql.NullTime   `json:"created_at"`
	Refunded        bool           `json:"refunded"`
}

// GetServicesByStatus gets services of given status
//   - service uuid
//   - service status
//   - appointment time
//   - customer name
//   - customer uuid
//   - customer avatar
//   - chatroom channel uuid to subscribe to chatroom
func (dao *ServiceDAO) GetServicesByStatus(providerID int, gender models.Gender, offset, perPage int, serviceType string, slist ...models.ServiceStatus) ([]ServiceResult, error) {
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

	// Incoming inquiries have to filter by chatroom deleted_at is null
	// Overdue inquiries no need filter by chatroom deleted_at is null
	joinChatroomClause := ""
	if serviceType == "incoming" {
		joinChatroomClause = "chatrooms.deleted_at IS null"
	} else {
		joinChatroomClause = "services.deleted_at IS null"
	}

	query := fmt.Sprintf(
		`
SELECT * FROM (
	SELECT * FROM (
		SELECT distinct ON(services.id)
			services.uuid as service_uuid,
			services.service_status,
			services.appointment_time,
			services.created_at,
			services.cancel_cause,
			users.username,
			users.uuid as user_uuid,
			users.avatar_url,
			chatrooms.channel_uuid,
			service_inquiries.uuid as inquiry_uuid,
			(
				CASE WHEN payments.refunded IS NULL
				THEN
					false
				ELSE
					payments.refunded::BOOLEAN
				END
			) AS refunded
		FROM services INNER JOIN users
			ON %s
		INNER JOIN service_inquiries
			ON services.inquiry_id = service_inquiries.id
		INNER JOIN chatrooms
			ON services.inquiry_id = chatrooms.inquiry_id AND
				%s
		LEFT JOIN payments ON payments.service_id = services.id
		WHERE %s
		AND %s
	) a ORDER BY created_at DESC
) b
	LIMIT $2
	OFFSET $3;
	`,
		joinTargetPersonClause,
		joinChatroomClause,
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
	rows, err := dao.DB.Queryx(query, uuid, perPage, offset)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	services := make([]models.Service, 0)

	for rows.Next() {
		service := models.Service{}

		if err := rows.StructScan(&service); err != nil {
			return nil, err
		}

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
	end_time = COALESCE($7, end_time),
	canceller_id = COALESCE($8, canceller_id),
	address = COALESCE($9, address),
	cancel_cause = COALESCE($10, cancel_cause),
	matching_fee = COALESCE($11, matching_fee)
WHERE id = $12
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
		params.CancellerId,
		params.Address,
		params.CancelCause,
		params.MatchingFee,
		params.ID,
	).StructScan(&service); err != nil {
		return (*models.Service)(nil), err
	}

	return &service, nil
}

func (dao *ServiceDAO) UpdateServiceByInquiryId(p contracts.UpdateServiceByInquiryIdParams) (*models.Service, error) {
	query := `
UPDATE services SET
	price = COALESCE($1, price),
	uuid = uuid,
	duration = COALESCE($2, duration),
	appointment_time = COALESCE($3, appointment_time),
	service_status = COALESCE($4, service_status),
	service_type = COALESCE($5, service_type),
	start_time = COALESCE($6, start_time),
	end_time = COALESCE($7, end_time),
	canceller_id = COALESCE($8, canceller_id)
WHERE inquiry_id = $9
RETURNING *;
`

	service := models.Service{}

	if err := dao.DB.QueryRowx(
		query,
		p.Price,
		p.Duration,
		p.Appointment,
		p.ServiceStatus,
		p.ServiceType,
		p.StartTime,
		p.EndTime,
		p.CancellerId,
		p.InquiryId,
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
func (dao *ServiceDAO) ScanExpiredServices() ([]*models.ServiceScannerData, error) {
	//return dao.ScanAndUpdateServiceStatusIfNeeded(
	//ScanAndUpdateServiceStatusIfNeededParams{
	//ScanStatus:     string(models.ServiceStatusToBeFulfilled),
	//UpdateToStatus: string(models.ServiceStatusExpired),
	//},
	//)
	query := fmt.Sprintf(
		`
WITH found_services AS (
	SELECT
		services.id,
		services.uuid,
		customers.username AS customer_username,
		customers.fcm_topic AS customer_fcm_topic,
		service_providers.username AS service_providers_username,
		service_providers.fcm_topic AS service_providers_fcm_topic
	FROM
		services
	INNER JOIN users AS customers ON services.customer_id = customers.id
	INNER JOIN users AS service_providers ON services.service_provider_id = service_providers.id
	WHERE
		services.service_status = $1 AND

		-- Allow 30 minutes buffer after appointment_time.
		now() >= appointment_time + (%d * interval '1 minute')
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
SELECT
	id,
	uuid,
	customer_username,
	customer_fcm_topic,
	service_providers_username,
	service_providers_fcm_topic
FROM
	found_services;
	`, 30,
	)

	rows, err := dao.DB.Queryx(
		query,
		string(models.ServiceStatusToBeFulfilled),
		string(models.ServiceStatusExpired),
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	srvs := make([]*models.ServiceScannerData, 0)

	for rows.Next() {
		srv := models.ServiceScannerData{}

		if err := rows.StructScan(&srv); err != nil {
			return nil, err
		}

		srvs = append(srvs, &srv)
	}

	return srvs, nil
}

// ScanCompletedServices scan those services with status `fulfilling`. If current time
// is greater than `end_time`, update the status to `completed`
func (dao *ServiceDAO) ScanCompletedServices() ([]*models.ServiceScannerData, error) {
	query := `
WITH found_services AS (
	SELECT
		services.id,
		services.uuid,
		customers.username AS customer_username,
		customers.fcm_topic AS customer_fcm_topic,
		service_providers.username AS service_providers_username,
		service_providers.fcm_topic AS service_providers_fcm_topic
	FROM
		services
	INNER JOIN users AS customers ON services.customer_id = customers.id
	INNER JOIN users AS service_providers ON services.service_provider_id = service_providers.id
	WHERE
		services.service_status = $1 AND
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
SELECT
	id,
	uuid,
	customer_username,
	customer_fcm_topic,
	service_providers_username,
	service_providers_fcm_topic
FROM
	found_services;
`
	rows, err := dao.DB.Queryx(
		query,
		string(models.ServiceStatusFulfilling),
		string(models.ServiceStatusCompleted),
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	srvs := make([]*models.ServiceScannerData, 0)

	for rows.Next() {
		srv := models.ServiceScannerData{}

		if err := rows.StructScan(&srv); err != nil {
			return nil, err
		}

		srvs = append(srvs, &srv)
	}

	return srvs, nil
}

// ScanInquiringServiceInquiries scan those service_inquiries with status `inquiring`.
// If created_at exceed 5 hour, update the status to `canceled`
func (dao *ServiceDAO) ScanInquiringServiceInquiries() ([]*models.ServiceScannerData, error) {
	query := `
WITH found_services AS (
	SELECT
		service_inquiries.id,
		service_inquiries.uuid
	FROM
		service_inquiries
	WHERE
		service_inquiries.inquiry_status = $1 AND
		
		-- Allow 5 hours buffer after created_at.
		now() >= created_at + interval '5 hour'
), updated AS (
	UPDATE
		service_inquiries
	SET
		inquiry_status = $2
	FROM
		 found_services
	WHERE
		found_services.id = service_inquiries.id
)
SELECT
	id,
	uuid
FROM
	found_services;
`
	rows, err := dao.DB.Queryx(
		query,
		string(models.InquiryStatusInquiring),
		string(models.InquiryStatusCanceled),
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	srvs := make([]*models.ServiceScannerData, 0)

	for rows.Next() {
		srv := models.ServiceScannerData{}

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

func (dao *ServiceDAO) GetServiceOptions() ([]*models.ServiceOption, error) {
	query := `
SELECT
	service_options.*
FROM
	service_options
WHERE
	service_options_type='default';
`

	rows, err := dao.DB.Queryx(query)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	srvOptions := make([]*models.ServiceOption, 0)

	for rows.Next() {
		srvOption := models.ServiceOption{}

		if err := rows.StructScan(&srvOption); err != nil {
			return nil, err
		}

		srvOptions = append(srvOptions, &srvOption)
	}

	return srvOptions, nil
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

const (
	DefaultBetweenServiceDuration    = 30
	DefaultAppointmentBufferDuration = 30
)

func (dao *ServiceDAO) GetOverlappedServices(p contracts.GetOverlappedServicesParams) ([]models.Service, error) {
	if p.BetweenServiceBufferDuration == 0 {
		p.BetweenServiceBufferDuration = DefaultBetweenServiceDuration
	}

	if p.AppointmentBufferDuration == 0 {
		p.AppointmentBufferDuration = DefaultAppointmentBufferDuration
	}

	// Retrieve all ongoing services that the user is currently engaging.
	query := `
SELECT
	*
FROM
	services
WHERE
	service_status NOT IN (
		'completed',
		'canceled',
		'expired',
		'payment_failed'
	) AND (
		customer_id = $1 OR
		service_provider_id = $1
	) AND uuid != $2 ;
`
	rows, err := dao.DB.Queryx(query, p.UserId, p.ExcludeServiceUuid)

	if err != nil {
		return nil, err
	}

	os := make([]models.Service, 0)

	for rows.Next() {
		var m models.Service

		if err := rows.StructScan(&m); err != nil {
			return nil, err
		}

		realEndTime := m.AppointmentTime.Time.Add(time.Duration(p.AppointmentBufferDuration) * time.Minute).Add(time.Duration(m.Duration.Int32))

		// Check inquiry appointment time is within the time inverval of the ongoing service.
		isAfOrEqAt := p.InquiryAppointmentTime.Equal(m.AppointmentTime.Time) || p.InquiryAppointmentTime.After(m.AppointmentTime.Time)
		isBfOrEqEt := p.InquiryAppointmentTime.Equal(realEndTime) || p.InquiryAppointmentTime.Before(
			realEndTime.Add(
				time.Duration(
					p.BetweenServiceBufferDuration,
				)*time.Minute,
			),
		)

		if isAfOrEqAt && isBfOrEqEt {
			os = append(os, m)
		}

	}

	return os, nil
}

func (dao *ServiceDAO) GetInquiryByServiceUuid(srvUuid string) (*models.ServiceInquiry, error) {
	query := `
SELECT service_inquiries.*
FROM service_inquiries
INNER JOIN services ON services.inquiry_id = service_inquiries.id
WHERE
	services.uuid = $1
ORDER BY service_inquiries.created_at DESC
LIMIT 1;
`

	var m models.ServiceInquiry

	if err := dao.DB.QueryRowx(query, srvUuid).StructScan(&m); err != nil {
		return &m, err
	}

	return &m, nil
}

func (dao *ServiceDAO) GetServiceProviderByServiceUUID(srvUUID string) (*models.User, error) {
	query := `
SELECT users.* FROM users
INNER JOIN services ON services.service_provider_id = users.id
WHERE services.uuid = $1;
	`

	var m models.User

	if err := dao.DB.QueryRowx(query, srvUUID).StructScan(&m); err != nil {
		return nil, err
	}

	return &m, nil
}

func (dao *ServiceDAO) CancelUnpaidServicesIfExceed30Minuties() ([]*models.CancelUnpaidServices, error) {
	query := `
WITH found_services AS (
SELECT
		services.id,
		services.uuid,
		customers.fcm_topic AS customer_fcm_topic,
		customers.username AS customer_name,
		service_providers.fcm_topic AS service_provider_fcm_topic,
		service_providers.username AS service_provider_name
	FROM
		services
	LEFT JOIN users AS customers ON customers.id = services.customer_id
	LEFT JOIN users AS service_providers ON service_providers.id = services.service_provider_id
	WHERE
		service_status = $1 AND
		CURRENT_TIMESTAMP > services.updated_at + (30 * interval '1 minute')
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

SELECT * FROM found_services;
	`

	ms := make([]*models.CancelUnpaidServices, 0)

	rows, err := dao.DB.Queryx(
		query,
		models.ServiceStatusUnpaid,
		models.ServiceStatusPaymentFailed,
	)

	if err != nil {
		return ms, err
	}

	for rows.Next() {
		var m models.CancelUnpaidServices

		if err := rows.StructScan(&m); err != nil {
			return ms, err
		}

		ms = append(ms, &m)
	}

	return ms, nil
}
