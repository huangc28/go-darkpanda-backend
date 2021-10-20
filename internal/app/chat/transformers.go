package chat

import (
	"time"

	"github.com/huangc28/go-darkpanda-backend/internal/app/models"
	darkfirestore "github.com/huangc28/go-darkpanda-backend/internal/app/pkg/dark_firestore"
)

type ChatTransformer struct{}

func NewTransformer() *ChatTransformer {
	return &ChatTransformer{}
}

type TransformedEmitTextMessage struct {
	ChannelUUID string                    `json:"channel_uuid"`
	Message     darkfirestore.ChatMessage `json:"message"`
}

func (t *ChatTransformer) TransformEmitTextMessage(channelUUID string, msg darkfirestore.ChatMessage) TransformedEmitTextMessage {
	return TransformedEmitTextMessage{
		ChannelUUID: channelUUID,
		Message:     msg,
	}
}

type TransformedInquiryChat struct {
	ServiceUuid   string               `json:"service_uuid"`
	ServiceType   models.InquiryStatus `json:"service_type"`
	Username      string               `json:"username"`
	AvatarURL     string               `json:"avatar_url"`
	ChannelUUID   string               `json:"channel_uuid"`
	ExpiredAt     time.Time            `json:"expired_at"`
	CreatedAt     time.Time            `json:"created_at"`
	InquiryUUID   string               `json:"inquiry_uuid"`
	InquirerUUID  string               `json:"inquirer_uuid"`
	PickerUUID    string               `json:"picker_uuid"`
	InquiryStatus string               `json:"inquiry_status"`

	// Messages only contains the latest message of the chatroom. It's an empty array
	// If the chatroom does not contain any message.
	Messages []*darkfirestore.ChatMessage `json:"messages"`
}

type TransformedInquiryChats struct {
	Chats []TransformedInquiryChat `json:"chats"`
}

func (t *ChatTransformer) TransformInquiryChats(chatModels []models.InquiryChatRoom, latestMessageMap map[string][]*darkfirestore.ChatMessage) TransformedInquiryChats {
	chats := make([]TransformedInquiryChat, 0)

	for _, m := range chatModels {
		chatMsgs := []*darkfirestore.ChatMessage{}

		if v, exists := latestMessageMap[m.ChannelUUID]; exists {
			chatMsgs = v
		}

		trfm := TransformedInquiryChat{
			ServiceUuid:   m.ServiceUuid,
			ServiceType:   m.ServiceType,
			Username:      m.Username,
			ChannelUUID:   m.ChannelUUID,
			ExpiredAt:     m.ExpiredAt,
			CreatedAt:     m.CreatedAt,
			Messages:      chatMsgs,
			InquiryUUID:   m.InquiryUUID,
			InquirerUUID:  m.InquirerUUID,
			PickerUUID:    m.PickerUUID,
			InquiryStatus: m.InquiryStatus,
		}

		if m.AvatarURL.Valid {
			trfm.AvatarURL = m.AvatarURL.String
		}

		chats = append(chats, trfm)
	}

	return TransformedInquiryChats{
		Chats: chats,
	}
}

type TransformedGetHistoricalMessages struct {
	Messages []interface{} `json:"messages"`
}

func (t *ChatTransformer) TransformGetHistoricalMessages(messageData []interface{}) TransformedGetHistoricalMessages {
	return TransformedGetHistoricalMessages{
		Messages: messageData,
	}
}

type RemovedUser struct {
	UUID string `json:"uuid"`
}

type RevertedInquiry struct {
	UUID          string `json:"uuid"`
	InquiryStatus string `json:"inquiry_status"`
}

type RemovedChatroom struct {
	ChannelUUID string `json:"channel_uuid"`
}

type TransformedRevertChatting struct {
	RemovedUsers    []RemovedUser   `json:"removed_users"`
	RemovedChatroom RemovedChatroom `json:"removed_chatroom"`
	RevertedInquiry RevertedInquiry `json:"reverted_inquiry"`
}

func TransformRevertChatting(removedUsers []models.User, inquiry models.ServiceInquiry, chatroom models.Chatroom) *TransformedRevertChatting {
	rusers := make([]RemovedUser, 0)

	for _, removedUser := range removedUsers {
		ruser := RemovedUser{
			UUID: removedUser.Uuid,
		}

		rusers = append(rusers, ruser)
	}

	return &TransformedRevertChatting{
		RemovedUsers: rusers,
		RemovedChatroom: RemovedChatroom{
			ChannelUUID: chatroom.ChannelUuid.String,
		},
		RevertedInquiry: RevertedInquiry{
			UUID:          inquiry.Uuid,
			InquiryStatus: inquiry.InquiryStatus.ToString(),
		},
	}
}
