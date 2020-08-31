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
	Uuid        string  `json:"uuid"`
	AvatarURL   *string `json:"avatar_url"`
	Nationality *string `json:"nationality"`
	Region      *string `json:"region"`
	Age         *int32  `json:"age"`
	Height      *string `json:"height"`
	Weight      *string `json:"weight"`
	BreastSize  *string `json:"breast_size"`
}

func (ut *UserTransform) TransformPatchedUser(user *models.User) *TransformedPatchedUser {
	t := &TransformedPatchedUser{
		Uuid: user.Uuid,
	}

	if user.AvatarUrl.Valid {
		t.AvatarURL = &user.AvatarUrl.String
	}

	if user.Nationality.Valid {
		t.Nationality = &user.Nationality.String
	}

	if user.Region.Valid {
		t.Region = &user.Region.String
	}

	if user.Age.Valid {
		t.Age = &user.Age.Int32
	}

	if user.Height.Valid {
		t.Height = &user.Height.String
	}

	if user.Weight.Valid {
		t.Weight = &user.Weight.String
	}

	if user.BreastSize.Valid {
		t.BreastSize = &user.BreastSize.String
	}

	return t
}