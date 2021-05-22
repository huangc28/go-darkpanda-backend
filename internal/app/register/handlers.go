package register

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin"
	container "github.com/golobby/container/pkg/container"
	"github.com/huangc28/go-darkpanda-backend/config"
	"github.com/huangc28/go-darkpanda-backend/db"
	"github.com/huangc28/go-darkpanda-backend/internal/app/apperr"
	"github.com/huangc28/go-darkpanda-backend/internal/app/contracts"
	"github.com/huangc28/go-darkpanda-backend/internal/app/models"
	genverifycode "github.com/huangc28/go-darkpanda-backend/internal/app/pkg/generate_verify_code"
	"github.com/huangc28/go-darkpanda-backend/internal/app/pkg/jwtactor"
	"github.com/huangc28/go-darkpanda-backend/internal/app/pkg/requestbinder"
	"github.com/huangc28/go-darkpanda-backend/internal/app/pkg/twilio"
	"github.com/jmoiron/sqlx"
	"github.com/teris-io/shortid"
)

type RegisterBody struct {
	Username  string `form:"username" uri:"username" json:"username" binding:"required"`
	Gender    string `form:"gender" uri:"gender" json:"gender" binding:"oneof='male' 'female'"`
	ReferCode string `form:"refer_code" uri:"refer_code" json:"refer_code" binding:"required"`
}

func RegisterHandler(c *gin.Context) {
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
	regSrv := NewRegisterService(NewRegisterDAO(db.GetDB()))

	if err := regSrv.ValidateReferralCode(ctx, body.ReferCode); err != nil {
		rcErr := err.(*ValidateReferralCodeError)

		var (
			errCode    string
			httpStatus int = http.StatusBadRequest
		)

		switch rcErr.ErrCode {
		case ReferralCodeNotExists:
			errCode = apperr.FailedToRetrieveReferCodeInfo
		case ReferralCodeExpired:
			errCode = apperr.ReferralCodeExpired
		case ReferralCodeOccupied:
			errCode = apperr.ReferCodeOccupied
		default:
			errCode = apperr.FailedToValidateReferralCode
			httpStatus = http.StatusInternalServerError
		}

		c.AbortWithError(
			httpStatus,
			apperr.NewErr(
				errCode,
				rcErr.Error(),
			),
		)

		return

	}

	// If refer code and username are all valid, create a new user.
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
	Username string `form:"username" json:"username" binding:"required,gt=0"`
}

func VerifyUsernameHandler(c *gin.Context, depCon container.Container) {
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

type VerifyReferralCodeBody struct {
	ReferralCode string `form:"referral_code" json:"referral_code" binding:"required,gt=0"`
}

func VerifyReferralCodeHandler(c *gin.Context, depCon container.Container) {
	var body VerifyReferralCodeBody

	if err := requestbinder.Bind(c, &body); err != nil {
		c.AbortWithError(
			http.StatusBadRequest,
			apperr.NewErr(
				apperr.FailedToVerifyReferralCode,
				err.Error(),
			),
		)
	}

	ctx := context.Background()
	srv := NewRegisterService(NewRegisterDAO(db.GetDB()))

	if err := srv.ValidateReferralCode(ctx, body.ReferralCode); err != nil {
		rcErr := err.(*ValidateReferralCodeError)
		var (
			errCode    string
			httpStatus int = http.StatusBadRequest
		)

		switch rcErr.ErrCode {
		case ReferralCodeNotExists:
			errCode = apperr.FailedToRetrieveReferCodeInfo
		case ReferralCodeExpired:
			errCode = apperr.ReferralCodeExpired
		case ReferralCodeOccupied:
			errCode = apperr.ReferCodeOccupied
		default:
			errCode = apperr.FailedToValidateReferralCode
			httpStatus = http.StatusInternalServerError
		}

		c.AbortWithError(
			httpStatus,
			apperr.NewErr(
				errCode,
				rcErr.Error(),
			),
		)

		return
	}

	c.JSON(http.StatusOK, struct{}{})
}

type SendMobileVerifyCodeHandlerBody struct {
	Uuid   string `json:"uuid" form:"uuid" binding:"required,gt=0"`
	Mobile string `json:"mobile" form:"mobile" binding:"required" binding:"required,gt=0"`
}

func SendMobileVerifyCodeHandler(c *gin.Context, depCon container.Container) {
	var body SendMobileVerifyCodeHandlerBody

	// Find user by uuid.
	if err := requestbinder.Bind(c, &body); err != nil {
		c.AbortWithError(
			http.StatusBadRequest,
			apperr.NewErr(
				apperr.FailedToSendMobileVerifyCode,
				err.Error(),
			),
		)

		return
	}

	var userDao contracts.UserDAOer
	depCon.Make(&userDao)

	user, err := userDao.GetUserByUuid(body.Uuid, "id")

	if err != nil {
		if err == sql.ErrNoRows {
			c.AbortWithError(
				http.StatusBadRequest,
				apperr.NewErr(
					apperr.UserNotFoundByUuid,
					err.Error(),
				),
			)

			return
		}

		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToGetUserByUuid,
				err.Error(),
			),
		)

		return
	}

	// Generate phone verify code and update user record.
	vs := genverifycode.GenVerifyCode()
	pVs := vs.BuildCode()

	if _, err = userDao.UpdateUserInfoByUuid(
		contracts.UpdateUserInfoParams{
			Uuid:            body.Uuid,
			PhoneVerifyCode: &pVs,
		},
	); err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToUpdateUserPhoneVerifyCode,
				err.Error(),
			),
		)

		return
	}

	// Send mobile verify code by twilio
	var tc twilio.TwilioServicer
	depCon.Make(&tc)
	smsResp, err := tc.SendSMS(
		config.GetAppConf().TwilioFrom,
		body.Mobile,
		fmt.Sprintf("your darkpanda verify code: \n\n %s", vs.BuildCode()),
	)

	if twilio.HandleSendTwilioError(c, err) != nil {
		return
	}

	log.
		WithFields(log.Fields{
			"user_uuid": user.Uuid,
			"mobile":    user.Mobile.String,
		}).
		Infof("sends twilio SMS success, login verify code created ! %v", smsResp.SID)

	c.JSON(
		http.StatusOK,
		NewTransform().TransformSendMobileVerifyCode(
			body.Uuid,
			vs.Chars,
		),
	)
}

