// Code generated by sqlc. DO NOT EDIT.
// source: inquiry.sql

package models

import (
	"context"
	"database/sql"
)

const checkUserOwnsInquiry = `-- name: CheckUserOwnsInquiry :exec
SELECT EXISTS (
	SELECT 1
	FROM service_inquiries
	JOIN users ON service_inquiries.inquirer_id = users.id
	WHERE users.uuid = $1
	AND service_inquiries.uuid = $2
) as exists
`

type CheckUserOwnsInquiryParams struct {
	Uuid   string `json:"uuid"`
	Uuid_2 string `json:"uuid_2"`
}

func (q *Queries) CheckUserOwnsInquiry(ctx context.Context, arg CheckUserOwnsInquiryParams) error {
	_, err := q.exec(ctx, q.checkUserOwnsInquiryStmt, checkUserOwnsInquiry, arg.Uuid, arg.Uuid_2)
	return err
}

const createInquiry = `-- name: CreateInquiry :one
INSERT INTO service_inquiries(
	uuid,
	inquirer_id,
	budget,
	service_type,
	inquiry_status,
	price,
	duration,
	appointment_time,
	lng,
	lat
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING id, inquirer_id, budget, service_type, inquiry_status, created_at, updated_at, deleted_at, uuid, price, duration, appointment_time, lng, lat
`

type CreateInquiryParams struct {
	Uuid            string         `json:"uuid"`
	InquirerID      sql.NullInt32  `json:"inquirer_id"`
	Budget          string         `json:"budget"`
	ServiceType     ServiceType    `json:"service_type"`
	InquiryStatus   InquiryStatus  `json:"inquiry_status"`
	Price           sql.NullString `json:"price"`
	Duration        sql.NullInt32  `json:"duration"`
	AppointmentTime sql.NullTime   `json:"appointment_time"`
	Lng             sql.NullString `json:"lng"`
	Lat             sql.NullString `json:"lat"`
}

func (q *Queries) CreateInquiry(ctx context.Context, arg CreateInquiryParams) (ServiceInquiry, error) {
	row := q.queryRow(ctx, q.createInquiryStmt, createInquiry,
		arg.Uuid,
		arg.InquirerID,
		arg.Budget,
		arg.ServiceType,
		arg.InquiryStatus,
		arg.Price,
		arg.Duration,
		arg.AppointmentTime,
		arg.Lng,
		arg.Lat,
	)
	var i ServiceInquiry
	err := row.Scan(
		&i.ID,
		&i.InquirerID,
		&i.Budget,
		&i.ServiceType,
		&i.InquiryStatus,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.DeletedAt,
		&i.Uuid,
		&i.Price,
		&i.Duration,
		&i.AppointmentTime,
		&i.Lng,
		&i.Lat,
	)
	return i, err
}

const getInquiryByInquirerID = `-- name: GetInquiryByInquirerID :one
SELECT id, inquirer_id, budget, service_type, inquiry_status, created_at, updated_at, deleted_at, uuid, price, duration, appointment_time, lng, lat FROM service_inquiries
WHERE inquirer_id = $1
AND inquiry_status = $2
`

type GetInquiryByInquirerIDParams struct {
	InquirerID    sql.NullInt32 `json:"inquirer_id"`
	InquiryStatus InquiryStatus `json:"inquiry_status"`
}

func (q *Queries) GetInquiryByInquirerID(ctx context.Context, arg GetInquiryByInquirerIDParams) (ServiceInquiry, error) {
	row := q.queryRow(ctx, q.getInquiryByInquirerIDStmt, getInquiryByInquirerID, arg.InquirerID, arg.InquiryStatus)
	var i ServiceInquiry
	err := row.Scan(
		&i.ID,
		&i.InquirerID,
		&i.Budget,
		&i.ServiceType,
		&i.InquiryStatus,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.DeletedAt,
		&i.Uuid,
		&i.Price,
		&i.Duration,
		&i.AppointmentTime,
		&i.Lng,
		&i.Lat,
	)
	return i, err
}

const getInquiryByUuid = `-- name: GetInquiryByUuid :one
SELECT id, inquirer_id, budget, service_type, inquiry_status, created_at, updated_at, deleted_at, uuid, price, duration, appointment_time, lng, lat FROM service_inquiries
WHERE uuid = $1
`

