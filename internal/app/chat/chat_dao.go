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

	defer rows.Close()

	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}

		ids = append(ids, id)
	}

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

	defer rows.Close()

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

// GetFemaleInquiryChatRooms retrieve all inquiry chatrooms for a female user.
//   - Those chatrooms's related inquiry status is chatting, wait_for_inquirer_approve
//   - Those chatrooms's related inquiry picker_id equals requester's id
func (dao *ChatDao) GetFemaleInquiryChatrooms(params contracts.GetFemaleInquiryChatrooms) ([]models.InquiryChatRoom, error) {
	// Retrieve all matching inquiry chatrooms
	inquiryUUIDQueryClause := "1 = 1"

	// Retrieve one inquiry chatroom that matches uuid = params.InquiryUUID
	if len(params.InquiryUUID) > 0 {
		inquiryUUIDQueryClause = fmt.Sprintf("si.uuid = '%s'", params.InquiryUUID)
	}

	query := fmt.Sprintf(`
SELECT
	inquirer.username,
	inquirer.uuid AS inquirer_uuid,
	inquirer.avatar_url,
	si.expect_service_type AS service_type,
	si.inquiry_status,
	si.uuid AS inquiry_uuid,
	inquirer.uuid AS inquirer_uuid,
	chatrooms.channel_uuid,
	chatrooms.expired_at,
	chatrooms.created_at,
	services.uuid AS service_uuid
FROM service_inquiries AS si
INNER JOIN chatrooms
	ON chatrooms.inquiry_id = si.id
	AND chatrooms.deleted_at IS NULL
INNER JOIN users AS inquirer
	ON inquirer.id = si.inquirer_id
INNER JOIN services
	ON si.id = services.inquiry_id
WHERE
	services.service_status = $1
AND
	picker_id = $2
AND
	%s
OFFSET $3
LIMIT $4;
	`, inquiryUUIDQueryClause)
	rows, err := dao.DB.Queryx(
		query,
		models.ServiceStatusNegotiating,
		params.UserID,
		params.Offset,
		params.PerPage,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

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

func (dao *ChatDao) GetMaleInquiryChatrooms(params contracts.GetMaleInquiryChatrooms) ([]models.InquiryChatRoom, error) {
	inquiryUUIDQueryClause := "1 = 1"

	if len(params.InquiryUUID) > 0 {
		inquiryUUIDQueryClause = fmt.Sprintf("si.uuid = '%s'", params.InquiryUUID)
	}

	query := fmt.Sprintf(`
SELECT
	pickers.username,
	pickers.uuid AS picker_uuid,
	pickers.avatar_url,
	si.expect_service_type AS service_type,
	si.inquiry_status,
	si.uuid AS inquiry_uuid,
	si.created_at,
	chatrooms.channel_uuid,
	services.uuid AS service_uuid
FROM service_inquiries AS si
INNER JOIN services ON si.id = services.inquiry_id
INNER JOIN chatrooms
	ON chatrooms.inquiry_id = si.id
	AND chatrooms.deleted_at IS NULL
INNER JOIN users AS pickers
	ON pickers.id = si.picker_id
WHERE
	services.service_status = $1
AND
	si.inquirer_id = $2
AND
	%s
ORDER BY si.created_at DESC
OFFSET $3
LIMIT $4;
`,
		inquiryUUIDQueryClause,
	)

	rows, err := dao.DB.Queryx(
		query,
		models.ServiceStatusNegotiating,
		params.UserID,
		params.Offset,
		params.PerPage,
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
	si.expect_service_type AS service_type,
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
	ORDER BY service_inquiries.created_at DESC
	LIMIT 1
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
	)
ORDER BY created_at DESC
LIMIT 1;
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

func (dao *ChatDao) DeleteChatroomByInquiryId(iqId int) error {
	query := `
UPDATE
	chatrooms
SET
	deleted_at = now()
WHERE
	inquiry_id = $1;
	`

	_, err := dao.DB.Exec(query, iqId)

	return err
}

//func (dao *ChatDao) GetOngoingChatroomsParam(p contracts.GetOngoingChatroomsParam) ([]models.InquiryChatRoom, error) {
//query := `
//SELECT
//pickers.username,
//pickers.uuid AS picker_uuid,
//pickers.avatar_url,
//si.expect_service_type AS service_type,
//si.inquiry_status,
//si.uuid AS inquiry_uuid,
//si.created_at,
//chatrooms.channel_uuid,
//services.uuid AS service_uuid
//FROM service_inquiries AS si
//INNER JOIN services ON si.id = services.inquiry_id
//INNER JOIN chatrooms
//ON chatrooms.inquiry_id = si.id
//AND chatrooms.deleted_at IS NULL
//INNER JOIN users AS pickers
//ON pickers.id = si.picker_id
//WHERE
//si.inquirer_id = $1
//AND (
//si.inquiry_status = $2 OR
//si.inquiry_status = $3
//)
//AND
//si.inquiry_type = $4
//ORDER BY si.created_at DESC
//OFFSET $5
//LIMIT $6;
//`

//rows, err := dao.DB.Queryx(
//query,
//p.InquirerID,
//models.InquiryStatusChatting,
//models.InquiryStatusWaitForInquirerApprove,
//models.InquiryTypeDirect,
//p.Offset,
//p.PerPage,
//)

//if err != nil {
//return nil, err
//}

//dics := make([]models.InquiryChatRoom, 0)

//for rows.Next() {
//dic := models.InquiryChatRoom{}

//if err := rows.StructScan(&dic); err != nil {
//return nil, err
//}

//dics = append(dics, dic)
//}

//return dics, nil
//}
