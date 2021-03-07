package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golobby/container/pkg/container"
	"github.com/huangc28/go-darkpanda-backend/db"
	"github.com/huangc28/go-darkpanda-backend/internal/app/apperr"
	"github.com/huangc28/go-darkpanda-backend/internal/app/contracts"
	"github.com/huangc28/go-darkpanda-backend/internal/app/models"
	"github.com/huangc28/go-darkpanda-backend/internal/app/pkg/requestbinder"
)

// GetListOfCurrentServicesHandler retrieve those service of the following status:
//   - unpaid
//   - to_be_fulfilled
type GetListOfCurrentServicesBody struct {
	Offset  int `form:"offset,default=0"`
	PerPage int `form:"perpage,default=5"`
}

func GetListOfCurrentServicesHandler(c *gin.Context, depCon container.Container) {
	body := GetListOfCurrentServicesBody{}

	if err := requestbinder.Bind(c, &body); err != nil {
		c.AbortWithError(
			http.StatusBadRequest,
			apperr.NewErr(
				apperr.FailedToBindInquiryUriParams,
				err.Error(),
			),
		)

		return
	}

	// Retrieve picker's uuid
	userUuid := c.GetString("uuid")

	var userDao contracts.UserDAOer
	depCon.Make(&userDao)

	user, err := userDao.GetUserByUuid(
		userUuid,
		"id",
		"gender",
	)

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

	// Retrieve list of incoming services
	var srvs []ServiceResult = make([]ServiceResult, 0)

	srvDao := NewServiceDAO(db.GetDB())

	if user.Gender == models.GenderFemale {
		srvs, err = srvDao.GetIncomingServicesByProviderId(
			GetServicesParams{
				UserID:  int(user.ID),
				Offset:  body.Offset,
				PerPage: body.PerPage,
			},
		)
	}

	if user.Gender == models.GenderMale {
		srvs, err = srvDao.GetOverduedServicesByCustomerId(
			GetServicesParams{
				UserID:  int(user.ID),
				Offset:  body.Offset,
				PerPage: body.PerPage,
			},
		)

	}

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToGetIncomingService,
				err.Error(),
			),
		)

		return
	}

	// Retrieve service provider uuid
	c.JSON(
		http.StatusOK,
		TransformGetServicesResults(srvs),
	)
}

type GetOverduedServicesBody struct {
	Offset  int `form:"offset,default=0"`
	PerPage int `form:"perpage,default=5"`
}

// GetOverduedServicesHandlers retrieve those service of the following status:
//  - canceled
//  - failed_due_to_both
//  - failed_due_to_girl
//  - failed_due_to_man
//  - completed
func GetOverduedServicesHandlers(c *gin.Context, depCon container.Container) {
	body := GetOverduedServicesBody{}

	if err := requestbinder.Bind(c, &body); err != nil {
		c.AbortWithError(
			http.StatusBadRequest,
			apperr.NewErr(
				apperr.FailedToBindApiBodyParams,
				err.Error(),
			),
		)

		return
	}

	// Retrieve picker's uuid
	var userDao contracts.UserDAOer
	depCon.Make(&userDao)

	user, err := userDao.GetUserByUuid(
		c.GetString("uuid"),
		"id",
		"gender",
	)

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

	var srvRes []ServiceResult = make([]ServiceResult, 0)

	// Retrieve list of overdued services
	srvDao := NewServiceDAO(db.GetDB())

	if user.Gender == models.GenderFemale {
		srvRes, err = srvDao.GetOverduedServicesByProviderId(
			GetServicesParams{
				UserID:  int(user.ID),
				Offset:  body.Offset,
				PerPage: body.PerPage,
			},
		)
	}

	if user.Gender == models.GenderMale {
		srvRes, err = srvDao.GetOverduedServicesByCustomerId(
			GetServicesParams{
				UserID:  int(user.ID),
				Offset:  body.Offset,
				PerPage: body.PerPage,
			},
		)
	}

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToGetOverdueService,
				err.Error(),
			),
		)

		return
	}

	c.JSON(
		http.StatusOK,
		TransformGetServicesResults(srvRes),
	)
}
