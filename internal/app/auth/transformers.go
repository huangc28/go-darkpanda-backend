package auth

import "github.com/huangc28/go-darkpanda-backend/internal/app/models"

type AuthTransformer interface {
	TransformUser(m *models.User) *TransformedUser
}

type AuthTransform struct{}

func NewTransform() *AuthTransform {
	return &AuthTransform{}
}

type TransformedSendMobileVerifyCode struct {
	UUID         string `json:"uuid"`
	VerifyPrefix string `json:"verify_prefix"`
}

func (at *AuthTransform) TransformSendMobileVerifyCode(uuid string, verifyPrefix string) TransformedSendMobileVerifyCode {
	return TransformedSendMobileVerifyCode{
		UUID:         uuid,
		VerifyPrefix: verifyPrefix,
	}
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
