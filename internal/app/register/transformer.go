package register

import "github.com/huangc28/go-darkpanda-backend/internal/app/models"

type RegisterTransform struct{}

func NewTransform() *RegisterTransform {
	return &RegisterTransform{}
}

type TransformedUser struct {
	Username      string `json:"username"`
	PhoneVerified bool   `json:"phone_verified"`
	Gender        string `json:"gender"`
	Uuid          string `json:"uuid"`
}

func (rt *RegisterTransform) TransformUser(m *models.User) *TransformedUser {
	tu := &TransformedUser{
		Uuid:          m.Uuid,
		Username:      m.Username,
		PhoneVerified: m.PhoneVerified,
		Gender:        string(m.Gender),
	}

	return tu
}
