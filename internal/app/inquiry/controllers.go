package inquiry

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/huangc28/go-darkpanda-backend/db"
	"github.com/huangc28/go-darkpanda-backend/internal/app/apperr"
	"github.com/huangc28/go-darkpanda-backend/internal/app/inquiry/util"
	"github.com/huangc28/go-darkpanda-backend/internal/models"
	"github.com/looplab/fsm"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"github.com/teris-io/shortid"
)

// @TODO budget received from client should be type float instead of string.
//       budget should be converted to type string before stored in DB.
type EmitInquiryBody struct {
	Budget      string `json:"budget" binding:"required"`
	ServiceType string `json:"service_type" binding:"required"`
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

func CancelInquiryHandler(c *gin.Context) {
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

	uriParams := eup.(*InquiryUriParams)
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

func ExpireInquiryHandler(c *gin.Context) {
	eup, uriParamExists := c.Get("uri_params")
	efsm, nFsmExists := c.Get("next_fsm_state")

	if !uriParamExists || !nFsmExists {
		c.AbortWithError(
			http.StatusBadRequest,
			apperr.NewErr(apperr.ParamsNotProperlySetInTheMiddleware),
		)

		return
	}

	uriParams := eup.(*InquiryUriParams)
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

func PickupInquiryHandler(c *gin.Context) {
	eup, uriParamExists := c.Get("uri_params")
	eiq, inquiryExists := c.Get("inquiry")
	efsm, nFsmExists := c.Get("next_fsm_state")

	if !uriParamExists || !nFsmExists || !inquiryExists {
		c.AbortWithError(
			http.StatusBadRequest,
			apperr.NewErr(apperr.ParamsNotProperlySetInTheMiddleware),
		)

		return
	}

	uriParams := eup.(*InquiryUriParams)
	fsm := efsm.(*fsm.FSM)
	iq := eiq.(models.ServiceInquiry)

	// ------------------- check if inquiry has already expired -------------------
	if util.IsExpired(iq.CreatedAt) {
		c.AbortWithError(
			http.StatusBadRequest,
			apperr.NewErr(apperr.CanNotPickupExpiredInquiry),
		)

		return
	}

	// ------------------- patch inquiry status -------------------
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

	// retrieve inquirier information
	inquirer, err := q.GetUserByID(ctx, int64(uiq.InquirerID.Int32))

	if err != nil {
		log.WithFields(log.Fields{
			"inquirer_id": inquirer.ID,
		}).Debugf("Failed to retrieve inquirer information %s", err.Error())

		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToGetInquiererByID,
				err.Error(),
			),
		)

		return
	}

	c.JSON(http.StatusOK, NewTransform().TransformPickupInquiry(uiq, inquirer))
}

// Girl has approved the inquiry, thus, update the inquiry content.
//   - price
//   - duration
//   - appointment time
//   - lng
//   - lat
type GirlApproveInquiryBody struct {
	Price           float64   `json:"price"`
	Duration        int       `json:"duration"`
	AppointmentTime time.Time `json:"appointment_time"`
	Lat             float64   `json:"lat"`
	Lng             float64   `json:"lng"`
}

func GirlApproveInquiryHandler(c *gin.Context) {
	ctx := context.Background()
	body := GirlApproveInquiryBody{}
	eup, _ := c.Get("uri_params")
	uriParams := eup.(*InquiryUriParams)

	if err := c.ShouldBindJSON(&body); err != nil {
		c.AbortWithError(
			http.StatusBadRequest,
			apperr.NewErr(
				apperr.GirlApproveInquiry,
				err.Error(),
			),
		)

		return
	}

	// ------------------- Updates inquiry content -------------------
	q := models.New(db.GetDB())
	efsm, _ := c.Get("next_fsm_state")
	fsm := efsm.(*fsm.FSM)

	latDec := decimal.NewFromFloat(body.Lng)
	lngDec := decimal.NewFromFloat(body.Lat)

	iq, err := q.UpdateInquiryByUuid(ctx, models.UpdateInquiryByUuidParams{
		Price: sql.NullString{
			String: fmt.Sprintf("%f", body.Price),
			Valid:  true,
		},

		Duration: sql.NullInt32{
			Int32: int32(body.Duration),
			Valid: true,
		},

		AppointmentTime: sql.NullTime{
			Time:  body.AppointmentTime,
			Valid: true,
		},

		Lng: sql.NullString{
			String: latDec.String(),
			Valid:  true,
		},

		Lat: sql.NullString{
			String: lngDec.String(),
			Valid:  true,
		},

		Uuid: uriParams.InquiryUuid,

		InquiryStatus: models.InquiryStatus(fsm.Current()),
	})

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToUpdateInquiryContent,
				err.Error(),
			),
		)

		return
	}

	// ------------------- Emit message to chatroom -------------------
	res, err := NewTransform().TransformGirlApproveInquiry(iq)

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToTransformGirlApproveInquiry,
				err.Error(),
			),
		)

		return
	}

	c.JSON(http.StatusOK, res)
}

// Emit event to girl for the purpose of notifying them the iquiry is booked by the man
type ManApproveInquiryBody struct {
	Price           float64   `json:"price"`
	Duration        int       `json:"duration"`
	AppointmentTime time.Time `json:appointment_time`
	Lng             float64   `json:"lng"`
	Lat             float64   `json:"lat"`
	ServiceType     string    `json:"service_type"`
}

func ManApproveInquiry(c *gin.Context) {
	eup, exists := c.Get("uri_params")

	if !exists {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(apperr.ParamsNotProperlySetInTheMiddleware),
		)

		return
	}

	body := ManApproveInquiryBody{}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToValidateBookInquiryParams,
				err.Error(),
			),
		)

		return
	}
	// Alter inquiry status to "booked"
	// Create a new service with "pending"
	ctx := context.Background()
	uriParams := eup.(*InquiryUriParams)

	tx, _ := db.GetDB().Begin()
	q := models.New(tx)

	iq, err := q.PatchInquiryStatusByUuid(ctx, models.PatchInquiryStatusByUuidParams{
		Uuid:          uriParams.InquiryUuid,
		InquiryStatus: models.InquiryStatusBooked,
	})

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(apperr.FailedToPatchInquiryStatus),
		)

		return
	}

	log.Printf("DEBUG 1 %v", iq)

	c.JSON(http.StatusOK, struct{}{})
}
