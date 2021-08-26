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
	PublishPickupInquiryNotification(ctx context.Context, m PublishPickupInquiryNotificationMessage) error
	PublishServicePaidNotification(ctx context.Context, serviceUUID string) error
}

type DPFirebaseMessage struct {
	c *messaging.Client
}

func New(c *messaging.Client) *DPFirebaseMessage {
	return &DPFirebaseMessage{
		c: c,
	}
}

func MakeDedicatedFCMTopicForUser(userUUID string) string {
	curts := time.Now().UTC().Unix()

	return fmt.Sprintf("user_%s_%d", userUUID, curts)
}

func MakeTopicName(inquiryUUID string) string {
	curts := time.Now().UTC().Unix()
	topicName := fmt.Sprintf("%s_%s_%d", "inquiry", inquiryUUID, curts)

	return topicName
}

type FCMType string

var (
	PickupInquiry FCMType = "pickup_inquiry"
	ServicePaid   FCMType = "service_paid"
)

type Notification struct {
	Type     FCMType     `json:"fcm_type"`
	Content  interface{} `json:"fcm_content"`
	DeepLink string      `json:"deep_link"`
}

const FCMImgUrl = "https://storage.googleapis.com/dark-panda-6fe35.appspot.com/fcm_logos/logo3.png"

type PublishPickupInquiryNotificationMessage struct {
	Topic      string
	PickerName string
	PickerUUID string
}

func (r *DPFirebaseMessage) PublishPickupInquiryNotification(ctx context.Context, m PublishPickupInquiryNotificationMessage) error {
	type Content struct {
		PickerName string `json:"picker_name"`
		PickerUuid string `json:"picker_uuid"`
	}

	c := Content{
		PickerName: m.PickerName,
		PickerUuid: m.PickerUUID,
	}

	n := Notification{
		Type:    PickupInquiry,
		Content: c,
	}

	bb, err := json.Marshal(&n)

	if err != nil {
		return err
	}

	res, err := r.c.Send(ctx, &messaging.Message{
		Topic: m.Topic,
		Notification: &messaging.Notification{
			Title:    fmt.Sprintf("%s responded to inquiry", m.PickerName),
			Body:     string(bb),
			ImageURL: FCMImgUrl,
		},
	})

	if err != nil {
		return err
	}

	log.Infof("FCM sent! %s", res)

	return err
}

type PublishServicePaidNotificationMessage struct {
	Topic       string
	ServiceUUID string
}

func (r *DPFirebaseMessage) PublishServicePaidNotification(ctx context.Context, m PublishServicePaidNotificationMessage) error {
	type Content struct {
		ServiceUuid string `json:"service_uuid"`
	}

	c := Content{
		ServiceUuid: m.ServiceUUID,
	}

	n := Notification{
		Type:     ServicePaid,
		Content:  c,
		DeepLink: "",
	}

	bd, err := json.Marshal(&n)

	if err != nil {
		return err
	}

	res, err := r.c.Send(ctx, &messaging.Message{
		Topic: m.Topic,
		Notification: &messaging.Notification{
			Title:    "服務預約完成",
			Body:     string(bd),
			ImageURL: FCMImgUrl,
		},
	})

	if err != nil {
		return err
	}

	log.Infof("FCM sent! %s", res)

	return nil
}
