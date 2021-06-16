package rate

import (
	"time"

	"github.com/huangc28/go-darkpanda-backend/internal/app/models"
)

type RateTransform struct{}

func NewTransform() *RateTransform {
	return &RateTransform{}
}

type TransformedRate struct {
	RaterUsername  string `json:"rater_username"`
	RaterUuid      string `json:"rater_uuid"`
	RaterAvatarUrl string `json:"rater_avatar_url"`

	Comment   string    `json:"comment"`
	Rating    int32     `json:"rating"`
	CreatedAt time.Time `json:"created_at"`
}

func (ba *RateTransform) TransformRate(raterInfo *models.User, srvRating *models.ServiceRating) *TransformedRate {
	return &TransformedRate{
		RaterUsername:  raterInfo.Username,
		RaterUuid:      raterInfo.Uuid,
		RaterAvatarUrl: raterInfo.AvatarUrl.String,

		Comment:   srvRating.Comments.String,
		Rating:    srvRating.Rating.Int32,
		CreatedAt: srvRating.CreatedAt,
	}
}
