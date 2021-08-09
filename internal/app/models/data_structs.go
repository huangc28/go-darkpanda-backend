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
	Price sql.NullFloat64 `json:"price"`

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
}
