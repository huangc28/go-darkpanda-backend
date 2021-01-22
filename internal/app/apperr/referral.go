package apperr

const (
	FailedToValidateVerifyReferralCodeParams = "1200001"
	FailedToGetReferralCode                  = "1200002"
	ReferralCodeNotFound                     = "1200003"
	ReferralCodeIsOccupied                   = "1200004"
	FailedToUpdateReferralcode               = "1200005"
	ReferralCodeExpired                      = "1200006"
)

var ReferralErrorMessageMap = map[string]string{
	ReferralCodeNotFound:   "Referral code given is not found",
	ReferralCodeIsOccupied: "Referral code is occupied",
	ReferralCodeExpired:    "Referral code has expired",
}
