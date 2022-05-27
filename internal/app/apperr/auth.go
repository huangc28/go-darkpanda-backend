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
	FailedToGetServiceOption                 = "1000046"
)

var AuthErrCodeMsgMap = map[string]string{
	ReferCodeOccupied:                       "參考代碼已被佔用",
	UsernameNotAvailable:                    "用戶名不可用",
	ReferCodeNotExist:                       "參考代碼不存在",
	UserHasPhoneVerified:                    "用戶已通過電話驗證",
	UserNotFoundByVerifyCode:                "給定的驗證碼找不到用戶資料",
	JWTNotProvided:                          "JWT token not exists",
	FailedToFindInquiryByInquiererID:        "查詢ID失敗",
	FailedToCheckSendLoginVerifyCodeParams:  "找不到用戶名發送驗證碼",
	UnableToSendVerifyCodeToUnverfiedNumber: "無法將登錄代碼發送到未經驗證的手機號碼。請聯繫客服",
	ExceedingLoginRetryLimit:                "嘗試登錄太多次。請稍後重試登錄",
	VerifyCodeUnmatched:                     "手機驗證碼不匹配",
	LoginVerifyCodeNotFound:                 "未找到驗證器的登錄驗證碼，請重新發送新的短信登錄碼",
	TokenIsInvalidated:                      "jwt token is invalid",
}
