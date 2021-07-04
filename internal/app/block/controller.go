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

type BlockUserBody struct {
	BlockeeUuid string `form:"blockee_uuid" json:"blockee_uuid"`
}

func BlockUserHandler(c *gin.Context, depCon container.Container) {
	body := BlockUserBody{}

	if err := requestbinder.Bind(c, &body); err != nil {
		c.AbortWithError(
			http.StatusBadRequest,
			apperr.NewErr(
				apperr.FailedToValidateEmitInquiryParams,
				err.Error(),
			),
		)

		return
	}

	var userDao contracts.UserDAOer
	depCon.Make(&userDao)

	// @TODO Create a method to retrieve multiple users.
	// Retrieve blocker ID.
	blocker, err := userDao.GetUserByUuid(c.GetString("uuid"), "id")

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

	// Retrieve blockee ID.
	blockee, err := userDao.GetUserByUuid(body.BlockeeUuid, "id")

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

	q := NewBlockDAO(db.GetDB())
	if err := q.BlockUser(
		BlockUserParams{
			BlockerId: int(blocker.ID),
			BlockeeId: int(blockee.ID),
		},
	); err != nil {
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

func UnBlockHandler(c *gin.Context, depCon container.Container) {
	// var blockeeUuid string = c.Param("uuid")

	// Check if the requester

	// q := NewBlockDAO(db.GetDB())
	// err := q.DeleteUserBlock(blockId)

	// if err != nil {
	// 	c.AbortWithError(
	// 		http.StatusInternalServerError,
	// 		apperr.NewErr(
	// 			apperr.FailedToGetUserByUuid,
	// 			err.Error(),
	// 		),
	// 	)

	// 	return
	// }

	c.JSON(http.StatusOK, struct{}{})
}
