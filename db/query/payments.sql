-- name: CreatePayment :one
INSERT INTO payments (
	payer_id,
	service_id,
	price
) VALUES ($1, $2, $3)
RETURNING *;
