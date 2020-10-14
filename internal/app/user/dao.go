package user

import (
	"github.com/huangc28/go-darkpanda-backend/internal/models"
	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

type User struct {
	models.User
	Inquiries []*models.ServiceInquiry `json:"inquiries"`
}

type PaymentDAOer interface {
	GetPaymentsByUuid(uuid string) ([]models.PaymentInfo, error)
}

type ServiceDAOer interface {
	GetUserHistoricalServicesByUuid(uuid string, perPage int, offset int) ([]models.Service, error)
}

type UserDAOer interface {
	GetUserInfoWithInquiryByUuid(ctx context.Context, uuid string, inquiryStatus models.InquiryStatus) (*User, error)
	UpdateUserInfoByUuid(ctx context.Context, p UpdateUserInfoParams) (*models.User, error)
	GetUserByUuid(ctx context.Context, uuid string) (*models.User, error)
	CheckIsMaleByUuid(uuid string) (bool, error)
	CheckIsFemaleByUuid(uuid string) (bool, error)
	GetUserImagesByUuid(uuid string, offset int, perPage int) ([]models.Image, error)
}

type UserDAO struct {
	db *sqlx.DB
}

func NewUserDAO(db *sqlx.DB) UserDAOer {
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

	return nil, nil
}

type UpdateUserInfoParams struct {
	AvatarURL   *string
	Nationality *string
	Region      *string
	Age         *int
	Height      *float64
	Weight      *float64
	Description *string
	BreastSize  *string
	Uuid        string
}

// https://stackoverflow.com/questions/13305878/dont-update-column-if-update-value-is-null
func (dao *UserDAO) UpdateUserInfoByUuid(ctx context.Context, p UpdateUserInfoParams) (*models.User, error) {
	sql := `
UPDATE users SET
	avatar_url = COALESCE($1, avatar_url),
	nationality = COALESCE($2, nationality),
	region = COALESCE($3, region),
	age = COALESCE($4, age),
	height = COALESCE($5, height),
	weight = COALESCE($6, weight),
	description = COALESCE($7, description),
	breast_size = COALESCE($8, breast_size)
WHERE uuid = $9
RETURNING
	id,
	username,
	phone_verified,
	gender,
	premium_type,
	premium_expiry_date,
	uuid,
	avatar_url,
	nationality,
	region,
	age,
	height,
	weight,
	habbits,
	description,
	breast_size;
`
	u := &models.User{}

	if err := dao.db.QueryRow(
		sql,
		p.AvatarURL,
		p.Nationality,
		p.Region,
		p.Age,
		p.Height,
		p.Weight,
		p.Description,
		p.BreastSize,
		p.Uuid,
	).Scan(
		&u.ID,
		&u.Username,
		&u.PhoneVerified,
		&u.Gender,
		&u.PremiumType,
		&u.PremiumExpiryDate,
		&u.Uuid,
		&u.AvatarUrl,
		&u.Nationality,
		&u.Region,
		&u.Age,
		&u.Height,
		&u.Weight,
		&u.Habbits,
		&u.Description,
		&u.BreastSize,
	); err != nil {
		log.Errorf("Failed to update user info %s", err.Error())

		return nil, err
	}

	return u, nil
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

func (dao *UserDAO) CheckIsFemaleByUuid(uuid string) (bool, error) {
	return dao.checkGender(uuid, models.GenderFemale)
}

func (dao *UserDAO) GetUserByUuid(ctx context.Context, uuid string) (*models.User, error) {
	sql := `
SELECT id, username, phone_verified, auth_sms_code, gender, premium_type, premium_expiry_date, created_at, updated_at, deleted_at, uuid, phone_verify_code, avatar_url, nationality, region, age, height, weight, habbits, description, breast_size, mobile FROM users
WHERE uuid = $1 LIMIT 1
	`
	i := models.User{}
	if err := dao.db.QueryRow(sql, uuid).Scan(
		&i.ID,
		&i.Username,
		&i.PhoneVerified,
		&i.AuthSmsCode,
		&i.Gender,
		&i.PremiumType,
		&i.PremiumExpiryDate,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.DeletedAt,
		&i.Uuid,
		&i.PhoneVerifyCode,
		&i.AvatarUrl,
		&i.Nationality,
		&i.Region,
		&i.Age,
		&i.Height,
		&i.Weight,
		&i.Habbits,
		&i.Description,
		&i.BreastSize,
		&i.Mobile,
	); err != nil {
		return nil, err
	}

	return &i, nil
}

func (dao *UserDAO) GetUserImagesByUuid(uuid string, offset int, perPage int) ([]models.Image, error) {

	sql := `
SELECT images.url
FROM images
LEFT JOIN users
	ON images.user_id = users.id
WHERE users.uuid = $1
ORDER BY images.created_at DESC
LIMIT $2
OFFSET $3;

`
	rows, err := dao.db.Query(
		sql,
		uuid,
		perPage,
		offset,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()
	images := make([]models.Image, 0)
	for rows.Next() {
		var image models.Image

		if err := rows.Scan(&image.Url); err != nil {
			return nil, err
		}

		images = append(images, image)
	}

	return images, nil
}
