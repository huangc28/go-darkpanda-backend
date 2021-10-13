package inquiry

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	cintrnal "github.com/golobby/container/pkg/container"
	"github.com/huangc28/go-darkpanda-backend/db"
	"github.com/huangc28/go-darkpanda-backend/internal/app/contracts"
	"github.com/huangc28/go-darkpanda-backend/internal/app/models"
)

type InquiryDAO struct {
	db db.Conn
}

func NewInquiryDAO(db db.Conn) *InquiryDAO {
	return &InquiryDAO{
		db: db,
	}
}

func InquiryDaoServiceProvider(c cintrnal.Container) func() error {
	return func() error {
		c.Transient(func() contracts.InquiryDAOer {
			return NewInquiryDAO(db.GetDB())
		})

		return nil
	}
}

func (dao *InquiryDAO) WithTx(db db.Conn) contracts.InquiryDAOer {
	dao.db = db

	return dao
}

func (dao *InquiryDAO) CheckHasActiveRandomInquiryByID(id int64) (bool, error) {
	sql := `
SELECT EXISTS(
	SELECT 1 FROM users
	LEFT JOIN service_inquiries as si ON si.inquirer_id = users.id
	WHERE users.id = $1
	AND inquiry_status='inquiring'
	AND inquiry_type='random'
) as exists;
`
	var exists bool

	err := dao.db.QueryRow(sql, id).Scan(&exists)

	return exists, err
}

func (dao *InquiryDAO) GetInquiries(p contracts.GetInquiriesParams) ([]*models.InquiryInfo, error) {
	if p.InquiryType == "" {
		p.InquiryType = models.InquiryTypeRandom
	}

	statusQuery := "1=1"

	if len(p.Statuses) > 0 {
		statusStrsArr := make([]string, len(p.Statuses))

		for _, status := range p.Statuses {
			statusStrsArr = append(
				statusStrsArr,
				fmt.Sprintf("si.inquiry_status = '%s' OR", string(status)),
			)
		}

		statusQuery = strings.TrimSuffix(
			strings.Join(
				statusStrsArr,
				" ",
			),
			"OR",
		)
	}

	query := fmt.Sprintf(
		`
WITH blocked_users AS (
	SELECT
		blocked_user_id
	FROM
		block_list
	WHERE
		user_id = $1 AND
		deleted_at IS NOT NULL
), ongoing_service AS (
	SELECT
		customer_id AS ongoing_customer_id
	FROM
		services
	WHERE
		service_provider_id = $1 AND
		service_status NOT IN (
			'completed',
			'expired',
			'canceled',
			'payment_failed'
		)
), inquiry_list AS (
	SELECT
		si.uuid AS inquiry_uuid,
		si.budget,
		si.service_type,
		si.duration,
		si.appointment_time,
		si.lng,
		si.lat,
		si.inquiry_status,

		users.uuid,
		users.username,
		users.avatar_url,
		users.nationality,

		services.uuid as service_uuid
	FROM service_inquiries AS si
	INNER JOIN users
		ON si.inquirer_id = users.id
	LEFT JOIN services ON services.inquiry_id = si.id 
	WHERE (%s)
	AND inquiry_type=$2 
	AND si.inquirer_id NOT IN (
		SELECT
			blocked_user_id
		FROM
			blocked_users
	)
	AND si.inquirer_id NOT IN (
		SELECT
			ongoing_customer_id
		FROM
			ongoing_service
	)
	ORDER BY si.created_at DESC
	LIMIT $3
	OFFSET $4
)

SELECT DISTINCT ON(inquiry_list.inquiry_uuid) * FROM inquiry_list;
`,
		statusQuery,
	)

	inquiries := make([]*models.InquiryInfo, 0)
	rows, err := dao.db.Queryx(
		query,
		p.UserID,
		models.InquiryTypeRandom,
		p.PerPage,
		p.Offset,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		iq := models.InquiryInfo{}
		inquirer := models.User{}
		serviceUuid := sql.NullString{}

		err := rows.Scan(
			&iq.Uuid,
			&iq.Budget,
			&iq.ServiceType,
			&iq.Duration,
			&iq.AppointmentTime,
			&iq.Lng,
			&iq.Lat,
			&iq.InquiryStatus,
			&inquirer.Uuid,
			&inquirer.Username,
			&inquirer.AvatarUrl,
			&inquirer.Nationality,
			&serviceUuid,
		)

		if err != nil {
			return nil, err
		}

		iq.Inquirer = inquirer
		iq.ServiceUuid = serviceUuid
		inquiries = append(inquiries, &iq)
	}

	return inquiries, nil
}

