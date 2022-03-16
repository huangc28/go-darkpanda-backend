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
			si.expect_service_type,
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
			&inquiry.ExpectServiceType,
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

func (dao *UserDAO) attachOngoingGirlsChannelUUID(girls []*models.RandomGirl) error {
	inquiryUUIDs := make([]string, 0)
	inquiryAndGirlMatch := make(map[string]*models.RandomGirl)

	// Get inquiry uuids for those girl that has ongoing girl or service with me.
	// Those inquiry uuids are used to retrieve channel uuids.
	for _, g := range girls {
		if g.HasInquiry || g.HasService {
			inquiryUUIDs = append(inquiryUUIDs, *g.InquiryUUID)
			inquiryAndGirlMatch[*g.InquiryUUID] = g
		}
	}

	if len(inquiryUUIDs) <= 0 {
		return nil
	}

	inquiryUUIDsStr := db.ComposeStringList(inquiryUUIDs...)

	query := fmt.Sprintf(
		`
SELECT
	DISTINCT ON (service_inquiries.uuid) uuid, chatrooms.created_at, chatrooms.channel_uuid
FROM
	chatrooms
INNER JOIN service_inquiries
	ON chatrooms.inquiry_id = service_inquiries.id
WHERE
	service_inquiries.uuid IN (%s)
ORDER BY service_inquiries.uuid, chatrooms.created_at DESC;

`, inquiryUUIDsStr)

	log.Printf("DEBUG*** %v", query)

	rows, err := dao.db.Queryx(query)

	if err != nil {
		return err
	}

	for rows.Next() {
		onGoingChatUUID := struct {
			InquiryUUID string    `json:"uuid"`
			CreatedAt   time.Time `json:"created_at"`
			ChannelUUID string    `json:"channel_uuid"`
		}{}

		if err := rows.StructScan(&onGoingChatUUID); err != nil {
			return err
		}

		if g, exists := inquiryAndGirlMatch[onGoingChatUUID.InquiryUUID]; exists {
			g.ChannelUUID = &onGoingChatUUID.ChannelUUID

		}

	}

	return nil
}

