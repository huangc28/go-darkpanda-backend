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
	Username      string `json:"username"`
	PhoneVerified bool   `json:"phone_verified"`
	Gender        string `json:"gender"`
	Uuid          string `json:"uuid"`
}

func (at *AuthTransform) TransformUser(m *models.User) *TransformedUser {
	tu := &TransformedUser{
		Uuid:          m.Uuid,
		Username:      m.Username,
		PhoneVerified: m.PhoneVerified,
		Gender:        string(m.Gender),
	}

	return tu
}

type TransformedSendLoginMobileVerifyCode struct {
	UUID         string `json:"uuid"`
	VerifyPrefix string `json:"verify_prefix"`
}

func (at *AuthTransform) TransformSendLoginMobileVerifyCode(uuid string, verifyPrefix string) TransformedSendLoginMobileVerifyCode {
	return TransformedSendLoginMobileVerifyCode{
		UUID:         uuid,
		VerifyPrefix: verifyPrefix,
	}
}