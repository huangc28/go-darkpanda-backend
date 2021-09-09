package user

import (
	"context"
	"database/sql"
	"fmt"

	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/golobby/container/pkg/container"
	"github.com/huangc28/go-darkpanda-backend/config"
	"github.com/huangc28/go-darkpanda-backend/db"
	"github.com/huangc28/go-darkpanda-backend/internal/app/apperr"
	"github.com/huangc28/go-darkpanda-backend/internal/app/contracts"
	"github.com/huangc28/go-darkpanda-backend/internal/app/models"
	genverifycode "github.com/huangc28/go-darkpanda-backend/internal/app/pkg/generate_verify_code"
	"github.com/huangc28/go-darkpanda-backend/internal/app/pkg/requestbinder"
	"github.com/huangc28/go-darkpanda-backend/internal/app/pkg/twilio"
	log "github.com/sirupsen/logrus"
)

type UserHandlers struct {
	Container container.Container
}

func (h *UserHandlers) GetMyProfileHandler(c *gin.Context) {
	var (
		uuid string          = c.GetString("uuid")
		ctx  context.Context = context.Background()
	)

	q := models.New(db.GetDB())
	usr, err := q.GetUserByUuid(ctx, uuid)

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

	c.JSON(http.StatusOK, NewTransform().TransformUser(&usr))
}

type GetUserProfileBody struct {
	UUID string `form:"uuid" json:"uuid" binding:"required,gt=0"`
}

func (h *UserHandlers) GetUserProfileHandler(c *gin.Context, depCon container.Container) {
	var (
		uuid string          = c.Param("uuid")
		ctx  context.Context = context.Background()
	)

	q := models.New(db.GetDB())
	user, err := q.GetUserByUuid(ctx, uuid)

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

	// Get user ratings.
	var srvDao contracts.UserDAOer
	depCon.Make(&srvDao)

	userRating, err := srvDao.GetRating(int(user.ID))

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToGetUserRating,
				err.Error(),
			),
		)

		return
	}

	tResp, err := NewTransform().TransformViewableUserProfile(user, *userRating)

	if err != nil {
		log.Fatal(err)
	}

	c.JSON(http.StatusOK, tResp)
}

type CreateImageParams struct {
	UserID int64
	URL    string
}

type PutUserInfoBody struct {
	AvatarURL    *string             `form:"avatar_url" json:"avatar_url"`
	Nationality  *string             `form:"nationality" json:"nationality"`
	Region       *string             `form:"region" json:"region"`
	Age          int                 `form:"age" json:"age"`
	Height       float64             `form:"height" json:"height"`
	Weight       float64             `form:"weight" json:"weight"`
	Habbits      *string             `form:"habbits" json:"habbits"`
	Description  *string             `form:"description" json:"description"`
	Images       []CreateImageParams `form:"imageList" json:"imageList"`
	RemoveImages []CreateImageParams `form:"removeImageList" json:"removeImageList"`
}

func (h *UserHandlers) PutUserInfo(c *gin.Context) {
	body := &PutUserInfoBody{}

	if err := requestbinder.Bind(c, &body); err != nil {
		c.AbortWithError(
			http.StatusBadRequest,
			apperr.NewErr(
				apperr.FailedToValidatePutUserParams,
				err.Error(),
			),
		)

		return
	}

	// ------------------- Update user info -------------------
	uuid := c.GetString("uuid")
	dao := NewUserDAO(db.GetDB())
	user, err := dao.UpdateUserInfoByUuid(contracts.UpdateUserInfoParams{
		Uuid:        uuid,
		AvatarURL:   body.AvatarURL,
		Nationality: body.Nationality,
		Region:      body.Region,
		Age:         &body.Age,
		Height:      &body.Height,
		Weight:      &body.Weight,
		Description: body.Description,
	})

	if err != nil {
		log.WithFields(log.Fields{
			"uuid": uuid,
		}).Errorf("Failed to patch user info by uuid %s", err.Error())

		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToPatchUserInfo,
				err.Error(),
			),
		)

		return
	}

	if len(body.RemoveImages) > 0 {

		for i := 0; i < len(body.RemoveImages); i++ {
			if err := dao.DeleteUserImages(body.RemoveImages[i].URL); err != nil {
				log.Fatalf("Failed to remove images %s", err.Error())
			}
		}
	}

	var imageDAO contracts.ImageDAOer
	h.Container.Make(&imageDAO)

	if len(body.Images) > 0 {
		imagesParams := make([]models.Image, 0)
		for i := 0; i < len(body.Images); i++ {
			imagesParams = append(imagesParams, models.Image{
				UserID: int32(user.ID),
				Url:    body.Images[i].URL,
			})
		}

		if err := imageDAO.CreateImages(imagesParams); err != nil {
			log.Fatalf("Failed to insert images %s", err.Error())

			c.AbortWithError(
				http.StatusBadRequest,
				apperr.NewErr(
					apperr.FailedToSendMobileVerifyCode,
					err.Error(),
				),
			)

			return
		}
		fmt.Print(imagesParams)
	}

	c.JSON(http.StatusOK, NewTransform().TransformPatchedUser(user))
}

