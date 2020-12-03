package auth

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/go-redis/redis/v8"
	"github.com/golobby/container/pkg/container"
	log "github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin"
	"github.com/huangc28/go-darkpanda-backend/config"
	"github.com/huangc28/go-darkpanda-backend/db"
	apperr "github.com/huangc28/go-darkpanda-backend/internal/app/apperr"
	"github.com/huangc28/go-darkpanda-backend/internal/app/auth/internal/twilio"
	"github.com/huangc28/go-darkpanda-backend/internal/app/models"
	genverifycode "github.com/huangc28/go-darkpanda-backend/internal/app/pkg/generate_verify_code"
	"github.com/huangc28/go-darkpanda-backend/internal/app/pkg/jwtactor"
	"github.com/huangc28/go-darkpanda-backend/internal/app/pkg/requestbinder"
	"github.com/huangc28/go-darkpanda-backend/internal/app/util"
	"github.com/spf13/viper"
	"github.com/ventu-io/go-shortid"
)

type AuthController struct {
	TwilioClient *twilio.TwilioClient
	Container    container.Container
}

type RegisterBody struct {
	Username  string `form:"username" uri:"username" json:"username" binding:"required"`
	Gender    string `form:"gender" uri:"gender" json:"gender" binding:"oneof='male' 'female'"`
	ReferCode string `form:"refer_code" uri:"refer_code" json:"refer_code" binding:"required"`
}

// We need the following to register new user
//   - reference code
//   - username
// @TODO check username is duplicated
func (ac *AuthController) RegisterHandler(c *gin.Context) {
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
	tx, err := db.GetDB().Begin()

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToBeginTx,
				err.Error(),
			),
		)

		return
	}

	newUser, err := q.WithTx(tx).CreateUser(c, models.CreateUserParams{
		Username:      body.Username,
		Gender:        models.Gender(body.Gender),
		Uuid:          uuid,
		PremiumType:   models.PremiumTypeNormal,
		PhoneVerified: false,
	})

	q.WithTx(tx).UpdateInviteeIDByRefCode(ctx, models.UpdateInviteeIDByRefCodeParams{
		InviteeID: sql.NullInt32{
			Int32: int32(newUser.ID),
			Valid: true,
		},

		RefCode: body.ReferCode,
	})

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToCreateUser,
				err.Error(),
			),
		)

		tx.Rollback()
		return
	}

	tx.Commit()
	c.JSON(http.StatusOK, NewTransform().TransformUser(&newUser))
}

// sends verification code to specified mobile number
//   - receive user uuid
//   - mobile number
//
// @TODOs
//   - verify code should be stored in redis instead of DB. It should stored in a key value pair like below:
//
//     {USER_UUID}: {verify_code}
//
//     Set the TTL to 1.5 mintues before expired.
// from the client.
type SendVerifyCodeBody struct {
	Uuid   string `form:"uuid" binding:"required,gt=0"`
	Mobile string `form:"mobile" json:"mobile" binding:"required,numeric,gt=0"`
}

