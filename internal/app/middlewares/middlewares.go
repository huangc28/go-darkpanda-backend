package middlewares

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/huangc28/go-darkpanda-backend/internal/app/apperr"
	"github.com/huangc28/go-darkpanda-backend/internal/app/contracts"
)

type UserDaoer interface {
	CheckIsMaleByUuid(uuid string) (bool, error)
	CheckIsFemaleByUuid(uuid string) (bool, error)
}

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func ResponseLogger(c *gin.Context) {
	blw := &bodyLogWriter{
		body:           bytes.NewBufferString(""),
		ResponseWriter: c.Writer,
	}
	c.Writer = blw
	c.Next()

	dst := &bytes.Buffer{}
	json.Indent(dst, blw.body.Bytes(), "", "  ")

	fmt.Println("Response body: " + string(dst.String()))
}

type InquiryUriParams struct {
	InquiryUuid string `uri:"inquiry_uuid" binding:"required"`
}

func ValidateInqiuryURIParams() gin.HandlerFunc {
	return func(c *gin.Context) {
		uriParams := &InquiryUriParams{}

		if err := c.ShouldBindUri(uriParams); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				apperr.NewErr(
					apperr.FailedToValidateCancelInquiryParams,
					err.Error(),
				),
			)

			return
		}

		c.Set("uri_params", uriParams)
		c.Next()
	}
}

func IsMale(userDao contracts.UserDAOer) gin.HandlerFunc {
	return func(c *gin.Context) {
		uuid := c.GetString("uuid")

		isMale, err := userDao.CheckIsMaleByUuid(uuid)

		if err != nil {
			c.AbortWithError(
				http.StatusInternalServerError,
				apperr.NewErr(
					apperr.FailedToCheckGender,
					err.Error(),
				),
			)

			return
		}

		if !isMale {
			c.AbortWithError(
				http.StatusBadRequest,
				apperr.NewErr(apperr.OnlyMaleCanBookService),
			)

			return
		}

		c.Next()
	}
}

func IsFemale(userDao contracts.UserDAOer) gin.HandlerFunc {
	return func(c *gin.Context) {
		uuid := c.GetString("uuid")

		isFemale, err := userDao.CheckIsFemaleByUuid(uuid)

		if err != nil {
			c.AbortWithError(
				http.StatusInternalServerError,
				apperr.NewErr(
					apperr.FailedToCheckGender,
					err.Error(),
				),
			)

			return
		}

		if !isFemale {
			c.AbortWithError(
				http.StatusBadRequest,
				apperr.NewErr(apperr.OnlyFemaleUserCanAccessAPI),
			)

			return
		}

		c.Next()
	}
}
