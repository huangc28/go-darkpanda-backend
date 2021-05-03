package rate

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golobby/container/pkg/container"
	"github.com/huangc28/go-darkpanda-backend/db"
	"github.com/huangc28/go-darkpanda-backend/internal/app/apperr"
	log "github.com/sirupsen/logrus"
)

type GetUserRatingBody struct {
	UUID string `form:"uuid" json:"uuid" binding:"required,gt=0"`
}

func GetUserRating(c *gin.Context, depCon container.Container) {
	var (
		uuid string = c.Param("uuid")
	)

	q := NewRateDAO(db.GetDB())
	bank, err := q.GetUserRating(uuid)

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

	tResp := NewTransform().TransformRate(bank)

	if err != nil {
		log.Fatal(err)
	}

	c.JSON(http.StatusOK, tResp)
}
