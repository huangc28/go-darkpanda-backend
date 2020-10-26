package inquiry

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/huangc28/go-darkpanda-backend/db"
	"github.com/huangc28/go-darkpanda-backend/internal/app/models"
)

type InquiryDAOer interface {
	CheckHasActiveInquiryByID(id int64) (bool, error)
	GetInquiries(status models.InquiryStatus, offset int, perpage int) ([]*InquiryInfo, error)
	HasMoreInquiries(offset int, perPage int) (bool, error)
	GetInquiryByUuid(iqUuid string, fields ...string) (*models.ServiceInquiry, error)
	IsInquiryExpired(inquiryID int64) (bool, error)
}

type InquiryDAO struct {
	db db.Conn
}

func NewInquiryDAO(db db.Conn) InquiryDAOer {
	return &InquiryDAO{
		db: db,
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
func (dao *InquiryDAO) GetInquiries(status models.InquiryStatus, offset int, perpage int) ([]*InquiryInfo, error) {
	log.Printf("DEBUG  GetInquiries %v %v", offset, perpage)
	sql := `
SELECT
	si.uuid,
	si.budget,
	si.service_type,
	si.price,
	si.duration,
	si.appointment_time,
	si.lng,
	si.lat,
	users.uuid,
	users.username,
	users.avatar_url,
	users.nationality
FROM service_inquiries AS si
INNER JOIN users
	ON si.inquirer_id = users.id
WHERE si.inquiry_status = $1
AND (
	si.expired_at > now() 
	OR  si.expired_at IS NULL
)
ORDER BY si.created_at DESC
LIMIT $2
OFFSET $3;
`
	inquiries := make([]*InquiryInfo, 0)
	rows, err := dao.db.Query(sql, status, perpage, offset)
	defer rows.Close()

	if err != nil {
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
