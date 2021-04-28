package bankAccount

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golobby/container/pkg/container"
	"github.com/huangc28/go-darkpanda-backend/db"
	"github.com/huangc28/go-darkpanda-backend/internal/app/apperr"
	"github.com/huangc28/go-darkpanda-backend/internal/app/contracts"
	"github.com/huangc28/go-darkpanda-backend/internal/app/models"
	"github.com/huangc28/go-darkpanda-backend/internal/app/pkg/requestbinder"
	log "github.com/sirupsen/logrus"
)

type GetUserBankAccountBody struct {
	UUID string `form:"uuid" json:"uuid" binding:"required,gt=0"`
}

func GetUserBankAccount(c *gin.Context, depCon container.Container) {
	var (
		uuid string = c.Param("uuid")
	)

	q := NewBankAccountDAO(db.GetDB())
	bank, err := q.GetUserBankAccount(uuid)

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToGetUserByUuid,
				err.Error(),
			),
		)

		return
	}

	tResp := NewTransform().TransformBankAccount(bank)

	if err != nil {
		log.Fatal(err)
	}

	c.JSON(http.StatusOK, tResp)
}

type InsertBankAccountBody struct {
	UUID          string              `form:"uuid" json:"uuid"`
	BankName      string              `form:"bank_name" json:"bank_name" body:"bank_name"`
	Branch        string              `form:"branch" json:"branch"`
	AccountNumber string              `form:"account_number" json:"account_number"`
	VerifyStatus  models.VerifyStatus `form:"verify_status" json:"verify_status"`
}

func InsertBankAccount(c *gin.Context, depCon container.Container) {
	body := &InsertBankAccountBody{}

	if err := requestbinder.Bind(c, body); err != nil {
		c.AbortWithError(
			http.StatusBadRequest,
			apperr.NewErr(
				apperr.FailedToValidateEmitInquiryParams,
				err.Error(),
			),
		)

		return
	}

	uuid := c.Param("uuid")

	q := NewBankAccountDAO(db.GetDB())
	bankExists, err := q.CheckHasBankAccountByUUID(uuid)

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToCheckActiveInquiry,
				err.Error(),
			),
		)

		return
	}

	var (
		userDao contracts.UserDAOer
	)

	depCon.Make(&userDao)

	user, err := userDao.GetUserByUuid(uuid, "id")

	if bankExists {
		err := q.PatchBankAccount(uuid, contracts.PatchBankAccountParams{
			UserID:        int(user.ID),
			BankName:      body.BankName,
			Branch:        body.Branch,
			AccountNumber: body.AccountNumber,
			VerifyStatus:  models.VerifyStatusPending,
		})

		if err != nil {
			c.AbortWithError(
				http.StatusInternalServerError,
				apperr.NewErr(
					apperr.FailedToPatchInquiry,
					err.Error(),
				),
			)

			return
		}
	} else {
		if err != nil {
			c.AbortWithError(
				http.StatusInternalServerError,
				apperr.NewErr(
					apperr.FailedToGetUserByUuid,
					err.Error(),
				),
			)

			return
		}

		err1 := q.InsertBankAccount(contracts.InsertBankAccountParams{
			UserID:        int(user.ID),
			BankName:      body.BankName,
			Branch:        body.Branch,
			AccountNumber: body.AccountNumber,
			VerifyStatus:  models.VerifyStatusPending,
		})

		if err1 != nil {
			c.AbortWithError(
				http.StatusInternalServerError,
				apperr.NewErr(
					apperr.FailedToGetUserByUuid,
					err1.Error(),
				),
			)

			return
		}
	}

	c.JSON(http.StatusOK, struct{}{})
}
