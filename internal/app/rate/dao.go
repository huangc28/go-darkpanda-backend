package rate

import (
	"github.com/golobby/container/pkg/container"
	"github.com/huangc28/go-darkpanda-backend/db"
	"github.com/huangc28/go-darkpanda-backend/internal/app/contracts"
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

func (dao *RateDAO) GetUserRating(uuid string) (*contracts.GetUserRatingParams, error) {
	query := `
		SELECT u2.id, u2.username, u2.avatar_url,
			ur.rating, ur."comments", ur.created_at 
		FROM user_ratings ur 
		INNER JOIN users u ON ur.to_user_id =u.id
		LEFT JOIN users u2 ON ur.from_user_id = u2.id
		WHERE u.uuid=$1;
	`

	rate := contracts.GetUserRatingParams{}

	if err := dao.db.QueryRow(query, uuid).Scan(
		&rate.ID,
		&rate.Username,
		&rate.AvatarUrl,
		&rate.Rating,
		&rate.Comments,
		&rate.CreatedAt,
	); err != nil {
		return nil, err
	}

	return &rate, nil
}
