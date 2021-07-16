package apperr

const (
	FailedToValidateVerifyUsernameParams   = "1800001"
	FailedToVerifyReferralCode             = "1800002"
	FailedToSendMobileVerifyCode           = "1800003"
	UserNotFoundByUuid                     = "1800004"
	FailedToCreateRegisterMobileVerifyCode = "1800005"
	UserAlreadyMobileVerified              = "1800006"
	PhoneVerifyCodeNotMatch                = "1800007"
	FailedToUpdateUserByUuid               = "1800008"
	FailedToUpdateInviteeIdByRefCode       = "1800009"
)

var RegisterErrCodeMsgMap = map[string]string{
	PhoneVerifyCodeNotMatch:   "mobile verify code does not match",
	UserAlreadyMobileVerified: "user mobile has already verified",
}
