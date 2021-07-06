package apperr

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// @TODO figure out why the following statement is able to convert gin.Error to apperr.Error
func HandleError() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		err := c.Errors.Last()

		if err != nil {
			srcErr := err.Err
			var parsedErr *Error

			switch srcErr := srcErr.(type) {
			case *Error:
				parsedErr = srcErr
			default:
				parsedErr = &Error{
					ErrCode: UnknownErrorToApplication,
					ErrMsg:  srcErr.Error(),
				}

				c.Writer.WriteHeader(http.StatusInternalServerError)
			}

			c.JSON(c.Writer.Status(), parsedErr)

			return
		}
	}
}
