package models

// InquiryStatus extension methods
func (s *InquiryStatus) IsValid() bool {
	switch *s {
	case InquiryStatusBooked, InquiryStatusCanceled, InquiryStatusExpired, InquiryStatusInquiring:
		return true
	default:
		return false
	}
}

func (s *InquiryStatus) ToString() string {
	return string(*s)
}
