package rate

import (
	"database/sql"
	"errors"

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

// GetServicePartnerInfo Retrieve the following info for user rating.
//   -  ratee username
//   -  ratee avatar_url
//   -  ratee uuid
//   -  rating
//   -  service uuid
//   -  comments
func (dao *RateDAO) GetServicePartnerInfo(p contracts.GetServicePartnerInfoParams) (*models.User, error) {
	query := `
WITH chatroom_partner AS (
	SELECT
		*
	FROM
		services
	WHERE
		uuid =  $1 AND
		(
			customer_id = $2 OR
			service_provider_id = $2
		)
)
SELECT
	users.id,
	users.username,
	users.uuid,
	users.avatar_url
FROM users
INNER JOIN  chatroom_partner ON
	users.id = chatroom_partner.customer_id OR
	users.id = chatroom_partner.service_provider_id;
	`
	rows, err := dao.db.Queryx(
		query,
		p.ServiceUuid,
		p.MyId,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var m models.User

	for rows.Next() {
		if err := rows.StructScan(&m); err != nil {
			return nil, err
		}

		// We've found the partner
		if m.ID != int64(p.MyId) {
			break
		}
	}

	return &m, nil
}

func (dao *RateDAO) GetServiceRating(p contracts.GetServiceRatingParams) (*models.ServiceRating, error) {
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

func (dao *RateDAO) IsServiceParticipant(pId int, srvUuid string) (bool, error) {
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

func (dao *RateDAO) IsServiceRatable(p contracts.IsServiceRatableParams) error {
	// Checks if the user is service participant
	isPar, err := dao.IsServiceParticipant(
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

func (dao *RateDAO) CreateServiceRating(p contracts.CreateServiceRatingParams) (*models.ServiceRating, error) {
	query := `
INSERT INTO service_ratings (rater_id, ratee_id, service_id, rating, comments)
SELECT
	$1 AS rater_id,
	$2 AS ratee_id,
	services.id AS service_id,
	$3 AS rating,
	$4 AS comments
FROM services
WHERE services.uuid = $5
RETURNING *;
`

	var m models.ServiceRating

	err := dao.db.QueryRowx(
		query,
		p.RaterId,
		p.RateeId,
		p.Rating,
		p.Comment,
		p.ServiceUuid,
	).StructScan(&m)

	if err != nil {
		return nil, err
	}

	return &m, nil
}

// @Deprecated Use GetRating in github.com/go-darkpanda-backend/internal/app/user/dao.go instead.
func (dao *RateDAO) GetUserRatings(p contracts.GetUserRatingsParams) ([]models.UserRatings, error) {
	query := `
SELECT
	comments,
	rating,
	service_ratings.created_at,
	raters.username AS rater_username,
	raters.uuid AS rater_uuid,
	raters.avatar_url AS rater_avatar_url
FROM
	service_ratings
INNER JOIN users AS raters ON raters.id = service_ratings.rater_id
WHERE
	service_ratings.ratee_id = $1
ORDER BY service_ratings.created_at
LIMIT $2
OFFSET $3;
`
	rows, err := dao.db.Queryx(
		query,
		p.UserID,
		p.PerPage,
		p.Offset,
	)

	ms := make([]models.UserRatings, 0)

	if err == sql.ErrNoRows {
		return ms, nil
	}

	if err != nil {

		return nil, err
	}

	defer rows.Close()

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
