package contracts

import (
	"time"

	"github.com/huangc28/go-darkpanda-backend/db"
	"github.com/huangc28/go-darkpanda-backend/internal/app/models"
)

type UpdateServiceByIDParams struct {
	ID int64

	Price         *float64
	Duration      *int
	Appointment   *time.Time
	ServiceType   *models.ServiceType
	ServiceStatus *models.ServiceStatus
	StartTime     *time.Time
	EndTime       *time.Time

	CancellerId *int64
}

type CreateServiceQRCodeParams struct {
	Uuid      string
	Url       string
	ServiceId int
}

type GetOverlappedServicesParams struct {
	UserId                       int64
	InquiryAppointmentTime       time.Time
	AppointmentBufferDuration    int64
	BetweenServiceBufferDuration int64
}

type UpdateServiceByInquiryIdParams struct {
	InquiryId int64

	Price         *float64
	Duration      *int
	Appointment   *time.Time
	ServiceType   *models.ServiceType
	ServiceStatus *models.ServiceStatus
	StartTime     *time.Time
	EndTime       *time.Time

	CancellerId *int64
}

type ServiceDAOer interface {
	GetUserHistoricalServicesByUuid(uuid string, perPage int, offset int) ([]models.Service, error)
	GetServiceByInquiryUUID(uuid string, fields ...string) (*models.Service, error)
	UpdateServiceByID(UpdateServiceByIDParams) (*models.Service, error)
	UpdateServiceByInquiryId(UpdateServiceByInquiryIdParams) (*models.Service, error)
	CreateServiceQRCode(params CreateServiceQRCodeParams) (*models.ServiceQrcode, error)
	WithTx(tx db.Conn) ServiceDAOer
	ScanExpiredServices() ([]*models.Service, error)
	ScanCompletedServices() ([]*models.Service, error)
	GetServiceByUuid(srvUuid string, fields ...string) (*models.Service, error)
	GetOverlappedServices(p GetOverlappedServicesParams) ([]models.Service, error)
}
