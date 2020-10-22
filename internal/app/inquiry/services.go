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

type ChatServicer interface {
	CreateAndJoinChatroom(inquiryID int64, userIDs ...int64) (*ChatroomInfo, error)
	WithTx(tx *sqlx.Tx) ChatServicer
}

type ChatServices struct {
	ChatDao ChatDaoer
}

type ChatroomInfo struct {
	ChannelUuid string
	ChatroomID  int64
}

func (cs *ChatServices) WithTx(tx *sqlx.Tx) ChatServicer {
	cs.ChatDao.WithTx(tx)

	return cs
}

func (cs *ChatServices) CreateAndJoinChatroom(inquiryID int64, userIDs ...int64) (*ChatroomInfo, error) {
	// Create chatroom
	chatInfo, err := cs.ChatDao.CreateChat(inquiryID)

	if err != nil {
		return (*ChatroomInfo)(nil), err
	}

	// Join chatroom
	if err := cs.ChatDao.JoinChat(chatInfo.ChatID, userIDs...); err != nil {
		return (*ChatroomInfo)(nil), err
	}

	return &ChatroomInfo{
		ChannelUuid: chatInfo.ChanelUuid,
		ChatroomID:  chatInfo.ChatID,
	}, nil
}
