package apperr

const (
	FailedToValidatePutUserParams           = "5000001"
	FailedToValidateUserURIParams           = "5000002"
	FailedToPatchUserInfo                   = "5000003"
	FailedToValidateGetUserProfileParams    = "5000004"
	FailedToGetImagesByUserUUID             = "5000005"
	FailedToValidateGetUserImagesParams     = "5000006"
	FailedToGetUserPayments                 = "5000007"
	FailedToTransformUserPayments           = "5000008"
	FailedToValidateGetServiceHistoryParams = "5000009"
	FailedToGetHistoricalServices           = "5000010"
	FailedToTransformHistoricalServices     = "5000011"
	FailedToGetUserByID                     = "5000012"
	FailedToCreateChangeMobileVerifyCode    = "5000013"
	FailedToSendTwilioMessage               = "5000014"
	FailedToGetChangeMobileVerifyCode       = "5000015"
	ChangeMobileVerifyCodeNotExists         = "5000016"
	ChangeMobileVerifyCodeNotMatching       = "5000017"
)

var userErrorCodeMsgMap = map[string]string{
	ChangeMobileVerifyCodeNotExists:   "verify code does not exist, please send verify code via mobile again.",
	ChangeMobileVerifyCodeNotMatching: "verify code does not match.",
}
