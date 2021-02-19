package inquiry

import (
	"time"

	darkfirestore "github.com/huangc28/go-darkpanda-backend/internal/app/pkg/dark_firestore"
	"github.com/jmoiron/sqlx"
)

type LobbyServicer interface {
	JoinLobby(inquiryID int64, df darkfirestore.DarkFireStorer) (string, error)
	AskingLobbyUser(inquiryID int64, df darkfirestore.DarkFireStorer) error
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

func (l *LobbyServices) JoinLobby(inquiryID int64, df darkfirestore.DarkFireStorer) (string, error) {
	//ctx := context.Background()
	//// Generate lobby key
	////  - Set countdown counter in the lobby record in the firestore.
	////  - Set status in the lobby record in the firestore
	//channelUUID, err := l.LobbyDao.JoinLobby(inquiryID)

	//if err != nil {
	//return "", err
	//}

	//_, _, err = df.CreateLobbyUser(
	//ctx,
	//darkfirestore.CreateLobbyUserParams{
	//LobbyUserChannelUUID: channelUUID,
	//LobbyUserStatus:      string(models.LobbyStatusWaiting),
	//// Timer we want to maintain the timer value in the firestore
	//// display in seconds. If we store the value using `time.Minute`,
	//// It displays seconds in terms of nanoseconds.
	//Timer: (time.Minute * 27) / time.Second,
	//},
	//)

	//if err != nil {
	//return "", err
	//}

	//return channelUUID, nil

	return "", nil
}

// AskLobbyUser when female user picks up an inquiry, we need to update the lobby status
// to `asking` on both the firestore and DB. Female user has to wait for the reply of
// Male user to alter the status to `left` which mean the Male has agreed on chatting
// with the female.
func (l *LobbyServices) AskingLobbyUser(inquiryID int64, df darkfirestore.DarkFireStorer) error {
	// Update lobby user status in DB to `asking`.
	//lobbyUser, err := l.LobbyDao.UpdateLobbyUserStatus(UpdateLobbyUserStatusParams{
	//InquiryID:       inquiryID,
	//LobbyUserStatus: models.LobbyStatusAsking,
	//})

	//if err != nil {
	//return err
	//}

	//log.Infof("asking lobby user %v", lobbyUser)

	//// Update lobby user status in firestore document to `asking`
	//ctx := context.Background()
	//err = df.AskingLobbyUser(
	//ctx,
	//darkfirestore.AskingLobbyUserParams{
	//LobbyUserChannelUUID: lobbyUser.ChannelUuid,
	//},
	//)

	//if err != nil {
	//return err
	//}

	return nil
}

func (l *LobbyServices) LeaveLobby(inquiryID int64) error {
	return l.LobbyDao.LeaveLobby(inquiryID)
}