func (ac *AuthController) SendVerifyCodeHandler(c *gin.Context) {
	var (
		body SendVerifyCodeBody
		ctx  context.Context = context.Background()
	)

	if err := requestbinder.Bind(c, &body); err != nil {
		c.AbortWithError(
			http.StatusBadRequest,
			apperr.NewErr(
				apperr.FailedToValidateSendVerifyCodeParams,
				err.Error(),
			),
		)

		return
	}

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
	if usr.PhoneVerified {
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

	// @TODO verify code should be stored in redis instead of DB.
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

	// ------------------- send verification code via twilio -------------------
	smsResp, err := ac.TwilioClient.SendSMS(
		viper.GetString("twilio.from"),
		body.Mobile,
		fmt.Sprintf("your darkpanda verify code: \n\n %d", verfDigs),
	)

	if twilio.HandleSendTwilioError(c, err) != nil {
		return
	}

	log.
		WithFields(log.Fields{
			"user_uuid": usr.Uuid,
			"mobile":    body.Mobile,
		}).
		Infof("sends twilio SMS success! %v", smsResp.SID)

	// ------------------- send sms code back -------------------
	c.JSON(http.StatusOK, NewTransform().TransformSendMobileVerifyCode(
		usr.Uuid,
		verPrefix,
	))
}

// ------------------- phone verification -------------------
// receive uuid
// receive verification code
// check the following to verify phone.
//   - phone_verified is false
//   - phone_verify_code stored in DB is the same as the one received from the client
// Sends jwt token back to client once its validated
type VerifyPhoneBody struct {
	Uuid       string `form:"uuid" json:"uuid" binding:"required,gt=0"`
	Mobile     string `form:"mobile" json:"mobile" binding:"required,gt=0"`
	VerifyCode string `form:"verify_code" json:"verify_code" binding:"required,gt=0"`
}

func (ac *AuthController) VerifyPhoneHandler(c *gin.Context) {
	var (
		ctx  context.Context = context.Background()
		body VerifyPhoneBody
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

	// ------------------- if user is already verified, return error -------------------
	if user.PhoneVerified {
		c.AbortWithError(
			http.StatusBadRequest,
			apperr.NewErr(apperr.UserHasPhoneVerified),
		)

		return
	}

	// ------------------- set the user to be phone verified -------------------
	if err := q.UpdateVerifyStatusById(ctx, models.UpdateVerifyStatusByIdParams{
		ID:            user.ID,
		PhoneVerified: true,
		Mobile: sql.NullString{
			String: body.Mobile,
			Valid:  true,
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

	c.JSON(http.StatusOK, struct {
		JwtToken string `json:"jwt"`
	}{token})
}

// store jwt token in redis and db
type RevokeJwtBody struct {
	Jwt string `json:"jwt" binding:"required,gt=0"`
}

func (ac *AuthController) RevokeJwtHandler(c *gin.Context) {
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

type SendLoginVerifyCodeBody struct {
	Username string `form:"username"  json:"username" binding:"required,gt=0"`
}

func (ac *AuthController) SendLoginVerifyCode(c *gin.Context) {
	var (
		body SendLoginVerifyCodeBody
		ctx  context.Context = context.Background()
	)

	// receive username
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

	// try finding username
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

	// Check if login record already exists in redis.
	authenticator, err := authDao.GetLoginRecord(ctx, user.Uuid)

	if err != nil {
		// If authenticator does not exists, that means this is the first time the user
		// performs login. We should create an authentication record in redis for this user.
		if err == redis.Nil {
			authenticator, err = authDao.CreateLoginVerifyCode(
				ctx,
				verifyCode.BuildCode(),
				user.Uuid,
			)

			if err != nil {
				c.AbortWithError(
					http.StatusInternalServerError,
					apperr.NewErr(
						apperr.UnableToCreateSendVerifyCode,
						err.Error(),
					),
				)

				return
			}

			// send mobile verify code via twilio
			smsResp, err := ac.TwilioClient.SendSMS(
				viper.GetString("twilio.from"),
				user.Mobile.String,
				fmt.Sprintf("your darkpanda verify code: \n\n %d", verifyCode.BuildCode()),
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

			c.JSON(http.StatusOK, NewTransform().TransformSendLoginMobileVerifyCode(
				user.Uuid,
				verifyCode.Chars,
				user.Mobile.String,
			))

			return
		} else {
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

	// send verify code via twilio
	smsResp, err := ac.TwilioClient.SendSMS(
		viper.GetString("twilio.from"),
		user.Mobile.String,
		fmt.Sprintf("your darkpanda verify code: \n\n %d", verifyCode.BuildCode()),
	)

	if twilio.HandleSendTwilioError(c, err) != nil {
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

func (ac *AuthController) VerifyLoginCode(c *gin.Context) {
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
	authRecord, _ := authDao.GetLoginRecord(ctx, user.Uuid)

	if authRecord.VerifyCode != fmt.Sprintf("%s-%d", body.VerifyChar, body.VerifyDig) {
		c.AbortWithError(
			http.StatusBadRequest,
			apperr.NewErr(apperr.VerifyCodeUnmatched),
		)

		return
	}

	// Verify code matches, Generate jwt token.
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
