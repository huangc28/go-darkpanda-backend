package contracts

import (
	"time"

	"github.com/huangc28/go-darkpanda-backend/db"
	"github.com/huangc28/go-darkpanda-backend/internal/app/models"
)

type UpdateServiceByIDParams struct {
	ID            int64
	Price         *float64
	Duration      *int
	Appointment   *time.Time
	ServiceType   *models.ServiceType
	ServiceStatus *models.ServiceStatus
}

type ServiceDAOer interface {
	GetUserHistoricalServicesByUuid(uuid string, perPage int, offset int) ([]models.Service, error)
	GetServiceByInquiryUUID(uuid string, fields ...string) (*models.Service, error)
	UpdateServiceByID(params UpdateServiceByIDParams) (*models.Service, error)
	WithTx(tx db.Conn) ServiceDAOer
}
