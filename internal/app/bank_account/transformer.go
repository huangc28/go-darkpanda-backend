package bank_account

import (
	"github.com/huangc28/go-darkpanda-backend/internal/app/models"
)

type BankAccountTransform struct{}

func NewTransform() *BankAccountTransform {
	return &BankAccountTransform{}
}

type TransformedBankAccount struct {
	BankName      string              `json:"bank_name"`
	Branch        string              `json:"branch"`
	AccountNumber string              `json:"account_number"`
	VerifyStatus  models.VerifyStatus `json:"verify_status"`
}

func (ba *BankAccountTransform) TransformBankAccount(bank *models.BankAccount) *TransformedBankAccount {
	return &TransformedBankAccount{
		BankName:      bank.BankName.String,
		Branch:        bank.Branch.String,
		AccountNumber: bank.AccountNumber.String,
		VerifyStatus:  models.VerifyStatus(bank.VerifyStatus.String),
	}
}
