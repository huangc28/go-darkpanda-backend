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
	ServiceType   models.InquiryStatus      `json:"service_type"`
	Username      string                    `json:"username"`
	AvatarURL     *string                   `json:"avatar_url"`
	ChannelUUID   string                    `json:"channel_uuid"`
	ExpiredAt     time.Time                 `json:"expired_at"`
	CreatedAt     time.Time                 `json:"created_at"`
	LatestMessage darkfirestore.ChatMessage `json:"latest_message"`
	InquiryUUID   string                    `json:"inquiry_uuid"`
}

type TransformedInquiryChats struct {
	Chats []TransformedInquiryChat `json:"chats"`
}

func (t *ChatTransformer) TransformInquiryChats(chatModels []models.InquiryChatRoom, latestMessageMap map[string]darkfirestore.ChatMessage) TransformedInquiryChats {
	chats := make([]TransformedInquiryChat, 0)

	for _, m := range chatModels {
		chatMsg := darkfirestore.ChatMessage{}

		if v, exists := latestMessageMap[m.ChannelUUID]; exists {
			chatMsg = v
		}

		trfm := TransformedInquiryChat{
			ServiceType:   m.ServiceType,
			Username:      m.Username,
			ChannelUUID:   m.ChannelUUID,
			ExpiredAt:     m.ExpiredAt,
			CreatedAt:     m.CreatedAt,
			LatestMessage: chatMsg,
			InquiryUUID:   m.InquiryUUID,
		}

		if m.AvatarURL.Valid {
			trfm.AvatarURL = &m.AvatarURL.String
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

type TransformedSendConfirmedServiceMessage struct {
}

func (t *ChatTransformer) TransformSendConfirmedServiceMessage() {}
