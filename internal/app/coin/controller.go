package coin

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/golobby/container/pkg/container"
	"github.com/huangc28/go-darkpanda-backend/config"
	"github.com/huangc28/go-darkpanda-backend/db"
	"github.com/huangc28/go-darkpanda-backend/internal/app/apperr"
	"github.com/huangc28/go-darkpanda-backend/internal/app/contracts"
	"github.com/huangc28/go-darkpanda-backend/internal/app/models"
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
	Name       string `form:"name" json:"name"`
	Email      string `form:"email" json:"email"`
	ZipCode    string `form:"zip_code" json:"zip_code"`
	Address    string `form:"address" json:"address"`
	NationalId string `form:"national_id" json:"national_id"`

	Month string `form:"month" json:"month" binding:"required,gt=0"`
	Year  string `form:"year" json:"year" binding:"required,gt=0"`
	Cvv   int    `form:"cvv" json:"cvv" binding:"required,gt=0"`
	Prime string `form:"prime" json:"prime" binding:"required,gt=0"`
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

	// Retrieve coin package info that the user wants to buy.
	coinPkgDao := NewCoinPackagesDAO(db.GetDB())
	pkg, err := coinPkgDao.GetPackageById(body.PackageId)

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToGetPackageInfo,
				err.Error(),
			),
		)

		return
	}

	// Insert into coin_orders with order_status "ordering".
	coinDao := NewCoinDAO(db.GetDB())
	coinOrder, err := coinDao.OrderCoin(
		contracts.OrderCoinParams{
			BuyerID:     int(user.ID),
			PackageId:   int(pkg.ID),
			Quantity:    1,
			OrderStatus: models.OrderStatusOrdering,
			Cost:        int(pkg.Cost.Int32),
		},
	)

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToCreateCoinOrder,
				err.Error(),
			),
		)

		return
	}

	tpCred := config.GetAppConf().TapPayCredential

	tpayer := NewTapPayer(
		TapPayerConf{
			Url:        tpCred.EndPoint,
			PartnerKey: tpCred.PartnerKey,
			MerchantId: tpCred.MerchantID,
		},
	)

	tpResp, respRaw, err := tpayer.PayByPrime(
		PayByPrimeParams{
			Prime:    body.Prime,
			Details:  "Tappay test",
			Amount:   strconv.Itoa(int(pkg.Cost.Int32)),
			Currency: "TWD",
			Cardholder: CardHolderParams{
				PhoneNumber: user.Mobile.String,
				Name:        body.Name,
				Email:       body.Email,
				ZipCode:     body.ZipCode,
				Address:     body.Address,
				NationalId:  body.NationalId,
			},
			Remember: false,
		},
	)

	if err != nil {
		// If payment failed to proceed, update the payment status to be 'failed'
		if _, err := coinDao.UpdateOrderCoinById(
			UpdateOrderCoinByIdParam{
				Id:          int(coinOrder.ID),
				OrderStatus: models.OrderStatusFailed,
				Raw:         respRaw,
			},
		); err != nil {
			c.AbortWithError(
				http.StatusInternalServerError,
				apperr.NewErr(
					apperr.FailedToUpdateCoinOrder,
					err.Error(),
				),
			)

			return
		}

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
	//   - Update order status to success.
	//   - Topup balance for the user.
	transResp := db.TransactWithFormatStruct(db.GetDB(), func(tx *sqlx.Tx) db.FormatResp {
		// We need to store RecTradeId and payment response json.
		coinDao.WithTx(tx)
		_, err := coinDao.UpdateOrderCoinById(
			UpdateOrderCoinByIdParam{
				Id:          int(coinOrder.ID),
				OrderStatus: models.OrderStatusSuccess,
				RecTradeId:  tpResp.RecTradeId,
				Raw:         respRaw,
			},
		)

		if err != nil {
			return db.FormatResp{
				Err:            err,
				ErrCode:        apperr.FailedToUpdateCoinOrder,
				HttpStatusCode: http.StatusInternalServerError,
			}
		}

		userBalDao := NewUserBalanceDAO(tx)
		userBal, err := userBalDao.CreateOrTopUpBalance(
			CreateOrTopUpBalanceParams{
				UserId:      int(user.ID),
				TopupAmount: float64(pkg.Cost.Int32),
			},
		)

		if err != nil {
			return db.FormatResp{
				Err:            err,
				ErrCode:        apperr.FailedToTopupUserBalance,
				HttpStatusCode: http.StatusInternalServerError,
			}

		}

		return db.FormatResp{
			Response: userBal.Balance,
		}
	})

	if transResp.Err != nil {
		c.AbortWithError(
			transResp.HttpStatusCode,
			apperr.NewErr(
				transResp.ErrCode,
				transResp.Err.Error(),
			),
		)

		return
	}

	respStruct, _ := TransformBuyCoin(transResp.Response.(string))

	// respond with the following:
	c.JSON(http.StatusOK, respStruct)
}

func GetConBalance(c *gin.Context) {
	//
}
