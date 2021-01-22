-- name: GetReferCodeInfoByRefcode :one
SELECT * FROM user_refcodes
WHERE ref_code = $1 LIMIT 1;

-- name: CreateRefcode :one
INSERT INTO user_refcodes (
	invitor_id,
	invitee_id,
	ref_code,
	ref_code_type,
	expired_at
) VALUES (
	$1,
	$2,
	$3,
	$4,
	$5
)
RETURNING *;

-- name: UpdateInviteeIDByRefCode :exec
UPDATE user_refcodes
SET
	invitee_id = $1
WHERE
	ref_code = $2;


