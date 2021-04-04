package darkfirestore

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"github.com/huangc28/go-darkpanda-backend/config"
	"github.com/huangc28/go-darkpanda-backend/internal/app/models"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"

	log "github.com/sirupsen/logrus"
)

var _darkFirestore *DarkFirestore

type MessageType string

const (
	Text             MessageType = "text"
	ServiceDetail    MessageType = "service_detail"
	ConfirmedService MessageType = "confirmed_service"
)

type DarkFireStorer interface {
	GetClient() *firestore.Client
	CreatePrivateChatRoom(ctx context.Context, params CreatePrivateChatRoomParams) error
	SendTextMessageToChatroom(ctx context.Context, params SendTextMessageParams) (ChatMessage, error)
	SendServiceDetailMessageToChatroom(ctx context.Context, params SendServiceDetailMessageParams) (ServiceDetailMessage, error)
	GetLatestMessageForEachChatroom(ctx context.Context, channelUUIDs []string) (map[string]ChatMessage, error)
	GetHistoricalMessages(ctx context.Context, params GetHistoricalMessagesParams) ([]interface{}, error)
	SendServiceConfirmedMessage(ctx context.Context, params SendServiceConfirmedMessageParams) (*firestore.DocumentRef, ServiceDetailMessage, error)
	CreateInquiringUser(ctx context.Context, params CreateInquiringUserParams) (*firestore.WriteResult, InquiringUserInfo, error)
	AskingInquiringUser(ctx context.Context, params AskingInquiringUserParams) error
	UpdateInquiryStatus(ctx context.Context, params UpdateInquiryStatusParams) error
}

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

func (df *DarkFirestore) GetClient() *firestore.Client {
	return _darkFirestore.Client
}

const (
	PrivateChatsCollectionName  = "private_chats"
	MessageSubCollectionName    = "messages"
	CreatePrivateChatBotContent = "Welcome! %s has picked up your inquiry."
)

// @TODO remove `To` column.
type ChatMessage struct {
	Type      MessageType `firestore:"type,omitempty" json:"type"`
	Content   interface{} `firestore:"content,omitempty" json:"content"`
	From      string      `firestore:"from,omitempty" json:"from"`
	To        string      `firestore:"to,omitempty" json:"to"`
	CreatedAt time.Time   `firestore:"created_at,omitempty" json:"created_at"`
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

	params.Data.Content = CreatePrivateChatBotContent

	if params.Data.Type == "" {
		params.Data.Type = Text
	}

	chat, err := df.SendTextMessageToChatroom(ctx, SendTextMessageParams{
		ChatroomName: params.ChatRoomName,
		Data:         params.Data,
	})

	log.WithFields(log.Fields{
		"chatroom_name": params.ChatRoomName,
		"updated_time":  chat.CreatedAt,
	}).Debug("Inquiry Chatroom created!")

	return err
}

type SendTextMessageParams struct {
	ChatroomName string
	Data         ChatMessage
}

func (df *DarkFirestore) SendTextMessageToChatroom(ctx context.Context, params SendTextMessageParams) (ChatMessage, error) {
	if params.Data.Type == "" {
		params.Data.Type = Text
	}

	params.Data.CreatedAt = time.Now()

	_, _, err := df.Client.
		Collection(PrivateChatsCollectionName).
		Doc(params.ChatroomName).
		Collection(MessageSubCollectionName).
		Add(ctx, params.Data)

	if err != nil {
		return params.Data, err
	}

	return params.Data, err
}

type ServiceDetailMessage struct {
	ChatMessage
	Price       float64 `firestore:"price,omitempty" json:"price"`
	Duration    int     `firestore:"duration,omitempty" json:"duration"`
	ServiceUUID string  `firestore:"service_uuid" json:"service_uuid"`
	ServiceTime int64   `firestore:"service_time,omitempty" json:"service_time"`
	ServiceType string  `firestore:"service_type,omitempty" json:"service_type"`
}

type SendServiceDetailMessageParams struct {
	ChatroomName string
	Data         ServiceDetailMessage
}

const (
	ServiceDetailMessageContent = "Service updated:\n"
)

