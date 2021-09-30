package apperr

const (
	FailedToInitAppcenterRequest        = "4000001"
	FailedToSendAppcenterOpenApiRequest = "4000002"
	FailedToScanAppcenterResponse       = "4000003"
)

var releaseErrorMap = map[string]string{
	FailedToInitAppcenterRequest: "failed to init appcenter request",
}
