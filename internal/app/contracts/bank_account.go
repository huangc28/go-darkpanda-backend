package contracts

import (
	"github.com/huangc28/go-darkpanda-backend/db"
	"github.com/huangc28/go-darkpanda-backend/internal/app/models"
)

type PatchBankAccountParams struct {
	UserID        int                 `form:"user_id" json:"user_id"`
	BankName      string              `form:"bank_name" json:"bank_name"`
	Branch        string              `form:"branch" json:"branch"`
	AccountNumber string              `form:"account_number" json:"account_number"`
	VerifyStatus  models.VerifyStatus `form:"verify_status" json:"verify_status"`
}

type InsertBankAccountParams struct {
	UserID        int                 `form:"user_id" json:"user_id"`
	BankName      string              `form:"bank_name" json:"bank_name"`
	Branch        string              `form:"branch" json:"branch"`
	AccountNumber string              `form:"account_number" json:"account_number"`
	VerifyStatus  models.VerifyStatus `form:"verify_status" json:"verify_status"`
}

type BankAccountDAOer interface {
	WithTx(tx db.Conn) BankAccountDAOer
	GetUserBankAccount(uuid string) (*models.BankAccount, error)
}