func (dao *InquiryDAO) GetInquiryByUuid(iqUuid string, fields ...string) (*contracts.InquiryResult, error) {
	if len(fields) == 0 {
		fields = append(fields, "service_inquiries.*")
	} else {
		for key, field := range fields {
			fields[key] = fmt.Sprintf("service_inquiries.%s", field)
		}
	}

	fieldsStr := strings.TrimSuffix(
		strings.Join(
			fields,
			",",
		),
		",",
	)

	baseQuery := `
SELECT
	%s,
	users.username,
	users.uuid as user_uuid,
	users.avatar_url
FROM service_inquiries
INNER JOIN users
	ON service_inquiries.inquirer_id = users.id
WHERE service_inquiries.uuid = $1
	`
	query := fmt.Sprintf(baseQuery, fieldsStr)

	inquiry := contracts.InquiryResult{}

	if err := dao.db.QueryRowx(query, iqUuid).StructScan(&inquiry); err != nil {
		return nil, err
	}

	return &inquiry, nil
}

func (dao *InquiryDAO) HasMoreInquiries(offset int, perPage int) (bool, error) {
	sql := `
SELECT
	count(full_count) AS num_records
FROM (
	SELECT
		COUNT(si.id) OVER() AS full_count
	FROM
		service_inquiries AS si
	WHERE
		si.inquiry_status = 'inquiring'
	LIMIT $1
	OFFSET $2
) AS records;
`
	var recordNum int

	if err := dao.
		db.QueryRow(sql, perPage, offset+perPage).
		Scan(&recordNum); err != nil {
		return false, err
	}

	return recordNum > 0, nil
}

func (dao *InquiryDAO) IsInquiryExpired(inquiryID int64) (bool, error) {
	sql := `
SELECT
	expired_at
FROM
	service_inquiries
WHERE
	id = $1
AND
	deleted_at IS NULL;
	`
	var expiredAt time.Time

	if err := dao.db.QueryRow(sql, inquiryID).Scan(&expiredAt); err != nil {
		return false, err
	}

	return time.Now().After(expiredAt), nil
}

// AskingInquiry Alters the inquiry status to `asking`. Meaning that the girl wants to chat with
// the inquirer and is waiting for the male user to reply.
func (dao *InquiryDAO) AskingInquiry(pickerID, inquiryID int64) (*models.ServiceInquiry, error) {
	sql := `
UPDATE service_inquiries
SET
	inquiry_status = $1,
	picker_id = $2
WHERE
	id = $3
RETURNING *;
	`

	var pickedInquiry models.ServiceInquiry

	if err := dao.
		db.
		QueryRowx(
			sql,
			models.InquiryStatusAsking,
			pickerID,
			inquiryID,
		).StructScan(&pickedInquiry); err != nil {
		return nil, err
	}

	return &pickedInquiry, nil
}

func (dao *InquiryDAO) PatchInquiryStatusByUUID(params contracts.PatchInquiryStatusByUUIDParams) error {
	sql := `
UPDATE service_inquiries
SET inquiry_status = $1
WHERE uuid = $2
`
	_, err := dao.db.Exec(sql, params.InquiryStatus, params.UUID)

	if err != nil {
		return err
	}

	return err
}

