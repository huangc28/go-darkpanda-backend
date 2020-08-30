package user

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/huangc28/go-darkpanda-backend/internal/app/apperr"
)

type UserUriParams struct {
	UserUuid string `uri:"uuid" binding:"uuid"`
}

func ValidateUserURIParams() gin.HandlerFunc {
	return func(c *gin.Context) {
		uriParams := &UserUriParams{}

		if err := c.ShouldBindUri(uriParams); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				apperr.NewErr(
					apperr.FailedToValidateUserURIParams,
					err.Error(),
				),
			)

			return
		}

		c.Set("uri_params", uriParams)
		c.Next()
	}
}
