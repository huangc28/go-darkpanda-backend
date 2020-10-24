package darkpubnub

import (
	"time"

	pubnub "github.com/pubnub/go"
)

var _darkPubNubInstance *DarkPubNub

type Config struct {
	PublishKey   string
	SubscribeKey string
	SecretKey    string
	UUID         string
}

type DarkPubNub struct {
	pn *pubnub.PubNub
}

func NewDarkPubNub(config Config) *DarkPubNub {
	initDarkPubNub(config)

	return _darkPubNubInstance
}

func initDarkPubNub(config Config) {
	punubConfig := pubnub.NewConfig()
	punubConfig.PublishKey = config.PublishKey
	punubConfig.SubscribeKey = config.SubscribeKey
	punubConfig.SecretKey = config.SecretKey
	punubConfig.UUID = config.UUID

	_darkPubNubInstance = &DarkPubNub{
		pn: pubnub.NewPubNub(punubConfig),
	}
}

func (p *DarkPubNub) SendTextMessage(channel string, m TextMessage) (time.Time, error) {
	tm := FormatTextMessage(m)
	pubResp, _, err := p.pn.
		Publish().
		Channel(channel).
		Message(tm).
		Execute()

	if err != nil {
		return time.Time{}, err
	}

	return time.Unix(PubnubTimestampToUnix(pubResp.Timestamp), 0), nil
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
