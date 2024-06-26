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
	log "github.com/sirupsen/logrus"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _darkFirestore *DarkFirestore

type MessageType string

const (
	Text                  MessageType = "text"
	BotInvitationChatText MessageType = "bot_invitation_chat_text"
	UpdateInquiryDetail   MessageType = "update_inquiry_detail"
	ServiceDetail         MessageType = "service_detail"
	ConfirmedService      MessageType = "confirmed_service"
	DisagreeInquiry       MessageType = "disagree_inquiry"
	QuitChatroom          MessageType = "quit_chatroom"
	CompletePayment       MessageType = "complete_payment"
	CancelService         MessageType = "cancel_service"
	StartService          MessageType = "start_service"
	Images                MessageType = "images"
)

const (
	// Key value of the `inquries` collection in firestore.
	InquiryCollectionName = "inquiries"

	// Key value of the `services` collection in firestore.
	ServiceCollectionName = "services"

	// Name of the column of `service` document in `services` collection that indicates the status.
	ServiceStatusFieldName = "status"

	// States the cause of service cancellation.
	ServiceCancelledCauseFieldName = "cancel_cause"

	// Name of the column of inquiry document in `inquiries` collection.
	ChannelUuidFieldName = "channel_uuid"

	// Column name of `inquiry` document in `inquiries` collection. We need to notify both chat partners
	// when male agrees to chat.
	ServiceUuidFieldName = "service_uuid"

	// Key value of the `private_chats` collection in firestore.
	PrivateChatsCollectionName = "private_chats"

	// Key value of the subcollection `messages` of `private_chats` collection.
	MessageSubCollectionName = "messages"

	// Message content template when inquiry is created
	BotInvitationChatContent = "Welcome! %s has picked up your inquiry"

	// Message content template when inquiry chatroom is created in Chinese.
	BotInvitationChatContentInZH = "開始聊聊吧!"

	// Message content template when female user has updated the service detail.
	ServiceDetailMessageContent = "Service updated:\n"
)

