package darkfirestore

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"github.com/huangc28/go-darkpanda-backend/config"
	"google.golang.org/api/iterator"
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
	Content   string    `firestore:"content,omitempty" json:"content"`
	From      string    `firestore:"from,omitempty" json:"from"`
	To        string    `firestore:"to,omitempty" json:"to"`
	CreatedAt time.Time `firestore:"created_at,omitempty" json:"created_at"`
}

type CreatePrivateChatRoomParams struct {
	ChatRoomName string
	Data         ChatMessage
}

func StructToMap(data interface{}) (map[string]interface{}, error) {
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

func MapToStruct(data map[string]interface{}, ts interface{}) error {
	dataByte, err := json.Marshal(data)

	if err != nil {
		return err
	}

	return json.Unmarshal(dataByte, ts)
}

func (df *DarkFirestore) CreatePrivateChatRoom(ctx context.Context, params CreatePrivateChatRoomParams) error {
	if params.Data.Content == "" {
		params.Data.Content = CreatePrivateChatBotContent
	}

	params.Data.CreatedAt = time.Now()

	_, _, err := df.Client.
		Collection(PrivateChatsCollectionName).
		Doc(params.ChatRoomName).
		Collection(MessageSubCollectionName).
		Add(ctx, params.Data)

	return err
}

func (df *DarkFirestore) GetLatestMessageForEachChatroom(ctx context.Context, channelUUIDs []string) (map[string]ChatMessage, error) {
	privateChatCollection := df.
		Client.
		Collection(PrivateChatsCollectionName)

	errChan := make(chan error)
	quitChan := make(chan struct{})
	dataChan := make(chan map[string]interface{})

	for _, channelUUID := range channelUUIDs {
		select {
		case <-quitChan:
			break
		default:
			go func(channelUUID string) {
				iter := privateChatCollection.
					Doc(channelUUID).
					Collection(MessageSubCollectionName).
					OrderBy("created_at", firestore.Desc).
					Limit(1).
					Documents(ctx)
				for {
					doc, err := iter.Next()

					if err == iterator.Done {
						break
					}

					if err != nil {
						errChan <- err

						break
					}

					data := doc.Data()
					data["channel_uuid"] = channelUUID

					dataChan <- data
				}

			}(channelUUID)
		}
	}

	channelMessageMap := make(map[string]ChatMessage)

	for range channelUUIDs {
		select {
		case err := <-errChan:
			close(quitChan)

			return nil, err
		case data := <-dataChan:
			m := ChatMessage{}
			if err := MapToStruct(data, &m); err != nil {
				close(quitChan)

				return nil, err
			}

			channelMessageMap[fmt.Sprintf("%s", data["channel_uuid"])] = m
		}
	}

	return channelMessageMap, nil
}
