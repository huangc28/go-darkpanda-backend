package models

// InquiryStatus extension methods
func (s *InquiryStatus) IsValid() bool {
	switch *s {
	case InquiryStatusBooked, InquiryStatusCanceled, InquiryStatusExpired, InquiryStatusInquiring, InquiryStatusChatting, InquiryStatusWaitForInquirerApprove:
		return true
	default:
		return false
	}
}

func (s *InquiryStatus) ToString() string {
	return string(*s)
}

// BreastSize extension methods
func (b *BreastSize) IsValid() bool {
	switch *b {
	case BreastSizeA, BreastSizeB, BreastSizeC, BreastSizeD, BreastSizeE, BreastSizeF, BreastSizeG:
		return true
	default:
		return false
	}
}

func (b *BreastSize) SQLNilBreastSize() {

}
