package apperr

var (
	FailedToGetServicePartnerInfo = "1900001"
	FailedToGetServiceRating      = "1900002"
	NotInvolveInService           = "1900003"
	ServiceNotRatable             = "1900004"
	FailedToCreateServiceRating   = "1900005"
	FailedToGetRatings            = "1900006"
)

var RatingErrCodeMsgMap = map[string]string{
	NotInvolveInService: "requester no involved in service.",
}
