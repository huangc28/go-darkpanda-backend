package apperr

const (
	FailedToValidateEmitInquiryParams = "3000001"
	OnlyMaleCanEmitInquiry            = "3000002"
	FailedToGetInquiryByInquirerID    = "3000003"
	UserAlreadyHasActiveInquiry       = "3000004"
	FailedToCreateInquiry             = "3000005"
)

var InquiryErrCodeMsgMap = map[string]string{
	OnlyMaleCanEmitInquiry:         "Only male user can emit inquiry",
	FailedToGetInquiryByInquirerID: "Failed to get inquiry by inquirer ID",
	UserAlreadyHasActiveInquiry:    "User already has active inquiry",
}
