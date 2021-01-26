package register

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golobby/container/pkg/container"
	"github.com/huangc28/go-darkpanda-backend/db"
	"github.com/huangc28/go-darkpanda-backend/internal/app/apperr"
	"github.com/huangc28/go-darkpanda-backend/internal/app/models"
	"github.com/huangc28/go-darkpanda-backend/internal/app/pkg/requestbinder"
	"github.com/jmoiron/sqlx"
	"github.com/teris-io/shortid"
)

type RegisterBody struct {
	Username  string `form:"username" uri:"username" json:"username" binding:"required"`
	Gender    string `form:"gender" uri:"gender" json:"gender" binding:"oneof='male' 'female'"`
	ReferCode string `form:"refer_code" uri:"refer_code" json:"refer_code" binding:"required"`
}

func HandleRegister(c *gin.Context) {
	var (
		body RegisterBody
		ctx  context.Context = context.Background()
	)

	if err := requestbinder.Bind(c, &body); err != nil {
		c.AbortWithError(
			http.StatusBadRequest,
			apperr.NewErr(
				apperr.FailedToValidateRegisterParams,
				err.Error(),
			),
		)

		return
	}

	// ------------------- check if username has been registered -------------------
	dao := NewRegisterDAO(db.GetDB())
	exists, err := dao.CheckUsernameExists(ctx, body.Username)

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

	if exists {
		c.AbortWithError(
			http.StatusBadRequest,
			apperr.NewErr(apperr.UsernameNotAvailable),
		)

		return
	}

	// ------------------- check if referral code exists -------------------
	refCodeExists, err := dao.CheckReferCodeExists(ctx, body.ReferCode)

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToCheckReferCodeExists,
				err.Error(),
			),
		)
	}

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToCheckReferCodeExists,
				err.Error(),
			),
		)
	}

	if !refCodeExists {
		c.AbortWithError(
			http.StatusNotFound,
			apperr.NewErr(apperr.ReferCodeNotExist),
		)

		return
	}

	// ------------------- check referral code is valid -------------------
	q := models.New(db.GetDB())
	urc, err := q.GetReferCodeInfoByRefcode(
		ctx,
		body.ReferCode,
	)

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

	// @TODO Update refer code alone with invitee ID. Should wrap operations in transactions.
	err, res := db.Transact(db.GetDB(), func(tx *sqlx.Tx) (error, interface{}) {
		q := models.New(tx)
		newUser, err := q.CreateUser(ctx, models.CreateUserParams{
			Username:      body.Username,
			Gender:        models.Gender(body.Gender),
			Uuid:          uuid,
			PremiumType:   models.PremiumTypeNormal,
			PhoneVerified: false,
		})

		if err != nil {
			return err, apperr.FailedToCreateUser
		}

		err = q.UpdateInviteeIDByRefCode(
			ctx,
			models.UpdateInviteeIDByRefCodeParams{
				InviteeID: sql.NullInt32{
					Int32: int32(newUser.ID),
					Valid: true,
				},
			},
		)

		if err != nil {
			return err, nil
		}

		return nil, newUser
	})

	if err != nil {
		c.AbortWithError(
			http.StatusBadRequest,
			apperr.NewErr(
				res.(string),
				err.Error(),
			),
		)

		return
	}

	newUser := res.(models.User)

	c.JSON(http.StatusOK, NewTransform().TransformUser(&newUser))
}

type VerifyUsernameBody struct {
	Username string `form:"username"`
}

func HandleVerifyUsername(c *gin.Context, depCon container.Container) {
	var (
		body VerifyUsernameBody
	)

	if err := requestbinder.Bind(c, &body); err != nil {
		c.AbortWithError(
			http.StatusBadRequest,
			apperr.NewErr(
				apperr.FailedToValidateVerifyUsernameParams,
				err.Error(),
			),
		)

		return
	}

	dao := NewRegisterDAO(db.GetDB())
	ctx := context.Background()

	exists, err := dao.CheckUsernameExists(ctx, body.Username)

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

	if exists {
		c.AbortWithError(
			http.StatusBadRequest,
			apperr.NewErr(apperr.UsernameNotAvailable),
		)

		return
	}

	c.JSON(http.StatusOK, struct{}{})
}

func HandleSendVerifyCode(c *gin.Context, depCon container.Container) {

}
