-- name: CreatePayment :one
INSERT INTO payments (
	payer_id,
	payee_id,
	service_id,
	price,
	rec_trade_id
) VALUES ($1, $2, $3, $4, $5)
RETURNING *;
