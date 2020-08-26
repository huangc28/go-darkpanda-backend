package inquiry

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/huangc28/go-darkpanda-backend/db"
	"github.com/huangc28/go-darkpanda-backend/internal/app/apperr"
	"github.com/huangc28/go-darkpanda-backend/internal/models"
)

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

func IsMale(dao UserDaoer) gin.HandlerFunc {
	return func(c *gin.Context) {
		uuid := c.GetString("uuid")

		isMale, err := dao.CheckIsMaleByUuid(uuid)

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

func IsFemale(dao UserDaoer) gin.HandlerFunc {
	return func(c *gin.Context) {
		uuid := c.GetString("uuid")

		isFemale, err := dao.CheckIsFemaleByUuid(uuid)

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
				apperr.NewErr(apperr.OnlyFemaleCanApproveInquiry),
			)

			return
		}

		c.Next()
	}
}

func ValidateBeforeAlterInquiryStatus(action InquiryActions) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.Background()
		usrUuid := c.GetString("uuid")

		eup, _ := c.Get("uri_params")
		uriParams := eup.(*InquiryUriParams)

		// ------------------- makesure the user owns the inquiry -------------------
		q := models.New(db.GetDB())
		err := q.CheckUserOwnsInquiry(ctx, models.CheckUserOwnsInquiryParams{

			Uuid:   usrUuid,
			Uuid_2: uriParams.InquiryUuid,
		})

		if err == sql.ErrNoRows {
			c.AbortWithError(
				http.StatusBadRequest,
				apperr.NewErr(apperr.UserNotOwnInquiry),
			)

			return
		}

		// ------------------- try to emit transition event  -------------------
		iq, err := q.GetInquiryByUuid(ctx, uriParams.InquiryUuid)
		fsm, _ := NewInquiryFSM(iq.InquiryStatus)

		if err := fsm.Event(action.ToString()); err != nil {
			c.AbortWithError(
				http.StatusBadRequest,
				apperr.NewErr(
					apperr.InquiryFSMTransitionFailed,
					err.Error(),
				),
			)

			return
		}

		c.Set("next_fsm_state", fsm)
		c.Set("inquiry", iq)

		c.Next()
	}
}
