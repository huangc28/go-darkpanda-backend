package service

import (
	"time"

	"github.com/huangc28/go-darkpanda-backend/config"
	"github.com/huangc28/go-darkpanda-backend/internal/app/models"
	darkfirestore "github.com/huangc28/go-darkpanda-backend/internal/app/pkg/dark_firestore"
	"github.com/shopspring/decimal"
)

type TransformedService struct {
	ServiceUuid     string    `json:"service_uuid"`
	ServiceStatus   string    `json:"service_status"`
	AppointmentTime time.Time `json:"appointment_time"`
	Username        string    `json:"chat_partner_username"`
	UserUuid        string    `json:"chat_partner_user_uuid"`
	AvatarUrl       string    `json:"chat_partner_avatar_url"`
	ChannelUuid     string    `json:"channel_uuid"`
	InquiryUuid     string    `json:"inquiry_uuid"`
}

type TransformedGetIncomingService struct {
	TransformedService

	// Messages only contains the latest message of the chatroom. It's an empty array
	// If the chatroom does not contain any message.
	Messages []*darkfirestore.ChatMessage `json:"messages"`
}

type TransformedGetIncomingServices struct {
	Services []TransformedGetIncomingService `json:"services"`
}

func TransformGetIncomingServices(results []ServiceResult, latestMessageMap map[string][]*darkfirestore.ChatMessage) TransformedGetIncomingServices {

	trfRes := make([]TransformedGetIncomingService, 0)

	for _, res := range results {
		chatMsgs := []*darkfirestore.ChatMessage{}

		if v, exists := latestMessageMap[res.ChannelUuid.String]; exists {
			chatMsgs = v
		}

		c := TransformedGetIncomingService{
			TransformedService{
				res.ServiceUuid.String,
				res.ServiceStatus.String,
				res.AppointmentTime.Time,
				res.Username.String,
				res.UserUuid.String,
				res.AvatarUrl.String,
				res.ChannelUuid.String,
				res.InquiryUuid.String,
			},
			chatMsgs,
		}

		trfRes = append(trfRes, c)
	}

	return TransformedGetIncomingServices{
		Services: trfRes,
	}
}

type TransformedGetOverdueServices struct {
	Services []TransformedService `json:"services"`
}

func TransformOverDueServices(results []ServiceResult) TransformedGetOverdueServices {
	trfRes := make([]TransformedService, 0)

	for _, res := range results {
		c := TransformedService{
			res.ServiceUuid.String,
			res.ServiceStatus.String,
			res.AppointmentTime.Time,
			res.Username.String,
			res.UserUuid.String,
			res.AvatarUrl.String,
			res.ChannelUuid.String,
			res.InquiryUuid.String,
		}

		trfRes = append(trfRes, c)
	}

	return TransformedGetOverdueServices{
		Services: trfRes,
	}
}

type TransformedScanServiceQrCode struct {
	Uuid          string    `json:"uuid"`
	ServiceStatus string    `json:"service_status"`
	StartTime     time.Time `json:"start_time"`
	EndTime       time.Time `json:"end_time"`
}

func TransformScanServiceQrCode(srv *models.Service) *TransformedScanServiceQrCode {
	loc, _ := time.LoadLocation(config.GetAppConf().AppTimeZone)
	tzSt := srv.StartTime.Time.In(loc)
	tzEt := srv.EndTime.Time.In(loc)

	return &TransformedScanServiceQrCode{
		Uuid:          srv.Uuid.String(),
		ServiceStatus: srv.ServiceStatus.ToString(),
		StartTime:     tzSt,
		EndTime:       tzEt,
	}
}

type TrfmedServiceName struct {
	ServiceName string `json:"name"`
}

type TrfmedServiceNames struct {
	ServiceNames []TrfmedServiceName `json:"service_names"`
}

func TransformServiceName(serviceNames []*models.ServiceName) TrfmedServiceNames {
	ns := make([]TrfmedServiceName, 0)

	for _, name := range serviceNames {
		n := TrfmedServiceName{
			ServiceName: string(name.ServiceName),
		}

		ns = append(ns, n)
	}

	return TrfmedServiceNames{
		ServiceNames: ns,
	}

}

type TrfedPaymentDetail struct {
	Price float64 `json:"price"`

	Address   string    `json:"address"`
	StartTime time.Time `json:"start_time"`
	Duration  *int64    `json:"duration"`

	PickerUuid      string  `json:"picker_uuid"`
	PickerUsername  string  `json:"picker_username"`
	PickerAvatarUrl *string `json:"picker_avatar_url"`

	HasCommented bool `json:"has_commented"`
}

func TrfPaymentDetail(m *models.ServicePaymentDetail, hasCommented bool) TrfedPaymentDetail {
	trf := TrfedPaymentDetail{
		Price: m.Price,

		Address:   m.Address,
		StartTime: m.StartTime,

		PickerUuid:     m.PickerUuid,
		PickerUsername: m.PickerUsername,

		HasCommented: hasCommented,
	}

	if m.Duration.Valid {
		trf.Duration = &m.Duration.Int64
	}

	if m.PickerAvatarUrl.Valid {
		trf.PickerAvatarUrl = &m.PickerAvatarUrl.String
	}

	return trf
}

func TrfServiceDetail(srv models.Service, matchingFee int) (interface{}, error) {
	decPrice, err := decimal.NewFromString(srv.Price.String)

	if err != nil {
		return nil, err
	}

	floatPrice, _ := decPrice.Float64()

	var (
		startTime *time.Time
		endTime   *time.Time
	)

	if srv.StartTime.Valid {
		startTime = &srv.StartTime.Time
	}

	if srv.EndTime.Valid {
		endTime = &srv.StartTime.Time
	}

	return struct {
		Uuid            string    `json:"uuid"`
		Price           float64   `json:"price"`
		Duration        int32     `json:"duration"`
		AppointmentTime time.Time `json:"appointment_time"`

		ServiceType   string     `json:"service_type"`
		ServiceStatus string     `json:"service_status"`
		Address       string     `json:"address"`
		StartTime     *time.Time `json:"start_time"`
		EndTime       *time.Time `json:"end_time"`
		MatchingFee   int        `json:"matching_fee"`
	}{
		srv.Uuid.String(),
		floatPrice,
		srv.Duration.Int32,
		srv.AppointmentTime.Time,
		srv.ServiceType.ToString(),
		srv.ServiceStatus.ToString(),
		srv.Address.String,
		startTime,
		endTime,
		matchingFee,
	}, nil
}

type TransformedRate struct {
	RaterUsername  string `json:"rater_username"`
	RaterUuid      string `json:"rater_uuid"`
	RaterAvatarUrl string `json:"rater_avatar_url"`

	Comment   string    `json:"comment"`
	Rating    int32     `json:"rating"`
	CreatedAt time.Time `json:"created_at"`
}

func TransformRate(raterInfo *models.User, srvRating *models.ServiceRating) *TransformedRate {
	return &TransformedRate{
		RaterUsername:  raterInfo.Username,
		RaterUuid:      raterInfo.Uuid,
		RaterAvatarUrl: raterInfo.AvatarUrl.String,

		Comment:   srvRating.Comments.String,
		Rating:    srvRating.Rating.Int32,
		CreatedAt: srvRating.CreatedAt,
	}
}
