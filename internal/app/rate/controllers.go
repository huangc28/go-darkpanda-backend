package rate

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golobby/container/pkg/container"
	"github.com/huangc28/go-darkpanda-backend/db"
	"github.com/huangc28/go-darkpanda-backend/internal/app/apperr"
	"github.com/huangc28/go-darkpanda-backend/internal/app/contracts"
)

func GetServiceRating(c *gin.Context, depCon container.Container) {
	var (
		srvUuid  string = c.Param("service_uuid")
		userUuid string = c.Param("uuid")
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
			PartnerId:   int(user.ID),
			ServiceUuid: srvUuid,
		},
	)

	if err != nil {
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

func CreateServiceRating(c *gin.Context) {

}
