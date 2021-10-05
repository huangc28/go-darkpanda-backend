package user

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	cintrnal "github.com/golobby/container/pkg/container"
	"github.com/huangc28/go-darkpanda-backend/db"
	"github.com/huangc28/go-darkpanda-backend/internal/app/contracts"
	"github.com/huangc28/go-darkpanda-backend/internal/app/models"
	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

type ServiceDAOer interface {
	GetUserHistoricalServicesByUuid(uuid string, perPage int, offset int) ([]models.Service, error)
}

type User struct {
	models.User
	Inquiries []*models.ServiceInquiry `json:"inquiries"`
}

type UserDAO struct {
	db db.Conn
}

func NewUserDAO(db *sqlx.DB) contracts.UserDAOer {
	return &UserDAO{
		db: db,
	}
}

func UserDaoServiceProvider(c cintrnal.Container) func() error {
	return func() error {
		c.Transient(func() contracts.UserDAOer {
			return NewUserDAO(db.GetDB())
		})

		return nil
	}
}

func (dao *UserDAO) WithTx(tx *sqlx.Tx) contracts.UserDAOer {
	dao.db = tx

	return dao
}

func (dao *UserDAO) WithDB(db *sqlx.DB) contracts.UserDAOer {
	dao.db = db

	return dao
}

// https://stackoverflow.com/questions/40093809/why-is-my-t-sql-left-join-not-working/40093841
// GetUserInfoWithInquiryByUuid
func (dao *UserDAO) GetUserInfoWithInquiryByUuid(ctx context.Context, uuid string, inquiryStatus models.InquiryStatus) (*models.UserWithInquiries, error) {
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

	if err != nil {
		return nil, err
	}

	defer rows.Close()

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
			return nil, err
		}

		user.Inquiries = append(user.Inquiries, inquiry)
	}

	return nil, nil
}

