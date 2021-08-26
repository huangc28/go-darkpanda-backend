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
	FCMTopioc     string `json:"fcm_topic"`
}

func (rt *RegisterTransform) TransformUser(m *models.User) *TransformedUser {
	tu := &TransformedUser{
		Uuid:          m.Uuid,
		Username:      m.Username,
		PhoneVerified: m.PhoneVerified,
		Gender:        string(m.Gender),
		FCMTopioc:     m.FcmTopic.String,
	}

	return tu
}

type TransformedSendMobileVerifyCode struct {
	UUID         string `json:"uuid"`
	VerifyPrefix string `json:"verify_prefix"`
}

func (at *RegisterTransform) TransformSendMobileVerifyCode(uuid string, verifyPrefix string) TransformedSendMobileVerifyCode {
	return TransformedSendMobileVerifyCode{
		UUID:         uuid,
		VerifyPrefix: verifyPrefix,
	}
}
