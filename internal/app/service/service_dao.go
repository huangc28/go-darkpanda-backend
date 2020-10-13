package service

import (
	"database/sql"

	"github.com/huangc28/go-darkpanda-backend/internal/models"
)

type ServiceDAO struct {
	DB *sql.DB
}

func (dao *ServiceDAO) GetUserHistoricalServicesByUuid(uuid string, perPage int, offset int) ([]models.Service, error) {
	query := `
SELECT
	services.uuid,
	price,
	service_type,
	service_status,
	services.created_at
FROM services
INNER JOIN users ON users.id = customer_id
WHERE users.uuid = $1
ORDER BY created_at DESC
LIMIT $2
OFFSET $3;
	`
	rows, err := dao.DB.Query(query, uuid, perPage, offset)

	if err != nil {
		return nil, err
	}

	services := make([]models.Service, 0)

	for rows.Next() {
		service := models.Service{}

		rows.Scan(
			&service.Uuid,
			&service.Price,
			&service.ServiceType,
			&service.ServiceStatus,
			&service.CreatedAt,
		)

		services = append(services, service)
	}

	return services, nil
}
