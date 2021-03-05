package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golobby/container/pkg/container"
	"github.com/huangc28/go-darkpanda-backend/db"
	"github.com/huangc28/go-darkpanda-backend/internal/app/apperr"
	"github.com/huangc28/go-darkpanda-backend/internal/app/contracts"
)

func GetListOfCurrentServicesHandler(c *gin.Context, depCon container.Container) {
	// Retrieve picker's uuid
	pickerUuid := c.GetString("uuid")

	var userDao contracts.UserDAOer
	depCon.Make(&userDao)

	picker, _ := userDao.GetUserByUuid(pickerUuid, "id")

	// Retrieve list of services
	srvDao := NewServiceDAO(db.GetDB())
	srvs, err := srvDao.GetIncomingServicesByProviderId(int(picker.ID))

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
		TransformGetIncomingServices(srvs),
	)
}
