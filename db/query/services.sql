-- name: CreateService :one
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
RETURNING *;
