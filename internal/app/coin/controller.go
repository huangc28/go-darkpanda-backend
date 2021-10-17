package coin

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golobby/container/pkg/container"
	"github.com/huangc28/go-darkpanda-backend/config"
	"github.com/huangc28/go-darkpanda-backend/db"
	"github.com/huangc28/go-darkpanda-backend/internal/app/apperr"
	"github.com/huangc28/go-darkpanda-backend/internal/app/contracts"
	"github.com/huangc28/go-darkpanda-backend/internal/app/models"
	dpfcm "github.com/huangc28/go-darkpanda-backend/internal/app/pkg/firebase_messaging"
	"github.com/huangc28/go-darkpanda-backend/internal/app/pkg/requestbinder"
	"github.com/jmoiron/sqlx"
	"github.com/shopspring/decimal"
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
	costDeci, err := decimal.NewFromString(pkg.Cost.String)

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToInitStringToDeci,
				err.Error(),
			),
		)

		return
	}

	costF, _ := costDeci.Float64()

	coinDao := NewCoinDAO(db.GetDB())
	coinOrder, err := coinDao.OrderCoin(
		contracts.OrderCoinParams{
			BuyerID:     int(user.ID),
			PackageId:   int(pkg.ID),
			Quantity:    1,
			OrderStatus: models.OrderStatusOrdering,
			Cost:        costF,
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

	appConf := config.GetAppConf()

	tpayer := NewTapPayer(
		TapPayerConf{
			Url:        appConf.TappayEndpoint,
			PartnerKey: appConf.TappayPartnerKey,
			MerchantId: appConf.TappayMerchantId,
		},
	)

	tpResp, respRaw, err := tpayer.PayByPrime(
		PayByPrimeParams{
			Prime:    body.Prime,
			Details:  "Tappay test",
			Amount:   pkg.Cost.String,
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
		// costDeci, _, := decimal.NewFromString(pkg.Cost.String)
		userBal, err := userBalDao.CreateOrTopUpBalance(
			contracts.CreateOrTopUpBalanceParams{
				UserID: int(user.ID),
				// TopupAmount: float64(pkg.Cost.Int32),
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

	// Emit message to female when male done service payment.
	var dpfcm dpfcm.DPFirebaseMessenger
	depCon.Make(&dpfcm)

	respStruct, _ := TransformBuyCoin(transResp.Response.(string))

	// respond with the following:
	c.JSON(http.StatusOK, respStruct)
}

func GetCoinBalance(c *gin.Context, depCon container.Container) {
	// Get requester uuid.
	uuid := c.GetString("uuid")

	var userDao contracts.UserDAOer
	depCon.Make(&userDao)

	user, err := userDao.GetUserByUuid(uuid, "id")

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

	// Retrieve user coin balance info. If balance record not exists, create the balance record for the user.
	userBalDao := NewUserBalanceDAO(db.GetDB())
	userBal, err := userBalDao.GetCoinBalanceByUserId(int(user.ID))

	if err == sql.ErrNoRows {
		userBal, err = userBalDao.CreateOrTopUpBalance(contracts.CreateOrTopUpBalanceParams{
			UserID: int(user.ID),
		})

		if err != nil {
			c.AbortWithError(
				http.StatusInternalServerError,
				apperr.NewErr(
					apperr.FailedToCreateUserBalance,
					err.Error(),
				),
			)

			return
		}
	}

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToGetUserBalance,
				err.Error(),
			),
		)

		return
	}

	respStruct, _ := TransformGetCoinBalance(userBal.Balance)

	c.JSON(http.StatusOK, respStruct)
}

func GetCoinPackages(c *gin.Context) {
	// Get list of coin packages.
	newConPkgDao := NewCoinPackagesDAO(db.GetDB())
	pkgs, err := newConPkgDao.GetPackages()

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToGetCoinPackages,
				err.Error(),
			),
		)

		return

	}

	trfPkgs, err := TransformCoinPakcages(pkgs)

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToTransformCoinPackage,
				err.Error(),
			),
		)

		return
	}

	c.JSON(http.StatusOK, trfPkgs)
}
