package rate

import (
	"time"

	"github.com/huangc28/go-darkpanda-backend/internal/app/contracts"
)

type RateTransform struct{}

func NewTransform() *RateTransform {
	return &RateTransform{}
}

type TransformedRate struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	AvatarUrl string    `json:"avatar_url"`
	Rating    int       `json:"rating"`
	Comments  string    `json:"comments"`
	CreatedAt time.Time `json:"created_at"`
}

func (ba *RateTransform) TransformRate(rate *contracts.GetUserRatingParams) *TransformedRate {
	return &TransformedRate{
		ID:        rate.ID,
		Username:  rate.Username,
		AvatarUrl: rate.AvatarUrl.String,
		Rating:    rate.Rating,
		Comments:  rate.Comments.String,
		CreatedAt: rate.CreatedAt,
	}
}
