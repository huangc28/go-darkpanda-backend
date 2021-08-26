package register

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-redis/redis/v8"
	log "github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin"
	container "github.com/golobby/container/pkg/container"
	"github.com/huangc28/go-darkpanda-backend/config"
	"github.com/huangc28/go-darkpanda-backend/db"
	"github.com/huangc28/go-darkpanda-backend/internal/app/apperr"
	"github.com/huangc28/go-darkpanda-backend/internal/app/contracts"
	"github.com/huangc28/go-darkpanda-backend/internal/app/models"
	dpfcm "github.com/huangc28/go-darkpanda-backend/internal/app/pkg/firebase_messaging"
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

func RegisterHandler(c *gin.Context, depCon container.Container) {
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

	userFcmTopic := dpfcm.MakeDedicatedFCMTopicForUser(uuid)

	txResp := db.TransactWithFormatStruct(db.GetDB(), func(tx *sqlx.Tx) db.FormatResp {
		q := models.New(tx)

		newUser, err := q.CreateUser(ctx, models.CreateUserParams{
			Username:      body.Username,
			Gender:        models.Gender(body.Gender),
			Uuid:          uuid,
			PremiumType:   models.PremiumTypeNormal,
			PhoneVerified: false,
			FcmTopic: sql.NullString{
				Valid:  true,
				String: userFcmTopic,
			},
		})

		if err != nil {
			return db.FormatResp{
				Err:            err,
				ErrCode:        apperr.FailedToCreateUser,
				HttpStatusCode: http.StatusInternalServerError,
			}
		}

		if err = q.UpdateInviteeIDByRefCode(
			ctx,
			models.UpdateInviteeIDByRefCodeParams{
				InviteeID: sql.NullInt32{
					Int32: int32(newUser.ID),
					Valid: true,
				},
				RefCode: body.ReferCode,
			},
		); err != nil {
			return db.FormatResp{
				Err:            err,
				ErrCode:        apperr.FailedToUpdateInviteeIdByRefCode,
				HttpStatusCode: http.StatusInternalServerError,
			}

		}

		return db.FormatResp{
			Response: &newUser,
		}
	})

	if txResp.Err != nil {
		c.AbortWithError(
			txResp.HttpStatusCode,
			apperr.NewErr(
				txResp.ErrCode,
				err.Error(),
			),
		)

		return
	}

	newUser := txResp.Response.(*models.User)

	c.JSON(http.StatusOK, NewTransform().TransformUser(newUser))
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
	Mobile string `json:"mobile" form:"mobile" binding:"required,gt=0"`
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

	user, err := userDao.GetUserByUuid(body.Uuid, "id", "username")

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
	ctx := context.Background()
	vs := genverifycode.GenVerifyCode()

	if err := CreateRegisterMobileVerifyCode(
		ctx, CreateRegisterMobileVerifyCodeParams{
			RedisCli:   db.GetRedis(),
			UserUuid:   body.Uuid,
			VerifyCode: vs.BuildCode(),
			Mobile:     body.Mobile,
		},
	); err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToCreateRegisterMobileVerifyCode,
				err.Error(),
			),
		)

		return
	}

	// Send mobile verify code by twilio. If uuid is in the white list,
	// we don't send real SMS message because it's too expensive
	regDao := NewRegisterDAO(db.GetDB())
	exists, err := regDao.CheckUserInSMSWhiteList(ctx, contracts.CheckUserInSMSWhiteListParams{
		RedisClient: db.GetRedis(),
		Username:    user.Username,
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

	var (
		tc   twilio.TwilioServicer
		from string = config.GetAppConf().TwilioFrom
	)
	depCon.Make(&tc)

	if exists {
		log.Info("DEV account, bypassing real sms sending")

		// Set twilio config to use DEV.
		tc.SetConfig(twilio.TwilioConf{
			AccountSID:   config.GetAppConf().TwilioDevAccountID,
			AccountToken: config.GetAppConf().TwilioDevAuthToken,
		})

		from = config.GetAppConf().TwilioDevFrom
	}

	smsResp, err := tc.SendSMS(
		from,
		body.Mobile,
		fmt.Sprintf("your darkpanda verify code: \n\n %s", vs.BuildCode()),
	)

	if twilio.HandleSendTwilioError(c, err) != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToSendTwilioSMS,
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

	c.JSON(
		http.StatusOK,
		NewTransform().TransformSendMobileVerifyCode(
			body.Uuid,
			vs.Chars,
		),
	)
}

type VerifyMobileBody struct {
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
		"phone_verified",
	)

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

	if user.PhoneVerified {
		c.AbortWithError(
			http.StatusBadRequest,
			apperr.NewErr(
				apperr.UserAlreadyMobileVerified,
			),
		)

		return
	}

	// Retrieve verify code from redis.
	ctx := context.Background()
	vc, err := GetRegisterMobileVerifyCode(ctx, GetRegisterMobileVerifyCodeParams{
		RedisCli: db.GetRedis(),
		UserUuid: user.Uuid,
	})

	if err == redis.Nil {
		c.AbortWithError(
			http.StatusBadRequest,
			apperr.NewErr(apperr.FailedToGetRegisterMobileVerifyCode),
		)

		return
	}

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToGetRegisterMobileVerifyCode,
				err.Error(),
			),
		)

		return
	}

	if vc.VerifyCode != body.VerifyCode {
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
		Mobile:        &vc.Mobile,
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
