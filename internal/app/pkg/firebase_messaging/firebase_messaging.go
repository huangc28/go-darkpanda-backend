package dpfcm

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"firebase.google.com/go/messaging"
	log "github.com/sirupsen/logrus"
)

type DPFirebaseMessenger interface {
	PublishPickupInquiryNotification(ctx context.Context, m Message) error
}

type DPFirebaseMessage struct {
	c *messaging.Client
}

func New(c *messaging.Client) *DPFirebaseMessage {
	return &DPFirebaseMessage{
		c: c,
	}
}

func MakeTopicName(inquiryUUID string) string {
	curts := time.Now().UTC().Unix()
	topicName := fmt.Sprintf("%s_%s_%d", "inquiry", inquiryUUID, curts)

	return topicName
}

type FCMType string

var (
	PickupInquiry FCMType = "pickup_inquiry"
)

type Message struct {
	TopicName  string
	PickerName string
}

const FCMImgUrl = "https://storage.googleapis.com/dark-panda-6fe35.appspot.com/fcm_logos/logo3.png"

func (r *DPFirebaseMessage) PublishPickupInquiryNotification(ctx context.Context, m Message) error {
	type Body struct {
		Type     FCMType `json:"fcm_type"`
		Content  string  `json:"fcm_content"`
		DeepLink string  `json:"deep_link"`
	}

	body := Body{
		Type:    PickupInquiry,
		Content: fmt.Sprintf("%s has picked up your inquiry.", m.PickerName),
	}

	bb, err := json.Marshal(&body)

	if err != nil {
		return err
	}

	res, err := r.c.Send(ctx, &messaging.Message{
		Topic: m.TopicName,
		Notification: &messaging.Notification{
			Title:    fmt.Sprintf("%s responds to inquiry", m.PickerName),
			Body:     string(bb),
			ImageURL: FCMImgUrl,
		},
	})

	log.Infof("FCM sent! %s", res)

	if err != nil {
		return err
	}

	return err
}
