package apperr

const (
	FailedToRetrieveFormFileFromRequest = "4000001"
	FailedToCopyFileToGCS               = "4000002"
	FailedToCloseObjectWriter           = "4000003"
	FailedToSetObjectPublic             = "4000004"
	FailedToGetObjectAttrs              = "4000005"
	FailedToInitGCSClient               = "4000006"
	FailedToParseMultipartForm          = "4000007"
	FailedToOpenMultipartFile           = "4000008"
	FailedToGetImagesByUserID           = "4000009"
	FailedToUploadImagesToGCS           = "4000010"
	FailedToSendImageMessage            = "4000011"
	FailedToCropImages                  = "4000012"
)

var ImageErrCodeMap = map[string]string{}
