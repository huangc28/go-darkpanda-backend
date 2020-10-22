package darkpubnub

import (
	pubnub "github.com/pubnub/go"
)

var _pubnubInstance *pubnub.PubNub

type Config struct {
	PublishKey   string
	SubscribeKey string
	SecretKey    string
	UUID         string
}

func NewPubnub(config Config) *pubnub.PubNub {
	initPubnub(config)

	return _pubnubInstance
}

func initPubnub(config Config) {
	punubConfig := pubnub.NewConfig()
	punubConfig.PublishKey = config.PublishKey
	punubConfig.SubscribeKey = config.SubscribeKey
	punubConfig.SecretKey = config.SecretKey
	punubConfig.UUID = config.UUID

	_pubnubInstance = pubnub.NewPubNub(punubConfig)
}

func GetPubnub() *pubnub.PubNub {
	return _pubnubInstance
}

type TextMessage struct {
	Content string
}

func FormatTextMessage(params TextMessage) map[string]interface{} {
	msg := map[string]interface{}{
		"content": params.Content,
	}

	return msg
}

func PubnubTimestampToUnix(t int64) int64 {
	return t / 1e7
}
