package contracts

import (
	"time"

	"github.com/huangc28/go-darkpanda-backend/db"
	"github.com/huangc28/go-darkpanda-backend/internal/app/models"
	"github.com/jmoiron/sqlx"
)

type ChatServicer interface {
	CreateAndJoinChatroom(inquiryID int64, userIDs ...int64) (*models.Chatroom, error)
	WithTx(tx *sqlx.Tx) ChatServicer
}

type UpdateChatByUuidParams struct {
	MessageCount *int
	ExpiredAt    *time.Time
	ChatroomType models.ChatroomType
	ChannelUuid  string
}

type GetFemaleInquiryChatrooms struct {
	UserID      int64
	InquiryUUID string
	Offset      int64
	PerPage     int64
}

type GetMaleInquiryChatrooms struct {
	UserID      int64
	InquiryUUID string
	Offset      int64
	PerPage     int64
}

type ChatDaoer interface {
	WithTx(tx *sqlx.Tx) ChatDaoer
	WithConn(conn db.Conn) ChatDaoer
	CreateChat(inquiryID int64) (*models.Chatroom, error)
	JoinChat(chatID int64, userIDs ...int64) error
	LeaveChat(chatID int64, userIDs ...int64) error
	LeaveAllMemebers(chatroomID int64) ([]models.User, error)
	GetChatRoomByChannelUUID(chanelUUID string, fields ...string) (*models.Chatroom, error)
	GetChatRoomByInquiryID(inquiryID int64, fields ...string) (*models.Chatroom, error)
	DeleteChatRoom(ID int64) error

	GetFemaleInquiryChatrooms(GetFemaleInquiryChatrooms) ([]models.InquiryChatRoom, error)
	GetMaleInquiryChatrooms(GetMaleInquiryChatrooms) ([]models.InquiryChatRoom, error)

	UpdateChatByUuid(params UpdateChatByUuidParams) (*models.Chatroom, error)
	IsUserInChatroom(userUuid string, chatroomUuid string) (bool, error)
	GetInquiryByChannelUuid(channelUuid string) (*models.ServiceInquiry, error)
	GetCompleteChatroomInfoById(id int) (*models.CompleteChatroomInfoModel, error)
	GetChatroomByServiceId(srvId int) (*models.Chatroom, error)
	DeleteChatroomByServiceId(srvId int) error
	DeleteChatroomByInquiryId(iqId int) error
}
