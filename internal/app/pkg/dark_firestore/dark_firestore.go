package darkfirestore

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"github.com/huangc28/go-darkpanda-backend/config"
	"google.golang.org/api/option"
)

var _darkFirestore *DarkFirestore

type DarkFirestore struct {
	Client *firestore.Client
}

func Get() *DarkFirestore {
	return _darkFirestore
}

type InitOptions struct {
	CredentialFile string
}

func InitFireStore(ctx context.Context, options InitOptions) error {
	sa := option.WithCredentialsFile(fmt.Sprintf("%s/%s", config.GetProjRootPath(), options.CredentialFile))
	app, err := firebase.NewApp(ctx, nil, sa)

	if err != nil {
		return err
	}

	firestore, err := app.Firestore(ctx)

	if err != nil {
		return err
	}

	_darkFirestore = &DarkFirestore{
		Client: firestore,
	}

	return nil
}

const (
	PrivateChatsCollectionName  = "private_chats"
	MessageSubCollectionName    = "messages"
	CreatePrivateChatBotContent = "Welcome! %s has picked up your inquiry."
)

type ChatMessage struct {
	Content   string    `json:"content"`
	From      string    `json:"from"`
	To        string    `json:"to"`
	CreatedAt time.Time `json:"created_at"`
}

type CreatePrivateChatRoomParams struct {
	ChatRoomName string
	Data         ChatMessage
}

func structToMap(data interface{}) (map[string]interface{}, error) {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	mapData := make(map[string]interface{})
	err = json.Unmarshal(dataBytes, &mapData)
	if err != nil {
		return nil, err
	}
	return mapData, nil
}

func (df *DarkFirestore) CreatePrivateChatRoom(ctx context.Context, params CreatePrivateChatRoomParams) error {
	if params.Data.Content == "" {
		params.Data.Content = CreatePrivateChatBotContent
	}

	params.Data.CreatedAt = time.Now()

	dataMap, err := structToMap(params.Data)

	if err != nil {
		return err
	}

	_, _, err = df.Client.
		Collection(PrivateChatsCollectionName).
		Doc(params.ChatRoomName).
		Collection(MessageSubCollectionName).
		Add(ctx, dataMap)

	return err
}
