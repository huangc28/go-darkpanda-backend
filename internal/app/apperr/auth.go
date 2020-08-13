package apperr

const (
	FailedToValidateRegisterParams = "1000001"
	FailedToRetrieveReferCodeInfo  = "1000002"
	FailedToCheckUsernameExistence = "1000003"
	UsernameNotAvailable           = "1000004"
	ReferCodeOccupied              = "1000005"
)

var ErrCodeMsgMap = map[string]string{
	ReferCodeOccupied:    "refer code already occupied",
	UsernameNotAvailable: "username is has been registered",
}
