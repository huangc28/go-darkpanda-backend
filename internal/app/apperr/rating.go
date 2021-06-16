package apperr

var (
	FailedToGetServicePartnerInfo = "1900001"
	FailedToGetServiceRating      = "1900002"
	NotInvolveInService           = "1900003"
)

var RatingErrCodeMsgMap = map[string]string{
	NotInvolveInService: "requester no involved in service.",
}
