package inquiry

import (
	"fmt"
	"time"

	"github.com/huangc28/go-darkpanda-backend/db"
	"github.com/jmoiron/sqlx"
	"github.com/teris-io/shortid"
)

type LobbyDaoer interface {
	JoinLobby(inquiryID int64) (string, error)
	LeaveLobby(inquiryID int64) error
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
	inquiry_id
) VALUES ($1, $2);
	`

	if _, err := l.DB.Exec(
		query,
		chanUuid,
		inquiryID,
	); err != nil {
		return "", err
	}

	return chanUuid, nil
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
