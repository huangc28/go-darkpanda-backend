package payment

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/golobby/container/pkg/container"
	"github.com/huangc28/go-darkpanda-backend/db"
	"github.com/huangc28/go-darkpanda-backend/internal/app/apperr"
	"github.com/huangc28/go-darkpanda-backend/internal/app/contracts"
	"github.com/huangc28/go-darkpanda-backend/internal/app/models"
	darkfirestore "github.com/huangc28/go-darkpanda-backend/internal/app/pkg/dark_firestore"
	"github.com/huangc28/go-darkpanda-backend/internal/app/pkg/requestbinder"
	"github.com/jmoiron/sqlx"
)

type CreatePaymentBody struct {
	ServiceUuid string `json:"service_uuid" form:"service_uuid" binding:"required,gt=0"`
}

// CreatePayment When service is confirmed. Male user would request this
// API to complete the payment and change the service status to `to_be_fulfilled`.
// The service status on fire store would also be synced once the payment is complete.
func CreatePayment(c *gin.Context, depCon container.Container) {
	body := CreatePaymentBody{}

	if err := requestbinder.Bind(c, &body); err != nil {
		c.AbortWithError(
			http.StatusBadRequest,
			apperr.NewErr(
				apperr.FailedToBindApiBodyParams,
				err.Error(),
			),
		)

		return
	}

	var (
		srvDao         contracts.ServiceDAOer
		userDao        contracts.UserDAOer
		userBalanceDao contracts.UserBalancer
	)

	depCon.Make(&srvDao)
	depCon.Make(&userDao)
	depCon.Make(&userBalanceDao)

	srv, err := srvDao.GetServiceByUuid(body.ServiceUuid)

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToGetServiceByUuid,
				err.Error(),
			),
		)

		return
	}

	user, err := userDao.GetUserByUuid(c.GetString("uuid"))

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

	// Check if the payer is the customer of the service.
	if srv.CustomerID.Int32 != int32(user.ID) {
		c.AbortWithError(
			http.StatusBadRequest,
			apperr.NewErr(apperr.PayerIsNotTheCustomerOfTheService),
		)

		return
	}

	// Makesure the service status is `unpaid`
	if srv.ServiceStatus != models.ServiceStatusUnpaid {
		c.AbortWithError(
			http.StatusBadRequest,
			apperr.NewErr(apperr.ServiceStatusInvalidForPayment),
		)

		return
	}

	// We only charge matching fee now.
	var coinPackageDaoer contracts.CoinPackageDAOer
	depCon.Make(&coinPackageDaoer)

	matchingFee, err := coinPackageDaoer.GetMatchingFee()

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToGetMatchingFee,
				err.Error(),
			),
		)

		return

	}

	// Check if the payer has enough balance.
	err = userBalanceDao.HasEnoughBalanceToCharge(int(user.ID), matchingFee)

	if err != nil {
		c.AbortWithError(
			http.StatusBadRequest,
			apperr.NewErr(
				apperr.FailedToCheckHasEnoughBalance,
				err.Error(),
			),
		)

		return
	}

	// Charge user, change the service status to `to_be_fulfilled` and create a payment record.
	ctx := context.Background()
	trxResp := db.TransactWithFormatStruct(
		db.GetDB(),
		func(tx *sqlx.Tx) db.FormatResp {
			// Deduct user balance by cost.
			_, err := userBalanceDao.
				WithTx(tx).
				DeductUserBalance(
					int(user.ID),
					matchingFee,
				)

			if err != nil {
				return db.FormatResp{
					Err:            err,
					ErrCode:        apperr.FailedToDeductBalance,
					HttpStatusCode: http.StatusInternalServerError,
				}
			}

			// Change service status to `to_be_fulfilled`.
			srvStatus := models.ServiceStatusToBeFulfilled
			_, err = srvDao.WithTx(tx).UpdateServiceByID(contracts.UpdateServiceByIDParams{
				ID:            srv.ID,
				ServiceStatus: &srvStatus,
			})

			if err != nil {
				return db.FormatResp{
					Err:            err,
					ErrCode:        apperr.FailedToUpdateService,
					HttpStatusCode: http.StatusInternalServerError,
				}
			}

			// Create a new payment record.
			q := models.New(tx)
			_, err = q.CreatePayment(
				ctx,
				models.CreatePaymentParams{
					PayerID:   int32(user.ID),
					ServiceID: int32(srv.ID),
					Price:     strconv.Itoa(int(matchingFee.Cost.Int32)),
				},
			)

			if err != nil {
				return db.FormatResp{
					Err:            err,
					ErrCode:        apperr.FailedToCreatePayment,
					HttpStatusCode: http.StatusInternalServerError,
				}
			}

			return db.FormatResp{}
		},
	)

	if trxResp.Err != nil {
		c.AbortWithError(
			trxResp.HttpStatusCode,
			apperr.NewErr(
				trxResp.ErrCode,
				trxResp.Err.Error(),
			),
		)

		return
	}

	// Emit firebase event to notify female user that the service status has been updated.
	df := darkfirestore.Get()
	if err := df.UpdateService(
		ctx,
		darkfirestore.UpdateServiceParams{
			ServiceUuid:   srv.Uuid.String(),
			ServiceStatus: string(models.ServiceStatusToBeFulfilled),
		},
	); err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FirestoreFailedToUpdateService,
				err.Error(),
			),
		)

		return
	}

	c.JSON(http.StatusOK, struct{}{})
}
