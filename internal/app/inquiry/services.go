package inquiry

import (
	"context"
	"time"

	"github.com/huangc28/go-darkpanda-backend/internal/app/models"
	darkfirestore "github.com/huangc28/go-darkpanda-backend/internal/app/pkg/dark_firestore"
	"github.com/jmoiron/sqlx"
)

type LobbyServicer interface {
	JoinLobby(inquiryID int64, df *darkfirestore.DarkFirestore) (string, error)
	LeaveLobby(inquiryID int64) error
	WithTx(tx *sqlx.Tx) LobbyServicer
}

func IsInquiryExpired(expT time.Time) bool {
	return expT.Before(time.Now())
}

type LobbyServices struct {
	LobbyDao LobbyDaoer
}

func NewLobbyService(dao LobbyDaoer) *LobbyServices {
	return &LobbyServices{
		LobbyDao: dao,
	}
}

func (l *LobbyServices) WithTx(tx *sqlx.Tx) LobbyServicer {
	l.LobbyDao.WithTx(tx)

	return l
}

// JoinLobby generates pubsub channel uuid for client to subscribe. We will create a new lobby record in
// `lobby_users` table

const InquiryTimerDuration = 27

func (l *LobbyServices) JoinLobby(inquiryID int64, df *darkfirestore.DarkFirestore) (string, error) {
	ctx := context.Background()
	// Generate lobby key
	//  - Set countdown counter in the lobby record in the firestore.
	//  - Set status in the lobby record in the firestore
	channelUUID, err := l.LobbyDao.JoinLobby(inquiryID)

	if err != nil {
		return "", err
	}

	_, _, err = df.CreateLobbyUser(
		ctx,
		darkfirestore.CreateLobbyUserParams{
			LobbyUserChannelUUID: channelUUID,
			LobbyUserStatus:      string(models.LobbyStatusWaiting),
			// Timer we want to maintain the timer value in the firestore
			// display in seconds. If we store the value using `time.Minute`,
			// It displays seconds in terms of nanoseconds.
			Timer: (time.Minute * 27) / time.Second,
		},
	)

	if err != nil {
		return "", err
	}

	return channelUUID, nil
}

func (l *LobbyServices) LeaveLobby(inquiryID int64) error {
	return l.LobbyDao.LeaveLobby(inquiryID)
}
