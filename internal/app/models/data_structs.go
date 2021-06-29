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
