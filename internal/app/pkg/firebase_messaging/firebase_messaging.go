package dpfcm

import (
	"context"
	"fmt"
	"time"

	"firebase.google.com/go/messaging"
	log "github.com/sirupsen/logrus"
)

type DPFirebaseMessenger interface {
	PublishMaleAgreeToChat(ctx context.Context, m PublishMaleAgreeToChatMessage) error
	PublishPickupInquiryNotification(ctx context.Context, m PublishPickupInquiryNotificationMessage) error
	PublishServicePaidNotification(ctx context.Context, m PublishServicePaidNotificationMessage) error
	PublishUnpaidServiceExpiredNotification(ctx context.Context, m PublishUnpaidServiceExpiredMessage) error
	PublishServiceCancelled(ctx context.Context, m PublishServiceCancelledMessage) error
	PublishServiceRefunded(ctx context.Context, m PublishServiceRefundedMessage) error
	PublishMaleSendDirectInquiryNotification(ctx context.Context, m PublishMaleSendDirectInquiryMessage) error
	PublishServiceCompletedNotification(ctx context.Context, m ServiceCompletedMessage) error
	PublishServiceExpiredNotification(ctx context.Context, m ServiceExpiredMessage) error
}

const FCMTypeFieldName = "fcm_type"

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
	PickupInquiry         FCMType = "pickup_inquiry"
	ServicePaid           FCMType = "service_paid"
	UnpaidServiceExpired  FCMType = "unpaid_service_expired"
	AgreeToChat           FCMType = "agree_to_chat"
	ServiceCancelled      FCMType = "service_cancelled"
	Refunded              FCMType = "refunded"
	MaleSendDirectInquiry FCMType = "male_send_direct_inquiry"
	ServiceEnded          FCMType = "service_ended"
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
	data := make(map[string]string)
	data[FCMTypeFieldName] = string(PickupInquiry)
	data["picker_name"] = m.PickerName
	data["picker_uuid"] = m.PickerUUID

	res, err := r.c.Send(ctx, &messaging.Message{
		Topic: m.Topic,
		Notification: &messaging.Notification{
			Title:    fmt.Sprintf("%s 回覆詢問", m.PickerName),
			Body:     fmt.Sprintf("%s 已回覆詢問", m.PickerName),
			ImageURL: FCMImgUrl,
		},
		Data: data,
	})

	if err != nil {
		return err
	}

	log.Infof("FCM sent! %s", res)

	return err
}

type PublishServicePaidNotificationMessage struct {
	Topic       string `json:"-"`
	PayerName   string `json:"payer_name"`
	ServiceUUID string `json:"service_uuid"`
}

