package darkfirestore

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"github.com/huangc28/go-darkpanda-backend/config"
	"github.com/huangc28/go-darkpanda-backend/internal/app/models"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	log "github.com/sirupsen/logrus"
)

var _darkFirestore *DarkFirestore

type MessageType string

const (
	Text                MessageType = "text"
	UpdateInquiryDetail MessageType = "update_inquiry_detail"
	ServiceDetail       MessageType = "service_detail"
	ConfirmedService    MessageType = "confirmed_service"
	DisagreeInquiry     MessageType = "disagree_inquiry"
	QuitChatroom        MessageType = "quit_chatroom"
	CompletePayment     MessageType = "complete_payment"
	CancelService       MessageType = "cancel_service"
)

const (
	// Key value of the `inquries` collection in firestore.
	InquiryCollectionName = "inquiries"

	// Key value of the `services` collection in firestore.
	ServiceCollectionName = "services"

	// Key value of the `private_chats` collection in firestore.
	PrivateChatsCollectionName = "private_chats"

	// Key value of the subcollection `messages` of `private_chats` collection.
	MessageSubCollectionName = "messages"

	// Message content template when inquiry is created
	CreatePrivateChatBotContent = "Welcome! %s has picked up your inquiry."

	// Message content template when female user has updated the service detail.
	ServiceDetailMessageContent = "Service updated:\n"
)

