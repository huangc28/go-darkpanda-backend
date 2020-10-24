package chat

import (
	"time"

	"github.com/huangc28/go-darkpanda-backend/internal/app/contracts"
	"github.com/huangc28/go-darkpanda-backend/internal/app/models"
	"github.com/jmoiron/sqlx"
)

const MaxMassageCount = 200

func IsExceedMaxMessageCount(count int) bool {
	return count+1 > MaxMassageCount
}

func IsChatroomExpired(expT time.Time) bool {
	return expT.Before(time.Now())
}

type ChatServices struct {
	ChatDao contracts.ChatDaoer
}

func NewChatServices(chatDao contracts.ChatDaoer) contracts.ChatServicer {
	return &ChatServices{
		ChatDao: chatDao,
	}
}

func (cs *ChatServices) WithTx(tx *sqlx.Tx) contracts.ChatServicer {
	cs.ChatDao.WithTx(tx)

	return cs
}

func (cs *ChatServices) CreateAndJoinChatroom(inquiryID int64, userIDs ...int64) (*models.ChatInfo, error) {
	// Create chatroom
	chatInfo, err := cs.ChatDao.CreateChat(inquiryID)

	if err != nil {
		return (*models.ChatInfo)(nil), err
	}

	// Join chatroom
	if err := cs.ChatDao.JoinChat(chatInfo.ChatID, userIDs...); err != nil {
		return (*models.ChatInfo)(nil), err
	}

	return &models.ChatInfo{
		ChanelUuid: chatInfo.ChanelUuid,
		ChatID:     chatInfo.ChatID,
	}, nil
}