func (h *UserHandlers) PatchUserImages(c *gin.Context) {
	c.JSON(http.StatusOK, struct{}{})
}

type GetUserImagesBody struct {
	Offset  int `form:"offset,default=0"`
	PerPage int `form:"perpage,default=9"`
}

func (h *UserHandlers) GetUserImagesHandler(c *gin.Context) {
	uuid := c.Param("uuid")

	body := &GetUserImagesBody{}

	if err := requestbinder.Bind(c, body); err != nil {
		c.AbortWithError(
			http.StatusBadRequest,
			apperr.NewErr(
				apperr.FailedToValidateGetUserImagesParams,
				err.Error(),
			),
		)

		return
	}

	// Get image link by user uuid
	images, err := NewUserDAO(db.GetDB()).GetUserImagesByUuid(
		uuid,
		body.Offset,
		body.PerPage,
	)

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToGetImagesByUserUUID,
				err.Error(),
			),
		)

		return
	}

	c.JSON(http.StatusOK, NewTransform().TransformUserImages(images))
}

func (h *UserHandlers) GetUserPayments(c *gin.Context) {
	uuid := c.Param("uuid")

	var paymentDAO contracts.PaymentDAOer
	h.Container.Make(&paymentDAO)

	paymentInfos, err := paymentDAO.GetPaymentsByUuid(uuid)

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToGetUserPayments,
				err.Error(),
			),
		)

		return

	}

	trfmPaymentInfo, err := NewTransform().TransformPaymentInfo(paymentInfos)

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToTransformUserPayments,
				err.Error(),
			),
		)

		return
	}

	c.JSON(http.StatusOK, trfmPaymentInfo)
}

type GetUserServiceHistoryRecords struct {
	Offset  int `form:"offset,default=0"`
	PerPage int `form:"perpage,default=5"`
}

func (h *UserHandlers) GetUserServiceHistory(c *gin.Context) {
	uuid := c.Param("uuid")
	body := GetUserServiceHistoryRecords{}

	if err := requestbinder.Bind(c, &body); err != nil {
		c.AbortWithError(
			http.StatusBadRequest,
			apperr.NewErr(
				apperr.FailedToValidateGetServiceHistoryParams,
				err.Error(),
			),
		)

		return
	}

	var serviceDAO contracts.ServiceDAOer
	h.Container.Make(&serviceDAO)

	// Retrieve past service records.
	services, err := serviceDAO.GetUserHistoricalServicesByUuid(
		uuid,
		body.PerPage,
		body.Offset,
	)

	if err != nil {
		c.AbortWithError(
			http.StatusBadRequest,
			apperr.NewErr(
				apperr.FailedToGetHistoricalServices,
				err.Error(),
			),
		)

		return
	}

	trfmSrvs, err := NewTransform().TransformHistoricalServices(services)

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToTransformHistoricalServices,
				err.Error(),
			),
		)

		return
	}

	c.JSON(http.StatusOK, trfmSrvs)
}

type GetUserRatingsBody struct {
	Offset  int `form:"offset,default=0"`
	PerPage int `form:"per_page,default=5"`
}

