-- name: CreateChatroom :one

INSERT INTO chatrooms(
    inquiry_id,
    channel_uuid,
    message_count
) VALUES ($1, $2, $3)
RETURNING *;
