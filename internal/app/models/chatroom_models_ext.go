package models

type CompleteChatroomInfoModel struct {
	Chatroom
	ServiceType   ServiceType   `json:"service_type"`
	InquiryStatus InquiryStatus `json:"inquiry_status"`
	InquiryUuid   string        `json:"inquiry_uuid"`
	InquirerId    int           `json:"inquirer_id"`
	PickerId      int           `json:"picker_id"`
}
