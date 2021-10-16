package models

import (
	"database/sql"
	"time"
)

type UserWithInquiries struct {
	User
	Inquiries []*ServiceInquiry `json:"inquiries"`
}

// InquiryChatRooms data models to be returned for method `GetFemaleInquiryChatRooms`.
type InquiryChatRoom struct {
	ServiceUuid   string         `json:"service_uuid"`
	ServiceType   InquiryStatus  `json:"service_type"`
	InquiryStatus string         `json:"inquiry_status"`
	InquiryUUID   string         `json:"inquiry_uuid"`
	InquirerUUID  string         `json:"inquirer_uuid"`
	PickerUUID    string         `json:"picker_uuid"`
	Username      string         `json:"username"`
	ChannelUUID   string         `json:"channel_uuid"`
	AvatarURL     sql.NullString `json:"avatar_url"`
	ExpiredAt     time.Time      `json:"expired_at"`
	CreatedAt     time.Time      `json:"created_at"`
}

type CompleteChatroomInfoModel struct {
	Chatroom
	ServiceType   ServiceType   `json:"service_type"`
	InquiryStatus InquiryStatus `json:"inquiry_status"`
	InquiryUuid   string        `json:"inquiry_uuid"`
	InquirerId    int           `json:"inquirer_id"`
	PickerId      int           `json:"picker_id"`
}

type ActiveInquiry struct {
	ServiceInquiry
	PickerUuid sql.NullString `json:"picker_uuid"`
}

type ServicePaymentDetail struct {
	PaymentID sql.NullInt64   `json:"payment_id"`
	Price     sql.NullFloat64 `json:"price"`

	Address         string        `json:"address"`
	AppointmentTime sql.NullTime  `json:"appointment_time"`
	Duration        sql.NullInt64 `json:"duration"`

	PickerUuid      string         `json:"picker_uuid"`
	PickerUsername  string         `json:"picker_username"`
	PickerAvatarUrl sql.NullString `json:"picker_avatar_url"`
}

type UserRatings struct {
	ServiceRating
	RaterUsername  string         `json:"rater_username"`
	RaterUuid      string         `json:"rater_uuid"`
	RaterAvatarUrl sql.NullString `json:"rater_avatar_url"`
}

type InquiryInfo struct {
	ServiceInquiry
	Inquirer    User
	ServiceUuid sql.NullString `json:"service_uuid"`
}

type PatchInquiryParams struct {
	Uuid            string         `json:"inquiry_uuid"`
	AppointmentTime *time.Time     `json:"appointment_time"`
	Budget          *float32       `json:"budget"`
	Price           *float32       `json:"price"`
	Duration        *int           `json:"duration"`
	ServiceType     *string        `json:"service_type"`
	InquiryStatus   *InquiryStatus `json:"inquiry_status"`
	Address         *string        `json:"address"`
	FcmTopic        *string        `json:"fcm_topic"`
}

type CancelUnpaidServices struct {
	Service
	CustomerFCMTopic        *string `json:"customer_fcm_topic"`
	CustomerName            string  `json:"customer_name"`
	ServiceProviderFCMTopic *string `json:"service_provider_fcm_topic"`
	ServiceProviderName     string  `json:"service_provider_name"`
}

type UserRating struct {
	RateeID int64 `json:"ratee_id"`

	// Average score of the services that the user has participated in.
	Score *float32 `json:"score"`

	// Number of services the score is calculated upon.
	NumberOfServices int `json:"number_of_services"`
}

type RandomGirl struct {
	Setseed *string `json:"setseed"`
	User

	// HasInquiry indicates whether male has had any inquiry with the girl.
	HasInquiry bool `json:"has_inquiry"`

	// InquiryUUID is the latest inquiry uuid that the male has with the girl.
	InquiryUUID *string `json:"inquiry_uuid"`

	// InquiryStatus is the latest inquiry status that the male
	InquiryStatus *string `json:"inquiry_status"`

	HasService  bool    `json:"has_service"`
	ChannelUUID *string `json:"channel_uuid"`
	ServiceUUID *string `json:"service_uuid"`

	Rating UserRating
}

type InquiryRequest struct {
	InquiryUUID   string    `json:"inquiry_uuid"`
	InquirerUUID  string    `json:"inquirer_uuid"`
	CreatedAt     time.Time `json:"created_at"`
	Username      string    `json:"username"`
	AvatarURL     string    `json:"avatar_url"`
	InquiryStatus string    `json:"inquiry_status"`
}
