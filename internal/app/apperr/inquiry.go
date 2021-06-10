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

	FailedToJoinLobby                        = "3000032"
	FailedToLeaveLobby                       = "3000033"
	FailedToCheckLobbyExpiry                 = "3000034"
	FailedToCreateAndJoinLobby               = "3000035"
	FailedToTransformResponse                = "3000036"
	FailedToPickupInquiry                    = "3000037"
	FailedToTransformServiceModel            = "3000038"
	FailedToGetInquirerByInquiryUUID         = "3000039"
	FailedToTransformInquirerResponse        = "3000040"
	FailedToGetLobbyUserByInquiryID          = "3000041"
	FailedToPickupStatusNotInquiring         = "3000042"
	FailedToAskInquiringUser                 = "3000043"
	FailedToBindInquiryUriParams             = "3000044"
	FailedToChangeFirestoreInquiryStatus     = "3000045"
	FailedToPickupStatusNotWaiting           = "3000046"
	FailedToCreateFSM                        = "3000047"
	FailedToCreatePrivateChatroomInFirestore = "3000048"
	FailedToUpdateInquiry                    = "3000049"
	InquiryUUIDNotInParams                   = "3000050"
	FailedToValidatePatchInquiryParams       = "3000051"
	FailedToPatchInquiry                     = "3000052"
	FailedToTransformUpdateInquiry           = "3000053"
	FailedToActiveInquiry                    = "3000054"
	FailedToTransformActiveInquiry           = "3000055"
	NoActiveInquiry                          = "3000056"
	InquiryHasNoPicker                       = "3000057"
)

var InquiryErrCodeMsgMap = map[string]string{
	OnlyMaleCanEmitInquiry:                 "only male user can emit inquiry",
	OnlyFemaleCanApproveInquiry:            "only female user can approve inquiry",
	FailedToGetInquiryByInquirerID:         "failed to get inquiry by inquirer ID",
	UserAlreadyHasActiveInquiry:            "user already has active inquiry",
	UserNotOwnInquiry:                      "user does not own the inquiry",
	FailedToPatchInquiryStatus:             "failed to patch inquiry status",
	ParamsNotProperlySetInTheMiddleware:    "params not properly set to the context in the previous middleware, please check",
	CanNotPickupExpiredInquiry:             "can not pickup expired inquiry",
	OnlyMaleCanBookService:                 "only male can book service",
	OnlyFemaleUserCanAccessAPI:             "only female user can access API",
	FailedToPickupInquiryDueToDirtyVersion: "inquiry has been modified by other requests. Please pick another request or try again later",
	FailedToPickupStatusNotWaiting:         "inquiry not available, status is not waiting",
	InquiryUUIDNotInParams:                 "Inquiry uuid not exists in uri param",
	FailedToPickupStatusNotInquiring:       "can not pickup inquiry since status is not inquiring",
	InquiryHasNoPicker:                     "can not start a chat since inquiry has not picker.",
}
