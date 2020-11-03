package models

import (
	"database/sql"
	"time"
)

type UserWithInquiries struct {
	User
	Inquiries []*ServiceInquiry `json:"inquiries"`
}

// InquiryChatRooms data models to be returned for method `GetFemaleInquiryChatRooms1.
type InquiryChatRoom struct {
	ServiceType   InquiryStatus  `json:"service_type"`
	InquiryStatus string         `json:"inquiry_status"`
	Username      string         `json:"username"`
	ChannelUUID   string         `json:"channel_uuid"`
	InquirerUUID  string         `json:"inquirer_uuid"`
	AvatarURL     sql.NullString `json:"avatar_url"`
	ExpiredAt     time.Time      `json:"expired_at"`
	CreatedAt     time.Time      `json:"created_at"`
}
