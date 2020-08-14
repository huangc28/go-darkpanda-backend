package auth

import "github.com/huangc28/go-darkpanda-backend/internal/models"

type AuthTransformer interface {
	TransformUser(m *models.User) *TransformedUser
}

type AuthTransform struct{}

func NewTransform() *AuthTransform {
	return &AuthTransform{}
}

type TransformedUser struct {
	ID            int64  `json:"id"`
	Username      string `json:"username"`
	PhoneVerified bool   `json:"phone_verified"`
	Gender        string `json:"gender"`
}

func (at *AuthTransform) TransformUser(m *models.User) *TransformedUser {
	tu := &TransformedUser{
		ID:            m.ID,
		Username:      m.Username,
		PhoneVerified: m.PhoneVerified.Bool,
		Gender:        string(m.Gender),
	}

	return tu
}
