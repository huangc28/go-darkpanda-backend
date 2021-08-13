package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-redis/redis/v8"
	"github.com/golobby/container/pkg/container"
	log "github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin"
	"github.com/huangc28/go-darkpanda-backend/config"
	"github.com/huangc28/go-darkpanda-backend/db"
	apperr "github.com/huangc28/go-darkpanda-backend/internal/app/apperr"
	"github.com/huangc28/go-darkpanda-backend/internal/app/contracts"
	"github.com/huangc28/go-darkpanda-backend/internal/app/models"
	genverifycode "github.com/huangc28/go-darkpanda-backend/internal/app/pkg/generate_verify_code"

	"github.com/huangc28/go-darkpanda-backend/internal/app/pkg/jwtactor"
	"github.com/huangc28/go-darkpanda-backend/internal/app/pkg/requestbinder"
	"github.com/huangc28/go-darkpanda-backend/internal/app/pkg/twilio"
)

type AuthController struct {
	TwilioClient *twilio.TwilioClient
	Container    container.Container
}

// store jwt token in redis and db
type RevokeJwtBody struct {
	Jwt string ` form:"jwt" json:"jwt" binding:"required,gt=0"`
}

func (ac *AuthController) RevokeJwtHandler(c *gin.Context) {
	var ctx context.Context = context.Background()

	jwt := c.GetString("jwt")

	// Retrieve expired timestamp of the jwt token
	claims, err := jwtactor.ParseToken(jwt, config.GetAppConf().JwtSecret)

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToParseJwtToken,
				err.Error(),
			),
		)

		return
	}

	// ------------------- invalidate jwt -------------------
	authDao := NewAuthDao(db.GetRedis())

	if err := authDao.RevokeJwt(ctx, jwt, claims.ExpiresAt); err != nil {
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

type SendLoginVerifyCodeBody struct {
	Username string `form:"username" json:"username" binding:"required,gt=0"`
}

func (ac *AuthController) SendVerifyCodeHandler(c *gin.Context, depCon container.Container) {
	var (
		body SendLoginVerifyCodeBody
		ctx  context.Context = context.Background()
	)

	if err := requestbinder.Bind(c, &body); err != nil {
		c.AbortWithError(
			http.StatusBadRequest,
			apperr.NewErr(
				apperr.FailedToCheckSendLoginVerifyCodeParams,
				err.Error(),
			),
		)

		return
	}

	q := models.New(db.GetDB())
	user, err := q.GetUserByUsername(ctx, body.Username)

	if err != nil {
		if err == sql.ErrNoRows {
			c.AbortWithError(
				http.StatusNotFound,
				apperr.NewErr(
					apperr.FailedToCheckSendLoginVerifyCodeParams,
				),
			)

			return
		}

		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToGetUserByUsername,
				err.Error(),
			),
		)

		return
	}

	// check if the user is phone verified. If not, the system can not send verify code.
	// The client should redirect the user to verify mobile page.
	if !user.PhoneVerified {
		c.AbortWithError(
			http.StatusBadRequest,
			apperr.NewErr(apperr.UnableToSendVerifyCodeToUnverfiedNumber),
		)

		return
	}

	// Generate an SMS login code.
	verifyCode := genverifycode.GenVerifyCode()
	authDao := NewAuthDao(db.GetRedis())

	var (
		tc     twilio.TwilioServicer
		regDao contracts.Registerar
		from   string = config.GetAppConf().TwilioFrom
	)

	ac.Container.Make(&tc)
	ac.Container.Make(&regDao)

	// If login user is in white list, we will send sandbox message instead
	// of real message to save some money.
	exists, err := regDao.CheckUserInSMSWhiteList(ctx, contracts.CheckUserInSMSWhiteListParams{
		RedisClient: db.GetRedis(),
		Username:    body.Username,
	})

	if err != nil && !errors.Is(err, redis.Nil) {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToCheckUserExistsInSMSWhiteList,
				err.Error(),
			),
		)

		return
	}

	if exists {
		log.Info("DEV account, bypassing real sms sending")

		tc.SetConfig(twilio.TwilioConf{
			AccountSID:   config.GetAppConf().TwilioDevAccountID,
			AccountToken: config.GetAppConf().TwilioDevAuthToken,
		})

		from = config.GetAppConf().TwilioDevFrom
	}

	// Check if login record already exists in redis.
	authenticator, err := authDao.GetLoginRecord(ctx, user.Uuid)

	if err != nil {
		// If authenticator does not exists, that means this is the first time the user
		// performs login. We should create an authentication record in redis for this user.
		if err == redis.Nil {
			log.Printf("DEBUG verify code 1 %s", verifyCode.BuildCode())

			ctx := context.Background()

			if _, err := authDao.CreateLoginVerifyCode(
				ctx,
				verifyCode.BuildCode(),
				user.Uuid,
			); err != nil {
				c.AbortWithError(
					http.StatusInternalServerError,
					apperr.NewErr(
						apperr.FailedToCreateAuthenticatorRecordInRedis,
						err.Error(),
					),
				)

				return
			}

			smsResp, err := tc.SendSMS(
				from,
				user.Mobile.String,
				fmt.Sprintf("your darkpanda verify code: \n\n %s", verifyCode.BuildCode()),
			)

			if err != nil {
				c.AbortWithError(
					http.StatusInternalServerError,
					apperr.NewErr(
						apperr.FailedToSendTwilioMessage,
						err.Error(),
					),
				)

				return
			}

			log.
				WithFields(log.Fields{
					"user_uuid": user.Uuid,
					"mobile":    user.Mobile.String,
				}).
				Infof("sends twilio SMS success, login verify code created ! %v", smsResp.SID)

			c.JSON(http.StatusOK, NewTransform().TransformSendLoginMobileVerifyCode(
				user.Uuid,
				verifyCode.Chars,
				user.Mobile.String,
			))

			return
		}

		// Error occurs when trying to get authentication record from redis, return error.
		c.AbortWithError(
			http.StatusBadRequest,
			apperr.NewErr(
				apperr.FailedToCreateAuthenticatorRecordInRedis,
				err.Error(),
			),
		)

		return
	}

	// Authenticator record is found. Check number of retries the user has attempt
	if authenticator.NumRetried >= LimitOnLoginRetry {
		c.AbortWithError(
			http.StatusBadRequest,
			apperr.NewErr(apperr.ExceedingLoginRetryLimit),
		)

		return
	}

	// Let's update the authenticator record in redis.
	if err := authDao.UpdateLoginRecord(ctx, user.Uuid, LoginAuthenticator{
		VerifyCode: verifyCode.BuildCode(),
		NumRetried: authenticator.NumRetried + 1,
	}); err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToUpdateAuthenticatorRecordInRedis,
				err.Error(),
			),
		)

		return
	}

	smsResp, err := tc.SendSMS(
		from,
		user.Mobile.String,
		fmt.Sprintf("[Darkpanda] login verify code: \n\n %s", verifyCode.BuildCode()),
	)

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToSendTwilioMessage,
				err.Error(),
			),
		)

		return
	}

	log.
		WithFields(log.Fields{
			"user_uuid": user.Uuid,
			"mobile":    user.Mobile.String,
		}).
		Infof("sends twilio SMS success, login verify code updated ! %v", smsResp.SID)

	c.JSON(http.StatusOK, NewTransform().TransformSendLoginMobileVerifyCode(
		user.Uuid,
		verifyCode.Chars,
		user.Mobile.String,
	))
}

