package inquiry

import (
	"fmt"
	"log"
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

func (dao *InquiryDAO) CheckHasActiveInquiryByID(id int64) (bool, error) {
	sql := `
SELECT EXISTS(
	SELECT 1 FROM users
	LEFT JOIN service_inquiries as si ON si.inquirer_id = users.id
	WHERE users.id = $1
	AND inquiry_status='inquiring'
) as exists;
`
	var exists bool

	err := dao.db.QueryRow(sql, id).Scan(&exists)

	return exists, err
}

type InquiryInfo struct {
	models.ServiceInquiry
	Inquirer models.User
}

// GetInquiries get list of inquiries with 7 records per page.
func (dao *InquiryDAO) GetInquiries(offset int, perpage int, statuses ...models.InquiryStatus) ([]*InquiryInfo, error) {
	statusQuery := "1=1"

	if len(statuses) > 0 {
		statusStrsArr := make([]string, len(statuses))

		for _, status := range statuses {
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

	log.Printf("DEBUG statusQuery %v", statusQuery)

	query := fmt.Sprintf(
		`
SELECT
	si.uuid,
	si.budget,
	si.service_type,
	si.price,
	si.duration,
	si.appointment_time,
	si.lng,
	si.lat,
	si.inquiry_status,
	users.uuid,
	users.username,
	users.avatar_url,
	users.nationality
FROM service_inquiries AS si
INNER JOIN users
	ON si.inquirer_id = users.id
WHERE (
	%s
)
AND (
	si.expired_at > now()
	OR  si.expired_at IS NULL
)
ORDER BY si.created_at DESC
LIMIT $1
OFFSET $2;
`,
		statusQuery,
	)

	inquiries := make([]*InquiryInfo, 0)
	rows, err := dao.db.Query(
		query,
		perpage,
		offset,
	)
	defer rows.Close()

	if err != nil {
		log.Printf("DEBUG inquiry error %v", err)

		return nil, err
	}

	for rows.Next() {
		iq := InquiryInfo{}
		inquirer := models.User{}

		err := rows.Scan(
			&iq.Uuid,
			&iq.Budget,
			&iq.ServiceType,
			&iq.Price,
			&iq.Duration,
			&iq.AppointmentTime,
			&iq.Lng,
			&iq.Lat,
			&iq.InquiryStatus,
			&inquirer.Uuid,
			&inquirer.Username,
			&inquirer.AvatarUrl,
			&inquirer.Nationality,
		)

		if err != nil {
			return nil, err
		}

		iq.Inquirer = inquirer
		inquiries = append(inquiries, &iq)
	}

	return inquiries, nil
}

func (dao *InquiryDAO) GetInquiryByUuid(iqUuid string, fields ...string) (*models.ServiceInquiry, error) {
	if len(fields) == 0 {
		fields = append(fields, "*")
	}

	fieldsStr := strings.TrimSuffix(strings.Join(fields, ","), ",")

	baseQuery := `
SELECT %s
FROM service_inquiries
WHERE uuid = $1
	`
	query := fmt.Sprintf(baseQuery, fieldsStr)

	var inquiry models.ServiceInquiry

	if err := dao.db.QueryRowx(query, iqUuid).StructScan(&inquiry); err != nil {
		return nil, err
	}

	return &inquiry, nil
}

func (dao *InquiryDAO) HasMoreInquiries(offset int, perPage int) (bool, error) {
	sql := `
SELECT count(full_count) as num_records FROM (
	SELECT COUNT(si.id) OVER() AS full_count
	FROM service_inquiries AS si
	WHERE si.inquiry_status = 'inquiring'
	LIMIT $1
	OFFSET $2
) AS records;
`
	var recordNum int

	if err := dao.db.QueryRow(sql, perPage, offset+perPage).Scan(&recordNum); err != nil {
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
