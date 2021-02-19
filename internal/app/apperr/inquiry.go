package apperr

const (
	FailedToValidateEmitInquiryParams      = "3000001"
	OnlyMaleCanEmitInquiry                 = "3000002"
	FailedToGetInquiryByInquirerID         = "3000003"
	UserAlreadyHasActiveInquiry            = "3000004"
	FailedToCreateInquiry                  = "3000005"
	FailedToValidateCancelInquiryParams    = "3000006"
	FailedToGetInquiryByUuid               = "3000007"
	FailedToGetUserIDByUuid                = "3000008"
	UserNotOwnInquiry                      = "3000009"
	FailedToPatchInquiryStatus             = "3000010"
	InquiryFSMTransitionFailed             = "3000011"
	FailedToCheckGender                    = "3000012"
	ParamsNotProperlySetInTheMiddleware    = "3000013"
	CanNotPickupExpiredInquiry             = "3000014"
	FailedToGetInquiererByID               = "3000016"
	GirlApproveInquiry                     = "3000017"
	FailedToUpdateInquiryContent           = "3000018"
	FailedToTransformGirlApproveInquiry    = "3000019"
	OnlyFemaleCanApproveInquiry            = "3000020"
	FailedToValidateBookInquiryParams      = "3000021"
	OnlyMaleCanBookService                 = "3000022"
	FSMNotSetInMiddleware                  = "3000023"
	FailedToCheckActiveInquiry             = "3000024"
	FailedToGetInquiryList                 = "3000025"
	FailedToValidateGetInquiryListParams   = "3000026"
	FailedToTransformGetInquiriesResponse  = "3000027"
	OnlyFemaleUserCanAccessAPI             = "3000028"
	FailedToCheckHasMoreInquiry            = "3000029"
	FailedToValidateGetInquiryParams       = "3000030"
	FailedToTransformGetInquiry            = "3000031"
	FailedToPickupInquiryDueToDirtyVersion = "3000032"

	FailedToJoinLobby                    = "3000032"
	FailedToLeaveLobby                   = "3000033"
	FailedToCheckLobbyExpiry             = "3000034"
	FailedToCreateAndJoinLobby           = "3000035"
	FailedToTransformResponse            = "3000036"
	FailedToPickupInquiry                = "3000037"
	FailedToTransformServiceModel        = "3000038"
	FailedToGetInquirerByInquiryUUID     = "3000039"
	FailedToTransformInquirerResponse    = "3000040"
	FailedToGetLobbyUserByInquiryID      = "3000041"
	FailedToPickupStatusNotInquiring     = "3000042"
	FailedToAskInquiringUser             = "3000043"
	FailedToBindInquiryUriParams         = "3000044"
	FailedToChangeFirestoreInquiryStatus = "3000045"
	FailedToPickupStatusNotWaiting       = "3000046"
)

var InquiryErrCodeMsgMap = map[string]string{
	OnlyMaleCanEmitInquiry:                 "Only male user can emit inquiry",
	OnlyFemaleCanApproveInquiry:            "Only female user can approve inquiry",
	FailedToGetInquiryByInquirerID:         "Failed to get inquiry by inquirer ID",
	UserAlreadyHasActiveInquiry:            "User already has active inquiry",
	UserNotOwnInquiry:                      "User does not own the inquiry",
	FailedToPatchInquiryStatus:             "Failed to patch inquiry status",
	ParamsNotProperlySetInTheMiddleware:    "Params not properly set to the context in the previous middleware, please check",
	CanNotPickupExpiredInquiry:             "Can not pickup expired inquiry",
	OnlyMaleCanBookService:                 "Only male can book service",
	OnlyFemaleUserCanAccessAPI:             "Only female user can access API",
	FailedToPickupInquiryDueToDirtyVersion: "Inquiry has been modified by other requests. Please pick another request or try again later",
	FailedToPickupStatusNotWaiting:         "Inquiry not available, status is not waiting",
}
