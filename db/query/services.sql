-- name: CreateService :one
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
RETURNING *;
