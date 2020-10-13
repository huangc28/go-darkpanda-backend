package models

type PaymentInfo struct {
	Payment
	Service Service
	Payer   User
}
