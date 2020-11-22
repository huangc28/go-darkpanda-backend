package apperr

const (
	FailedToCreateService = "1100001"
	FailedToUpdateService = "1100002"
)

var ServiceErrorMessageMap = map[string]string{
	FailedToCreateService: "Failed to create service",
}
