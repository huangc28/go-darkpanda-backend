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

func (s *Service) IsOneOfStatus(types ...ServiceStatus) bool {
	for _, st := range types {
		if s.ServiceStatus == st {
			return true
		}
	}

	return false
}

func (s *Service) IsNotOneOfStatus(types ...ServiceStatus) bool {
	return !s.IsOneOfStatus(types...)
}

func (s *Service) GetPartnerId(myId int64) int32 {
	if s.CustomerID.Int32 == int32(myId) {
		return s.ServiceProviderID.Int32
	}

	return s.ServiceProviderID.Int32
}
