package rate

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golobby/container/pkg/container"
	"github.com/huangc28/go-darkpanda-backend/db"
	"github.com/huangc28/go-darkpanda-backend/internal/app/apperr"
	"github.com/huangc28/go-darkpanda-backend/internal/app/contracts"
	"github.com/huangc28/go-darkpanda-backend/internal/app/pkg/requestbinder"
)

func GetServiceRating(c *gin.Context, depCon container.Container) {
	var (
		srvUuid  string = c.Param("service_uuid")
		userUuid string = c.GetString("uuid")
		userDao  contracts.UserDAOer
	)

	depCon.Make(&userDao)
	user, err := userDao.GetUserByUuid(userUuid)

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

	rateDao := NewRateDAO(db.GetDB())
	partnerInfo, err := rateDao.GetServicePartnerInfo(
		GetServicePartnerInfoParams{
			Gender:      user.Gender,
			MyId:        int(user.ID),
			ServiceUuid: srvUuid,
		},
	)

	if err != nil {
		if err == sql.ErrNoRows {
			c.AbortWithError(
				http.StatusBadRequest,
				apperr.NewErr(apperr.NotInvolveInService),
			)

			return
		}

		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToGetServicePartnerInfo,
				err.Error(),
			),
		)

		return
	}

	// Get service rating made by the chat partner.
	srvRating, err := rateDao.GetServiceRating(
		GetServiceRatingParams{
			ServiceUuid: srvUuid,
			RaterId:     int(partnerInfo.ID),
		},
	)

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToGetServiceRating,
				err.Error(),
			),
		)

		return
	}

	tResp := NewTransform().TransformRate(partnerInfo, srvRating)

	c.JSON(http.StatusOK, tResp)
}

type CreateServiceRatingparams struct {
	ServiceUuid string `json:"service_uuid" form:"service_uuid" binding:"required,gt=0"`
	Rating      int    `json:"rating" form:"rating" binding:"required,gt=0"`
	Comment     string `json:"comment" form:"comment"`
}

func CreateServiceRating(c *gin.Context, depCon container.Container) {
	var body CreateServiceRatingparams

	if err := requestbinder.Bind(c, &body); err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToBindBodyParams,
				err.Error(),
			),
		)

		return
	}

	var userDao contracts.UserDAOer
	depCon.Make(&userDao)
	usr, err := userDao.GetUserByUuid(c.GetString("uuid"), "id")

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

	// Check the request is the participant of the service.
	rateDao := NewRateDAO(db.GetDB())

	if err := rateDao.IsServiceRatable(IsServiceRatableParams{
		ParticipantId: int(usr.ID),
		ServiceUuid:   body.ServiceUuid,
	}); err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.ServiceNotRatable,
				err.Error(),
			),
		)

		return
	}

	// Create rating record.
	if err := rateDao.CreateServiceRating(
		CreateServiceRatingParams{
			Rating:      body.Rating,
			RaterId:     int(usr.ID),
			ServiceUuid: body.ServiceUuid,
			Comment:     body.Comment,
		},
	); err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToCreateServiceRating,
				err.Error(),
			),
		)

		return
	}

	c.JSON(http.StatusOK, struct{}{})
}