type VerifyMobileBody struct {
	Mobile     string `form:"mobile" json:"mobile" bind:"required,gt=0"`
	UUID       string `form:"uuid" json:"uuid" bind:"required,gt=0"`
	VerifyCode string `form:"verify_code" json:"verify_code" bind:"required,gt=0"`
}

func VerifyMobileHandler(c *gin.Context, depCon container.Container) {
	var body VerifyMobileBody

	if err := requestbinder.Bind(c, &body); err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToValidateRequestBody,
				err.Error(),
			),
		)

		return
	}

	var userDao contracts.UserDAOer
	depCon.Make(&userDao)

	user, err := userDao.GetUserByUuid(
		body.UUID,
		"id",
		"uuid",
		"phone_verify_code",
		"phone_verified",
	)

	if err != nil {
		if err == sql.ErrNoRows {
			c.AbortWithError(
				http.StatusBadRequest,
				apperr.NewErr(
					apperr.UserNotFoundByUuid,
					err.Error(),
				),
			)

			return

		}

		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToGetUserByUuid,
				err.Error(),
			),
		)

		return
	}

	if user.PhoneVerified {
		c.AbortWithError(
			http.StatusBadRequest,
			apperr.NewErr(
				apperr.UserAlreadyMobileVerified,
			),
		)

		return
	}

	if user.PhoneVerifyCode.String != body.VerifyCode {
		c.AbortWithError(
			http.StatusBadRequest,
			apperr.NewErr(
				apperr.PhoneVerifyCodeNotMatch,
			),
		)

		return
	}

	phoneVerified := true
	if _, err := userDao.UpdateUserInfoByUuid(contracts.UpdateUserInfoParams{
		Uuid:          body.UUID,
		PhoneVerified: &phoneVerified,
		Mobile:        body.Mobile,
	}); err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToUpdateUserByUuid,
				err.Error(),
			),
		)

		return
	}

	// Retreive verifier by user uuid
	// If user already mobile verified, respond with ok status
	// If user is not mobile verified, try compare the verify code
	// If verify code matches, update the user to be mobile verified
	// If verify code isn't match, respond bad request.
	// Response with auth jwt token.

	jwt, err := jwtactor.CreateToken(
		user.Uuid,
		config.GetAppConf().JwtSecret,
	)

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToCreateJWTToken,
				err.Error(),
			),
		)

		return
	}

	c.JSON(http.StatusOK, struct {
		Jwt string `json:"jwt"`
	}{jwt})
}