func (h *UserHandlers) GetUserRatings(c *gin.Context, depCon container.Container) {
	body := GetUserRatingsBody{}

	if err := requestbinder.Bind(c, &body); err != nil {
		c.AbortWithError(
			http.StatusBadRequest,
			apperr.NewErr(
				apperr.FailedToBindBodyParams,
				err.Error(),
			),
		)

		return

	}

	userUuid := c.Param("uuid")
	userDao := NewUserDAO(db.GetDB())
	targetUser, err := userDao.GetUserByUuid(userUuid, "id")

	if err != nil {

		if err == sql.ErrNoRows {
			c.AbortWithError(
				http.StatusBadRequest,
				apperr.NewErr(
					apperr.FailedToGetUserByUuid,
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

	// Retrieve all rating of services that I have participated in.
	var rateDao contracts.RateDAOer
	depCon.Make(&rateDao)
	ratings, err := rateDao.GetUserRatings(
		contracts.GetUserRatingsParams{
			UserId:  int(targetUser.ID),
			PerPage: body.PerPage,
			Offset:  body.Offset,
		},
	)

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToGetRatings,
				err.Error(),
			),
		)

		return
	}

	trms := TrfGetUserRatings(ratings)

	c.JSON(http.StatusOK, struct {
		Ratings []TrfmedUserRating `json:"ratings"`
	}{
		Ratings: trms,
	})
}

type ChangeMobileVerifyCodeParams struct {
	Mobile string `json:"mobile" form:"mobile" binding:"required,gt=0"`
}

func ChangeMobileVerifyCodeHandler(c *gin.Context, depCon container.Container) {
	body := ChangeMobileVerifyCodeParams{}

	if err := requestbinder.Bind(c, &body); err != nil {
		c.AbortWithError(
			http.StatusBadRequest,
			apperr.NewErr(
				apperr.FailedToBindBodyParams,
				err.Error(),
			),
		)

		return
	}

	ctx := context.Background()

	// Generate verify code.
	verifyCode := genverifycode.GenVerifyCode()

	// Store verify code in redis.
	if err := CreateChangeMobileVerifyCode(
		ctx,
		CreateChangeMobileVerifyCodeParams{
			RedisCli:   db.GetRedis(),
			VerifyCode: verifyCode.BuildCode(),
			UserUuid:   c.GetString("uuid"),
			Mobile:     body.Mobile,
		},
	); err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToCreateChangeMobileVerifyCode,
				err.Error(),
			),
		)

		return
	}

	// Send verify code via twilio.
	var tc twilio.TwilioServicer
	depCon.Make(&tc)

	smsResp, err := tc.SendSMS(
		config.GetAppConf().TwilioFrom,
		body.Mobile,
		fmt.Sprintf("[Darkpanda] Here is your change mobile verify code: \n\n %s", verifyCode.BuildCode()),
	)

	if twilio.HandleSendTwilioError(c, err) != nil {
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
			"user_uuid": c.GetString("uuid"),
			"mobile":    body.Mobile,
		}).
		Infof("sends twilio SMS success, login verify code created ! %v", smsResp.SID)

	c.JSON(
		http.StatusOK,
		struct {
			VerifyPrefix string `json:"verify_prefix"`
			Mobile       string `json:"mobile"`
		}{
			verifyCode.Chars,
			body.Mobile,
		},
	)
}

type VerifyMobileVerifyCodeParams struct {
	VerifyCode string `json:"verify_code" form:"verify_code" binding:"required,gt=0"`
}

func VerifyMobileVerifyCodeHandler(c *gin.Context, depCon container.Container) {
	body := VerifyMobileVerifyCodeParams{}

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

	// Get change mobile verify record from redis, if it exists.
	ctx := context.Background()
	m, err := GetChangeMobileVerifyCode(ctx, GetChangeMobileVerifyCodeParams{
		RedisCli: db.GetRedis(),
		UserUuid: c.GetString("uuid"),
	})

	if err == redis.Nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(apperr.ChangeMobileVerifyCodeNotExists),
		)

		return
	}

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToGetChangeMobileVerifyCode,
				err.Error(),
			),
		)

		return
	}

	if body.VerifyCode != m.VerifyCode {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(apperr.ChangeMobileVerifyCodeNotMatching),
		)

		return
	}

	var userDao contracts.UserDAOer
	depCon.Make(&userDao)

	pv := true
	if _, err := userDao.UpdateUserInfoByUuid(contracts.UpdateUserInfoParams{
		Uuid:          c.GetString("uuid"),
		Mobile:        &m.Mobile,
		PhoneVerified: &pv,
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

	c.JSON(http.StatusOK, struct {
		Mobile string `json:"mobile"`
	}{
		m.Mobile,
	})
}
