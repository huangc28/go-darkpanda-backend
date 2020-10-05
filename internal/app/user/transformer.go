package user

import (
	"github.com/huangc28/go-darkpanda-backend/internal/models"

	convertnullsql "github.com/huangc28/go-darkpanda-backend/internal/app/pkg/convert_null_sql"
)

type UserTransformer interface {
	TransformUserWithInquiry(m *models.User, iq *models.ServiceInquiry) *UserTransform
}

type UserTransform struct{}

func NewTransform() *UserTransform {
	return &UserTransform{}
}

type TransformedUser struct {
	Username    string        `json:"username"`
	Gender      models.Gender `json:"gender"`
	Uuid        string        `json:"uuid"`
	AvatarUrl   string        `json:"avatar_url"`
	Nationality string        `json:"nationality"`
	Region      string        `json:"region"`
	Age         int           `json:"age"`
	Height      float32       `json:"height"`
	Weight      float32       `json:"weight"`
	Description string        `json:"description"`
}

func (ut *UserTransform) TransformUser(m *models.User) *TransformedUser {
	return &TransformedUser{
		Username:  m.Username,
		Gender:    m.Gender,
		Uuid:      m.Uuid,
		AvatarUrl: m.AvatarUrl.String,
	}
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

type TransformedMaleUser struct {
	Username    string        `json:"username"`
	Gender      models.Gender `json:"gender"`
	Uuid        string        `json:"uuid"`
	AvatarUrl   string        `json:"avatar_url"`
	Nationality string        `json:"nationality"`
	Region      string        `json:"region"`
	Age         int           `json:"age"`
	Height      *float32      `json:"height"`
	Weight      *float32      `json:"weight"`
	Description string        `json:"description"`
}

func (ut *UserTransform) TransformMaleUser(user models.User) (*TransformedMaleUser, error) {
	var (
		weightF *float32
		heightF *float32
		err     error
	)

	if user.Weight.Valid {
		weightF, err = convertnullsql.ConvertSqlNullStringToFloat32(user.Weight)

		if err != nil {
			return nil, err
		}
	}

	if user.Height.Valid {
		heightF, err = convertnullsql.ConvertSqlNullStringToFloat32(user.Height)

		if err != nil {
			return nil, err
		}
	}

	return &TransformedMaleUser{
		Uuid:        user.Uuid,
		Username:    user.Username,
		Gender:      user.Gender,
		AvatarUrl:   user.AvatarUrl.String,
		Nationality: user.Nationality.String,
		Region:      user.Region.String,
		Age:         int(user.Age.Int32),
		Height:      heightF,
		Weight:      weightF,
		Description: user.Description.String,
	}, nil
}
