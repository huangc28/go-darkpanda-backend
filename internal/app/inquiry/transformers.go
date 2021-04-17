package inquiry

import (
	"fmt"
	"strconv"
	"time"

	"github.com/huangc28/go-darkpanda-backend/internal/app/contracts"
	"github.com/huangc28/go-darkpanda-backend/internal/app/inquiry/util"
	"github.com/huangc28/go-darkpanda-backend/internal/app/models"
	convertnullsql "github.com/huangc28/go-darkpanda-backend/internal/app/pkg/convert_null_sql"
	"github.com/shopspring/decimal"
)

type InquiryTransform struct{}

func NewTransform() *InquiryTransform {
	return &InquiryTransform{}
}

type TransformedInquiry struct {
	Uuid          string    `json:"inquiry_uuid"`
	Budget        float64   `json:"budget"`
	ServiceType   string    `json:"service_type"`
	InquiryStatus string    `json:"inquiry_status"`
	CreatedAt     time.Time `json:"created_at"`
}

func (t *InquiryTransform) TransformEmitInquiry(m models.ServiceInquiry) (TransformedInquiry, error) {
	budget, err := strconv.ParseFloat(m.Budget, 64)

	if err != nil {
		return TransformedInquiry{}, err
	}

	tiq := TransformedInquiry{
		Uuid:          m.Uuid,
		Budget:        budget,
		ServiceType:   string(m.ServiceType),
		InquiryStatus: string(m.InquiryStatus),
		CreatedAt:     m.CreatedAt,
	}

	return tiq, nil
}

func (t *InquiryTransform) TransformInquiry(m models.ServiceInquiry) (TransformedInquiry, error) {
	// Convert string to float.
	badget, err := strconv.ParseFloat(m.Budget, 64)

	if err != nil {
		return TransformedInquiry{}, err

	}

	tiq := TransformedInquiry{
		Uuid:          m.Uuid,
		Budget:        badget,
		ServiceType:   string(m.ServiceType),
		InquiryStatus: string(m.InquiryStatus),
		CreatedAt:     m.CreatedAt,
	}

	return tiq, nil
}

type TransformedService struct {
	Uuid          string              `json:"uuid"`
	ServiceStatus string              `json:"service_status"`
	ServiceType   string              `json:"service_type"`
	User          TransformedInquirer `json:"inquirer"`
}

type TransformedInquirer struct {
	Uuid        string `json:"uuid"`
	Username    string `json:"username"`
	PremiumType string `json:"premium_type"`
}

func (t *InquiryTransform) TransformService(m models.Service, iqer models.User) TransformedService {
	return TransformedService{
		Uuid:          m.Uuid.String(),
		ServiceStatus: string(m.ServiceStatus),
		ServiceType:   string(m.ServiceType),
		User: TransformedInquirer{
			Uuid:        iqer.Uuid,
			Username:    iqer.Username,
			PremiumType: string(iqer.PremiumType),
		},
	}
}

type TransformedPickupInquiry struct {
	ServiceType   string    `json:"service_type"`
	InquiryUUID   string    `json:"inquiry_uuid"`
	InquiryStatus string    `json:"inquiry_status"`
	ExpiredAt     time.Time `json:"expired_at"`
	CreatedAt     time.Time `json:"created_at"`
}

func (t *InquiryTransform) TransformPickupInquiry(iq models.ServiceInquiry) TransformedPickupInquiry {
	return TransformedPickupInquiry{
		ServiceType:   iq.ServiceType.ToString(),
		InquiryStatus: iq.InquiryStatus.ToString(),
		InquiryUUID:   iq.Uuid,
		ExpiredAt:     iq.ExpiredAt.Time,
		CreatedAt:     iq.CreatedAt,
	}
}

type TransformedGirlApproveInquiry struct {
	TransformedInquiry
	Price           string    `json:"price"`
	Duration        int32     `json:"duration"`
	AppointmentTime time.Time `json:"appointment_time"`
	Lat             string    `json:"lat"`
	Lng             string    `json:"lng"`
}

