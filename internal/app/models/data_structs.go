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
	ServiceType InquiryStatus  `json:"service_type"`
	Username    string         `json:"username"`
	AvatarURL   sql.NullString `json:"avatar_url"`
	ChannelUUID string         `json:"channel_uuid"`
	ExpiredAt   time.Time      `json:"expired_at"`
	CreatedAt   time.Time      `json:"created_at"`
}
