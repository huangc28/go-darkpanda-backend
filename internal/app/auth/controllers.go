package auth

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin"
	"github.com/huangc28/go-darkpanda-backend/config"
	"github.com/huangc28/go-darkpanda-backend/db"
	apperr "github.com/huangc28/go-darkpanda-backend/internal/app/apperr"
	"github.com/huangc28/go-darkpanda-backend/internal/app/auth/internal/twilio"
	"github.com/huangc28/go-darkpanda-backend/internal/app/pkg/jwtactor"
	"github.com/huangc28/go-darkpanda-backend/internal/app/util"
	"github.com/huangc28/go-darkpanda-backend/internal/models"
	"github.com/spf13/viper"
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
	dao := NewUserCheckerDAO(db.GetDB())
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
	Username string `json:"username" binding:"required,gt=0"`
	Mobile   string `json:"mobile" binding:"required,numeric,gt=0"`
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
	usr, err := q.GetUserByUsername(ctx, body.Username)

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
			apperr.NewErr(apperr.UserHasPhoneVerified),
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
	twilioClient := twilio.New(twilio.TwilioConf{
		AccountSID:   viper.GetString("twilio.account_id"),
		AccountToken: viper.GetString("twilio.auth_token"),
	})

	smsResp, err := twilioClient.SendSMS(
		viper.GetString("twilio.from"),
		body.Mobile,
		fmt.Sprintf("your darkpanda verify code: \n\n %d", verfDigs),
	)

	if err != nil {
		if _, isTwilioErr := err.(*twilio.SMSError); isTwilioErr {
			log.Fatalf("twilio sends back failed response %s", err.Error())

			c.AbortWithError(
				http.StatusBadRequest,
				apperr.NewErr(
					apperr.TwilioRespErr,
					err.Error(),
				),
			)

			return
		}

		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToSendTwilioSMSErr,
				err.Error(),
			),
		)

		return
	}

	log.
		WithFields(log.Fields{
			"user_uuid": usr.Uuid,
			"mobile":    body.Mobile,
		}).
		Infof("sends twilio SMS success! %v", smsResp.SID)

	// ------------------- send sms code back -------------------
	res := struct {
		Uuid         string `json:"uuid"`
		VerifyPrefix string `json:"verify_prefix"`
		VerifySuffix int    `json:"verify_suffix"`
	}{
		usr.Uuid,
		verPrefix,
		verfDigs,
	}

	c.JSON(http.StatusOK, &res)
}

// ------------------- phone verification -------------------
// receive uuid
// receive verification code
// check the following to verify phone.
//   - phone_verified is false
//   - phone_verify_code is the same as the one received from client
// Sends jwt token back to client once its validated
type VerifyPhoneBody struct {
	Uuid       string `json:"uuid" binding:"required,gt=0"`
	VerifyCode string `json:"verify_code" binding:"required,gt=0"`
}

func VerifyPhoneHandler(c *gin.Context) {
	var (
		ctx  context.Context = context.Background()
		body VerifyPhoneBody
	)

	if err := c.ShouldBindJSON(&body); err != nil {
		c.AbortWithError(
			http.StatusBadRequest,
			apperr.NewErr(
				apperr.FailedToValidateVerifyPhoneParams,
				err.Error(),
			),
		)

		return
	}

	// ------------------- check if the given verify code exists in DB -------------------
	q := models.New(db.GetDB())
	user, err := q.GetUserByVerifyCode(ctx, sql.NullString{
		String: body.VerifyCode,
		Valid:  true,
	})

	if err != nil {
		if err == sql.ErrNoRows {
			c.AbortWithError(
				http.StatusNotFound,
				apperr.NewErr(apperr.UserNotFoundByVerifyCode),
			)

			return
		}

		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToGetUserByVerifyCode,
				err.Error(),
			),
		)

		return
	}

	// ------------------- makesure verify code given matches -------------------
	if user.PhoneVerifyCode.String != body.VerifyCode {
		c.AbortWithError(
			http.StatusBadRequest,
			apperr.NewErr(apperr.VerifyCodeNotMatching),
		)

		return
	}

	// ------------------- set the user to be phone verified -------------------
	if user.PhoneVerified.Bool == false {
		if err := q.UpdateVerifyStatusById(ctx, models.UpdateVerifyStatusByIdParams{
			ID: user.ID,
			PhoneVerified: sql.NullBool{
				Bool:  true,
				Valid: true,
			},
		}); err != nil {
			c.AbortWithError(
				http.StatusInternalServerError,
				apperr.NewErr(
					apperr.FailedToUpdateVerifyStatus,
					err.Error(),
				),
			)

			return
		}
	}

	// ------------------- generate jwt token and return it -------------------
	token, err := jwtactor.CreateToken(
		user.Uuid,
		config.GetAppConf().JwtSecret,
	)

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToGenerateJwtToken,
				err.Error(),
			),
		)

		return
	}

	resp := struct {
		JwtToken string `json:"jwt"`
	}{token}

	c.JSON(http.StatusOK, &resp)
}

// store jwt token in redis and db
type RevokeJwtBody struct {
	Jwt string `json:"jwt" binding:"required,gt=0"`
}

func RevokeJwtHandler(c *gin.Context) {
	var (
		ctx context.Context = context.Background()
	)

	jwt := c.GetString("jwt")

	// ------------------- invalidate jwt -------------------
	authDao := NewAuthDao(db.GetRedis())

	if err := authDao.RevokeJwt(ctx, jwt); err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToInvalidateSignature,
				err.Error(),
			),
		)

		return
	}

	c.JSON(http.StatusOK, struct{}{})
}
