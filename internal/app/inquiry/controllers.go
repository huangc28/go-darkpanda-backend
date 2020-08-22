package inquiry

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/huangc28/go-darkpanda-backend/db"
	"github.com/huangc28/go-darkpanda-backend/internal/app/apperr"
	"github.com/huangc28/go-darkpanda-backend/internal/app/inquiry/util"
	"github.com/huangc28/go-darkpanda-backend/internal/models"
	"github.com/looplab/fsm"
	log "github.com/sirupsen/logrus"
	"github.com/teris-io/shortid"
)

type EmitInquiryBody struct {
	Budget      float64 `json:"budget" binding:"required"`
	ServiceType string  `json:"service_type" binding:"required"`
}

func EmitInquiryHandler(c *gin.Context) {
	// retrieve user
	// check if user is male
	// check if user has active inquiry already
	body := &EmitInquiryBody{}
	ctx := context.Background()

	if err := c.ShouldBindJSON(&body); err != nil {
		c.AbortWithError(
			http.StatusBadRequest,
			apperr.NewErr(
				apperr.FailedToValidateEmitInquiryParams,
				err.Error(),
			),
		)

		return
	}

	uuid := c.GetString("uuid")
	q := models.New(db.GetDB())
	usr, err := q.GetUserByUuid(ctx, uuid)

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

	// ------------------- only male user can emit inquiry -------------------
	if usr.Gender != models.GenderMale {
		c.AbortWithError(
			http.StatusBadRequest,
			apperr.NewErr(apperr.OnlyMaleCanEmitInquiry),
		)

		return
	}

	// ------------------- check if the user already has active inquiry -------------------
	resIq, err := q.GetInquiryByInquirerID(ctx, models.GetInquiryByInquirerIDParams{
		InquirerID: sql.NullInt32{
			Int32: int32(usr.ID),
			Valid: true,
		},
		InquiryStatus: models.InquiryStatusInquiring,
	})

	if err != nil && err != sql.ErrNoRows {
		log.WithFields(log.Fields{
			"uuid":  usr.Uuid,
			"error": err.Error(),
		}).Debug("User has active inquiry.")

		c.AbortWithError(
			http.StatusBadRequest,
			apperr.NewErr(apperr.UserAlreadyHasActiveInquiry),
		)

		return
	}

	// we have to makesure the retrieved inquiry it's still within 27 mins.
	// if it is not, change the inquiry status to expired.
	if util.IsExpired(resIq.CreatedAt) {
		q.PatchInquiryStatus(ctx, models.PatchInquiryStatusParams{
			ID:            resIq.ID,
			InquiryStatus: models.InquiryStatusExpired,
		})
	}

	// ------------------- create a new inquiry -------------------
	sid, _ := shortid.Generate()
	iq, err := q.CreateInquiry(ctx, models.CreateInquiryParams{
		Uuid: sid,
		InquirerID: sql.NullInt32{
			Int32: int32(usr.ID),
			Valid: true,
		},
		Budget:        body.Budget,
		ServiceType:   models.ServiceType(body.ServiceType),
		InquiryStatus: models.InquiryStatusInquiring,
	})

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToCreateInquiry,
				err.Error(),
			),
		)

		return
	}

	c.JSON(http.StatusOK, NewTransform().TransformInquiry(iq))
}

type CancelInquiryUriParam struct {
	InquiryUuid string `uri:"inquiry_uuid" binding:"required"`
}

func CancelInquiry(c *gin.Context) {
	// ------------------- gather information from middleware -------------------
	eup, uriParamExists := c.Get("uri_params")
	efsm, nFsmExists := c.Get("next_fsm_state")

	if !uriParamExists || !nFsmExists {
		c.AbortWithError(
			http.StatusBadRequest,
			apperr.NewErr(apperr.ParamsNotProperlySetInTheMiddleware),
		)

		return
	}

	uriParams := eup.(*CancelInquiryUriParam)
	fsm := efsm.(*fsm.FSM)

	// ------------------- Update inquiry status to cancel  -------------------
	ctx := context.Background()
	q := models.New(db.GetDB())

	uiq, err := q.PatchInquiryStatusByUuid(ctx, models.PatchInquiryStatusByUuidParams{
		InquiryStatus: models.InquiryStatus(fsm.Current()),
		Uuid:          uriParams.InquiryUuid,
	})

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(apperr.FailedToPatchInquiryStatus),
		)

		return
	}

	c.JSON(http.StatusOK, NewTransform().TransformInquiry(uiq))
}

func ExpireInquiry(c *gin.Context) {
	eup, uriParamExists := c.Get("uri_params")
	efsm, nFsmExists := c.Get("next_fsm_state")

	if !uriParamExists || !nFsmExists {
		c.AbortWithError(
			http.StatusBadRequest,
			apperr.NewErr(apperr.ParamsNotProperlySetInTheMiddleware),
		)

		return
	}

	uriParams := eup.(*CancelInquiryUriParam)
	fsm := efsm.(*fsm.FSM)

	// ------------------- Update inquiry status to expire  -------------------
	ctx := context.Background()
	q := models.New(db.GetDB())

	uiq, err := q.PatchInquiryStatusByUuid(ctx, models.PatchInquiryStatusByUuidParams{
		InquiryStatus: models.InquiryStatus(fsm.Current()),
		Uuid:          uriParams.InquiryUuid,
	})

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(apperr.FailedToPatchInquiryStatus),
		)

		return
	}

	c.JSON(http.StatusOK, NewTransform().TransformInquiry(uiq))
}
