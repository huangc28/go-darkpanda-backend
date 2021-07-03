package apperr

const (
	FailedToValidateVerifyReferralCodeParams = "1700001"
	FailedToGetReferralCode                  = "1700002"
	ReferralCodeNotFound                     = "1700003"
	ReferralCodeIsOccupied                   = "1700004"
	FailedToUpdateReferralcode               = "1700005"
	ReferralCodeExpired                      = "1700006"
	FailedToGetOccupiedRefcode               = "1700007"
	FailedToCreateReferralCode               = "1700008"
)

var ReferralErrorMessageMap = map[string]string{
	ReferralCodeNotFound:   "Referral code given is not found",
	ReferralCodeIsOccupied: "Referral code is occupied",
	ReferralCodeExpired:    "Referral code has expired",
}