func (t *InquiryTransform) TransformGirlApproveInquiry(iq models.ServiceInquiry) (*TransformedGirlApproveInquiry, error) {
	tiq, err := t.TransformInquiry(iq)

	if err != nil {
		return nil, err
	}

	tPrice, err := strconv.ParseFloat(
		iq.Price.String,
		64,
	)

	if err != nil {
		return nil, err
	}

	latDec, err := decimal.NewFromString(iq.Lat.String)

	if err != nil {
		return nil, err
	}

	lngDec, err := decimal.NewFromString(iq.Lng.String)

	if err != nil {
		return nil, err
	}

	return &TransformedGirlApproveInquiry{
		tiq,
		fmt.Sprintf("%.2f", util.RoundDown2Deci(tPrice)),
		iq.Duration.Int32,
		iq.AppointmentTime.Time,
		latDec.String(),
		lngDec.String(),
	}, nil
}

// TransformBookService response with the information of booked service and the information about
// the service provider.
// @TODO information of service provider should include provider image.
type TransformedServiceProvider struct {
	Uuid     string `json:"uuid"`
	Username string `json:"username"`
}

type TransformedBookedService struct {
	Uuid            string                     `json:"uuid"`
	ServiceProvider TransformedServiceProvider `json:"service_provider"`
	Price           string                     `json:"price"`
	Duration        int32                      `json:"duration"`
	AppointmentTime time.Time                  `json:"appointment_time"`
	Lng             string                     `json:"lng"`
	Lat             string                     `json:"lat"`
	ServiceType     string                     `json:"service_type"`
	CreatedAt       time.Time                  `json:"created_at"`
}

func (t *InquiryTransform) TransformBookedService(srv models.Service, userProvider models.User) *TransformedBookedService {
	tsrv := &TransformedBookedService{
		Uuid:            srv.Uuid.String(),
		Price:           srv.Price.String,
		Duration:        srv.Duration.Int32,
		AppointmentTime: srv.AppointmentTime.Time,
		Lng:             srv.Lng.String,
		Lat:             srv.Lat.String,
		ServiceType:     string(srv.ServiceType),
		CreatedAt:       srv.CreatedAt,
		ServiceProvider: TransformedServiceProvider{
			Uuid:     userProvider.Uuid,
			Username: userProvider.Username,
		},
	}

	return tsrv
}

// Transformed object for GET /v1/inquiries
type TransformedGetInquiryInquirer struct {
	Uuid        string `json:"uuid"`
	Username    string `json:"username"`
	AvatarURL   string `json:"avatar_url"`
	Nationality string `json:"nationality"`
}

type TransformedGetInquiryWithInquirer struct {
	Uuid          string                        `json:"uuid"`
	Budget        float64                       `json:"budget"`
	ServiceType   string                        `json:"service_type"`
	Price         *float64                      `json:"price"`
	Duration      int32                         `json:"duration"`
	Appointment   time.Time                     `json:"appointment_time"`
	Lng           *float32                      `json:"lng"`
	Lat           *float32                      `json:"lat"`
	InquiryStatus string                        `json:"inquiry_status"`
	Inquirer      TransformedGetInquiryInquirer `json:"inquirer"`
}

type TransformedInquiries struct {
	Inquiries []TransformedGetInquiryWithInquirer `json:"inquiries"`
	HasMore   bool                                `json:"has_more"`
}

