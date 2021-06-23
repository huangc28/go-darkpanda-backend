package apperr

const (
	UnknownErrorToApplication           = "0000001"
	FailedToBeginTx                     = "0000002"
	FailedToCommitTx                    = "0000003"
	FailedToValidateRequestBody         = "0000004"
	FailedToConvertNullSQLStringToFloat = "0000005"
	FailedToBindJwtInHeader             = "0000006"
	MissingAuthToken                    = "0000007"
	FailedToBindBodyParams              = "0000008"
	FailedToParsePaginateParams         = "0000009"
	AssetNotFound                       = "0000010"
)

var GeneralErrorMessageMap = map[string]string{
	MissingAuthToken: "missing auth token",
	AssetNotFound:    "query results no asset found.",
}
