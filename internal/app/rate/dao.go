package rate

import (
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/golobby/container/pkg/container"
	"github.com/huangc28/go-darkpanda-backend/db"
	"github.com/huangc28/go-darkpanda-backend/internal/app/contracts"
	"github.com/huangc28/go-darkpanda-backend/internal/app/models"
)

type RateDAO struct {
	db db.Conn
}

func NewRateDAO(db db.Conn) *RateDAO {
	return &RateDAO{
		db: db,
	}
}

func RateDAOServiceProvider(c container.Container) func() error {
	return func() error {
		c.Transient(func() contracts.RateDAOer {
			return NewRateDAO(db.GetDB())
		})

		return nil
	}
}

func (dao *RateDAO) WithTx(tx db.Conn) contracts.RateDAOer {
	dao.db = tx

	return dao
}

type GetServicePartnerInfoParams struct {
	Gender      models.Gender
	MyId        int
	ServiceUuid string
}

// GetServicePartnerInfo Retrieve the following info for user rating.
//   -  ratee username
//   -  ratee avatar_url
//   -  ratee uuid
//   -  rating
//   -  service uuid
//   -  comments
func (dao *RateDAO) GetServicePartnerInfo(p GetServicePartnerInfoParams) (*models.User, error) {
	partnerCriteria := "service_provider_id"

	if p.Gender == models.GenderFemale {
		partnerCriteria = "customer_id"
	}

	myCrteria := "customer_id"

	if p.Gender == models.GenderFemale {
		myCrteria = "service_provider_id"
	}

	query := fmt.Sprintf(
		`
SELECT
	users.username,
	users.uuid,
	users.avatar_url
FROM users
INNER JOIN services ON services.%s = users.id
	AND services.%s = $1
	AND services.uuid = $2;`,
		partnerCriteria,
		myCrteria,
	)

	var m models.User

	if err := dao.db.QueryRowx(
		query,
		p.MyId,
		p.ServiceUuid,
	).StructScan(&m); err != nil {
		return nil, err
	}

	return &m, nil
}

type GetServiceRatingParams struct {
	ServiceUuid string
	RaterId     int
}

func (dao *RateDAO) GetServiceRating(p GetServiceRatingParams) (*models.ServiceRating, error) {
	query := `
SELECT
	service_ratings.rating,
	service_ratings.comments,
	service_ratings.created_at
FROM
	service_ratings
INNER JOIN services ON
	services.id = service_ratings.service_id
WHERE
	rater_id = $1 AND
	services.uuid = $2;
`

	var m models.ServiceRating

	if err := dao.db.QueryRowx(
		query,
		p.RaterId,
		p.ServiceUuid,
	).StructScan(&m); err != nil {
		return nil, err

	}

	return &m, nil
}

func (dao *RateDAO) isServiceParticipant(pId int, srvUuid string) (bool, error) {
	query := `
SELECT EXISTS (
	SELECT
		1
	FROM
		services
	WHERE
		uuid = $1 AND (
			customer_id = $2 OR
			service_provider_id = $2
		)
);
`
	var exists bool

	if err := dao.db.QueryRowx(
		query,
		srvUuid,
		pId,
	).Scan(&exists); err != nil {
		return false, err

	}

	return exists, nil
}

func (dao *RateDAO) hasRated(raterId int, serviceUuid string) (bool, error) {
	query := `
SELECT EXISTS (
	SELECT
		1
	FROM
		service_ratings
	INNER JOIN services ON
		services.id = service_ratings.service_id AND
		services.uuid = $2
	WHERE
		rater_id = $1
);
`
	var exists bool
	if err := dao.db.QueryRowx(
		query,
		raterId,
		serviceUuid,
	).Scan(&exists); err != nil {
		return false, err
	}

	return exists, nil
}

// Check if service status is one of the following:
//   - completed
//   - expired
//   - canceled
func (dao *RateDAO) isServiceStatusRatable(serviceUuid string) (bool, error) {
	query := `
SELECT EXISTS (
	SELECT
		1
	FROM
		services
	WHERE
		uuid = $1 AND
		(
			service_status = 'completed' OR
			service_status = 'expired' OR
			service_status = 'canceled'
		)
);
	`

	var ratable bool

	if err := dao.db.QueryRowx(query, serviceUuid).Scan(&ratable); err != nil {
		return false, err
	}

	return ratable, nil
}

type IsServiceRatableParams struct {
	ParticipantId int
	ServiceUuid   string
}

func (dao *RateDAO) IsServiceRatable(p IsServiceRatableParams) error {
	// Checks if the user is service participant
	isPar, err := dao.isServiceParticipant(
		p.ParticipantId,
		p.ServiceUuid,
	)

	if err != nil {
		return err
	}

	if !isPar {
		return errors.New("user is not a service participant.")

	}

	statusRatable, err := dao.isServiceStatusRatable(p.ServiceUuid)

	if err != nil {
		return err
	}

	if !statusRatable {
		return errors.New("service status is not ratable")
	}

	// Checks if the participant has rated the service already.
	hasRated, err := dao.hasRated(p.ParticipantId, p.ServiceUuid)

	if err != nil {
		return err
	}

	if hasRated {
		return errors.New("participant has already reated the service")
	}

	return nil
}

type CreateServiceRatingParams struct {
	Rating      int
	RaterId     int
	ServiceUuid string
	Comment     string
}

func (dao *RateDAO) CreateServiceRating(p CreateServiceRatingParams) error {
	query := `
INSERT INTO service_ratings (rater_id, service_id, rating, comments)
SELECT
	$1 AS rater_id,
	services.id AS service_id,
	$2 AS rating,
	$3 AS comments
FROM services
WHERE services.uuid = $4
RETURNING *;
`
	_, err := dao.db.Exec(
		query,
		p.RaterId,
		p.Rating,
		p.Comment,
		p.ServiceUuid,
	)

	return err
}

func (dao *RateDAO) GetUserRatings(p contracts.GetUserRatingsParams) ([]models.UserRatings, error) {
	query := `
WITH participate_services AS (
	SELECT
		id,
		created_at
	FROM
		services
	WHERE
		customer_id = $1 OR
		service_provider_id =$1
	ORDER BY created_at DESC
	LIMIT $1
	OFFSET $2
)
SELECT
	service_ratings.*,
	users.username AS rater_username,
	users.uuid AS rater_uuid,
	users.avatar_url AS rater_avatar_url
FROM
	service_ratings
INNER JOIN users ON service_ratings.rater_id = users.id
WHERE service_id IN (
	SELECT
		id
	FROM
		participate_services
)
AND NOT service_ratings.rater_id = $3
ORDER BY created_at;
`
	rows, err := dao.db.Queryx(
		query,
		p.PerPage,
		p.Offset,
		p.UserId,
	)

	ms := make([]models.UserRatings, 0)

	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("DEBUG spot 1 %v", ms)

			return ms, nil
		}

		return nil, err
	}

	for rows.Next() {
		var m models.UserRatings

		if err := rows.StructScan(&m); err != nil {
			return nil, err
		}

		ms = append(ms, m)
	}

	return ms, nil
}

func (dao *RateDAO) HasCommented(serviceId, userId int) (bool, error) {
	query := `
SELECT EXISTS (
	SELECT 1
	FROM service_ratings
	WHERE
		rater_id = $1 AND
		service_id = $2
);
`
	var exists bool

	if err := dao.db.QueryRowx(
		query,
		userId,
		serviceId,
	).Scan(&exists); err != nil {
		return false, err
	}

	return exists, nil
}
