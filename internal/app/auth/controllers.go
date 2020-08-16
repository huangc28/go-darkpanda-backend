package auth

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/huangc28/go-darkpanda-backend/db"
	apperr "github.com/huangc28/go-darkpanda-backend/internal/app/apperr"
	"github.com/huangc28/go-darkpanda-backend/internal/app/util"
	"github.com/huangc28/go-darkpanda-backend/internal/models"
	"github.com/ventu-io/go-shortid"
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

	// if refer code and username are all valid, create a new user.
	// generates uuid for new user.
	uuid, err := shortid.Generate()

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToGenerateUuid,
				err.Error(),
			),
		)
	}

	newUser, err := q.CreateUser(c, models.CreateUserParams{
		Username:      body.Username,
		Gender:        models.Gender(body.Gender),
		Uuid:          uuid,
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

// sends verification code to specified mobile number
//   - receive user uuid
//   - mobile number
// from the client.
type SendVerifyCodeBody struct {
	Uuid   string `json:"uuid" binding:"required,gt=0"`
	Mobile string `json:"mobile" binding:"required,numeric,gt=0"`
}

func SendVerifyCodeHandler(c *gin.Context) {
	var (
		body SendVerifyCodeBody
		ctx  context.Context = context.Background()
	)

	if err := c.ShouldBindJSON(&body); err != nil {
		c.AbortWithError(
			http.StatusBadRequest,
			apperr.NewErr(
				apperr.FailedToValidateSendVerifyCodeParams,
				err.Error(),
			),
		)

		return
	}

	// retrieve user by uuid
	q := models.New(db.GetDB())
	usr, err := q.GetUserByUuid(ctx, body.Uuid)

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

	// -------------------  user is not phone verified -------------------
	if usr.PhoneVerified.Bool {
		c.AbortWithError(
			http.StatusBadRequest,
			apperr.NewErr(apperr.UserIsPhoneVerified),
		)

		return
	}

	// ------------------- sends verify code to specified number -------------------
	// generate 4 digit code
	verPrefix := util.GenRandStringRune(3)
	verfDigs := util.Gen4DigitNum(1000, 9999)

	// store verify prefix and verify digits to db in a form of ccc-3333
	err = q.UpdateVerifyCodeById(ctx, models.UpdateVerifyCodeByIdParams{
		PhoneVerifyCode: sql.NullString{
			String: fmt.Sprintf("%s-%d", verPrefix, verfDigs),
			Valid:  true,
		},
		ID: usr.ID,
	})

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToUpdateVerifyCode,
				err.Error(),
			),
		)

		return
	}

	// ------------------- send verification code via twillio -------------------
	res := struct {
		Uuid         string `json:"uuid"`
		VerifyPrefix string `json:"verify_prefix"`
		VerifySuffix int    `json:"verify_suffix "`
	}{
		usr.Uuid,
		verPrefix,
		verfDigs,
	}

	log.Printf("DEBUG 1 %v", res)

	c.JSON(http.StatusOK, &res)
}

func VerifyPhoneHandler(c *gin.Context) {

}
