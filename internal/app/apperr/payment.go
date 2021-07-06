package apperr

var (
	PayerIsNotTheCustomerOfTheService      = "2000001"
	ServiceStatusInvalidForPayment         = "2000002"
	FailedToCheckHasEnoughBalance          = "2000003"
	FailedToGetMatchingFee                 = "2000004"
	FailedToDeductBalance                  = "2000005"
	FailedToCreatePayment                  = "2000006"
	FailedToTransfromCreatePaymentResponse = "2000007"
)

var PaymentErrCodeMsgMap = map[string]string{
	PayerIsNotTheCustomerOfTheService: "payer is not the customer of the service",
	ServiceStatusInvalidForPayment:    "service status invalid to pay",
}
