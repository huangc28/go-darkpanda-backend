package inquiry

import (
	"github.com/jmoiron/sqlx"
)

type LobbyServicer interface {
	JoinLobby(inquiryID int64) (string, error)
	LeaveLobby(inquiryID int64) error
	WithTx(tx *sqlx.Tx) LobbyServicer
}

type LobbyServices struct {
	LobbyDao LobbyDaoer
}

// LobbyKey lobby key is composed of user uuid and current unix timestamp
const (
	LobbyKey     = "lobby_%s"
	JoinedAtKey  = "joined_at"
	ExpiredAtKey = "expired_at"
)

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
