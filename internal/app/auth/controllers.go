package auth

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/huangc28/go-darkpanda-backend/db"
	apperr "github.com/huangc28/go-darkpanda-backend/internal/app/apperr"
	"github.com/huangc28/go-darkpanda-backend/internal/models"
)

type RegisterBody struct {
	ReferCode string `json:"refer_code" binding:"required"`
	Username  string `json:"username" binding:"required"`
	Gender    string `json:"gender" binding:"oneof='male' 'female'"`
}

// We need the following to register new user
//   - reference code
//   - username
// @TODO check username is duplicated
func RegisterHandler(c *gin.Context) {
	var (
		body RegisterBody
		ctx  context.Context = context.Background()
	)

	if err := c.ShouldBindJSON(&body); err != nil {
		c.AbortWithError(
			http.StatusBadRequest,
			apperr.NewErr(
				apperr.FailedToValidateRegisterParams,
				err.Error(),
			),
		)

		return
	}

	// ------------------- check username is not registered already -------------------
	dao := NewAuthDao(db.GetDB())
	usernameExists, err := dao.CheckUsernameExists(ctx, body.Username)

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToCheckUsernameExistence,
				err.Error(),
			),
		)

		return
	}

	if usernameExists {
		c.AbortWithError(
			http.StatusBadRequest,
			apperr.NewErr(apperr.UsernameNotAvailable),
		)

		return
	}

	// ------------------- check if refercode exists -------------------
	referCodeExists, err := dao.CheckReferCodeExists(ctx, body.ReferCode)

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToCheckReferCodeExists,
				err.Error(),
			),
		)
	}

	if !referCodeExists {
		c.AbortWithError(
			http.StatusNotFound,
			apperr.NewErr(apperr.ReferCodeNotExist),
		)

		return
	}

	// ------------------- check refercode is valid -------------------
	// check if reference code exists and invitee id is null
	q := models.New(db.GetDB())

	urc, err := q.GetReferCodeInfoByRefcode(
		ctx,
		body.ReferCode,
	)

	// @TODO handle error using middleware
	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToRetrieveReferCodeInfo,
				err.Error(),
			),
		)

		return
	}

	// if inviteeID has been occupied, the given refer code can't be used anymore
	if urc.InviteeID.Valid {
		c.AbortWithError(
			http.StatusBadRequest,
			apperr.NewErr(apperr.ReferCodeOccupied),
		)

		return
	}

	// if refer code and username are all valid, create a new user
	newUser, err := q.CreateUser(c, models.CreateUserParams{
		Username:      body.Username,
		Gender:        models.Gender(body.Gender),
		PremiumType:   models.PremiumTypeNormal,
		PhoneVerified: sql.NullBool{Bool: false, Valid: true},
	})

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToCreateUser,
				err.Error(),
			),
		)

		return
	}

	c.JSON(http.StatusOK, NewTransform().TransformUser(&newUser))
}
