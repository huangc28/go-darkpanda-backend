package apperr

const (
	FailedToValidateEmitInquiryParams   = "3000001"
	OnlyMaleCanEmitInquiry              = "3000002"
	FailedToGetInquiryByInquirerID      = "3000003"
	UserAlreadyHasActiveInquiry         = "3000004"
	FailedToCreateInquiry               = "3000005"
	FailedToValidateCancelInquiryParams = "3000006"
	FailedToGetInquiryByUuid            = "3000007"
	FailedToGetUserIDByUuid             = "3000008"
	UserNotOwnInquiry                   = "3000009"
	FailedToPatchInquiryStatus          = "3000010"
	InquiryFSMTransitionFailed          = "3000011"
	FailedToCheckGender                 = "3000012"
	ParamsNotProperlySetInTheMiddleware = "3000013"
)

var InquiryErrCodeMsgMap = map[string]string{
	OnlyMaleCanEmitInquiry:              "Only male user can emit inquiry",
	FailedToGetInquiryByInquirerID:      "Failed to get inquiry by inquirer ID",
	UserAlreadyHasActiveInquiry:         "User already has active inquiry",
	UserNotOwnInquiry:                   "User does not own the inquiry",
	FailedToPatchInquiryStatus:          "Failed to patch inquiry status",
	ParamsNotProperlySetInTheMiddleware: "Params not properly set to the context in the previous middleware, please check",
}
