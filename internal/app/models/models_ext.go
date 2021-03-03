package models

// InquiryStatus extension methods
func (s *InquiryStatus) IsValid() bool {
	switch *s {
	case InquiryStatusBooked,
		InquiryStatusCanceled,
		InquiryStatusExpired,
		InquiryStatusInquiring,
		InquiryStatusChatting,
		InquiryStatusWaitForInquirerApprove,
		InquiryStatusAsking:

		return true
	default:
		return false
	}
}

func (s InquiryStatus) ToString() string {
	return string(s)
}

func (st *ServiceType) ToString() string {
	return string(*st)
}

func (st *ServiceStatus) ToString() string {
	return string(*st)
}
