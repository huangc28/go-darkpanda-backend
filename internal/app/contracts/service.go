package contracts

import (
	"time"

	"github.com/huangc28/go-darkpanda-backend/db"
	"github.com/huangc28/go-darkpanda-backend/internal/app/models"
)

type UpdateServiceByIDParams struct {
	ID int64

	Price         *float64
	MatchingFee   *float64
	Duration      *int
	Appointment   *time.Time
	ServiceType   *models.ServiceType
	ServiceStatus *models.ServiceStatus
	StartTime     *time.Time
	EndTime       *time.Time
	Address       *string

	CancellerId *int64
	CancelCause *models.CancelCause
}

type CreateServiceQRCodeParams struct {
	Uuid      string
	Url       string
	ServiceId int
}

type GetOverlappedServicesParams struct {
	ExcludeServiceUuid           string
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

type CreateUserServiceOptionParams struct {
	UserID          int
	ServiceOptionID int
}

type ServiceDAOer interface {
	GetUserHistoricalServicesByUuid(uuid string, perPage, offset int) ([]models.Service, error)
	GetServiceByInquiryUUID(string, ...string) (*models.Service, error)
	UpdateServiceByID(UpdateServiceByIDParams) (*models.Service, error)
	UpdateServiceByInquiryId(UpdateServiceByInquiryIdParams) (*models.Service, error)
	CreateServiceQRCode(CreateServiceQRCodeParams) (*models.ServiceQrcode, error)
	WithTx(db.Conn) ServiceDAOer
	ScanExpiredServices() ([]*models.ServiceScannerData, error)
	ScanCompletedServices() ([]*models.ServiceScannerData, error)
	ScanInquiringServiceInquiries() ([]*models.ServiceScannerData, error)
	GetServiceByUuid(string, ...string) (*models.Service, error)
	GetOverlappedServices(GetOverlappedServicesParams) ([]models.Service, error)
	GetInquiryByServiceUuid(srvUuid string) (*models.ServiceInquiry, error)
	GetServiceProviderByServiceUUID(srvUUID string) (*models.User, error)
	CancelUnpaidServicesIfExceed30Minuties() ([]*models.CancelUnpaidServices, error)
}
