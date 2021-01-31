package auth

import "github.com/huangc28/go-darkpanda-backend/internal/app/models"

type TransformedUser struct {
	Username      string `json:"username"`
	PhoneVerified bool   `json:"phone_verified"`
	Gender        string `json:"gender"`
	Uuid          string `json:"uuid"`
}

type AuthTransformer interface {
	TransformUser(m *models.User) *TransformedUser
}

type AuthTransform struct{}

func NewTransform() *AuthTransform {
	return &AuthTransform{}
}

type TransformedSendLoginMobileVerifyCode struct {
	UUID         string `json:"uuid"`
	VerifyPrefix string `json:"verify_prefix"`
	Mobile       string `json:"mobile"`
}

func (at *AuthTransform) TransformSendLoginMobileVerifyCode(uuid string, verifyPrefix string, mobile string) TransformedSendLoginMobileVerifyCode {
	return TransformedSendLoginMobileVerifyCode{
		UUID:         uuid,
		VerifyPrefix: verifyPrefix,
		Mobile:       mobile,
	}
}