func (t *InquiryTransform) TransformInquiryList(inquiryList []*InquiryInfo, hasMore bool) (TransformedInquiries, error) {
	trfedIqs := make([]TransformedGetInquiryWithInquirer, 0)
	for _, oi := range inquiryList {
		price, err := convertnullsql.ConvertSqlNullStringToFloat64(oi.Price)

		if err != nil {
			return TransformedInquiries{}, err
		}

		budget, err := strconv.ParseFloat(oi.Budget, 64)

		if err != nil {
			return TransformedInquiries{}, err
		}

		lng, err := convertnullsql.ConvertSqlNullStringToFloat32(oi.Lng)

		if err != nil {
			return TransformedInquiries{}, err
		}

		lat, err := convertnullsql.ConvertSqlNullStringToFloat32(oi.Lat)

		if err != nil {
			return TransformedInquiries{}, err
		}

		trfedIq := TransformedGetInquiryWithInquirer{
			Uuid:          oi.Uuid,
			Budget:        budget,
			ServiceType:   oi.ServiceType.ToString(),
			Price:         price,
			Duration:      oi.Duration.Int32,
			Appointment:   oi.AppointmentTime.Time,
			Lng:           lng,
			Lat:           lat,
			InquiryStatus: oi.InquiryStatus.ToString(),
			Inquirer: TransformedGetInquiryInquirer{
				Uuid:        oi.Inquirer.Uuid,
				Username:    oi.Inquirer.Username,
				AvatarURL:   oi.Inquirer.AvatarUrl.String,
				Nationality: oi.Inquirer.Nationality.String,
			},
		}

		trfedIqs = append(trfedIqs, trfedIq)
	}

	return TransformedInquiries{
		Inquiries: trfedIqs,
		HasMore:   hasMore,
	}, nil
}

type Inquirer struct {
	Username  string `json:"username"`
	UserUuid  string `json:"user_uuid"`
	AvatarUrl string `json:"avatar_url"`
}

type TransformedGetInquiry struct {
	Uuid          string    `json:"uuid"`
	Budget        float64   `json:"budget"`
	ServiceType   string    `json:"service_type"`
	InquiryStatus string    `json:"inquiry_status"`
	Price         *float64  `json:"price"`
	Duration      int32     `json:"duration"`
	Appointment   time.Time `json:"appoinment_time"`
	Lng           *float32  `json:"lng"`
	Lat           *float32  `json:"lat"`
	Address       string    `json:"address"`
	Inquirer      Inquirer  `json:"inquirer"`
}

func (t *InquiryTransform) TransformGetInquiry(iq contracts.InquiryResult) (*TransformedGetInquiry, error) {
	budget, err := strconv.ParseFloat(iq.Budget, 64)

	if err != nil {
		return nil, err
	}

	price, err := strconv.ParseFloat(iq.Budget, 64)

	if err != nil {
		return nil, err
	}

	lng, err := convertnullsql.ConvertSqlNullStringToFloat32(iq.Lng)

	if err != nil {
		return nil, err
	}

	lat, err := convertnullsql.ConvertSqlNullStringToFloat32(iq.Lat)

	if err != nil {
		return nil, err
	}

	return &TransformedGetInquiry{
		iq.Uuid,
		budget,
		iq.ServiceType.ToString(),
		iq.InquiryStatus.ToString(),
		&price,
		iq.Duration.Int32,
		iq.AppointmentTime.Time,
		lng,
		lat,
		iq.Address.String,
		Inquirer{
			Username:  iq.Username,
			UserUuid:  iq.UserUuid,
			AvatarUrl: iq.AvatarUrl.String,
		},
	}, nil
}

type RemovedUser struct {
	UUID string `json:"uuid"`
}

type RevertedInquiry struct {
	UUID          string `json:"uuid"`
	InquiryStatus string `json:"inquiry_status"`
}

type RemovedChatRoom struct {
	ChanelUUID string `json:"chanel_uuid"`
}
type TransformedRevertChatting struct {
	RemovedUsers    []RemovedUser   `json:"removed_users"`
	RemovedChatRoom RemovedChatRoom `json:"removed_chatroom"`
	RevertedInquiry RevertedInquiry `json:"reverted_inquiry"`
}

