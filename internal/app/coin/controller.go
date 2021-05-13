package coin

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/golobby/container/pkg/container"
	"github.com/huangc28/go-darkpanda-backend/config"
	"github.com/huangc28/go-darkpanda-backend/db"
	"github.com/huangc28/go-darkpanda-backend/internal/app/apperr"
	"github.com/huangc28/go-darkpanda-backend/internal/app/contracts"
	"github.com/huangc28/go-darkpanda-backend/internal/app/pkg/requestbinder"
	"github.com/jmoiron/sqlx"
)

type PaymentResponse struct {
	Status int    `json:"status"`
	Msg    string `json:"msg"`
}

type BuyCoinBody struct {
	PackageId  int    `form:"package_id" json:"package_id" binding:"required"`
	CardNumber string `form:"card_number" json:"card_number" binding:"required,gt=0"`
	Name       string `form:"name" json:"name" binding:"required,gt=0"`
	Month      string `form:"month" json:"month" binding:"required,gt=0"`
	Year       string `form:"year" json:"year" binding:"required,gt=0"`
	Cvv        int    `form:"cvv" json:"cvv" binding:"required,gt=0"`
	Prime      string `form:"prime" json:"prime" binding:"required,gt=0"`
}

func BuyCoin(c *gin.Context, depCon container.Container) {
	body := &BuyCoinBody{}

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

	uuid := c.GetString("uuid")

	var userDao contracts.UserDAOer
	depCon.Make(&userDao)

	// Get User ID
	user, err := userDao.GetUserByUuid(uuid, "id")

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(apperr.FailedToGetUserByUuid),
		)

		return
	}

	// Insert into coin_order with order_status=ordering
	//q := NewCoinDAO(db.GetDB())
	//coinModel, err := q.OrderCoin(contracts.OrderCoinParams{
	//BuyerID:     int(user.ID),
	//Amount:      body.Amount,
	//Cost:        body.Cost,
	//OrderStatus: models.OrderStatusOrdering,
	//})

	//if err != nil {
	//c.AbortWithError(
	//http.StatusInternalServerError,
	//apperr.NewErr(apperr.FailedToGetUserByUuid),
	//)

	//return
	//}

	tpCred := config.GetAppConf().TapPayCredential

	tpayer := NewTapPayer(
		TapPayerConf{
			Url:        tpCred.EndPoint,
			PartnerKey: tpCred.PartnerKey,
			MerchantId: tpCred.MerchantID,
		},
	)

	tpResp, err := tpayer.PayByPrime(
		PayByPrimeParams{
			Prime:    body.Prime,
			Details:  "Tappay test",
			Amount:   strconv.Itoa(100),
			Currency: "TWD",
			Cardholder: CardHolderParams{
				PhoneNumber: user.Mobile.String,
				Name:        "",
				Email:       "",
				ZipCode:     "",
				Address:     "",
				NationalId:  "",
			},
			Remember: false,
		},
	)

	if err != nil {
		// If payment failed to proceed, update the payment status to be 'failed'

		c.AbortWithError(
			http.StatusBadRequest,
			apperr.NewErr(
				apperr.TapPayFailedToPayByPrime,
				err.Error(),
			),
		)

		return
	}

	// TapPay API request success, we now perform the following:
	// - Change the order status to success.
	// - Topup balance for the user.
	db.TransactWithFormatStruct(db.GetDB(), func(tx *sqlx.Tx) db.FormatResp {
		coinDao := NewCoinDAO(tx)

		//coinDao

		return db.FormatResp{}
	})
	log.Printf("DEBUG tpResp %v", tpResp)

	//if m.Status == 0 {
	//// Update coin_order order_status=success
	//errUpdate := q.UpdateOrderCoinStatus(contracts.UpdateOrderCoinStatusParams{
	//ID:          int(coin.ID),
	//OrderStatus: models.OrderStatusSuccess,
	//})

	//if errUpdate != nil {
	//c.AbortWithError(
	//http.StatusInternalServerError,
	//apperr.NewErr(
	//apperr.FailedToGetUserByUuid,
	//errUpdate.Error(),
	//),
	//)

	//return
	//}
	//} else {
	//// Update coin_order order_status=failed
	//errUpdate := q.UpdateOrderCoinStatus(contracts.UpdateOrderCoinStatusParams{
	//ID:          int(coin.ID),
	//OrderStatus: models.OrderStatusFailed,
	//})

	//if errUpdate != nil {
	//c.AbortWithError(
	//http.StatusInternalServerError,
	//apperr.NewErr(
	//apperr.FailedToGetUserByUuid,
	//errUpdate.Error(),
	//),
	//)

	//return
	//}

	//c.AbortWithError(
	//http.StatusInternalServerError,
	//apperr.NewErr(
	//strconv.Itoa(m.Status),
	//m.Msg,
	//),
	//)

	//return
	//}

	//}

	c.JSON(http.StatusOK, struct{}{})
}
