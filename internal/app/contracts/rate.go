package contracts

import (
	"github.com/huangc28/go-darkpanda-backend/db"
	"github.com/huangc28/go-darkpanda-backend/internal/app/models"
)

type GetUserRatingsParams struct {
	UserId  int
	PerPage int
	Offset  int
}

type RateDAOer interface {
	WithTx(tx db.Conn) RateDAOer
	GetUserRatings(p GetUserRatingsParams) ([]models.UserRatings, error)
}
