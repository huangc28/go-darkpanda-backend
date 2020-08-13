package apperr

const (
	FailedToValidateRegisterParams = "1000001"
	FailedToRetrieveReferCodeInfo  = "1000002"
	ReferCodeOccupied              = "1000002"
)

var ErrCodeMsgMap = map[string]string{
	ReferCodeOccupied: "refer code already occupied",
}
