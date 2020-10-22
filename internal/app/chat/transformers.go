package chat

import "time"

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
