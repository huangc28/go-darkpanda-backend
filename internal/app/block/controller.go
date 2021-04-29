package block

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golobby/container/pkg/container"
	"github.com/huangc28/go-darkpanda-backend/db"
	"github.com/huangc28/go-darkpanda-backend/internal/app/apperr"
	"github.com/huangc28/go-darkpanda-backend/internal/app/contracts"
	"github.com/huangc28/go-darkpanda-backend/internal/app/pkg/requestbinder"
)

type GetUserBankAccountBody struct {
	UUID string `form:"uuid" json:"uuid" binding:"required,gt=0"`
}

func GetUserBlock(c *gin.Context, depCon container.Container) {
	var (
		uuid string = c.Param("uuid")
	)

	q := NewBlockDAO(db.GetDB())
	blocks, err := q.GetUserBlock(uuid)

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

	tResp, err := NewTransform().TransformBlock(blocks)

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToTransformUserPayments,
				err.Error(),
			),
		)

		return
	}

	c.JSON(http.StatusOK, tResp)
}

func InsertUserBlock(c *gin.Context, depCon container.Container) {
	body := contracts.InsertUserBlockListParams{}

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

	q := NewBlockDAO(db.GetDB())
	err := q.InsertUserBlock(body)

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

func DeleteUserBlock(c *gin.Context, depCon container.Container) {
	var (
		blockId string = c.Param("id")
	)

	q := NewBlockDAO(db.GetDB())
	err := q.DeleteUserBlock(blockId)

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
