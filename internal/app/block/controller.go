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

func GetBlockedUsersHandler(c *gin.Context, depCon container.Container) {
	userUuid := c.GetString("uuid")

	q := NewBlockDAO(db.GetDB())
	blockedUsers, err := q.GetBlockedUsers(userUuid)

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

	c.JSON(http.StatusOK, NewTransform().TransformBlockedUser(blockedUsers))
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
			apperr.NewErr(apperr.UnableToFindBlockee),
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
				apperr.FailedToBlockUser,
				err.Error(),
			),
		)

		return
	}

	c.JSON(http.StatusOK, struct{}{})
}

type UnblockBody struct {
	BlockeeUuid string `form:"blockee_uuid" json:"blockee_uuid"`
}

func UnblockHandler(c *gin.Context, depCon container.Container) {
	body := UnblockBody{}

	if err := requestbinder.Bind(c, &body); err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToBindApiBodyParams,
				err.Error(),
			),
		)

		return
	}

	// Check if the requester has ever blocked the user.
	q := NewBlockDAO(db.GetDB())
	hasBlocked, err := q.HasBlockedByUser(contracts.HasBlockedByUserParams{
		BlockerUuid: c.GetString("uuid"),
		BlockeeUuid: body.BlockeeUuid,
	})

	if !hasBlocked {
		c.AbortWithError(
			http.StatusBadRequest,
			apperr.NewErr(apperr.FailedToUnblockNotBlockedUser),
		)

		return
	}

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToCheckHasBlockedUser,
				err.Error(),
			),
		)

		return
	}

	if err := q.Unblock(UnblockParams{
		BlockerUuid: c.GetString("uuid"),
		BlockeeUuid: body.BlockeeUuid,
	}); err != nil {
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
