package user

import (
	"database/sql"

	"github.com/huangc28/go-darkpanda-backend/internal/models"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

type User struct {
	models.User
	Inquiries []*models.ServiceInquiry `json:"inquiries"`
}

type UserDAOer interface {
	GetUserInfoWithInquiryByUuid(ctx context.Context, uuid string) (User, error)
}

type UserDAO struct {
	db *sql.DB
}

func NewUserDAO(db *sql.DB) *UserDAO {
	return &UserDAO{
		db: db,
	}
}

// https://stackoverflow.com/questions/40093809/why-is-my-t-sql-left-join-not-working/40093841
// GetUserInfoWithInquiryByUuid
func (dao *UserDAO) GetUserInfoWithInquiryByUuid(ctx context.Context, uuid string, inquiryStatus models.InquiryStatus) (*User, error) {
	sql := `
		SELECT
			users.username,
			users.uuid,
			users.gender,
			si.budget,
			si.service_type,
			si.inquiry_status
		FROM users
		LEFT JOIN service_inquiries AS si
			ON users.id = si.inquirer_id
			AND si.inquiry_status = $2
		WHERE users.uuid = $1;
	`

	rows, err := dao.db.Query(sql, uuid, inquiryStatus)
	defer rows.Close()

	if err != nil {
		return nil, err
	}

	user := &User{}

	for rows.Next() {
		inquiry := &models.ServiceInquiry{}

		if err := rows.Scan(
			&user.Username,
			&user.Uuid,
			&user.Gender,
			&inquiry.Budget,
			&inquiry.ServiceType,
			&inquiry.InquiryStatus,
		); err != nil {
			return (*User)(nil), err
		}

		user.Inquiries = append(user.Inquiries, inquiry)
	}

	log.Printf("DEBUG 999 %v", user.Inquiries[0].Budget)

	return nil, nil
}

func (dao *UserDAO) checkGender(uuid string, gender models.Gender) (bool, error) {
	sql := `
		SELECT EXISTS (
			SELECT 1 FROM users
			WHERE uuid = $1
			AND gender = $2
		) AS exists;
`

	var exists bool

	if err := dao.db.QueryRow(sql, uuid, string(gender)).Scan(&exists); err != nil {
		return false, err
	}

	return exists, nil
}

func (dao *UserDAO) CheckIsMaleByUuid(uuid string) (bool, error) {
	return dao.checkGender(uuid, models.GenderMale)
}

func (dao *UserDAO) CheckIsFeMaleByUuid(uuid string) (bool, error) {
	return dao.checkGender(uuid, models.GenderFemale)
}
