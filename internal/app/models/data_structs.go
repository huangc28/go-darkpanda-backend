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
	Price      float64 `json:"price"`
	RecTradeID string  `json:"rec_trade_id"`

	Address   string        `json:"address"`
	StartTime time.Time     `json:"start_time"`
	Duration  sql.NullInt64 `json:"duration"`

	PickerUuid      string         `json:"picker_uuid"`
	PickerUsername  string         `json:"picker_username"`
	PickerAvatarUrl sql.NullString `json:"picker_avatar_url"`
}
