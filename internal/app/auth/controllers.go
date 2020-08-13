package auth

import (
	"context"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/huangc28/go-darkpanda-backend/db"
	apperr "github.com/huangc28/go-darkpanda-backend/internal/app/apperr"
	"github.com/huangc28/go-darkpanda-backend/internal/models"
)

type RegisterBody struct {
	ReferCode string        `json:"refer_code" binding:"required"`
	Username  string        `json:"username" binding:"required"`
	Gender    models.Gender `json:"gender" binding:"required"`
}

// We need the following to register new user
//   - reference code
//   - username
func RegisterHandler(c *gin.Context) {
	var body RegisterBody

	//abortHelper := apperr.AbortWithResponse(c)

	if err := c.ShouldBindJSON(&body); err != nil {
		log.Println("failed to bind")
		c.AbortWithError(
			http.StatusBadRequest,
			apperr.NewErr(
				apperr.FailedToValidateRegisterParams,
				err.Error(),
			),
		)
		//abortHelper(
		//http.StatusBadRequest,
		//apperr.FailedToValidateRegisterParams,
		//err.Error(),
		//).SetType(gin.ErrorTypePublic)

		return
	}

	// check if reference code exists and invitee id is null
	q := models.New(db.GetDB())

	urc, err := q.GetReferCodeInfoByRefcode(
		context.Background(),
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
		).SetType(gin.ErrorTypePublic)

		return
	}

	// if inviteeID has been occupied, the given refer code can't be used anymore
	if urc.InviteeID.Valid == true {
		c.AbortWithError(
			http.StatusBadRequest,
			apperr.NewErr(
				apperr.ReferCodeOccupied,
				"refer code already occupied",
			),
		).SetType(gin.ErrorTypePublic)

		return
	}

	// if refer code and username are all valid, create a new user
	//newUser, err = q.CreateUser(c, models.CreateUserParams{
	//Username: body.Username,
	//})

	c.String(http.StatusOK, "register handler")
}
