package chat

import (
	"fmt"
	"strings"
	"time"

	"github.com/golobby/container/pkg/container"
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

func ChatDaoServiceProvider(c container.Container) func() error {
	return func() error {
		c.Transient(func() contracts.ChatDaoer {
			return NewChatDao(db.GetDB())
		})

		return nil
	}
}

const (
	PrivateChatKey = "private_chat:%s"
)

func (dao *ChatDao) WithTx(tx *sqlx.Tx) contracts.ChatDaoer {
	dao.DB = tx

	return dao
}

func (dao *ChatDao) WithConn(conn db.Conn) contracts.ChatDaoer {
	dao.DB = conn

	return dao
}

func (dao *ChatDao) CreateChat(inquiryID int64) (*models.Chatroom, error) {
	// Create chatroom record.
	sid, err := shortid.Generate()

	if err != nil {
		return nil, err
	}

	channelUuid := fmt.Sprintf(PrivateChatKey, sid)
	messageCount := 0
	expiredAt := time.Now().Add(time.Minute * 27)

	var chatroom models.Chatroom

	query := `
INSERT INTO chatrooms (
	inquiry_id,
	channel_uuid,
	message_count,
	expired_at
) VALUES ($1, $2, $3, $4)
RETURNING *;
	`

	if err := dao.
		DB.QueryRowx(query, inquiryID, channelUuid, messageCount, expiredAt).StructScan(&chatroom); err != nil {
		return nil, err
	}

	return &chatroom, nil
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
UPDATE 
	chatroom_users 
SET 
	deleted_at = now()
WHERE 
	user_id IN (%s)
AND 
	chatroom_id = $1;
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

func (dao *ChatDao) GetChatRoomByChannelUUID(channelUuid string, fields ...string) (*models.Chatroom, error) {
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
UPDATE
	chatrooms
SET
	deleted_at = now()
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

// GetFemaleInquiryChatRooms retrieve all inquiry chatrooms of a female user.
//   - Those chatrooms's related inquiry status is chatting
//   - Those chatrooms's related inquiry picker_id equals requester's id
// Inquiry title --- not exists yet
// Inquiry type
// inquirer name
// inquirer avatar
// chatroom channel uuid
// chatroom created_at
// @TODOs Add pagination.
func (dao *ChatDao) GetFemaleInquiryChatRooms(userID int64) ([]models.InquiryChatRoom, error) {
	query := `
SELECT
	si.service_type,
	si.inquiry_status,
	si.uuid AS inquiry_uuid,
	inquirer.username,
	inquirer.uuid AS inquirer_uuid,
	inquirer.avatar_url,
	chatrooms.channel_uuid,
	chatrooms.expired_at,
	chatrooms.created_at
FROM service_inquiries	AS si
INNER JOIN chatrooms
	ON chatrooms.inquiry_id = si.id
	AND chatrooms.deleted_at IS NULL
INNER JOIN users AS inquirer
	ON inquirer.id = si.inquirer_id
WHERE
	si.inquiry_status = $1 OR
	si.inquiry_status = $2
AND picker_id = $3
	`

	rows, err := dao.DB.Queryx(
		query,
		models.InquiryStatusChatting,
		models.InquiryStatusWaitForInquirerApprove,
		userID,
	)

	if err != nil {
		return nil, err
	}

	chatrooms := make([]models.InquiryChatRoom, 0)
	for rows.Next() {
		cr := models.InquiryChatRoom{}
		if err := rows.StructScan(&cr); err != nil {
			return nil, err
		}

		chatrooms = append(chatrooms, cr)
	}

	return chatrooms, nil
}

func (dao *ChatDao) UpdateChatByUuid(params contracts.UpdateChatByUuidParams) (*models.Chatroom, error) {
	query := `
UPDATE chatrooms SET
	message_count = COALESCE($1, message_count),
	expired_at = COALESCE($3, expired_at),
	chatroom_type = COALESCE($4, chatroom_type)
WHERE channel_uuid = $5
RETURNING *;
`
	var chatroom models.Chatroom

	if err := dao.
		DB.
		QueryRowx(
			query,
			params.MessageCount,
			params.ExpiredAt,
			params.ChatroomType,
			params.ChannelUuid,
		).StructScan(&chatroom); err != nil {

		return nil, err
	}

	return &chatroom, nil
}

// IsUserInTheChatroom checks if a given user is in the chatroom.
func (dao *ChatDao) IsUserInChatroom(userUuid string, chatroomUuid string) (bool, error) {
	query := `
	SELECT EXISTS(
		SELECT
			1
		FROM
			chatroom_users
		INNER JOIN users AS u
			ON u.id = chatroom_users.user_id
		INNER JOIN chatrooms AS c
			ON c.id = chatroom_users.chatroom_id
		WHERE
			u.uuid = $1 AND
			c.channel_uuid = $2
	);

`
	var exists bool
	if err := dao.DB.QueryRow(query, userUuid, chatroomUuid).Scan(&exists); err != nil {
		return exists, err
	}

	return exists, nil

}

func (dao *ChatDao) GetInquiryByChannelUuid(channelUuid string) (*models.ServiceInquiry, error) {
	query := `
SELECT
	service_inquiries.*
FROM
	service_inquiries
INNER JOIN chatrooms AS c
	ON service_inquiries.id = c.inquiry_id
WHERE c.channel_uuid = $1;
	`
	var iqModel models.ServiceInquiry

	if err := dao.DB.QueryRowx(query, channelUuid).StructScan(&iqModel); err != nil {
		return nil, err
	}

	return &iqModel, nil
}

// GetIntactChatroomById gets more complete information about the chatroom.
func (dao *ChatDao) GetCompleteChatroomInfoById(id int) (*models.CompleteChatroomInfoModel, error) {
	query := `
SELECT
	si.service_type,
	si.inquiry_status,
	si.uuid AS inquiry_uuid,
	si.inquirer_id,
	si.picker_id,
	chatrooms.*
FROM chatrooms
INNER JOIN service_inquiries AS si ON si.id = chatrooms.inquiry_id
WHERE chatrooms.id = $1;
`

	var m models.CompleteChatroomInfoModel

	if err := dao.DB.QueryRowx(query, id).StructScan(&m); err != nil {
		return nil, err
	}

	return &m, nil
}

func (dao *ChatDao) GetChatroomByServiceId(srvId int) (*models.Chatroom, error) {
	query := `
WITH related_inquiry AS (
	SELECT
		service_inquiries.id
	FROM
		service_inquiries
	INNER JOIN
		services ON services.inquiry_id = service_inquiries.id AND
		services.id = $1
)

SELECT
	*
FROM
	chatrooms
WHERE
	inquiry_id IN (
		SELECT
			id
		FROM
			related_inquiry
	);
`

	var m models.Chatroom

	if err := dao.DB.QueryRowx(
		query,
		srvId,
	).StructScan(&m); err != nil {
		return nil, err
	}

	return &m, nil
}

func (dao *ChatDao) DeleteChatroomByServiceId(srvId int) error {
	query := `
WITH related_inquiry AS (
	SELECT
		service_inquiries.id
	FROM
		service_inquiries
	INNER JOIN
		services ON services.inquiry_id = service_inquiries.id AND
		services.id = $1
)

UPDATE 
	chatrooms 
SET
	deleted_at = now()
WHERE 
	inquiry_id IN (
		SELECT id	
		FROM related_inquiry
	);
`

	_, err := dao.DB.Exec(query, srvId)

	return err
}
