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

	FirestoreFailedToUpdateService  = "1100011"
	FailedToGetQrCodeByServiceUuid  = "1100012"
	NoQRCodeFound                   = "1100013"
	FailedToGetServiceNames         = "1100014"
	FailedToGetServiceByUuid        = "1100015"
	ServiceNotYetEnd                = "1100016"
	FailedToGetPaymentByServiceUuid = "1100017"
	FailedToCheckHasCommented       = "1100018"
)

var ServiceErrorMessageMap = map[string]string{
	FailedToCreateService:             "failed to create service.",
	FailedServiceQrCodeSecretNotMatch: "qr code secret does not match.",
	NotAServiceParticipant:            "scanner is not a service participant.",
	ServiceNotYetEnd:                  "no payment detail since service has not ended yet.",
}
