// Code generated by sqlc. DO NOT EDIT.
// source: services.sql

package models

import (
	"context"
	"database/sql"
)

const createService = `-- name: CreateService :one
INSERT INTO services(
	customer_id,
	service_provider_id,
	inquiry_id,
	service_status,
	budget,
	price,
	duration,
	appointment_time,
	lng,
	lat,
	service_type,
	girl_ready,
	man_ready
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
RETURNING id, uuid, customer_id, service_provider_id, price, duration, appointment_time, lng, lat, service_type, girl_ready, man_ready, created_at, updated_at, deleted_at, budget, inquiry_id, service_status
`

type CreateServiceParams struct {
	CustomerID        sql.NullInt32  `json:"customer_id"`
	ServiceProviderID sql.NullInt32  `json:"service_provider_id"`
	InquiryID         int32          `json:"inquiry_id"`
	ServiceStatus     ServiceStatus  `json:"service_status"`
	Budget            sql.NullString `json:"budget"`
	Price             sql.NullString `json:"price"`
	Duration          sql.NullInt32  `json:"duration"`
	AppointmentTime   sql.NullTime   `json:"appointment_time"`
	Lng               sql.NullString `json:"lng"`
	Lat               sql.NullString `json:"lat"`
	ServiceType       ServiceType    `json:"service_type"`
	GirlReady         sql.NullBool   `json:"girl_ready"`
	ManReady          sql.NullBool   `json:"man_ready"`
}

func (q *Queries) CreateService(ctx context.Context, arg CreateServiceParams) (Service, error) {
	row := q.queryRow(ctx, q.createServiceStmt, createService,
		arg.CustomerID,
		arg.ServiceProviderID,
		arg.InquiryID,
		arg.ServiceStatus,
		arg.Budget,
		arg.Price,
		arg.Duration,
		arg.AppointmentTime,
		arg.Lng,
		arg.Lat,
		arg.ServiceType,
		arg.GirlReady,
		arg.ManReady,
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
		&i.Lng,
		&i.Lat,
		&i.ServiceType,
		&i.GirlReady,
		&i.ManReady,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.DeletedAt,
		&i.Budget,
		&i.InquiryID,
		&i.ServiceStatus,
	)
	return i, err
}
