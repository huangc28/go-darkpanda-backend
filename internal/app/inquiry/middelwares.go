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

func IsMale(dao UserDaoer) gin.HandlerFunc {
	return func(c *gin.Context) {
		uuid := c.GetString("uuid")

		isMale, err := dao.CheckIsMaleByUuid(uuid)

		if !isMale {
			c.AbortWithError(
				http.StatusBadRequest,
				apperr.NewErr(
					apperr.OnlyMaleCanEmitInquiry,
					err.Error(),
				),
			)

			return
		}

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

		c.Next()
	}
}

func IsFemale(dao UserDaoer) gin.HandlerFunc {
	return func(c *gin.Context) {}
}

func ValidateBeforeAlterInquiryStatus(action InquiryActions) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.Background()
		usrUuid := c.GetString("uuid")
		uriParams := &CancelInquiryUriParam{}

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

		q := models.New(db.GetDB())

		// ------------------- makesure the user owns the inquiry -------------------
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
	}
}