// GetGirls retrieve list of girl profile who wants their profile to be viewed publically.
// It also retrieve latest inquiry made between each girl with the male user. If no inquiry
// has ever existed,
func (dao *UserDAO) GetGirls(p contracts.GetGirlsParams) ([]*models.RandomGirl, error) {
	query := `
SELECT
	username,
	users.uuid,
	avatar_url,
	age,
	height,
	weight,
	breast_size,
	description,
	CASE WHEN si.id IS NOT NULL
    THEN true
    ELSE false
    END has_inquiry,

    CASE WHEN services.uuid IS NOT NULL
    THEN true
    ELSE false
    END has_service,

    si.uuid AS inquiry_uuid,
	si.inquiry_status,
	si.expect_service_type,

	services.uuid AS service_uuid,
	services.service_status
FROM users

-- Retrieve related inquiries if any
LEFT JOIN service_inquiries AS si
	ON si.inquirer_id = $1
	AND si.picker_id = users.id
	AND si.inquiry_status NOT IN (
		'canceled',
		'booked'
	)
	AND si.created_at=(
	 	SELECT max(created_at)
		FROM service_inquiries
        	WHERE inquirer_id=5
        	AND picker_id = users.id
	)

-- Retrieve related services if any
LEFT JOIN services
	ON services.inquiry_id = si.id
	AND services.service_status NOT IN (
		'canceled',
		'completed',
		'expired'
	)
WHERE
	gender='female'
ORDER BY users.id % 4, users.id
LIMIT $2
OFFSET $3;
	`

	gs := make([]*models.RandomGirl, 0)

	rows, err := dao.db.Queryx(query, p.InquirerID, p.Limit, p.Offset)

	if err != nil {
		return gs, err
	}

	girlIDs := make([]int64, 0)

	for rows.Next() {
		var g models.RandomGirl

		if err := rows.StructScan(&g); err != nil {
			return gs, err
		}

		gs = append(gs, &g)
		girlIDs = append(girlIDs, g.ID)
	}

	// If no girls are loaded, we don't have to fetch rating for the girls.
	if len(gs) == 0 {
		return gs, nil
	}

	// Compose a query to retrieve girls rating.
	girlIDsStr := strings.Trim(strings.Join(strings.Fields(fmt.Sprint(girlIDs)), ","), "[]")

	ratingQuery := fmt.Sprintf(`
SELECT
	ratee_id,
	ROUND(AVG(rating), 2) AS score,
	COUNT(1) AS number_of_services
FROM
	service_ratings
INNER JOIN users ON ratee_id = users.id
WHERE
	ratee_id IN (%s)
GROUP BY ratee_id;
	`, girlIDsStr)

	ratingRows, err := dao.db.Queryx(ratingQuery)

	if err != nil {
		return nil, err
	}

	// Map to record ratee ID with rating info, we will be
	// using this data structure to find appropriate female
	// rating and assign it to "gs (slice of girls)".
	// ex:
	//  [1] => UserRating for girl ID 1
	//  [2] => UserRating for girl ID 2
	//  [3] => UserRating for girl ID 3
	//  ...
	ratingGirlIDMap := make(map[int64]*models.UserRating)

	for ratingRows.Next() {
		ur := models.UserRating{}

		if err := ratingRows.StructScan(&ur); err != nil {
			return nil, err
		}

		ratingGirlIDMap[ur.RateeID] = &ur
	}

	for _, g := range gs {
		if r, ok := ratingGirlIDMap[g.ID]; ok {
			g.Rating = *r
		}
	}

	// Compose a query to retrieve chatroom uuid for those inquiry and service that are still ongoing.
	// Iterate through all girls, collect those girls that has ongoing inquiry or service with me. Retrieve channel uuid of those.

	if err := dao.attachOngoingGirlsChannelUUID(gs); err != nil {
		return nil, err
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

func (dao *UserDAO) GetUserServiceOption(userID int) ([]models.UserServiceOptionData, error) {
	query := `
	select
		so."name" ,
		so.price ,
		so."service_options_type" ,
		so.description,
		so.duration,
		so.id service_option_id
	from user_service_options uso
	inner join service_options so
	on uso.service_option_id = so.id
	where uso.users_id=$1 AND uso.deleted_at IS null
	order by so.created_at`

	services := make([]models.UserServiceOptionData, 0)

	rows, err := dao.db.Queryx(query, userID)

	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var service models.UserServiceOptionData

		if err := rows.Scan(
			&service.ServiceName,
			&service.Price, &service.OptionType,
			&service.Description,
			&service.Duration,
			&service.ServiceOptionID,
		); err != nil {
			return nil, err
		}

		services = append(services, service)
	}

	return services, nil
}

func (dao *UserDAO) CreateServiceOption(params contracts.CreateServiceOptionsParams) (*models.ServiceOption, error) {
	query := `
	INSERT INTO service_options (
		name,
		description,
		price,
		service_options_type,
		duration
	) VALUES ($1, $2, $3, $4, $5)
	RETURNING *;
	`
	var m models.ServiceOption

	if err := dao.db.QueryRowx(
		query,
		params.Name,
		params.Description,
		params.Price,
		params.ServiceOptionsType,
		params.Duration,
	).StructScan(&m); err != nil {
		return nil, err
	}

	return &m, nil
}

func (dao *UserDAO) CreateUserServiceOption(params contracts.CreateServiceOptionParams) (*models.UserServiceOption, error) {
	query := `
	INSERT INTO user_service_options (
		users_id,
		service_option_id
	) VALUES ($1, $2)
	RETURNING *;
	`
	var m models.UserServiceOption

	if err := dao.db.QueryRowx(
		query,
		params.UserID,
		params.ServiceOptionID,
	).StructScan(&m); err != nil {
		return nil, err
	}

	return &m, nil
}

func (dao *UserDAO) DeleteUserServiceOption(userID int, serviceOptionID int) error {
	query := `
		UPDATE user_service_options
		SET deleted_at = now()
		WHERE users_id = $1 AND service_option_id = $2
	`

	_, err := dao.db.Exec(
		query,
		userID,
		serviceOptionID,
	)

	if err != nil {
		return err
	}

	return err
}

func (dao *UserDAO) CheckServiceOptionExists(userID int, username string) (bool, error) {
	query := `SELECT EXISTS(
				SELECT 1
				FROM service_options so
				INNER JOIN user_service_options uso on so.id = uso.service_option_id
				WHERE so."name" = $1 and uso.users_id = $2 and uso.deleted_at IS NULL
			) AS "exists"`
	var exists bool

	if err := dao.db.QueryRow(query, username, userID).Scan(&exists); err != nil {
		return false, err
	}

	return exists, nil
}
