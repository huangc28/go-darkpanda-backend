package user

import (
	"github.com/huangc28/go-darkpanda-backend/internal/models"
)

type UserTransformer interface {
	TransformUserWithInquiry(m *models.User, iq *models.ServiceInquiry) *UserTransform
}

type UserTransform struct{}

func NewTransform() *UserTransform {
	return &UserTransform{}
}

type TransformUserWithInquiryData struct {
	Username  string                   `json:"username"`
	Gender    models.Gender            `json:"gender"`
	Uuid      string                   `json:"uuid"`
	Inquiries []*models.ServiceInquiry `json:"inquiries"`
}

func (ut *UserTransform) TransformUserWithInquiry(m *models.User, si []*models.ServiceInquiry) *TransformUserWithInquiryData {
	t := &TransformUserWithInquiryData{
		Username:  m.Username,
		Gender:    m.Gender,
		Uuid:      m.Uuid,
		Inquiries: si,
	}

	return t
}

type TransformedPatchedUser struct {
	Uuid      string  `json:"uuid"`
	AvatarURL *string `json:"avatar_url"`
}

func (ut *UserTransform) TransformPatchedUser(user *models.User) *TransformedPatchedUser {
	t := &TransformedPatchedUser{
		Uuid: user.Uuid,
	}

	if user.AvatarUrl.Valid {
		t.AvatarURL = &user.AvatarUrl.String
	}

	return t
}
