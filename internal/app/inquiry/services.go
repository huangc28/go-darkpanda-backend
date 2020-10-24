package inquiry

import (
	"time"

	"github.com/jmoiron/sqlx"
)

type LobbyServicer interface {
	JoinLobby(inquiryID int64) (string, error)
	LeaveLobby(inquiryID int64) error
	WithTx(tx *sqlx.Tx) LobbyServicer
}

func IsInquiryExpired(expT time.Time) bool {
	return expT.Before(time.Now())
}

type LobbyServices struct {
	LobbyDao LobbyDaoer
}

func (l *LobbyServices) WithTx(tx *sqlx.Tx) LobbyServicer {
	l.LobbyDao.WithTx(tx)

	return l
}

// JoinLobby generates pubsub channel uuid for client to subscribe. We will create a new lobby record in
// `lobby_users` table
func (l *LobbyServices) JoinLobby(inquiryID int64) (string, error) {
	// Generate lobby key
	return l.LobbyDao.JoinLobby(inquiryID)
}

func (l *LobbyServices) LeaveLobby(inquiryID int64) error {
	return l.LobbyDao.LeaveLobby(inquiryID)
}
