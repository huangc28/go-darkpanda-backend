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

type GetServicePartnerInfoParams struct {
	MyId        int
	ServiceUuid string
}

type GetServiceRatingParams struct {
	ServiceUuid string
	RaterId     int
}

type IsServiceRatableParams struct {
	ParticipantId int
	ServiceUuid   string
}

type CreateServiceRatingParams struct {
	Rating      int
	RaterId     int
	RateeId     int
	ServiceUuid string
	Comment     string
}

type RateDAOer interface {
	WithTx(tx db.Conn) RateDAOer
	GetUserRatings(p GetUserRatingsParams) ([]models.UserRatings, error)
	HasCommented(serviceId, raterId int) (bool, error)
	GetServicePartnerInfo(p GetServicePartnerInfoParams) (*models.User, error)
	GetServiceRating(p GetServiceRatingParams) (*models.ServiceRating, error)
	IsServiceRatable(p IsServiceRatableParams) error
	CreateServiceRating(p CreateServiceRatingParams) (*models.ServiceRating, error)
	IsServiceParticipant(pId int, srvUuid string) (bool, error)
}
