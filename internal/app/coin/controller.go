package coin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/golobby/container/pkg/container"
	"github.com/huangc28/go-darkpanda-backend/db"
	"github.com/huangc28/go-darkpanda-backend/internal/app/apperr"
	"github.com/huangc28/go-darkpanda-backend/internal/app/contracts"
	"github.com/huangc28/go-darkpanda-backend/internal/app/models"
	"github.com/huangc28/go-darkpanda-backend/internal/app/pkg/requestbinder"
)

type PaymentResponse struct {
	Status int    `json:"status"`
	Msg    string `json:"msg"`
}

type PaymentParam struct {
	Prime       string     `json:"prime"`
	Partner_key string     `json:"partner_key"`
	Merchant_id string     `json:"merchant_id"`
	Details     string     `json:"details"`
	Amount      string     `json:"amount"`
	Currency    string     `json:"currency"`
	Cardholder  CardHolder `json:"cardholder"`
	Remember    bool       `json:"remember"`
}
type CardHolder struct {
	Phone_number string `json:"phone_number"`
	Name         string `json:"name"`
	Email        string `json:"email"`
	Zip_code     string `json:"zip_code"`
	Address      string `json:"address"`
	National_id  string `json:"national_id"`
}

type PaymentCard struct {
	Number string `form:"number" json:"number"`
	Name   string `form:"name" json:"name"`
	Month  string `form:"month" json:"month"`
	Year   string `form:"year" json:"year"`
	Cvv    int    `form:"cvv" json:"cvv"`
	Prime  string `form:"prime" json:"prime"`
}

type OrderCoinBody struct {
	UUID        string             `form:"uuid"`
	Amount      int                `form:"amount"`
	Cost        int                `form:"cost"`
	OrderStatus models.OrderStatus `form:"order_status"`
	PaymentCard PaymentCard        `form:"paymentCard"`
}

func OrderCoin(c *gin.Context, depCon container.Container) {
	body := &OrderCoinBody{}

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
	user, errUser := userDao.GetUserByUuid(uuid)

	if errUser != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToGetUserByUuid,
				errUser.Error(),
			),
		)

		return
	}

	// Insert into coin_order with order_status=ordering
	q := NewCoinDAO(db.GetDB())
	coin, errCoin := q.OrderCoin(contracts.OrderCoinParams{
		BuyerID:     int(user.ID),
		Amount:      body.Amount,
		Cost:        body.Cost,
		OrderStatus: models.OrderStatusOrdering,
	})

	if errCoin != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToGetUserByUuid,
				errCoin.Error(),
			),
		)

		return
	}

	// Call TapPay API - Pay By Prime
	fmt.Println(body.PaymentCard.Prime)
	url := "https://sandbox.tappaysdk.com/tpc/payment/pay-by-prime"
	values := &PaymentParam{
		Prime:       body.PaymentCard.Prime,
		Partner_key: "partner_tzcz4DPLdBH86XxaQtDtPHyXmpx5M5Edn6EuIVJMpQ77hz0qlROGCapa",
		Merchant_id: "huangc28_CTBC",
		Details:     "Tappay test",
		Amount:      strconv.Itoa(body.Amount),
		Currency:    "TWD",
		Cardholder: CardHolder{
			Phone_number: user.Mobile.String,
			Name:         "",
			Email:        "",
			Zip_code:     "",
			Address:      "",
			National_id:  "",
		},
		Remember: false,
	}

	jsonValue, err := json.Marshal(values)

	if err != nil {
		fmt.Println(err)
		return
	}

	// create client
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonValue))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("x-api-key", "partner_tzcz4DPLdBH86XxaQtDtPHyXmpx5M5Edn6EuIVJMpQ77hz0qlROGCapa")

	// make a request
	client := &http.Client{}
	resp, errResp := client.Do(req)

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToGetUserByUuid,
				errResp.Error(),
			),
		)

		return
	}

	if resp.StatusCode == http.StatusOK {
		var (
			m PaymentResponse
		)

		ma := json.NewDecoder(resp.Body)

		if errMa := ma.Decode(&m); err != nil {
			c.AbortWithError(
				http.StatusInternalServerError,
				apperr.NewErr(
					apperr.FailedToGetUserByUuid,
					errMa.Error(),
				),
			)

			return
		}

		if m.Status == 0 {
			// Update coin_order order_status=success
			errUpdate := q.UpdateOrderCoinStatus(contracts.UpdateOrderCoinStatusParams{
				ID:          int(coin.ID),
				OrderStatus: models.OrderStatusSuccess,
			})

			if errUpdate != nil {
				c.AbortWithError(
					http.StatusInternalServerError,
					apperr.NewErr(
						apperr.FailedToGetUserByUuid,
						errUpdate.Error(),
					),
				)

				return
			}
		} else {
			// Update coin_order order_status=failed
			errUpdate := q.UpdateOrderCoinStatus(contracts.UpdateOrderCoinStatusParams{
				ID:          int(coin.ID),
				OrderStatus: models.OrderStatusFailed,
			})

			if errUpdate != nil {
				c.AbortWithError(
					http.StatusInternalServerError,
					apperr.NewErr(
						apperr.FailedToGetUserByUuid,
						errUpdate.Error(),
					),
				)

				return
			}

			c.AbortWithError(
				http.StatusInternalServerError,
				apperr.NewErr(
					strconv.Itoa(m.Status),
					m.Msg,
				),
			)

			return
		}

	}

	c.JSON(http.StatusOK, struct{}{})
}
