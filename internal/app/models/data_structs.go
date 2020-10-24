package models

type ChatInfo struct {
	ChanelUuid string
	ChatID     int64
}

type UserWithInquiries struct {
	User
	Inquiries []*ServiceInquiry `json:"inquiries"`
}
