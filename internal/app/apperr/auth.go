package apperr

const (
	FailedToValidateRegisterParams       = "1000001"
	FailedToRetrieveReferCodeInfo        = "1000002"
	FailedToCheckUsernameExistence       = "1000003"
	UsernameNotAvailable                 = "1000004"
	FailedToCheckReferCodeExists         = "1000005"
	ReferCodeOccupied                    = "1000006"
	ReferCodeNotExist                    = "1000007"
	FailedToCreateUser                   = "1000008"
	FailedToGenerateUuid                 = "1000009"
	FailedToValidateSendVerifyCodeParams = "1000010"
	FailedToGetUserByUuid                = "1000011"
	UserHasPhoneVerified                 = "1000012"
	FailedToUpdateVerifyCode             = "1000013"
	TwilioRespErr                        = "1000014"
	FailedToSendTwilioSMSErr             = "1000015"
	FailedToValidateVerifyPhoneParams    = "1000016"
	FailedToGetUserByVerifyCode          = "1000017"
	UserNotFoundByVerifyCode             = "1000018"
	VerifyCodeNotMatching                = "1000019"
	FailedToUpdateVerifyStatus           = "1000020"
)

var ErrCodeMsgMap = map[string]string{
	ReferCodeOccupied:        "refer code already occupied",
	UsernameNotAvailable:     "username is has been registered",
	ReferCodeNotExist:        "refer code does't exist",
	UserHasPhoneVerified:     "user is phone verified",
	UserNotFoundByVerifyCode: "user not found by the given verify code",
}
