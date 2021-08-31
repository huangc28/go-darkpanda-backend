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
	PublishServicePaidNotification(ctx context.Context, m PublishServicePaidNotificationMessage) error
	PublishUnpaidServiceExpiredNotification(ctx context.Context, m PublishUnpaidServiceExpiredMessage) error
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
	PickupInquiry        FCMType = "pickup_inquiry"
	ServicePaid          FCMType = "service_paid"
	UnpaidServiceExpired FCMType = "unpaid_service_expired"
)

type Notification struct {
	Type     FCMType     `json:"fcm_type"`
	Content  interface{} `json:"fcm_content"`
	DeepLink string      `json:"deep_link"`
}

const FCMImgUrl = "https://storage.googleapis.com/dark-panda-6fe35.appspot.com/fcm_logos/logo3.png"

type PublishPickupInquiryNotificationMessage struct {
	Topic      string `json:"-"`
	PickerName string `json:"picker_name"`
	PickerUUID string `json:"picker_uuid"`
}

func (r *DPFirebaseMessage) PublishPickupInquiryNotification(ctx context.Context, m PublishPickupInquiryNotificationMessage) error {
	n := Notification{
		Type:    PickupInquiry,
		Content: m,
	}

	bb, err := json.Marshal(&n)

	if err != nil {
		return err
	}

	res, err := r.c.Send(ctx, &messaging.Message{
		Topic: m.Topic,
		Notification: &messaging.Notification{
			Title:    fmt.Sprintf("%s 回覆詢問", m.PickerName),
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
	Topic       string `json:"-"`
	ServiceUUID string `json:"service_uuid"`
}

func (r *DPFirebaseMessage) PublishServicePaidNotification(ctx context.Context, m PublishServicePaidNotificationMessage) error {
	n := Notification{
		Type:     ServicePaid,
		Content:  m,
		DeepLink: "",
	}

	bd, err := json.Marshal(&n)

	if err != nil {
		return err
	}

	res, err := r.c.Send(ctx, &messaging.Message{
		Topic: m.Topic,
		Notification: &messaging.Notification{
			Title:    "服務付款完成",
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

type PublishUnpaidServiceExpiredMessage struct {
	Topic               string `json:"-"`
	ServiceUUID         string `json:"service_uuid"`
	CustomerName        string `json:"customer_name"`
	ServiceProviderName string `json:"service_provider_name"`
}

func (r *DPFirebaseMessage) PublishUnpaidServiceExpiredNotification(ctx context.Context, m PublishUnpaidServiceExpiredMessage) error {
	n := Notification{
		Type:     UnpaidServiceExpired,
		Content:  m,
		DeepLink: "",
	}

	bd, err := json.Marshal(&n)

	if err != nil {
		return err
	}

	res, err := r.c.Send(ctx, &messaging.Message{
		Topic: m.Topic,
		Notification: &messaging.Notification{
			Title:    "未付款服務已過期",
			Body:     string(bd),
			ImageURL: FCMImgUrl,
		},
	})

	if err != nil {
		return err
	}

	log.Infof("unpaid service expired FCM sent! %s", res)

	return nil
}
