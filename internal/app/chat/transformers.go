package chat

import (
	"time"

	"github.com/huangc28/go-darkpanda-backend/internal/app/models"
)

type ChatTransformer struct{}

func NewTransformer() *ChatTransformer {
	return &ChatTransformer{}
}

type TransformedEmitTextMessage struct {
	Timestamp time.Time
	Content   string
}

type TransformEmitTextMessageParams struct {
	Timestamp time.Time
	Content   string
}

func (t *ChatTransformer) TransformEmitTextMessage(params TransformEmitTextMessageParams) TransformedEmitTextMessage {
	return TransformedEmitTextMessage{
		Timestamp: params.Timestamp,
		Content:   params.Content,
	}
}

type TransformedInquiryChat struct {
	ServiceType models.InquiryStatus `json:"service_type"`
	Username    string               `json:"username"`
	AvatarURL   *string              `json:"avatar_url"`
	ChannelUUID string               `json:"channel_uuid"`
	ExpiredAt   time.Time            `json:"expired_at"`
	CreatedAt   time.Time            `json:"created_at"`
}

type TransformedInquiryChats struct {
	Chats []TransformedInquiryChat `json:"chats"`
}

func (t *ChatTransformer) TransformInquiryChats(chatModels []models.InquiryChatRoom) TransformedInquiryChats {
	chats := make([]TransformedInquiryChat, 0)

	for _, m := range chatModels {
		trfm := TransformedInquiryChat{
			ServiceType: m.ServiceType,
			Username:    m.Username,
			ChannelUUID: m.ChannelUUID,
			ExpiredAt:   m.ExpiredAt,
			CreatedAt:   m.CreatedAt,
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
