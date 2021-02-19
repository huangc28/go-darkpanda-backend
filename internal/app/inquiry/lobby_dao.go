package inquiry

import (
	"fmt"
	"time"

	"github.com/huangc28/go-darkpanda-backend/db"
	"github.com/huangc28/go-darkpanda-backend/internal/app/models"
	"github.com/jmoiron/sqlx"
	"github.com/teris-io/shortid"
)

type LobbyDaoer interface {
	JoinLobby(inquiryID int64) (string, error)
	LeaveLobby(inquiryID int64) error
	UpdateLobbyUserStatus(params UpdateLobbyUserStatusParams) (*models.LobbyUser, error)
	WithTx(tx *sqlx.Tx)
}

type LobbyDao struct {
	DB db.Conn
}

func NewLobbyDao(DB db.Conn) *LobbyDao {
	return &LobbyDao{
		DB: DB,
	}
}

func (l *LobbyDao) WithTx(tx *sqlx.Tx) {
	l.DB = tx
}

func (l *LobbyDao) JoinLobby(inquiryID int64) (string, error) {
	uuid, err := shortid.Generate()
	chanUuid := fmt.Sprintf("lobby_%s", uuid)

	if err != nil {
		return "", err
	}

	query := `
INSERT INTO lobby_users (
	channel_uuid,
	inquiry_id,
	lobby_status
) VALUES ($1, $2, $3);
	`

	if _, err := l.DB.Exec(
		query,
		chanUuid,
		inquiryID,
		models.LobbyStatusWaiting,
	); err != nil {
		return "", err
	}

	return chanUuid, nil
}

type UpdateLobbyUserStatusParams struct {
	InquiryID       int64
	LobbyUserStatus models.LobbyStatus
}

func (l *LobbyDao) UpdateLobbyUserStatus(params UpdateLobbyUserStatusParams) (*models.LobbyUser, error) {
	query := `
UPDATE
	lobby_users
SET
	lobby_status = $1
WHERE
	inquiry_id = $2
RETURNING *;
`
	var m models.LobbyUser

	if err := l.DB.QueryRowx(
		query,
		params.LobbyUserStatus,
		params.InquiryID,
	).StructScan(&m); err != nil {
		return nil, err
	}

	return &m, nil
}

func (dao *LobbyDao) LeaveLobby(inquiryID int64) error {
	sql := `
UPDATE
	lobby_users
SET
	deleted_at = $1
WHERE
	inquiry_id = $2
AND
	deleted_at IS NULL;
`
	leaveTime := time.Now()

	if _, err := dao.DB.Exec(sql, leaveTime, inquiryID); err != nil {
		return err
	}

	return nil
}

func (dao *LobbyDao) GetLobbyUserByInquiryID(inquiryID int64) (*models.LobbyUser, error) {
	query := `
SELECT
	*
FROM
	lobby_users
WHERE
	inquiry_id = $1;
`
	var m models.LobbyUser

	if err := dao.DB.QueryRowx(
		query,
		inquiryID,
	).StructScan(&m); err != nil {
		return nil, err
	}

	return &m, nil
}
