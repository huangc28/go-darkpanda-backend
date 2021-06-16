package rate

import (
	"fmt"

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
	PartnerId   int
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
	objCriteria := "service_provider_id"

	if p.Gender == models.GenderFemale {
		objCriteria = " customer_id"
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
		objCriteria,
		objCriteria,
	)

	var m models.User

	if err := dao.db.QueryRowx(
		query,
		p.PartnerId,
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
	rating,
	comments,
	created_at
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
