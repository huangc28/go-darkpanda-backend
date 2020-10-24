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
	LeaveChat(chatID int64, userIDs ...int64) error
	LeaveAllMemebers(chatroomID int64) ([]models.User, error)
	GetChatRoomByChannelID(chanelUUID string, fields ...string) (*models.Chatroom, error)
	GetChatRoomByInquiryID(inquiryID int64, fields ...string) (*models.Chatroom, error)
	DeleteChatRoom(ID int64) error
	WithTx(tx *sqlx.Tx) ChatDaoer
}