func (df *DarkFirestore) SendServiceDetailMessageToChatroom(ctx context.Context, params SendServiceDetailMessageParams) (ServiceDetailMessage, error) {
	if params.Data.Type == "" {
		params.Data.Type = ServiceDetail
	}

	if params.Data.Content == "" {
		params.Data.Content = ServiceDetailMessageContent
	}

	params.Data.CreatedAt = time.Now()

	// Service detail content has different
	_, _, err := df.Client.
		Collection(PrivateChatsCollectionName).
		Doc(params.ChatroomName).
		Collection(MessageSubCollectionName).
		Add(ctx, params.Data)

	if err != nil {
		return params.Data, err
	}

	return params.Data, nil
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

type GetHistoricalMessagesParams struct {
	Offset      int
	Limit       int
	ChannelUUID string
}

// GetHistoricalMessages retrieve historical message from firestore. The return format would be slice of
// empty interfaces.
func (df *DarkFirestore) GetHistoricalMessages(ctx context.Context, params GetHistoricalMessagesParams) ([]interface{}, error) {
	currBatch := df.
		Client.
		Collection(PrivateChatsCollectionName).
		Doc(params.ChannelUUID).
		Collection(MessageSubCollectionName).
		OrderBy("created_at", firestore.Desc).
		Offset(params.Offset).
		Limit(params.Limit).
		Documents(ctx)

	currDocs, err := currBatch.GetAll()

	if err != nil {
		return nil, err
	}

	msgs := make([]interface{}, 0)

	for _, doc := range currDocs {
		msgs = append(msgs, doc.Data())
	}

	return msgs, nil
}

type SendServiceConfirmedMessageParams struct {
	ChannelUUID string
	Data        ServiceDetailMessage
}

// EmitServiceConfirmedMessage male user would emit a service confirmed message to notify female user that the
// service is confirmed by the client.
func (df *DarkFirestore) SendServiceConfirmedMessage(ctx context.Context, params SendServiceConfirmedMessageParams) (*firestore.DocumentRef, ServiceDetailMessage, error) {
	if params.Data.Type == "" {
		params.Data.Type = ConfirmedService
	}

	params.Data.CreatedAt = time.Now()

	ref, _, err := df.Client.
		Collection(PrivateChatsCollectionName).
		Doc(params.ChannelUUID).
		Collection(MessageSubCollectionName).
		Add(ctx, params.Data)

	if err != nil {
		return nil, params.Data, err
	}

	return ref, params.Data, nil
}

const (
	InquiryCollectionName = "inquiries"
)

type CreateInquiringUserParams struct {
	InquiryUUID string
}

type InquiringUserInfo struct {
	InquiryUUID string `firestore:"inquiry_uuid,omitempty"`
	Status      string `firestore:"status,omitempty"`
}

// CreateInquiry Adds user into lobby by creating a user record in the firestore.
// User record includes following info:
//   - timer countdown in second.
//   - lobby status.
func (df *DarkFirestore) CreateInquiringUser(ctx context.Context, params CreateInquiringUserParams) (*firestore.WriteResult, InquiringUserInfo, error) {
	data := InquiringUserInfo{
		InquiryUUID: params.InquiryUUID,
		Status:      string(models.InquiryStatusInquiring),
	}

	wres, err := df.
		Client.
		Collection(InquiryCollectionName).
		Doc(params.InquiryUUID).
		Set(ctx, data)

	if err != nil {
		return nil, data, err
	}

	return wres, data, nil
}

type UpdateInquiryStatusParams struct {
	InquiryUUID string
	Status      models.InquiryStatus
}

func (df *DarkFirestore) UpdateInquiryStatus(ctx context.Context, params UpdateInquiryStatusParams) error {
	_, err := df.
		Client.
		Collection(InquiryCollectionName).
		Doc(params.InquiryUUID).
		Update(ctx, []firestore.Update{
			{
				Path:  "status",
				Value: params.Status,
			},
		})

	if err != nil {
		return err
	}

	return err
}

type AskingInquiringUserParams struct {
	InquiryUUID string
}

// AskingLobbyUser updates the status of lobby user document
// to be `asking` to notify male user to diplay a popup.
func (df *DarkFirestore) AskingInquiringUser(ctx context.Context, params AskingInquiringUserParams) error {
	return df.UpdateInquiryStatus(
		ctx,
		UpdateInquiryStatusParams{
			InquiryUUID: params.InquiryUUID,
			Status:      models.InquiryStatusAsking,
		},
	)
}

type ChatInquiringUserParams struct {
	InquiryUUID string
}

func (df *DarkFirestore) ChatInquiringUser(ctx context.Context, params ChatInquiringUserParams) error {
	return df.UpdateInquiryStatus(
		ctx,
		UpdateInquiryStatusParams{
			InquiryUUID: params.InquiryUUID,
			Status:      models.InquiryStatusChatting,
		},
	)
}