// GetInquirerByInquiryUUID gets the inquirer information given inquiry UUID. If no fields is given,
// it retrieves all field in regards to that inquirer.
func (dao *InquiryDAO) GetInquirerByInquiryUUID(uuid string, fields ...string) (*models.User, error) {
	if len(fields) == 0 {
		fields = append(fields, "users.*")
	}

	fieldsStr := strings.TrimSuffix(strings.Join(fields, ","), ",")

	baseQuery := `
SELECT %s
FROM users
INNER JOIN service_inquiries
	ON service_inquiries.inquirer_id = users.id
WHERE service_inquiries.uuid = $1;
	`

	query := fmt.Sprintf(baseQuery, fieldsStr)

	inquirer := models.User{}

	if err := dao.db.QueryRowx(query, uuid).StructScan(&inquirer); err != nil {
		return nil, err
	}

	return &inquirer, nil
}

func (dao *InquiryDAO) PatchInquiryByInquiryUUID(params models.PatchInquiryParams) (*models.ServiceInquiry, error) {
	query := `
UPDATE service_inquiries SET
	appointment_time = COALESCE($1, appointment_time),
	service_type = COALESCE($2, service_type),
	inquiry_status = COALESCE($3, inquiry_status),
	budget = COALESCE($4, budget),
	duration = COALESCE($5, duration),
	address = COALESCE($6, address)
WHERE
	uuid = $7
RETURNING
	id,
	uuid,
	appointment_time,
	service_type,
	budget,
	duration,
	address;
`
	inquiry := models.ServiceInquiry{}

	err := dao.db.QueryRowx(
		query,
		params.AppointmentTime,
		params.ServiceType,
		params.InquiryStatus,
		params.Budget,
		params.Duration,
		params.Address,
		params.Uuid,
	).StructScan(&inquiry)

	if err != nil {
		return nil, err
	}

	return &inquiry, nil
}

func (dao *InquiryDAO) GetActiveInquiry(inquirerId int) (*models.ActiveInquiry, error) {
	query := `
SELECT
	service_inquiries.*,
	users.uuid AS picker_uuid
FROM service_inquiries
LEFT JOIN users ON service_inquiries.picker_id = users.id
WHERE
	inquirer_id = $1
AND
	(
		inquiry_status='inquiring' OR
		inquiry_status='asking'
	)
AND
	inquiry_type=$2
ORDER BY created_at DESC
LIMIT 1;
	`
	var m models.ActiveInquiry
	if err := dao.db.QueryRowx(query, inquirerId, models.InquiryTypeRandom).StructScan(&m); err != nil {
		return nil, err
	}

	return &m, nil
}

func (dao *InquiryDAO) GetInquiryByChannelUuid(channelUuid string) (*models.ServiceInquiry, error) {
	query := `
SELECT service_inquiries.* FROM service_inquiries
INNER JOIN
	chatrooms ON chatrooms.inquiry_id = service_inquiries.id
	AND chatrooms.channel_uuid = $1;
`

	var m models.ServiceInquiry

	if err := dao.db.QueryRowx(query, channelUuid).StructScan(&m); err != nil {
		return nil, err

	}

	return &m, nil
}

func (dao *InquiryDAO) GetInquiryRequests(p contracts.GetInquiryRequestsParams) ([]models.InquiryRequest, error) {
	query := `
SELECT	
	si.uuid AS inquiry_uuid,
	si.created_at,
	si.inquiry_status,
	users.username,
	users.avatar_url,
	users.uuid AS inquirer_uuid
FROM
	service_inquiries AS si
INNER JOIN users ON users.id = si.inquirer_id
WHERE 
	inquiry_type=$1
AND 
	inquiry_status=$2
AND
	picker_id=$3
ORDER BY si.created_at
OFFSET $4
LIMIT $5;
	`

	rows, err := dao.db.Queryx(
		query,
		models.InquiryTypeDirect,
		models.InquiryStatusAsking,
		p.UserID,
		p.Offset,
		p.PerPage,
	)

	if err != nil {
		return nil, err
	}

	irs := make([]models.InquiryRequest, 0)

	for rows.Next() {
		ir := models.InquiryRequest{}

		if err := rows.StructScan(&ir); err != nil {
			return irs, nil
		}

		irs = append(irs, ir)
	}

	return irs, nil
}
