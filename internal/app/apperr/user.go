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
	FailedToGetRegisterMobileVerifyCode     = "5000018"
	FailedToGetUserRating                   = "5000019"
	FailedToGetUserProfiles                 = "5000020"
	FailedToTransformGirlProfile            = "5000021"
	FailedToGetGirlIDOfDirectInquiry        = "5000022"
	FailedToCreateUserServiceOption         = "5000023"
	FailedToCreateServiceOption             = "5000024"
	FailedToGetUserServiceOption            = "5000025"
	FailedToCheckServiceOptionExistence     = "5000026"
	ServiceOptionNotAvailable               = "5000027"
	FailedToGetGirlsInfo                    = "5000028"
)

var userErrorCodeMsgMap = map[string]string{
	ChangeMobileVerifyCodeNotExists:     "verify code does not exist, please send verify code via mobile again",
	ChangeMobileVerifyCodeNotMatching:   "verify code does not match",
	FailedToGetRegisterMobileVerifyCode: "verify code not found, please resend verify code again",
	ServiceOptionNotAvailable:           "Service option exists",
}