// https://stackoverflow.com/questions/13305878/dont-update-column-if-update-value-is-null
func (dao *UserDAO) UpdateUserInfoByUuid(p contracts.UpdateUserInfoParams) (*models.User, error) {
	sql := `
UPDATE users SET
	avatar_url = COALESCE($1, avatar_url),
	nationality = COALESCE($2, nationality),
	region = COALESCE($3, region),
	age = COALESCE($4, age),
	height = COALESCE($5, height),
	weight = COALESCE($6, weight),
	description = COALESCE($7, description),
	breast_size = COALESCE($8, breast_size),
	phone_verified = COALESCE($9, phone_verified),
	mobile = COALESCE($10, mobile)
WHERE uuid = $11
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
	breast_size,
	mobile;
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
		p.PhoneVerified,
		p.Mobile,
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
		&u.Mobile,
	); err != nil {
		log.Errorf("Failed to update user info %s", err.Error())

		return nil, err
	}

	return u, nil
}

func (dao *UserDAO) checkGender(uuid string, gender models.Gender) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT 1 FROM users
			WHERE uuid = $1
			AND gender = $2
		) AS exists;
`

	var exists bool

	if err := dao.db.QueryRow(query, uuid, string(gender)).Scan(&exists); err != nil {
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

func (dao *UserDAO) GetUserByUuid(uuid string, fields ...string) (*models.User, error) {
	if len(fields) == 0 {
		fields = append(fields, "*")
	}

	fieldsStr := strings.TrimSuffix(strings.Join(fields, ","), ",")

	baseQuery := `
SELECT %s
FROM users
WHERE uuid = $1
`
	query := fmt.Sprintf(baseQuery, fieldsStr)

	var user models.User

	if err := dao.db.QueryRowx(query, uuid).StructScan(&user); err != nil {
		return nil, err
	}

	return &user, nil
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

func (dao *UserDAO) GetUserByID(ID int64, fields ...string) (*models.User, error) {
	baseQuery := `
SELECT %s
FROM users
WHERE id = $1
`
	query := fmt.Sprintf(baseQuery, db.ComposeFieldsSQLString(fields...))

	var user models.User

	if err := dao.db.QueryRowx(query, ID).StructScan(&user); err != nil {
		return nil, err
	}

	return &user, nil
}

func (dao *UserDAO) GetUserByUsername(username string, fields ...string) (*models.User, error) {
	baseQuery := `
SELECT %s
FROM users
WHERE username = $1
`
	query := fmt.Sprintf(baseQuery, db.ComposeFieldsSQLString(fields...))

	var user models.User

	if err := dao.db.QueryRowx(query, username).StructScan(&user); err != nil {
		return nil, err
	}

	return &user, nil
}

func (dao *UserDAO) DeleteUserImages(url string) error {
	query := `
		DELETE FROM images
		WHERE url=$1
	`

	_, err := dao.db.Exec(
		query,
		url,
	)

	if err != nil {
		return err
	}

	return err
}

// GetRating calculates the average rating that the user has participated in.
func (dao *UserDAO) GetRating(userID int) (*models.UserRating, error) {
	query := `
SELECT 	
	AVG(rating)::numeric(10,2) AS score,
	count(DISTINCT service_id) AS number_of_services 
FROM service_ratings
WHERE ratee_id = $1;
	`

	var rating models.UserRating

	if err := dao.db.QueryRowx(query, userID).StructScan(&rating); err != nil {
		return &rating, err
	}

	return &rating, nil
}

// GetGirls retrieve list of girl profile who wants their profile to be viewed publically.

func (dao *UserDAO) GetGirls(p contracts.GetGirlsParams) ([]*models.RandomGirl, error) {
	ranSeed := 0.1

	query := fmt.Sprintf(` 
SELECT 
	setseed(%f),
	uuid,
	username,
	avatar_url,
	age,
	height,
	weight,
	breast_size
FROM users 
WHERE 
	gender = 'female'
ORDER BY random()
LIMIT %d 
OFFSET %d;
	`, ranSeed, p.Limit, p.Offset)

	gs := make([]*models.RandomGirl, 0)

	rows, err := dao.db.Queryx(query)

	if err != nil {
		return gs, err
	}

	for rows.Next() {
		var g models.RandomGirl

		if err := rows.StructScan(&g); err != nil {
			return gs, err
		}

		gs = append(gs, &g)
	}

	return gs, nil
}

const (
	ChangeMobileVerifyCodeHashName = "change_mobile_verify_code:%s"
	ChangeMobileVerifyCodeFieldKey = "verify_code"
	ChangeMobileNumberFieldKey     = "mobile"
)

type CreateChangeMobileVerifyCodeParams struct {
	RedisCli   *redis.Client
	VerifyCode string
	UserUuid   string
	Mobile     string
}

func CreateChangeMobileVerifyCode(ctx context.Context, p CreateChangeMobileVerifyCodeParams) error {
	if p.RedisCli == nil {
		return errors.New("redis client is not provided")
	}

	pipe := p.RedisCli.TxPipeline()
	defer pipe.Close()

	pipe.HSet(
		ctx,
		fmt.Sprintf(ChangeMobileVerifyCodeHashName, p.UserUuid),
		ChangeMobileVerifyCodeFieldKey,
		p.VerifyCode,
		ChangeMobileNumberFieldKey,
		p.Mobile,
	)

	pipe.Expire(
		ctx,
		fmt.Sprintf(ChangeMobileVerifyCodeHashName, p.UserUuid),
		5*time.Minute,
	)

	if _, err := pipe.Exec(ctx); err != nil {
		return err
	}

	return nil
}

type ChangeMobileVerifyCodeModel struct {
	Mobile     string
	VerifyCode string
}

func parseChangeMobileVerifyCodeResult(res map[string]string) *ChangeMobileVerifyCodeModel {
	m := &ChangeMobileVerifyCodeModel{}

	if v, ok := res[ChangeMobileVerifyCodeFieldKey]; ok {
		m.VerifyCode = v
	}

	if v, ok := res[ChangeMobileNumberFieldKey]; ok {
		m.Mobile = v
	}

	return m
}

type GetChangeMobileVerifyCodeParams struct {
	RedisCli *redis.Client
	UserUuid string
}

func GetChangeMobileVerifyCode(ctx context.Context, p GetChangeMobileVerifyCodeParams) (*ChangeMobileVerifyCodeModel, error) {
	val, err := p.RedisCli.HGetAll(ctx, fmt.Sprintf(ChangeMobileVerifyCodeHashName, p.UserUuid)).Result()

	if err != nil {
		return nil, err
	}

	m := parseChangeMobileVerifyCodeResult(val)

	return m, nil
}
