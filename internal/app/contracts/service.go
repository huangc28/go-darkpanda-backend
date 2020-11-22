package contracts

import (
	"time"

	"github.com/huangc28/go-darkpanda-backend/internal/app/models"
)

type UpdateServiceByIDParams struct {
	ID          int64
	Price       *float64
	Duration    *int
	Appointment *time.Time
	ServiceType *models.ServiceType
}

type ServiceDAOer interface {
	GetUserHistoricalServicesByUuid(uuid string, perPage int, offset int) ([]models.Service, error)
	GetServiceByInquiryUUID(uuid string) (*models.Service, error)
	UpdateServiceByID(params UpdateServiceByIDParams) (*models.Service, error)
}
