package apperr

const (
	FailedToRetrieveFormFileFromRequest = "4000001"
	FailedToCopyFileToGCS               = "4000002"
	FailedToCloseObjectWriter           = "4000003"
	FailedToSetObjectPublic             = "4000004"
	FailedToGetObjectAttrs              = "4000005"
	FailedToInitGCSClient               = "4000006"
)

var ImageErrCodeMap = map[string]string{}
