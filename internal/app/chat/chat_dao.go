package chat

import (
	"fmt"
	"strings"
	"time"

	"github.com/huangc28/go-darkpanda-backend/db"
	"github.com/huangc28/go-darkpanda-backend/internal/app/contracts"
	"github.com/huangc28/go-darkpanda-backend/internal/app/models"
	"github.com/jmoiron/sqlx"
	"github.com/teris-io/shortid"
)

type ChatDao struct {
	DB db.Conn
}

func NewChatDao(db db.Conn) contracts.ChatDaoer {
	return &ChatDao{
		DB: db,
	}
}

const (
	PrivateChatKey = "private_chat:%s"
)

func (dao *ChatDao) WithTx(tx *sqlx.Tx) {
	dao.DB = tx
}

func (dao *ChatDao) CreateChat(inquiryID int64) (*models.ChatInfo, error) {
	// Create chatroom record.
	sid, err := shortid.Generate()

	if err != nil {
		return nil, err
	}

	channelUuid := fmt.Sprintf(PrivateChatKey, sid)
	messageCount := 0
	enabled := true
	expiredAt := time.Now().Add(time.Minute * 27)

	var id int64

	query := `
INSERT INTO chatrooms (
	inquiry_id,
	channel_uuid,
	message_count,
	enabled,
	expired_at
) VALUES ($1, $2, $3, $4, $5)
RETURNING id;
	`

	if err := dao.
		DB.QueryRow(query, inquiryID, channelUuid, messageCount, enabled, expiredAt).Scan(&id); err != nil {
		return nil, err
	}

	return &models.ChatInfo{
		ChanelUuid: channelUuid,
		ChatID:     id,
	}, nil
}

func (dao *ChatDao) JoinChat(chatID int64, userIDs ...int64) error {
	// Join multiple users to chat
	sqlStr := `
INSERT INTO chatroom_users (
	chatroom_id,
	user_id
) VALUES
`
	vals := []interface{}{}

	for _, userID := range userIDs {
		sqlStr += " (?, ?),"
		vals = append(vals, chatID, userID)

	}

	sqlStr = strings.TrimSuffix(sqlStr, ",")
	pgStr := db.ReplaceSQLPlaceHolderWithPG(sqlStr, "?")

	stmt, err := dao.DB.Prepare(pgStr)

	if err != nil {
		return err
	}

	if _, err := stmt.Exec(vals...); err != nil {
		return err
	}

	return nil
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
