package contracts

import (
	"github.com/huangc28/go-darkpanda-backend/internal/app/models"
	"github.com/jmoiron/sqlx"
)

type ChatServicer interface {
	CreateAndJoinChatroom(inquiryID int64, userIDs ...int64) (*models.ChatInfo, error)
	WithTx(tx *sqlx.Tx) ChatServicer
}

type ChatDaoer interface {
	CreateChat(inquiryID int64) (*models.ChatInfo, error)
	JoinChat(chatID int64, userIDs ...int64) error
	GetChatRoomByChannelID(channelUuid string, fields ...string) (*models.Chatroom, error)
	WithTx(tx *sqlx.Tx)
}
