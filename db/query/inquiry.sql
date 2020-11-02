-- name: GetInquiryByInquirerID :one
SELECT * FROM service_inquiries
WHERE inquirer_id = $1
AND inquiry_status = $2;

-- name: GetInquiryByUuid :one
SELECT * FROM service_inquiries
WHERE uuid = $1;

-- name: CreateInquiry :one
INSERT INTO service_inquiries(
	uuid,
	inquirer_id,
	picker_id,	
	budget,
	service_type,
	inquiry_status,
	price,
	duration,
	appointment_time,
	lng,
	lat,
	expired_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
RETURNING *;

-- name: PatchInquiryStatus :exec
UPDATE service_inquiries
SET inquiry_status = $1
WHERE id = $2;

-- name: PatchInquiryStatusByUuid :one
UPDATE service_inquiries
SET inquiry_status = $1
WHERE uuid = $2
RETURNING *;

-- name: CheckUserOwnsInquiry :exec
SELECT EXISTS (
	SELECT 1
	FROM service_inquiries
	JOIN users ON service_inquiries.inquirer_id = users.id
	WHERE users.uuid = $1
	AND service_inquiries.uuid = $2
) as exists;

-- name: UpdateInquiryByUuid :one
UPDATE  service_inquiries
SET price = $1, duration = $2, appointment_time = $3, lng = $4, lat = $5, inquiry_status = $6
WHERE uuid = $7
RETURNING *;
