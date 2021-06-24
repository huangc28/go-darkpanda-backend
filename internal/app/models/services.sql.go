// Code generated by sqlc. DO NOT EDIT.
// source: services.sql

package models

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
)

const createService = `-- name: CreateService :one
INSERT INTO services(
	uuid,
	customer_id,
	service_provider_id,
	price,
	duration,
	appointment_time,
	inquiry_id,
	service_status,
	budget,
	service_type,
	address
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
RETURNING id, uuid, customer_id, service_provider_id, price, duration, appointment_time, service_type, created_at, updated_at, deleted_at, budget, inquiry_id, service_status, address, start_time, end_time
`

type CreateServiceParams struct {
	Uuid              uuid.UUID      `json:"uuid"`
	CustomerID        sql.NullInt32  `json:"customer_id"`
	ServiceProviderID sql.NullInt32  `json:"service_provider_id"`
	Price             sql.NullString `json:"price"`
	Duration          sql.NullInt32  `json:"duration"`
	AppointmentTime   sql.NullTime   `json:"appointment_time"`
	InquiryID         int32          `json:"inquiry_id"`
	ServiceStatus     ServiceStatus  `json:"service_status"`
	Budget            sql.NullString `json:"budget"`
	ServiceType       ServiceType    `json:"service_type"`
	Address           sql.NullString `json:"address"`
}

func (q *Queries) CreateService(ctx context.Context, arg CreateServiceParams) (Service, error) {
	row := q.queryRow(ctx, q.createServiceStmt, createService,
		arg.Uuid,
		arg.CustomerID,
		arg.ServiceProviderID,
		arg.Price,
		arg.Duration,
		arg.AppointmentTime,
		arg.InquiryID,
		arg.ServiceStatus,
		arg.Budget,
		arg.ServiceType,
		arg.Address,
	)
	var i Service
	err := row.Scan(
		&i.ID,
		&i.Uuid,
		&i.CustomerID,
		&i.ServiceProviderID,
		&i.Price,
		&i.Duration,
		&i.AppointmentTime,
		&i.ServiceType,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.DeletedAt,
		&i.Budget,
		&i.InquiryID,
		&i.ServiceStatus,
		&i.Address,
		&i.StartTime,
		&i.EndTime,
	)
	return i, err
}
