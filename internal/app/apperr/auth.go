package apperr

const (
	FailedToValidateRegisterParams           = "1000001"
	FailedToRetrieveReferCodeInfo            = "1000002"
	FailedToCheckUsernameExistence           = "1000003"
	UsernameNotAvailable                     = "1000004"
	FailedToCheckReferCodeExists             = "1000005"
	ReferCodeOccupied                        = "1000006"
	ReferCodeNotExist                        = "1000007"
	FailedToCreateUser                       = "1000008"
	FailedToGenerateUuid                     = "1000009"
	FailedToValidateSendVerifyCodeParams     = "1000010"
	FailedToGetUserByUuid                    = "1000011"
	UserHasPhoneVerified                     = "1000012"
	FailedToUpdateVerifyCode                 = "1000013"
	TwilioRespErr                            = "1000014"
	FailedToSendTwilioSMSErr                 = "1000015"
	FailedToValidateVerifyPhoneParams        = "1000016"
	FailedToGetUserByVerifyCode              = "1000017"
	UserNotFoundByVerifyCode                 = "1000018"
	VerifyCodeNotMatching                    = "1000019"
	FailedToUpdateVerifyStatus               = "1000020"
	FailedToGenerateJwtToken                 = "1000021"
	FailedToValidateRevokeJwtParams          = "1000022"
	InvalidSignature                         = "1000023"
	FailedToParseSignature                   = "1000024"
	InvalidSigature                          = "1000025"
	FailedToInvalidateSignature              = "1000026"
	JWTNotProvided                           = "1000027"
	FailedToFindInquiryByInquiererID         = "1000028"
	FailedToCheckSendLoginVerifyCodeParams   = "1000029"
	FailedToGetUserByUsername                = "1000030"
	UnableToSendVerifyCodeToUnverfiedNumber  = "1000031"
	UnableToCreateSendVerifyCode             = "1000032"
	FailedToCreateAuthenticatorRecordInRedis = "1000033"
	ExceedingLoginRetryLimit                 = "1000034"
	FailedToUpdateAuthenticatorRecordInRedis = "1000035"
	FailedToValidateVerifyLoginParams        = "1000036"
	VerifyCodeUnmatched                      = "1000037"
	FailedToCreateJWTToken                   = "1000038"
	FailedToValidateReferralCode             = "1000039"
	FailedToValidateFindByUsernameParams     = "1000040"
	LoginVerifyCodeNotFound                  = "1000041"
	FailedToGetAuthenticatorRecord           = "1000042"
	FailedToParseJwtToken                    = "1000043"
	FailedToValidateToken                    = "1000044"
	TokenIsInvalidated                       = "1000045"
)

var AuthErrCodeMsgMap = map[string]string{
	ReferCodeOccupied:                       "refer code already occupied",
	UsernameNotAvailable:                    "username is not available",
	ReferCodeNotExist:                       "refer code does't exist",
	UserHasPhoneVerified:                    "user is phone verified",
	UserNotFoundByVerifyCode:                "user not found by the given verify code",
	JWTNotProvided:                          "JWT token not exists",
	FailedToFindInquiryByInquiererID:        "failed to find inquiry by inquirer ID",
	FailedToCheckSendLoginVerifyCodeParams:  "failed to find username to send verify code ",
	UnableToSendVerifyCodeToUnverfiedNumber: "can not send login code to an unverified mobile number. Please contact customer service",
	ExceedingLoginRetryLimit:                "attempt login too many times. Please retry login later",
	VerifyCodeUnmatched:                     "mobile verify code not matched",
	LoginVerifyCodeNotFound:                 "login verify code not found for the authenticator, please send a new sms login code again",
}
