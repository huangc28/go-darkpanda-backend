package service

import (
	"time"

	"github.com/huangc28/go-darkpanda-backend/config"
	"github.com/huangc28/go-darkpanda-backend/internal/app/models"
	darkfirestore "github.com/huangc28/go-darkpanda-backend/internal/app/pkg/dark_firestore"
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
