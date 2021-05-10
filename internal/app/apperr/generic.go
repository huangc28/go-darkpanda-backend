package apperr

const (
	UnknownErrorToApplication           = "0000001"
	FailedToBeginTx                     = "0000002"
	FailedToCommitTx                    = "0000003"
	FailedToValidateRequestBody         = "0000004"
	FailedToConvertNullSQLStringToFloat = "0000005"
	FailedToBindJwtInHeader             = "0000006"
	MissingAuthToken                    = "0000007"
)

var GeneralErrorMessageMap = map[string]string{
	MissingAuthToken: "Missing auth token",
}
