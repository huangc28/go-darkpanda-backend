package service

import "time"

type TransformedGetIncomingService struct {
	ServiceUuid     string    `json:"service_uuid"`
	ServiceStatus   string    `json:"service_status"`
	AppointmentTime time.Time `json:"appointment_time"`
	Username        string    `json:"username"`
	UserUuid        string    `json:"user_uuid"`
	AvatarUrl       string    `json:"avatar_url"`
	ChannelUuid     string    `json:"channel_uuid"`
}

func TransformGetIncomingServices(results []IncomingServiceResult) []TransformedGetIncomingService {
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
		}

		trfRes = append(trfRes, c)
	}

	return trfRes
}
