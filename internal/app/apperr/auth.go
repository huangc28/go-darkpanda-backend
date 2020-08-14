package apperr

const (
	FailedToValidateRegisterParams = "1000001"
	FailedToRetrieveReferCodeInfo  = "1000002"
	FailedToCheckUsernameExistence = "1000003"
	UsernameNotAvailable           = "1000004"
	FailedToCheckReferCodeExists   = "1000005"
	ReferCodeOccupied              = "1000006"
	ReferCodeNotExist              = "1000007"
	FailedToCreateUser             = "1000008"
)

var ErrCodeMsgMap = map[string]string{
	ReferCodeOccupied:    "refer code already occupied",
	UsernameNotAvailable: "username is has been registered",
	ReferCodeNotExist:    "refer code does't exist",
}
