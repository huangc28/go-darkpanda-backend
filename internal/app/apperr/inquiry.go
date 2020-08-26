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
	CanNotPickupExpiredInquiry          = "3000014"
	FailedToCreateService               = "3000015"
	FailedToGetInquiererByID            = "3000016"
	GirlApproveInquiry                  = "3000017"
	FailedToUpdateInquiryContent        = "3000018"
	FailedToTransformGirlApproveInquiry = "3000019"
	OnlyFemaleCanApproveInquiry         = "3000020"
	FailedToValidateBookInquiryParams   = "3000021"
)

var InquiryErrCodeMsgMap = map[string]string{
	OnlyMaleCanEmitInquiry:              "Only male user can emit inquiry",
	OnlyFemaleCanApproveInquiry:         "Only male user can approve inquiry",
	FailedToGetInquiryByInquirerID:      "Failed to get inquiry by inquirer ID",
	UserAlreadyHasActiveInquiry:         "User already has active inquiry",
	UserNotOwnInquiry:                   "User does not own the inquiry",
	FailedToPatchInquiryStatus:          "Failed to patch inquiry status",
	ParamsNotProperlySetInTheMiddleware: "Params not properly set to the context in the previous middleware, please check",
	CanNotPickupExpiredInquiry:          "Can not pickup expired inquiry",
	FailedToCreateService:               "Failed to create service",
}
