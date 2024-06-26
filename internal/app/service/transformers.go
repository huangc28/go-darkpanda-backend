package service

import (
	"time"

	"github.com/huangc28/go-darkpanda-backend/internal/app/models"
	darkfirestore "github.com/huangc28/go-darkpanda-backend/internal/app/pkg/dark_firestore"
	"github.com/shopspring/decimal"
)

type TransformedService struct {
	ServiceUuid     string    `json:"service_uuid"`
	Status          string    `json:"status"`
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

type TrfedOverDueService struct {
	TransformedService
	CancelCause string `json:"cancel_cause"`
	Refunded    bool   `json:"refunded"`
}

type TransformedGetOverdueServices struct {
	Services []TrfedOverDueService `json:"services"`
}

func TransformOverDueServices(results []ServiceResult) TransformedGetOverdueServices {
	trfRes := make([]TrfedOverDueService, 0)

	for _, res := range results {
		c := TrfedOverDueService{
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
			string(res.CancelCause),
			res.Refunded,
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
	return &TransformedScanServiceQrCode{
		Uuid:          srv.Uuid.String,
		ServiceStatus: srv.ServiceStatus.ToString(),
		StartTime:     srv.StartTime.Time,
		EndTime:       srv.EndTime.Time,
	}
}

type TrfmedServiceOption struct {
	ServiceOption string `json:"name"`
}

type TrfmedServiceOptions struct {
	ServiceNames []TrfmedServiceOption `json:"service_names"`
}

func TransformServiceOptions(serviceOptions []*models.ServiceOption) TrfmedServiceOptions {
	os := make([]TrfmedServiceOption, 0)

	for _, option := range serviceOptions {
		o := TrfmedServiceOption{
			ServiceOption: option.Name,
		}

		os = append(os, o)
	}

	return TrfmedServiceOptions{
		ServiceNames: os,
	}
}

type TrfedPaymentDetail struct {
	Price *float64 `json:"price"`

	ServiceType string     `json:"service_type"`
	Address     string     `json:"address"`
	StartTime   *time.Time `json:"start_time"`
	Duration    *int64     `json:"duration"`
	Refunded    bool       `json:"refunded"`

	PickerUuid      string  `json:"picker_uuid"`
	PickerUsername  string  `json:"picker_username"`
	PickerAvatarUrl *string `json:"picker_avatar_url"`
	CancelCause     string  `json:"cancel_cause"`
	Currency        string  `json:"currency"`

	HasCommented bool    `json:"has_commented"`
	HasBlocked   bool    `json:"has_blocked"`
	MatchingFee  float64 `json:"matching_fee"`
}

type TrfPaymentDetailParams struct {
	PaymentDetail *models.ServicePaymentDetail
	MatchingFee   float64
	HasCommented  bool
	HasBlocked    bool
}

func TrfPaymentDetail(p TrfPaymentDetailParams) TrfedPaymentDetail {
	trf := TrfedPaymentDetail{
		ServiceType: p.PaymentDetail.ServiceType,
		Address:     p.PaymentDetail.Address,
		Refunded:    p.PaymentDetail.Refunded,

		PickerUuid:     p.PaymentDetail.PickerUuid,
		PickerUsername: p.PaymentDetail.PickerUsername,
		CancelCause:    p.PaymentDetail.CancelCause,

		HasCommented: p.HasCommented,
		HasBlocked:   p.HasBlocked,
		MatchingFee:  p.MatchingFee,
	}

	if p.PaymentDetail.Duration.Valid {
		trf.Duration = &p.PaymentDetail.Duration.Int64
	}

	if p.PaymentDetail.PickerAvatarUrl.Valid {
		trf.PickerAvatarUrl = &p.PaymentDetail.PickerAvatarUrl.String
	}

	if p.PaymentDetail.AppointmentTime.Valid {
		trf.StartTime = &p.PaymentDetail.AppointmentTime.Time
	}

	if p.PaymentDetail.Price.Valid {
		trf.Price = &p.PaymentDetail.Price.Float64
	}

	if p.PaymentDetail.Currency.Valid {
		trf.Currency = p.PaymentDetail.Currency.String
	}

	return trf
}

func TrfServiceDetail(srv models.Service, matchingFee float64) (interface{}, error) {
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
		endTime = &srv.EndTime.Time
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
		MatchingFee   float64    `json:"matching_fee"`
		CreatedAt     time.Time  `json:"created_at"`
	}{
		srv.Uuid.String,
		floatPrice,
		srv.Duration.Int32,
		srv.AppointmentTime.Time,
		srv.ServiceType,
		srv.ServiceStatus.ToString(),
		srv.Address.String,
		startTime,
		endTime,
		matchingFee,
		srv.CreatedAt,
	}, nil
}

type TransformedRate struct {
	Comment   string    `json:"comment"`
	Rating    int32     `json:"rating"`
	CreatedAt time.Time `json:"created_at"`
}

func TransformRate(srvRating *models.ServiceRating) *TransformedRate {
	return &TransformedRate{
		Comment:   srvRating.Comments.String,
		Rating:    srvRating.Rating.Int32,
		CreatedAt: srvRating.CreatedAt,
	}
}