func (r *DPFirebaseMessage) PublishServicePaidNotification(ctx context.Context, m PublishServicePaidNotificationMessage) error {
	data := make(map[string]string)
	data[FCMTypeFieldName] = string(ServicePaid)
	data["payer_name"] = m.PayerName
	data["service_uuid"] = m.ServiceUUID

	res, err := r.c.Send(ctx, &messaging.Message{
		Topic: m.Topic,
		Notification: &messaging.Notification{
			Title:    "服務付款完成",
			Body:     fmt.Sprintf("%s 服務付款完成", m.PayerName),
			ImageURL: FCMImgUrl,
		},
		Data: data,
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
	data := make(map[string]string)
	data[FCMTypeFieldName] = string(UnpaidServiceExpired)
	data["service_uuid"] = m.ServiceUUID
	data["customer_name"] = m.CustomerName
	data["service_provider_name"] = m.ServiceProviderName

	res, err := r.c.Send(ctx, &messaging.Message{
		Topic: m.Topic,
		Notification: &messaging.Notification{
			Title:    "未付款服務已過期",
			Body:     fmt.Sprintf("與 %s 的未付款服務已過期", m.ServiceProviderName),
			ImageURL: FCMImgUrl,
		},
		Data: data,
	})

	if err != nil {
		return err
	}

	log.Infof("unpaid service expired FCM sent! %s", res)

	return nil
}

type PublishMaleAgreeToChatMessage struct {
	Topic          string `json:"-"`
	InquiryUuid    string `json:"inquiry_uuid"`
	MaleUsername   string `json:"male_username"`
	FemaleUsername string `json:"female_username"`
}

func (r *DPFirebaseMessage) PublishMaleAgreeToChat(ctx context.Context, m PublishMaleAgreeToChatMessage) error {
	data := make(map[string]string)
	data[FCMTypeFieldName] = string(AgreeToChat)
	data["inquiry_uuid"] = m.InquiryUuid
	data["male_username"] = m.MaleUsername
	data["female_username"] = m.FemaleUsername

	res, err := r.c.Send(ctx, &messaging.Message{
		Topic: m.Topic,
		Notification: &messaging.Notification{
			Title:    fmt.Sprintf("%s 接受聊天", m.MaleUsername),
			Body:     fmt.Sprintf("開始與 %s 聊聊吧", m.MaleUsername),
			ImageURL: FCMImgUrl,
		},
		Data: data,
	})

	if err != nil {
		return err
	}

	log.Infof("FCM sent! %s", res)

	return err
}

// Service with 'xxx' has been cancelled by ...
type PublishServiceCancelledMessage struct {
	Topics []string `json:"-"`

	ServiceUUID string `json:"service_uuid"`

	CancellerUUID string `json:"canceller_uuid"`

	CancellerUsername string `json:"canceller_username"`
}

// PublishServiceCancelled emits service cancelled message to both parties of the service.
func (r *DPFirebaseMessage) PublishServiceCancelled(ctx context.Context, m PublishServiceCancelledMessage) error {
	// Publish a message to customer
	data := make(map[string]string)
	data[FCMTypeFieldName] = string(ServiceCancelled)
	data["service_uuid"] = m.ServiceUUID
	data["canceller_uuid"] = m.CancellerUUID
	data["canceller_username"] = m.CancellerUsername

	for _, topic := range m.Topics {
		res, err := r.c.Send(ctx, &messaging.Message{
			Topic: topic,
			Notification: &messaging.Notification{
				Title:    fmt.Sprintf("%s 取消服務", m.CancellerUsername),
				Body:     fmt.Sprintf("%s 已取消服務", m.CancellerUsername),
				ImageURL: FCMImgUrl,
			},
			Data: data,
		})

		if err != nil {
			return err
		}

		log.Infof("cancel service FCM sent! %s", res)
	}

	return nil
}

type PublishServiceRefundedMessage struct {
	Topic       string `json:"-"`
	ServiceUUID string `json:"service_uuid"`
}

func (r *DPFirebaseMessage) PublishServiceRefunded(ctx context.Context, m PublishServiceRefundedMessage) error {
	data := make(map[string]string)
	data[FCMTypeFieldName] = string(Refunded)
	data["service_uuid"] = m.ServiceUUID

	res, err := r.c.Send(ctx, &messaging.Message{
		Topic: m.Topic,
		Notification: &messaging.Notification{
			Title:    "服務退款通知",
			Body:     "服務退款通知",
			ImageURL: FCMImgUrl,
		},

		Data: data,
	})

	if err != nil {
		return err
	}

	log.Infof("cancel service FCM sent! %s", res)

	return nil
}

type PublishMaleSendDirectInquiryMessage struct {
	Topic          string `json:"-"`
	InquiryUUID    string `json:"inquiry_uuid"`
	Femaleusername string `json:"female_username"`
	MaleUsername   string `json:"male_username"`
}

func (r *DPFirebaseMessage) PublishMaleSendDirectInquiryNotification(ctx context.Context, m PublishMaleSendDirectInquiryMessage) error {
	data := make(map[string]string)
	data[FCMTypeFieldName] = string(MaleSendDirectInquiry)
	data["inquiry_uuid"] = m.InquiryUUID
	data["female_username"] = m.Femaleusername
	data["male_username"] = m.MaleUsername

	res, err := r.c.Send(ctx, &messaging.Message{
		Topic: m.Topic,
		Notification: &messaging.Notification{
			Title:    "男生詢問 ",
			Body:     fmt.Sprintf("%s 男生向您提出詢問", m.MaleUsername),
			ImageURL: FCMImgUrl,
		},
		Data: data,
	})

	if err != nil {
		return err
	}

	log.Infof("[fcm_info] direct inquiry sent %s", res)

	return err
}

type ServiceCompletedMessage struct {
	Topic               string
	CounterPartUsername string
	ServiceUUID         string
}

// PublishServiceEndedNotification notifies both customer and service provider that the
// service has ended.
func (r *DPFirebaseMessage) PublishServiceCompletedNotification(ctx context.Context, m ServiceCompletedMessage) error {
	res, err := r.publishServiceScannedNotification(ctx, serviceScannedMessage{
		Topic:       m.Topic,
		Title:       "服務結束",
		Body:        fmt.Sprintf("您與 %s 的服務已經結束", m.CounterPartUsername),
		ServiceUUID: m.ServiceUUID,
	})

	if err != nil {
		return err
	}

	log.Infof("[fcm_info] service completed message sent %s", res)
	return nil
}

type ServiceExpiredMessage struct {
	Topic               string
	CounterPartUsername string
	ServiceUUID         string
}

func (r *DPFirebaseMessage) PublishServiceExpiredNotification(ctx context.Context, m ServiceExpiredMessage) error {
	res, err := r.publishServiceScannedNotification(ctx, serviceScannedMessage{
		Topic:       m.Topic,
		Title:       "服務過期",
		Body:        fmt.Sprintf("您與 %s 的服務已經過期", m.CounterPartUsername),
		ServiceUUID: m.ServiceUUID,
	})

	if err != nil {
		return err
	}

	log.Infof("[fcm_info] service expired message sent %s", res)

	return nil
}

type serviceScannedMessage struct {
	Topic       string
	Title       string
	Body        string
	ServiceUUID string
}

func (r *DPFirebaseMessage) publishServiceScannedNotification(ctx context.Context, m serviceScannedMessage) (string, error) {
	data := make(map[string]string)
	data[FCMTypeFieldName] = string(ServiceEnded)
	data["service_uuid"] = m.ServiceUUID

	res, err := r.c.Send(ctx, &messaging.Message{
		Topic: m.Topic,
		Notification: &messaging.Notification{
			Title:    m.Title,
			Body:     m.Body,
			ImageURL: FCMImgUrl,
		},
		Data: data,
	})

	if err != nil {
		return "", err
	}

	return res, nil
}
