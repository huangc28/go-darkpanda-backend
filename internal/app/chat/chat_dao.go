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

func (dao *ChatDao) WithTx(tx *sqlx.Tx) contracts.ChatDaoer {
	dao.DB = tx

	return dao
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

func (dao *ChatDao) LeaveChat(chatID int64, userIDs ...int64) error {
	baseQuery := `
UPDATE chatroom_users SET deleted_at = now()	
WHERE user_id IN (%s) 
AND chatroom_id = $1;
	`
	idStr := ""
	for _, id := range userIDs {
		idStr += fmt.Sprintf("%d,", id)
	}

	idStr = strings.TrimSuffix(idStr, ",")
	query := fmt.Sprintf(baseQuery, idStr)

	if _, err := dao.DB.Exec(query, chatID); err != nil {
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

	if err := dao.DB.QueryRowx(
		fmt.Sprintf(query, db.ComposeFieldsSQLString(fields...)),
		channelUuid,
	).StructScan(&chatroom); err != nil {
		return (*models.Chatroom)(nil), err
	}

	return &chatroom, nil
}

func (dao *ChatDao) GetChatRoomByInquiryID(inquiryID int64, fields ...string) (*models.Chatroom, error) {
	query := `
SELECT %s
FROM chatrooms
WHERE inquiry_id = $1;
	`

	var chatroom models.Chatroom

	if err := dao.DB.QueryRowx(
		fmt.Sprintf(query, db.ComposeFieldsSQLString(fields...)),
		inquiryID,
	).StructScan(&chatroom); err != nil {
		return (*models.Chatroom)(nil), err
	}

	return &chatroom, nil
}

func (dao *ChatDao) DeleteChatRoom(ID int64) error {
	sql := `
UPDATE chatrooms 
SET deleted_at = now()
WHERE id = $1;
	`
	if _, err := dao.DB.Exec(sql, ID); err != nil {
		return err
	}

	return nil
}

func (dao *ChatDao) LeaveAllMemebers(chatroomID int64) ([]models.User, error) {
	sql := `
UPDATE chatroom_users 	
SET deleted_at = now()
WHERE chatroom_id = $1 
RETURNING user_id
	`
	var ids []int
	rows, err := dao.DB.Query(sql, chatroomID)

	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}

		ids = append(ids, id)
	}

	// log.Printf("DEBUG & 2 %v ", users)
	return dao.GetUserUUIDsByIDs(ids...)
}

func (dao *ChatDao) GetUserUUIDsByIDs(IDs ...int) ([]models.User, error) {
	baseQuery := `
SELECT 
	id,
	uuid 
FROM 
	users	
WHERE
	id IN (%s)
	`
	idStr := ""

	for _, id := range IDs {
		idStr += fmt.Sprintf("%d,", id)
	}

	idStr = strings.TrimSuffix(idStr, ",")
	query := fmt.Sprintf(baseQuery, idStr)

	rows, err := dao.DB.Query(query)

	if err != nil {
		return nil, err
	}

	var users []models.User
	for rows.Next() {
		var user models.User

		if err := rows.Scan(&user.ID, &user.Uuid); err != nil {
			return nil, err
		}

		users = append(users, user)
	}

	return users, nil
}

// func (dao *ChatDao) RevertChatByInquiryUuid(inquiryUuid string)