type VerifyLoginCodeBody struct {
	Mobile     string `form:"mobile" json:"mobile" bind:"required,gt=0"`
	UUID       string `form:"uuid" json:"uuid" bind:"required,gt=0"`
	VerifyChar string `form:"verify_char" json:"verify_char" bind:"required,gt=0"`
	VerifyDig  int    `form:"verify_dig" json:"verify_dig" bind:"required,gt=0"`
}

func (ac *AuthController) VerifyLoginCode(c *gin.Context, depCon container.Container) {
	var (
		body VerifyLoginCodeBody
		ctx  context.Context = context.Background()
	)

	if err := requestbinder.Bind(c, &body); err != nil {
		c.AbortWithError(
			http.StatusBadRequest,
			apperr.NewErr(
				apperr.FailedToValidateVerifyPhoneParams,
				err.Error(),
			),
		)

		return
	}

	// Try to retrieve authenticator record from redis via UUID.
	// If given key does not exist, response with error.
	q := models.New(db.GetDB())
	user, err := q.GetUserByUuid(ctx, body.UUID)

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

	// Retrieve auth record from redis
	authDao := NewAuthDao(db.GetRedis())
	authRecord, err := authDao.GetLoginRecord(ctx, user.Uuid)

	if err != nil {
		if err == redis.Nil {
			c.AbortWithError(
				http.StatusBadRequest,
				apperr.NewErr(apperr.LoginVerifyCodeNotFound),
			)

			return
		}

		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToGetAuthenticatorRecord,
				err.Error(),
			),
		)

		return

	}

	if authRecord.VerifyCode != fmt.Sprintf("%s-%d", body.VerifyChar, body.VerifyDig) {
		c.AbortWithError(
			http.StatusBadRequest,
			apperr.NewErr(apperr.VerifyCodeUnmatched),
		)

		return
	}

	// Verify code matches, generate jwt token.
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
		JWT string `json:"jwt"`
	}{jwt})
}
