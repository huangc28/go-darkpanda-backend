package apperr

const (
	FailedToCreateService      = "1100001"
	FailedToUpdateService      = "1100002"
	FailedToGetIncomingService = "1100003"
	FailedToGetOverdueService  = "1100004"
	FailedToBindApiBodyParams  = "1100005"
)

var ServiceErrorMessageMap = map[string]string{
	FailedToCreateService: "Failed to create service",
}
