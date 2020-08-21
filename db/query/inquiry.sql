-- name: GetInquiryByInquirerID :one
SELECT * FROM service_inquiries
WHERE inquirer_id = $1
AND inquiry_status = $2;

-- name: CreateInquiry :one
INSERT INTO service_inquiries(
	uuid,
	inquirer_id,
	budget,
	service_type,
	inquiry_status
) VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: PatchInquiryStatus :exec
UPDATE service_inquiries
SET inquiry_status = $1
WHERE id = $2;