func (t *InquiryTransform) TransformRevertChatting(removedUsers []models.User, inquiry models.ServiceInquiry, chatroom models.Chatroom) *TransformedRevertChatting {
	rusers := make([]RemovedUser, 0)

	for _, removedUser := range removedUsers {
		ruser := RemovedUser{
			UUID: removedUser.Uuid,
		}

		rusers = append(rusers, ruser)
	}

	return &TransformedRevertChatting{
		RemovedUsers: rusers,
		RemovedChatRoom: RemovedChatRoom{
			ChanelUUID: chatroom.ChannelUuid.String,
		},
		RevertedInquiry: RevertedInquiry{
			UUID:          inquiry.Uuid,
			InquiryStatus: inquiry.InquiryStatus.ToString(),
		},
	}
}

type TransformedGetServiceByInquiryUUID struct {
	UUID            string    `json:"uuid"`
	ServiceType     string    `json:"service_type"`
	Price           float64   `json:"price"`
	Duration        int32     `json:"duration"`
	AppointmentTime time.Time `json:"appointment_time"`
}

func (t *InquiryTransform) TransformGetServiceByInquiryUUID(srv models.Service) (*TransformedGetServiceByInquiryUUID, error) {
	price, err := strconv.ParseFloat(srv.Price.String, 64)

	if err != nil {
		return nil, err
	}

	return &TransformedGetServiceByInquiryUUID{
		UUID:            srv.Uuid.String(),
		ServiceType:     srv.ServiceType.ToString(),
		Price:           price,
		Duration:        srv.Duration.Int32,
		AppointmentTime: srv.AppointmentTime.Time,
	}, nil
}

type TransformGetInquirerInfo struct {
	Username    string         `json:"username"`
	UUID        string         `json:"uuid"`
	AvatarURL   string         `json:"avatar_url"`
	Nationality string         `json:"nationality"`
	Region      string         `json:"region"`
	Age         *int32         `json:"age"`
	Height      *float64       `json:"height"`
	Weight      *float64       `json:"weight"`
	Description string         `json:"description"`
	Image       []models.Image `json:"images"`
}

func (t *InquiryTransform) TransformGetInquirerInfo(inquirer models.User, images []models.Image) (*TransformGetInquirerInfo, error) {
	var (
		err    error
		age    *int32
		height *float64
		weight *float64
	)

	if inquirer.Age.Valid != false {
		age = &inquirer.Age.Int32
	}

	if inquirer.Height.Valid != false {
		*height, err = strconv.ParseFloat(inquirer.Height.String, 64)

		if err != nil {
			return nil, err
		}
	}

	if inquirer.Weight.Valid != false {
		*weight, err = strconv.ParseFloat(inquirer.Weight.String, 64)

		if err != nil {
			return nil, err
		}
	}

	return &TransformGetInquirerInfo{
		Username:    inquirer.Username,
		UUID:        inquirer.Uuid,
		AvatarURL:   inquirer.AvatarUrl.String,
		Nationality: inquirer.Nationality.String,
		Age:         age,
		Height:      height,
		Weight:      weight,
		Description: inquirer.Description.String,
		Image:       images,
	}, nil
}

type ServiceProvider struct {
	Username    string `json:"username"`
	AvatarUrl   string `json:"avatar_url"`
	Uuid        string `json:"uuid"`
	Rating      int    `json:"rating"`
	Description string `json:"description"`
}

type TransformedAgreePickupInquiry struct {
	ServiceProvider ServiceProvider `json:"service_provider"`
	ChannelUuid     string          `json:"channel_uuid"`
}

// TransformAgreePickupInquiry respond with the following data
//   - service provider's info
//      - username
//      - avatar url
//      - user uuid
//      - rating
//      - description
//   - inquiry info
//   - private chat uuid in firestore for inquirer to subscribe
func (t *InquiryTransform) TransformAgreePickupInquiry(picker models.User, pcUuid string) TransformedAgreePickupInquiry {
	trf := TransformedAgreePickupInquiry{
		ServiceProvider: ServiceProvider{
			Username:    picker.Username,
			AvatarUrl:   picker.AvatarUrl.String,
			Uuid:        picker.Uuid,
			Description: picker.Description.String,
		},
		ChannelUuid: pcUuid,
	}

	return trf
}