type DarkFireStorer interface {
	GetClient() *firestore.Client

	CreatePrivateChatRoom(ctx context.Context, params CreatePrivateChatRoomParams) error

	SendTextMessageToChatroom(ctx context.Context, params SendTextMessageParams) (ChatMessage, error)
	SendServiceDetailMessageToChatroom(ctx context.Context, params SendServiceDetailMessageParams) (ServiceDetailMessage, error)
	SendServiceConfirmedMessage(ctx context.Context, params SendServiceConfirmedMessageParams) (*firestore.DocumentRef, ServiceDetailMessage, error)

	GetLatestMessageForEachChatroom(ctx context.Context, channelUUIDs []string) (map[string][]*ChatMessage, error)
	GetHistoricalMessages(ctx context.Context, params GetHistoricalMessagesParams) ([]interface{}, error)

	CreateInquiringUser(ctx context.Context, params CreateInquiringUserParams) (*firestore.WriteResult, InquiringUserInfo, error)
	AskingInquiringUser(ctx context.Context, params AskingInquiringUserParams) error
	UpdateInquiryStatus(ctx context.Context, params UpdateInquiryStatusParams) error

	CreateService(ctx context.Context, params CreateServiceParams) error
	UpdateService(ctx context.Context, params UpdateServiceParams) error
	CancelService() error
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

// @TODO remove `To` column.
type ChatMessage struct {
	Type      MessageType `firestore:"type,omitempty" json:"type"`
	Content   interface{} `firestore:"content,omitempty" json:"content"`
	From      string      `firestore:"from,omitempty" json:"from"`
	CreatedAt time.Time   `firestore:"created_at,omitempty" json:"created_at"`
	Empty     bool        `json:"empty"`
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

	// Create private chatroom by adding a dummy field "last_touched".
	_, err := df.Client.
		Collection(PrivateChatsCollectionName).
		Doc(params.ChatroomName).
		Set(ctx, map[string]interface{}{
			"last_touched": time.Now(),
		})

	if err != nil {
		return params.Data, err

	}

	// Once private chatroom is created, we send a welcome message here.
	msgDoc := df.
		Client.
		Collection(PrivateChatsCollectionName).
		Doc(params.ChatroomName).
		Collection(MessageSubCollectionName).
		NewDoc()

	if _, err := msgDoc.Set(ctx, params.Data); err != nil {
		return params.Data, err
	}

	return params.Data, nil
}

type ServiceDetailMessage struct {
	ChatMessage
	Price       float64 `firestore:"price,omitempty" json:"price"`
	MatchingFee int     `firestore:"matching_fee" json:"matching_fee"`
	Duration    int     `firestore:"duration,omitempty" json:"duration"`
	ServiceUUID string  `firestore:"service_uuid" json:"service_uuid"`
	ServiceTime int64   `firestore:"service_time,omitempty" json:"service_time"`
	ServiceType string  `firestore:"service_type,omitempty" json:"service_type"`
}

type SendServiceDetailMessageParams struct {
	ChatroomName string
	Data         ServiceDetailMessage
}

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

func (df *DarkFirestore) GetLatestMessageForEachChatroom(ctx context.Context, channelUUIDs []string) (map[string][]*ChatMessage, error) {
	privateChatCollection := df.
		Client.
		Collection(PrivateChatsCollectionName)

	errChan := make(chan error)
	quitChan := make(chan struct{})
	dataChan := make(chan []map[string]interface{})

	for _, channelUUID := range channelUUIDs {
		select {
		case <-quitChan:
			break
		default:
			go func(channelUUID string) {
				// What happens if channelUUID does not exists in firestore?
				_, err := privateChatCollection.Doc(channelUUID).Get(ctx)

				if grpc.Code(err) == codes.NotFound {
					errChan <- errors.New(fmt.Sprintf("error chatroom channel: %s not found", channelUUID))

					return
				}

				iter := privateChatCollection.
					Doc(channelUUID).
					Collection(MessageSubCollectionName).
					OrderBy("created_at", firestore.Desc).
					Limit(1).
					Documents(ctx)

				messagesArr := make([]map[string]interface{}, 0)

				for {
					doc, err := iter.Next()

					if err == iterator.Done {
						empty := make(map[string]interface{})
						empty["channel_uuid"] = channelUUID
						empty["empty"] = true

						messagesArr = append(messagesArr, empty)

						break
					}

					if err != nil {
						errChan <- err

						break
					}

					data := doc.Data()
					data["channel_uuid"] = channelUUID
					data["empty"] = false

					messagesArr = append(messagesArr, data)
				}

				dataChan <- messagesArr

			}(channelUUID)
		}
	}

	channelMessageMap := make(map[string][]*ChatMessage)

	for range channelUUIDs {
		select {
		case err := <-errChan:
			close(quitChan)

			if err == iterator.Done {
				return nil, errors.New("")
			}

			return nil, err
		case data := <-dataChan:
			msgArr := make([]*ChatMessage, 0)

			firstMsg := data[0]
			m := &ChatMessage{}

			if firstMsg["empty"] == false {
				if err := MapToStruct(firstMsg, m); err != nil {
					close(quitChan)

					return nil, err
				}

				msgArr = append(msgArr, m)
			}

			channelMessageMap[fmt.Sprintf("%s", firstMsg["channel_uuid"])] = msgArr
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

type InquiryDetailMessage struct {
	ChatMessage
	Price           float64 `firestore:"price,omitempty" json:"price"`
	MatchingFee     int     `firestore:"matching_fee,omitempty" json:"matching_fee"`
	Duration        int     `firestore:"duration,omitempty" json:"duration"`
	InquiryUuid     string  `firestore:"inquiry_uuid" json:"inquiry_uuid"`
	AppointmentTime int64   `firestore:"appointment_time,omitempty" json:"appointment_time"`
	ServiceType     string  `firestore:"service_type,omitempty" json:"service_type"`
	Address         string  `firestore:"address,omitempty" json:"address"`
}

type UpdateInquiryMessage struct {
	ChannelUuid string
	Data        InquiryDetailMessage
}

func (df *DarkFirestore) GetMessageColllectionRef(channelUuid string) *firestore.CollectionRef {
	ref := df.Client.
		Collection(PrivateChatsCollectionName).
		Doc(channelUuid).
		Collection(MessageSubCollectionName)

	return ref
}

func (df *DarkFirestore) SendUpdateInquiryMessage(ctx context.Context, params UpdateInquiryMessage) (*firestore.DocumentRef, InquiryDetailMessage, error) {
	if params.Data.Type == "" {
		params.Data.Type = UpdateInquiryDetail
	}

	params.Data.CreatedAt = time.Now()

	msgColRef := df.GetMessageColllectionRef(params.ChannelUuid)

	msgDocRef, _, err := msgColRef.Add(ctx, params.Data)

	if err != nil {
		return nil, params.Data, err
	}

	return msgDocRef, params.Data, nil
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

type CompletePaymentParams struct {
	ChannelUuid string
	Data        ChatMessage
}

func (df *DarkFirestore) SendCompletePaymentMessage(ctx context.Context, p CompletePaymentParams) (*firestore.DocumentRef, ChatMessage, error) {
	p.Data.Type = CompletePayment
	p.Data.CreatedAt = time.Now()

	ref, _, err := df.Client.
		Collection(PrivateChatsCollectionName).
		Doc(p.ChannelUuid).
		Collection(MessageSubCollectionName).
		Add(ctx, p.Data)

	if err != nil {
		return nil, p.Data, err
	}

	return ref, p.Data, nil
}

type QuitChatroomMessageParams struct {
	ChannelUuid string
	Data        ChatMessage
}

func (df *DarkFirestore) SendQuitChatroomMessage(ctx context.Context, params QuitChatroomMessageParams) (*firestore.DocumentRef, ChatMessage, error) {
	if params.Data.Type == "" {
		params.Data.Type = QuitChatroom
	}

	params.Data.CreatedAt = time.Now()

	ref, _, err := df.Client.
		Collection(PrivateChatsCollectionName).
		Doc(params.ChannelUuid).
		Collection(MessageSubCollectionName).
		Add(ctx, params.Data)

	if err != nil {
		return nil, params.Data, err
	}

	return ref, params.Data, nil
}

type SendDisagreeInquiryMessageParams struct {
	ChannelUuid string
	Data        ChatMessage
}

func (df *DarkFirestore) SendDisagreeInquiryMessage(ctx context.Context, params SendDisagreeInquiryMessageParams) (*firestore.DocumentRef, ChatMessage, error) {
	if params.Data.Type == "" {
		params.Data.Type = DisagreeInquiry
	}

	params.Data.CreatedAt = time.Now()

	ref, _, err := df.
		Client.
		Collection(PrivateChatsCollectionName).
		Doc(params.ChannelUuid).
		Collection(MessageSubCollectionName).
		Add(ctx, params.Data)

	if err != nil {
		return nil, params.Data, nil
	}

	return ref, params.Data, nil
}

type CreateInquiringUserParams struct {
	InquiryUUID string
}

type InquiringUserInfo struct {
	InquiryUUID string `firestore:"inquiry_uuid,omitempty"`
	Status      string `firestore:"status,omitempty"`
}

// CreateInquiry creates new record in inquiries collection.
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
	InquiryUuid string
	Status      models.InquiryStatus
	PickerUuid  string
}

func (df *DarkFirestore) UpdateInquiryStatus(ctx context.Context, params UpdateInquiryStatusParams) error {
	_, err := df.
		Client.
		Collection(InquiryCollectionName).
		Doc(params.InquiryUuid).
		Update(ctx, []firestore.Update{
			{
				Path:  "status",
				Value: params.Status,
			},
			{
				Path:  "picker_uuid",
				Value: params.PickerUuid,
			},
		})

	if err != nil {
		return err
	}

	return err
}

type AskingInquiringUserParams struct {
	InquiryUuid string
	PickerUuid  string
}

// AskingLobbyUser updates the status of lobby user document
// to be `asking` to notify male user to diplay a popup.
func (df *DarkFirestore) AskingInquiringUser(ctx context.Context, params AskingInquiringUserParams) error {
	return df.UpdateInquiryStatus(
		ctx,
		UpdateInquiryStatusParams{
			InquiryUuid: params.InquiryUuid,
			Status:      models.InquiryStatusAsking,
			PickerUuid:  params.PickerUuid,
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
			InquiryUuid: params.InquiryUUID,
			Status:      models.InquiryStatusChatting,
		},
	)
}

type CreateServiceParams struct {
	ServiceUuid   string `firestore:"service_uuid,omitempty" json:"service_uuid"`
	ServiceStatus string `firestore:"service_status,omitempty" json:"service_status"`
}

func (df *DarkFirestore) CreateService(ctx context.Context, params CreateServiceParams) error {
	// Create a service record.
	_, err := df.
		Client.
		Collection(ServiceCollectionName).
		Doc(params.ServiceUuid).
		Set(ctx, params)

	if err != nil {
		return err

	}

	return nil
}

type UpdateServiceParams struct {
	ServiceUuid   string `firestore:"service_uuid,omitempty" json:"service_uuid"`
	ServiceStatus string `firestore:"service_status,omitempty" json:"service_status"`
}

func (df *DarkFirestore) UpdateService(ctx context.Context, params UpdateServiceParams) error {
	_, err := df.
		Client.
		Collection(ServiceCollectionName).
		Doc(params.ServiceUuid).
		Update(
			ctx,
			[]firestore.Update{
				{
					Path:  "service_status",
					Value: params.ServiceStatus,
				},
			},
		)

	if err != nil {
		return err
	}

	return nil
}

type UpdateMultipleServiceStatusParams struct {
	ServiceUuids  []string
	ServiceStatus string
}

func (df *DarkFirestore) UpdateMultipleServiceStatus(ctx context.Context, params UpdateMultipleServiceStatusParams) error {
	batch := df.Client.Batch()

	for _, sUuid := range params.ServiceUuids {
		docRef := df.
			Client.
			Collection(ServiceCollectionName).
			Doc(sUuid)

		batch.Set(docRef, map[string]interface {
		}{
			"service_status": params.ServiceStatus,
		}, firestore.MergeAll)
	}

	_, err := batch.Commit(ctx)

	return err
}

type CancelServiceParams struct {
	ChannelUuid string
	ServiceUuid string
	Data        ChatMessage
}

// CancelService consist of 2 actions.
//   - Update service status in `services` collection
//   - Send cancel service message to chatroom in `private_chats` collection
func (df *DarkFirestore) CancelService(ctx context.Context, p CancelServiceParams) error {
	chatroomRef := df.Client.
		Collection(PrivateChatsCollectionName).Doc(p.ChannelUuid)

	srvRef := df.Client.
		Collection(ServiceCollectionName).
		Doc(p.ServiceUuid)

	p.Data.Type = CancelService

	err := df.Client.RunTransaction(
		ctx,
		func(ctx context.Context, tx *firestore.Transaction) error {
			if err := tx.Update(srvRef, []firestore.Update{
				{
					Path:  "service_status",
					Value: models.ServiceStatusCanceled,
				},
			}); err != nil {
				return err
			}

			if err := tx.Set(chatroomRef, p.Data); err != nil {
				return err
			}

			return nil
		},
	)

	if err != nil {
		return err
	}

	return nil
}