type DarkFireStorer interface {
	GetClient() *firestore.Client

	SendTextMessageToChatroom(ctx context.Context, params SendTextMessageParams) (ChatMessage, error)
	SendServiceDetailMessageToChatroom(ctx context.Context, params SendServiceDetailMessageParams) (ServiceDetailMessage, error)
	SendServiceConfirmedMessage(ctx context.Context, params SendServiceConfirmedMessageParams) (*firestore.DocumentRef, ServiceDetailMessage, error)

	GetLatestMessageForEachChatroom(ctx context.Context, channelUUIDs []string, userUUID string) (map[string][]*ChatMessage, error)
	GetHistoricalMessages(ctx context.Context, params GetHistoricalMessagesParams) ([]interface{}, error)

	CreateInquiringUser(ctx context.Context, params CreateInquiringUserParams) (*firestore.WriteResult, InquiringUserInfo, error)
	AskingInquiringUser(ctx context.Context, params AskingInquiringUserParams) error
	UpdateInquiryStatus(ctx context.Context, params UpdateInquiryStatusParams) (*firestore.WriteResult, error)
	DisagreeInquiry(ctx context.Context, params DisagreeInquiryParams) (ChatMessage, error)
	UpdateInquiryDetail(ctx context.Context, params UpdateInquiryDetailParams) (InquiryDetailMessage, error)

	UpdateService(ctx context.Context, params UpdateServiceParams) error
	CancelService(ctx context.Context, p CancelServiceParams) error

	UpdateIsReadToTrue(ctx context.Context, p UpdateIsReadParams) error
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

type ChatMessage struct {
	Type      MessageType `firestore:"type,omitempty" json:"type"`
	Content   interface{} `firestore:"content,omitempty" json:"content"`
	From      string      `firestore:"from,omitempty" json:"from"`
	Username  string      `firestore:"username,omitempty" json:"username"`
	CreatedAt time.Time   `firestore:"created_at,omitempty" json:"created_at"`
	IsRead    bool        `firestore:"is_read,omitempty" json:"is_read"`
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

type SendTextMessageParams struct {
	ChannelUuid string
	InquiryUuid string
	Data        ChatMessage
}

func (df *DarkFirestore) SendTextMessageToChatroom(ctx context.Context, params SendTextMessageParams) (ChatMessage, error) {
	if params.Data.Type == "" {
		params.Data.Type = Text
	}

	params.Data.CreatedAt = time.Now()

	df.UpdateIsReadToFalse(ctx, UpdateIsReadParams{
		ChannelUuid: params.ChannelUuid,
		UserUuid:    params.Data.From,
	})

	msgDoc := df.
		Client.
		Collection(PrivateChatsCollectionName).
		Doc(params.ChannelUuid).
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
	MatchingFee float64 `firestore:"matching_fee" json:"matching_fee"`
	Duration    int     `firestore:"duration,omitempty" json:"duration"`
	ServiceUUID string  `firestore:"service_uuid" json:"service_uuid"`
	ServiceTime int64   `firestore:"service_time,omitempty" json:"service_time"`
	ServiceType string  `firestore:"service_type,omitempty" json:"service_type"`
}

type SendServiceDetailMessageParams struct {
	ChannelUuid string
	Data        ServiceDetailMessage
}

func (df *DarkFirestore) SendServiceDetailMessageToChatroom(ctx context.Context, params SendServiceDetailMessageParams) (ServiceDetailMessage, error) {
	params.Data.Type = ServiceDetail
	params.Data.Content = ServiceDetailMessageContent
	params.Data.CreatedAt = time.Now()

	// Service detail content has different
	msgRef := df.getNewChatroomMsgRef(params.ChannelUuid)
	_, err := msgRef.Set(ctx, params.Data)

	if err != nil {
		return params.Data, err
	}

	return params.Data, nil
}

func (df *DarkFirestore) GetLatestMessageForEachChatroom(ctx context.Context, channelUUIDs []string, userUUID string) (map[string][]*ChatMessage, error) {
	privateChatCollection := df.
		Client.
		Collection(PrivateChatsCollectionName)

	errChan := make(chan error)
	quitChan := make(chan struct{})
	dataChan := make(chan []map[string]interface{})

OuterLoop:
	for _, channelUUID := range channelUUIDs {
		select {
		case <-quitChan:
			break OuterLoop
		default:
			go func(channelUUID string) {
				// What happens if channelUUID does not exists in firestore?
				_, err := privateChatCollection.Doc(channelUUID).Get(ctx)

				if status.Code(err) == codes.NotFound {
					errChan <- fmt.Errorf("error chatroom channel: %s not found", channelUUID)

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
					parentDoc, err := doc.Ref.Parent.Parent.Get(ctx)

					if err != nil {
						errChan <- err

						break
					}
					parentData := parentDoc.Data()

					data["channel_uuid"] = channelUUID
					data["empty"] = false

					// check which inquirer or picker is me,
					// get the is read
					if parentData["inquirer_uuid"] == userUUID {
						data["is_read"] = parentData["inquirer_is_read"]
					}

					if parentData["picker_uuid"] == userUUID {
						data["is_read"] = parentData["picker_is_read"]
					}

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

func (df *DarkFirestore) GetEachChatroom(ctx context.Context, channelUUIDs []string) (map[string][]*ChatMessage, error) {
	privateChatCollection := df.
		Client.
		Collection(PrivateChatsCollectionName)

	errChan := make(chan error)
	quitChan := make(chan struct{})
	dataChan := make(chan []map[string]interface{})

OuterLoop:
	for _, channelUUID := range channelUUIDs {
		select {
		case <-quitChan:
			break OuterLoop
		default:
			go func(channelUUID string) {
				// What happens if channelUUID does not exists in firestore?
				_, err := privateChatCollection.Doc(channelUUID).Get(ctx)

				if status.Code(err) == codes.NotFound {
					errChan <- fmt.Errorf("error chatroom channel: %s not found", channelUUID)

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
	MatchingFee     float64 `firestore:"matching_fee,omitempty" json:"matching_fee"`
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
	ServiceUuid string
	Username    string
	From        string
}

func (df *DarkFirestore) CompletePayment(ctx context.Context, p CompletePaymentParams) (ChatMessage, error) {
	msg := ChatMessage{
		Type:      CompletePayment,
		CreatedAt: time.Now(),
		From:      p.From,
		Username:  p.Username,
	}

	srvRef := df.getServiceRef(p.ServiceUuid)
	chatRef := df.getNewChatroomMsgRef(p.ChannelUuid)

	err := df.Client.RunTransaction(
		ctx,
		func(ctx context.Context, tx *firestore.Transaction) error {
			if err := tx.Update(
				srvRef,
				[]firestore.Update{
					{
						Path:  ServiceStatusFieldName,
						Value: models.ServiceStatusToBeFulfilled,
					},
				},
			); err != nil {
				return err
			}

			if err := tx.Set(chatRef, msg); err != nil {
				return nil
			}

			return nil
		},
	)

	return msg, err
}

type QuitChatroomMessageParams struct {
	ChannelUuid string
	InquiryUuid string
	Data        ChatMessage
}

func (df *DarkFirestore) QuitChatroom(ctx context.Context, p QuitChatroomMessageParams) (ChatMessage, error) {
	iqRef := df.getInquiryRef(p.InquiryUuid)
	chatRef := df.getNewChatroomMsgRef(p.ChannelUuid)

	p.Data.Type = QuitChatroom
	p.Data.CreatedAt = time.Now()

	err := df.Client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		if err := tx.Update(iqRef, []firestore.Update{
			{
				Path:  "status",
				Value: models.InquiryStatusInquiring,
			},
			{
				Path:  "picker_uuid",
				Value: "",
			},
		}); err != nil {
			return err
		}

		if err := tx.Set(chatRef, p.Data); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return p.Data, err
	}

	return p.Data, nil
}

type DisagreeInquiryParams struct {
	ChannelUuid string
	InquiryUuid string
	Data        ChatMessage
}

func (df *DarkFirestore) DisagreeInquiry(ctx context.Context, params DisagreeInquiryParams) (ChatMessage, error) {

	params.Data.Type = DisagreeInquiry
	params.Data.CreatedAt = time.Now()

	iqRef := df.getInquiryRef(params.InquiryUuid)
	chatRef := df.getNewChatroomMsgRef(params.ChannelUuid)

	err := df.Client.RunTransaction(ctx,
		func(ctx context.Context, tx *firestore.Transaction) error {
			if err := tx.Update(iqRef, []firestore.Update{
				{
					Path:  ServiceStatusFieldName,
					Value: models.InquiryStatusChatting,
				},
			}); err != nil {
				return err
			}

			if err := tx.Set(chatRef, params.Data); err != nil {
				return err
			}

			return nil
		},
	)

	if err != nil {
		return params.Data, err
	}

	return params.Data, nil
}

type CreateInquiringUserParams struct {
	InquiryUuid  string
	InquirerUuid string
	InquiryType  string
	Status       string
}

type InquiringUserInfo struct {
	InquiryUuid  string `firestore:"inquiry_uuid,omitempty"`
	InquirerUuid string `firestore:"inquirer_uuid,omitempty"`
	InquiryType  string `firestore:"inquiry_type,omitempty"`
	Status       string `firestore:"status,omitempty"`
	ServiceUuid  string `firestore:"service_uuid"`
}

// CreateInquiry creates new record in inquiries collection. Inquiry can
func (df *DarkFirestore) CreateInquiringUser(ctx context.Context, params CreateInquiringUserParams) (*firestore.WriteResult, InquiringUserInfo, error) {
	if params.Status == "" {
		params.Status = string(models.InquiryStatusInquiring)
	}

	data := InquiringUserInfo{
		InquiryUuid:  params.InquiryUuid,
		InquirerUuid: params.InquirerUuid,
		Status:       params.Status,
		ServiceUuid:  "",
	}

	wres, err := df.
		Client.
		Collection(InquiryCollectionName).
		Doc(params.InquiryUuid).
		Set(
			ctx,
			data,
		)

	if err != nil {
		return nil, data, err
	}

	return wres, data, nil
}

type UpdateInquiryStatusParams struct {
	InquiryUuid string
	Status      models.InquiryStatus

	PickerUuid     string
	PickerUsername string
	ChannelUuid    string
}

func (df *DarkFirestore) UpdateInquiryStatus(ctx context.Context, p UpdateInquiryStatusParams) (*firestore.WriteResult, error) {
	iqRef := df.getInquiryRef(p.InquiryUuid)

	rw, err := iqRef.Update(ctx, []firestore.Update{
		{
			Path:  ServiceStatusFieldName,
			Value: p.Status,
		},
		{
			Path:  "picker_uuid",
			Value: p.PickerUuid,
		},
		{
			Path:  "picker_username",
			Value: p.PickerUsername,
		},
		{
			Path:  ChannelUuidFieldName,
			Value: p.ChannelUuid,
		},
	})

	return rw, err
}

type UpdateMultipleInquiryStatusParams struct {
	InquiryUuids []string
	Status       string
}

func (df *DarkFirestore) UpdateMultipleInquiryStatus(ctx context.Context, params UpdateMultipleInquiryStatusParams) error {
	batch := df.Client.Batch()

	for _, sUuid := range params.InquiryUuids {
		iqRef := df.getInquiryRef(sUuid)

		batch.Set(iqRef, map[string]interface {
		}{
			ServiceStatusFieldName: params.Status,
		}, firestore.MergeAll)
	}

	_, err := batch.Commit(ctx)

	return err
}

type UpdateInquiryDetailParams struct {
	InquiryUuid string
	ChannelUuid string
	Status      models.InquiryStatus
	PickerUuid  string
	Data        InquiryDetailMessage
}

func (df *DarkFirestore) UpdateInquiryDetail(ctx context.Context, params UpdateInquiryDetailParams) (InquiryDetailMessage, error) {
	iqRef := df.getInquiryRef(params.InquiryUuid)
	chatRef := df.getNewChatroomMsgRef(params.ChannelUuid)

	params.Data.Type = UpdateInquiryDetail
	params.Data.CreatedAt = time.Now()

	err := df.Client.RunTransaction(
		ctx,
		func(ctx context.Context, tx *firestore.Transaction) error {
			err := tx.Update(iqRef, []firestore.Update{
				{
					Path:  ServiceStatusFieldName,
					Value: params.Status,
				},
				{
					Path:  "picker_uuid",
					Value: params.PickerUuid,
				},
			})

			if err != nil {
				return fmt.Errorf("failed to update inquiry status %s", err.Error())
			}

			if err := tx.Set(chatRef, params.Data); err != nil {
				return fmt.Errorf("failed to send inquiry update message %s", err.Error())
			}

			return nil
		},
	)

	return params.Data, err
}

type AskingInquiringUserParams struct {
	InquiryUuid    string
	PickerUuid     string
	PickerUsername string
}

// AskingLobbyUser updates the status of lobby user document
// to be `asking` to notify male user to diplay a popup.
func (df *DarkFirestore) AskingInquiringUser(ctx context.Context, params AskingInquiringUserParams) error {
	_, err := df.UpdateInquiryStatus(
		ctx,
		UpdateInquiryStatusParams{
			InquiryUuid:    params.InquiryUuid,
			Status:         models.InquiryStatusAsking,
			PickerUuid:     params.PickerUuid,
			PickerUsername: params.PickerUsername,
		},
	)

	return err
}

type UpdateServiceParams struct {
	ServiceUuid   string `firestore:"service_uuid,omitempty" json:"service_uuid"`
	ServiceStatus string `firestore:"status,omitempty" json:"status"`
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
					Path:  ServiceStatusFieldName,
					Value: params.ServiceStatus,
				},
			},
		)

	return err
}

type UpdateMultipleServiceStatusParams struct {
	ServiceUuids  []string
	ServiceStatus string
}

func (df *DarkFirestore) UpdateMultipleServiceStatus(ctx context.Context, params UpdateMultipleServiceStatusParams) error {
	batch := df.Client.Batch()

	for _, sUuid := range params.ServiceUuids {
		docRef := df.getServiceRef(sUuid)

		batch.Set(docRef, map[string]interface {
		}{
			ServiceStatusFieldName: params.ServiceStatus,
		}, firestore.MergeAll)
	}

	_, err := batch.Commit(ctx)

	return err
}

func (df *DarkFirestore) getChatroomRef(channelUuid string) *firestore.DocumentRef {
	return df.
		Client.
		Collection(PrivateChatsCollectionName).
		Doc(channelUuid)
}

func (df *DarkFirestore) getNewChatroomMsgRef(channelUuid string) *firestore.DocumentRef {
	return df.Client.
		Collection(PrivateChatsCollectionName).
		Doc(channelUuid).
		Collection(MessageSubCollectionName).
		NewDoc()
}

func (df *DarkFirestore) getInquiryRef(inquiryUuid string) *firestore.DocumentRef {
	return df.
		Client.
		Collection(InquiryCollectionName).
		Doc(inquiryUuid)
}

func (df *DarkFirestore) getServiceRef(serviceUuid string) *firestore.DocumentRef {
	return df.
		Client.
		Collection(ServiceCollectionName).
		Doc(serviceUuid)
}

type CancelServiceParams struct {
	ChannelUuid string
	ServiceUuid string
	Cause       string
	Data        ChatMessage
}

// CancelService consist of 2 actions.
//   - Update service status in `services` collection
//   - Send cancel service message to chatroom in `private_chats` collection
func (df *DarkFirestore) CancelService(ctx context.Context, p CancelServiceParams) error {
	chatroomRef := df.getNewChatroomMsgRef(p.ChannelUuid)
	srvRef := df.getServiceRef(p.ServiceUuid)

	p.Data.CreatedAt = time.Now()
	p.Data.Type = CancelService

	err := df.Client.RunTransaction(
		ctx,
		func(ctx context.Context, tx *firestore.Transaction) error {
			if err := tx.Update(srvRef, []firestore.Update{
				{
					Path:  ServiceStatusFieldName,
					Value: models.ServiceStatusCanceled,
				},
				{
					Path:  ServiceCancelledCauseFieldName,
					Value: p.Cause,
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

type StartServiceParams struct {
	ChannelUuid   string
	ServiceUuid   string
	ServiceStatus models.ServiceStatus
	Data          ChatMessage
}

func (df *DarkFirestore) StartService(ctx context.Context, p StartServiceParams) error {
	srvRef := df.getServiceRef(p.ServiceUuid)
	chatRef := df.getNewChatroomMsgRef(p.ChannelUuid)

	p.Data.Type = StartService
	p.Data.CreatedAt = time.Now()

	err := df.Client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		if err := tx.Update(srvRef, []firestore.Update{
			{
				Path:  ServiceStatusFieldName,
				Value: p.ServiceStatus,
			},
		}); err != nil {
			return err
		}

		if err := tx.Set(chatRef, p.Data); err != nil {
			return nil
		}

		return nil
	})

	return err
}

type ImageMessage struct {
	ChatMessage
	ImageUrls []string `firestore:"image_urls,omitempty" json:"image_urls"`
}
type SendImageMessageParams struct {
	ChannelUuid string
	ImageUrls   []string
	From        string
}

func (df *DarkFirestore) SendImageMessage(ctx context.Context, p SendImageMessageParams) error {
	chatRef := df.getNewChatroomMsgRef(p.ChannelUuid)

	df.UpdateIsReadToFalse(ctx, UpdateIsReadParams{
		ChannelUuid: p.ChannelUuid,
		UserUuid:    p.From,
	})

	msg := ImageMessage{
		ChatMessage{
			Type:      Images,
			From:      p.From,
			Content:   "",
			CreatedAt: time.Now(),
		},
		p.ImageUrls,
	}

	if _, err := chatRef.Set(ctx, msg); err != nil {
		return nil
	}

	return nil
}

type CreateServiceParams struct {
	ServiceUuid   string `firestore:"service_uuid,omitempty" json:"service_uuid"`
	ServiceStatus string `firestore:"status,omitempty" json:"status"`
}
type PrepareToStartInquiryChatParams struct {
	InquiryUuid      string
	PickerUsername   string
	InquirerUsername string
	PickerUuid       string
	InquirerUuid     string

	// Message sender's uuid
	SenderUUID  string
	ChannelUuid string
	ServiceUuid string
}

type StartInquiryChatMessage struct {
	ChatMessage
	PickerUsername   string `firestore:"picker_username,omitempty" json:"picker_username"`
	InquirerUsername string `firestore:"inquirer_username,omitempty" json:"inquirer_username"`
}

func (df *DarkFirestore) PrepareToStartInquiryChat(ctx context.Context, p PrepareToStartInquiryChatParams) error {
	iqRef := df.getInquiryRef(p.InquiryUuid)
	chatRef := df.getChatroomRef(p.ChannelUuid)
	chatMsgRef := df.getNewChatroomMsgRef(p.ChannelUuid)
	srvRef := df.getServiceRef(p.ServiceUuid)

	err := df.Client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		// Update inquiry status.
		if err := tx.Update(
			iqRef,
			[]firestore.Update{
				{
					Path:  ChannelUuidFieldName,
					Value: p.ChannelUuid,
				},
				{
					Path:  ServiceStatusFieldName,
					Value: models.InquiryStatusChatting,
				},
				{
					Path:  ServiceUuidFieldName,
					Value: p.ServiceUuid,
				},
			},
		); err != nil {
			return err
		}

		// Create chatroom & send first message
		if err := tx.Set(chatRef, map[string]interface{}{
			"last_touch":       time.Now(),
			"inquirer_uuid":    p.InquirerUuid,
			"picker_uuid":      p.PickerUuid,
			"inquirer_is_read": true,
			"picker_is_read":   true,
		}); err != nil {
			return err
		}

		msg := StartInquiryChatMessage{
			ChatMessage: ChatMessage{
				Content:   BotInvitationChatContentInZH,
				From:      p.SenderUUID,
				Type:      BotInvitationChatText,
				CreatedAt: time.Now(),
			},
			InquirerUsername: p.InquirerUsername,
			PickerUsername:   p.PickerUsername,
		}

		if err := tx.Set(
			chatMsgRef, msg,
		); err != nil {
			return nil
		}

		log.WithFields(log.Fields{
			"chatroom_name": p.ChannelUuid,
			"updated_time":  msg.CreatedAt,
		}).Debug("Inquiry Chatroom created!")

		// Create service record
		if err := tx.Set(srvRef, CreateServiceParams{
			ServiceUuid:   p.ServiceUuid,
			ServiceStatus: string(models.ServiceStatusNegotiating),
		}); err != nil {
			return err
		}

		return nil
	})

	return err
}

type UpdateIsReadParams struct {
	ChannelUuid string
	UserUuid    string
}

func (df *DarkFirestore) UpdateIsReadToTrue(ctx context.Context, p UpdateIsReadParams) error {

	// Check whether inquirer or picker is me
	// if inquirer is me then set inquirer_is_read to true
	// if picker is me then set picker_is_read to true
	chatroomRef := df.getChatroomRef(p.ChannelUuid)
	chatroom, chatErr := chatroomRef.Get(ctx)
	if chatErr != nil {
		return nil
	}
	chatroomData := chatroom.Data()
	if chatroomData["inquirer_uuid"] == p.UserUuid {

		//  private chatroom update field "inquirer_is_read".
		_, err := chatroomRef.Update(ctx, []firestore.Update{
			{
				Path:  "inquirer_is_read",
				Value: true,
			},
		})

		if err != nil {
			return nil
		}
	}
	if chatroomData["picker_uuid"] == p.UserUuid {

		//  private chatroom update field "picker_is_read".
		_, err := chatroomRef.Update(ctx, []firestore.Update{
			{
				Path:  "picker_is_read",
				Value: true,
			},
		})

		if err != nil {
			return nil
		}
	}

	return nil
}

func (df *DarkFirestore) UpdateIsReadToFalse(ctx context.Context, p UpdateIsReadParams) error {

	// Check whether inquirer or picker is me
	// if inquirer is me then set picker_is_read to false
	// if picker is me then set inquirer_is_read to false
	senderUuid := p.UserUuid
	chatroomRef := df.getChatroomRef(p.ChannelUuid)
	chatroom, chatroomErr := chatroomRef.Get(ctx)
	if chatroomErr != nil {
		return nil
	}

	chatroomData := chatroom.Data()
	if chatroomData["inquirer_uuid"] == senderUuid {

		//  private chatroom update field "picker_is_read".
		_, err := chatroomRef.Update(ctx, []firestore.Update{
			{
				Path:  "picker_is_read",
				Value: false,
			},
		})

		if err != nil {
			return nil
		}
	}
	if chatroomData["picker_uuid"] == senderUuid {

		//  private chatroom update field "inquirer_is_read".
		_, err := chatroomRef.Update(ctx, []firestore.Update{
			{
				Path:  "inquirer_is_read",
				Value: false,
			},
		})

		if err != nil {
			return nil
		}
	}

	return nil
}
