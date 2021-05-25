package service

import (
	"time"

	"github.com/huangc28/go-darkpanda-backend/config"
	"github.com/huangc28/go-darkpanda-backend/internal/app/models"
)

type TransformedGetIncomingService struct {
	ServiceUuid     string    `json:"service_uuid"`
	ServiceStatus   string    `json:"service_status"`
	AppointmentTime time.Time `json:"appointment_time"`
	Username        string    `json:"username"`
	UserUuid        string    `json:"user_uuid"`
	AvatarUrl       string    `json:"avatar_url"`
	ChannelUuid     string    `json:"channel_uuid"`
	InquiryUuid     string    `json:"inquiry_uuid"`
}

type TransformedGetIncomingServices struct {
	Services []TransformedGetIncomingService `json:"services"`
}

func TransformGetServicesResults(results []ServiceResult) TransformedGetIncomingServices {
	trfRes := make([]TransformedGetIncomingService, 0)

	for _, res := range results {
		c := TransformedGetIncomingService{
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

	return TransformedGetIncomingServices{
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
