package coin

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golobby/container/pkg/container"
	"github.com/huangc28/go-darkpanda-backend/db"
	"github.com/huangc28/go-darkpanda-backend/internal/app/apperr"
	"github.com/huangc28/go-darkpanda-backend/internal/app/contracts"
	"github.com/huangc28/go-darkpanda-backend/internal/app/models"
	"github.com/huangc28/go-darkpanda-backend/internal/app/pkg/requestbinder"
)

type OrderCoinBody struct {
	BuyerID     int32              `form:"buyer_id" json:"buyer_id"`
	Amount      float32            `form:"amount" json:"amount"`
	Cost        float32            `form:"cost" json:"cost"`
	OrderStatus models.OrderStatus `form:"order_status" json:"order_status"`
}

func OrderCoin(c *gin.Context, depCon container.Container) {
	body := OrderCoinBody{}

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

	q := NewCoinDAO(db.GetDB())
	err := q.OrderCoin(contracts.OrderCoinParams{
		BuyerID:     body.BuyerID,
		Amount:      body.Amount,
		Cost:        body.Cost,
		OrderStatus: models.OrderStatusOrdering,
	})

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

	c.JSON(http.StatusOK, struct{}{})
}
