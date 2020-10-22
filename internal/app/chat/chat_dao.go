package chat

import (
	"fmt"

	"github.com/huangc28/go-darkpanda-backend/db"
	"github.com/huangc28/go-darkpanda-backend/internal/app/models"
)

type ChatDaoer interface {
	GetChatRoomByChannelID(channelUuid string, fields ...string) (*models.Chatroom, error)
}

type ChatDao struct {
	DB db.Conn
}

func (dao *ChatDao) GetChatRoomByChannelID(channelUuid string, fields ...string) (*models.Chatroom, error) {
	query := `
SELECT %s
FROM chatrooms
WHERE channel_uuid = $1;
	`

	var chatroom models.Chatroom

	if err := dao.DB.QueryRow(
		fmt.Sprintf(query, db.ComposeFieldsSQLString(fields...)),
		channelUuid,
	).Scan(&chatroom); err != nil {
		return (*models.Chatroom)(nil), err
	}

	return &chatroom, nil
}