func (q *Queries) GetInquiryByUuid(ctx context.Context, uuid string) (ServiceInquiry, error) {
	row := q.queryRow(ctx, q.getInquiryByUuidStmt, getInquiryByUuid, uuid)
	var i ServiceInquiry
	err := row.Scan(
		&i.ID,
		&i.InquirerID,
		&i.Budget,
		&i.ServiceType,
		&i.InquiryStatus,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.DeletedAt,
		&i.Uuid,
		&i.Price,
		&i.Duration,
		&i.AppointmentTime,
		&i.Lng,
		&i.Lat,
	)
	return i, err
}

const patchInquiryStatus = `-- name: PatchInquiryStatus :exec
UPDATE service_inquiries
SET inquiry_status = $1
WHERE id = $2
`

type PatchInquiryStatusParams struct {
	InquiryStatus InquiryStatus `json:"inquiry_status"`
	ID            int64         `json:"id"`
}

func (q *Queries) PatchInquiryStatus(ctx context.Context, arg PatchInquiryStatusParams) error {
	_, err := q.exec(ctx, q.patchInquiryStatusStmt, patchInquiryStatus, arg.InquiryStatus, arg.ID)
	return err
}

const patchInquiryStatusByUuid = `-- name: PatchInquiryStatusByUuid :one
UPDATE service_inquiries
SET inquiry_status = $1
WHERE uuid = $2
RETURNING id, inquirer_id, budget, service_type, inquiry_status, created_at, updated_at, deleted_at, uuid, price, duration, appointment_time, lng, lat
`

type PatchInquiryStatusByUuidParams struct {
	InquiryStatus InquiryStatus `json:"inquiry_status"`
	Uuid          string        `json:"uuid"`
}

func (q *Queries) PatchInquiryStatusByUuid(ctx context.Context, arg PatchInquiryStatusByUuidParams) (ServiceInquiry, error) {
	row := q.queryRow(ctx, q.patchInquiryStatusByUuidStmt, patchInquiryStatusByUuid, arg.InquiryStatus, arg.Uuid)
	var i ServiceInquiry
	err := row.Scan(
		&i.ID,
		&i.InquirerID,
		&i.Budget,
		&i.ServiceType,
		&i.InquiryStatus,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.DeletedAt,
		&i.Uuid,
		&i.Price,
		&i.Duration,
		&i.AppointmentTime,
		&i.Lng,
		&i.Lat,
	)
	return i, err
}

const updateInquiryByUuid = `-- name: UpdateInquiryByUuid :one
UPDATE  service_inquiries
SET price = $1, duration = $2, appointment_time = $3, lng = $4, lat = $5, inquiry_status = $6
WHERE uuid = $7
RETURNING id, inquirer_id, budget, service_type, inquiry_status, created_at, updated_at, deleted_at, uuid, price, duration, appointment_time, lng, lat
`

type UpdateInquiryByUuidParams struct {
	Price           sql.NullString `json:"price"`
	Duration        sql.NullInt32  `json:"duration"`
	AppointmentTime sql.NullTime   `json:"appointment_time"`
	Lng             sql.NullString `json:"lng"`
	Lat             sql.NullString `json:"lat"`
	InquiryStatus   InquiryStatus  `json:"inquiry_status"`
	Uuid            string         `json:"uuid"`
}

func (q *Queries) UpdateInquiryByUuid(ctx context.Context, arg UpdateInquiryByUuidParams) (ServiceInquiry, error) {
	row := q.queryRow(ctx, q.updateInquiryByUuidStmt, updateInquiryByUuid,
		arg.Price,
		arg.Duration,
		arg.AppointmentTime,
		arg.Lng,
		arg.Lat,
		arg.InquiryStatus,
		arg.Uuid,
	)
	var i ServiceInquiry
	err := row.Scan(
		&i.ID,
		&i.InquirerID,
		&i.Budget,
		&i.ServiceType,
		&i.InquiryStatus,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.DeletedAt,
		&i.Uuid,
		&i.Price,
		&i.Duration,
		&i.AppointmentTime,
		&i.Lng,
		&i.Lat,
	)
	return i, err
}