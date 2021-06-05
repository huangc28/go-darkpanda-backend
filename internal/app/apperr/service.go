package apperr

const (
	FailedToCreateService      = "1100001"
	FailedToUpdateService      = "1100002"
	FailedToGetIncomingService = "1100003"
	FailedToGetOverdueService  = "1100004"
	FailedToBindApiBodyParams  = "1100005"

	FailedServiceQrCodeSecretNotMatch = "1100006"
	FailedToGetServiceByQrCodeUuid    = "1100007"
	NotAServiceParticipant            = "1100008"
	InvalidServiceStatus              = "1100009"
	FailedToChangeServiceStatus       = "1100010"

	FirestoreFailedToUpdateService = "1100011"
	FailedToGetQrCodeByServiceUuid = "1100012"
	NoQRCodeFound                  = "1100013"
	FailedToGetServiceNames        = "1100014"
)

var ServiceErrorMessageMap = map[string]string{
	FailedToCreateService:             "Failed to create service.",
	FailedServiceQrCodeSecretNotMatch: "QR code secret does not match.",
	NotAServiceParticipant:            "Scanner is not a service participant",
}
