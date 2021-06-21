package contracts

import (
	"time"

	"github.com/huangc28/go-darkpanda-backend/internal/app/models"
	"github.com/jmoiron/sqlx"
)

type ChatServicer interface {
	CreateAndJoinChatroom(inquiryID int64, userIDs ...int64) (*models.Chatroom, error)
	WithTx(tx *sqlx.Tx) ChatServicer
}

type UpdateChatByUuidParams struct {
	MessageCount *int
	Enabled      *bool
	ExpiredAt    *time.Time
	ChatroomType models.ChatroomType
	ChannelUuid  string
}

type ChatDaoer interface {
	WithTx(tx *sqlx.Tx) ChatDaoer
	CreateChat(inquiryID int64) (*models.Chatroom, error)
	JoinChat(chatID int64, userIDs ...int64) error
	LeaveChat(chatID int64, userIDs ...int64) error
	LeaveAllMemebers(chatroomID int64) ([]models.User, error)
	GetChatRoomByChannelUUID(chanelUUID string, fields ...string) (*models.Chatroom, error)
	GetChatRoomByInquiryID(inquiryID int64, fields ...string) (*models.Chatroom, error)
	DeleteChatRoom(ID int64) error
	GetFemaleInquiryChatRooms(userID int64) ([]models.InquiryChatRoom, error)
	UpdateChatByUuid(params UpdateChatByUuidParams) (*models.Chatroom, error)
	IsUserInChatroom(userUuid string, chatroomUuid string) (bool, error)
	GetInquiryByChannelUuid(channelUuid string) (*models.ServiceInquiry, error)
	GetCompleteChatroomInfoById(id int) (*models.CompleteChatroomInfoModel, error)
	GetChatroomByServiceId(srvId int) (*models.Chatroom, error)
}
