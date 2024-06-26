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
	ServiceStatusNotValidToCancel   = "1100019"
	ServiceHasBeenCanceled          = "1100020"

	FailedToDeleteChatroomByServiceId         = "1100021"
	FailedToSendCancelMessage                 = "1100022"
	FailedToStartService                      = "1100023"
	FailedToMarshQRCodeInfo                   = "1100024"
	FailedToSendServiceConfirmedMsg           = "1100025"
	FailedToSendServiceDetailMsg              = "1100026"
	FailedToGetOverlappedServices             = "1100027"
	OverlappingService                        = "1100028"
	FailedToGetServiceProviderByServiceUUID   = "1100029"
	FailedToPerformRefundCustomerIfRefundable = "1100030"
	FailedToSendServiceCancelledFCM           = "1100031"
	FailedToSendRefundedFCM                   = "1100032"
	FailedToCalcServiceMatchingFee            = "1100033"
)

var ServiceErrorMessageMap = map[string]string{
	FailedToCreateService:             "failed to create service",
	FailedServiceQrCodeSecretNotMatch: "qr code secret does not match",
	NotAServiceParticipant:            "scanner is not a service participant",
	ServiceNotYetEnd:                  "no payment detail since service has not ended yet",
	ServiceStatusNotValidToCancel:     "service status is not valid for canceling",
	ServiceHasBeenCanceled:            "service has been canceld by partner",
	OverlappingService:                "service can not be booked due to overlapping service, pick another time",
}
