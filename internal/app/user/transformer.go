package user

import (
	"time"

	"github.com/huangc28/go-darkpanda-backend/internal/app/models"

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
	FCMTopic    string        `json:"fcm_topic"`
}

func (ut *UserTransform) TransformUser(m *models.User) *TransformedUser {
	return &TransformedUser{
		Username:  m.Username,
		Gender:    m.Gender,
		Uuid:      m.Uuid,
		AvatarUrl: m.AvatarUrl.String,
		FCMTopic:  m.FcmTopic.String,
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

// Format user traits to array of traits.
type TraitType string

const (
	Age    TraitType = "age"
	Height TraitType = "height"
	Weight TraitType = "weight"
)

type Trait struct {
	Type  TraitType   `json:"type"`
	Value interface{} `json:"value"`
}

func formatUserTraits(user models.User) ([]Trait, error) {
	traits := make([]Trait, 0)

	if user.Age.Valid {
		traits = append(traits, Trait{
			Type:  Age,
			Value: user.Age.Int32,
		})
	}

	if user.Height.Valid {
		height, err := convertnullsql.ConvertSqlNullStringToFloat32(user.Height)

		if err != nil {
			return nil, err
		}

		traits = append(traits, Trait{
			Type:  Height,
			Value: height,
		})
	}

	if user.Weight.Valid {
		weight, err := convertnullsql.ConvertSqlNullStringToFloat32(user.Weight)

		if err != nil {
			return nil, err
		}

		traits = append(traits, Trait{
			Type:  Weight,
			Value: weight,
		})

	}

	return traits, nil
}

type TransformedViewableUserProfile struct {
	Username    string        `json:"username"`
	Gender      models.Gender `json:"gender"`
	Uuid        string        `json:"uuid"`
	AvatarUrl   string        `json:"avatar_url"`
	Nationality string        `json:"nationality"`
	Region      string        `json:"region"`
	Description string        `json:"description"`
	Traits      []Trait       `json:"traits"`
}

func (ut *UserTransform) TransformViewableUserProfile(user models.User) (*TransformedViewableUserProfile, error) {
	traits, err := formatUserTraits(user)

	if err != nil {
		return nil, err
	}

	return &TransformedViewableUserProfile{
		Uuid:        user.Uuid,
		Username:    user.Username,
		Gender:      user.Gender,
		AvatarUrl:   user.AvatarUrl.String,
		Nationality: user.Nationality.String,
		Region:      user.Region.String,
		Traits:      traits,
		Description: user.Description.String,
	}, nil
}

type TransformedUserImage struct {
	Url string `json:"url"`
}

type TransformedUserImages struct {
	Images []TransformedUserImage `json:"images"`
}

func (ut *UserTransform) TransformUserImages(imgs []models.Image) TransformedUserImages {
	trfImgs := make([]TransformedUserImage, 0)

	for _, img := range imgs {
		trfImg := TransformedUserImage{
			Url: img.Url,
		}

		trfImgs = append(trfImgs, trfImg)
	}

	return TransformedUserImages{
		trfImgs,
	}
}

type TransformedPaymentInfos struct {
	Payments []TransformedPaymentInfo `json:"payments"`
}

type TransformedPaymentInfo struct {
	Price   *float32 `json:"price"`
	Service TransformedPaymentServiceInfo
	Payer   TransformedPaymentPayerInfo
}

type TransformedPaymentServiceInfo struct {
	Uuid        string `json:"uuid"`
	ServiceType string `json:"service_type"`
}

type TransformedPaymentPayerInfo struct {
	Uuid      string `json:"uuid"`
	Username  string `json:"username"`
	AvatarURL string `json:"avatar_url"`
}

func (ut *UserTransform) TransformPaymentInfo(infos []models.PaymentInfo) (*TransformedPaymentInfos, error) {
	trfmInfos := make([]TransformedPaymentInfo, 0)

	for _, info := range infos {

		trfmSrv := TransformedPaymentServiceInfo{
			info.Service.Uuid.String,
			info.Service.ServiceType.ToString(),
		}

		trfmUser := TransformedPaymentPayerInfo{
			info.Payer.Uuid,
			info.Payer.Username,
			info.Payer.AvatarUrl.String,
		}

		paymentPrice, err := convertnullsql.ConvertSqlNullStringToFloat32(info.Service.Price)

		if err != nil {
			return nil, err
		}

		trfmInfo := TransformedPaymentInfo{
			paymentPrice,
			trfmSrv,
			trfmUser,
		}

		trfmInfos = append(trfmInfos, trfmInfo)

	}

	return &TransformedPaymentInfos{
		trfmInfos,
	}, nil
}

type TransformedHistoricalServices struct {
	Services []TransformedService `json:"services"`
}

type TransformedService struct {
	Uuid          string    `json:"uuid"`
	Price         *float32  `json:"price"`
	ServiceType   string    `json:"service_type"`
	ServiceStatus string    `json:"service_status"`
	CreatedAt     time.Time `json:"created_at"`
}

func (ut *UserTransform) TransformHistoricalServices(services []models.Service) (*TransformedHistoricalServices, error) {
	trfmServices := make([]TransformedService, 0)

	for _, srv := range services {
		srvPrice, err := convertnullsql.ConvertSqlNullStringToFloat32(srv.Price)

		if err != nil {
			return nil, err
		}

		trfmSrv := TransformedService{
			Uuid:          srv.Uuid.String,
			Price:         srvPrice,
			ServiceType:   srv.ServiceType.ToString(),
			ServiceStatus: srv.ServiceStatus.ToString(),
			CreatedAt:     srv.CreatedAt,
		}

		trfmServices = append(trfmServices, trfmSrv)
	}

	return &TransformedHistoricalServices{
		Services: trfmServices,
	}, nil
}

type TrfmedUserRating struct {
	Comment        *string `json:"comment"`
	Rating         int32   `json:"rating"`
	RaterUsername  string  `json:"rater_username"`
	RaterUuid      string  `json:"rater_uuid"`
	RaterAvatarUrl *string `json:"rater_avatar_url"`
}

func TrfGetUserRatings(ms []models.UserRatings) []TrfmedUserRating {
	trfms := make([]TrfmedUserRating, 0)

	for _, m := range ms {
		trfm := TrfmedUserRating{
			nil,
			m.Rating.Int32,
			m.RaterUsername,
			m.RaterUuid,
			nil,
		}

		if m.Comments.Valid {
			trfm.Comment = &m.Comments.String
		}

		if m.RaterAvatarUrl.Valid {
			trfm.RaterAvatarUrl = &m.RaterUuid
		}

		trfms = append(trfms, trfm)
	}

	return trfms
}
