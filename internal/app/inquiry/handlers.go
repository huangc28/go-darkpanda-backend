package inquiry

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/golobby/container/pkg/container"
	"github.com/huangc28/go-darkpanda-backend/db"
	"github.com/huangc28/go-darkpanda-backend/internal/app/apperr"
	"github.com/huangc28/go-darkpanda-backend/internal/app/contracts"
	"github.com/huangc28/go-darkpanda-backend/internal/app/inquiry/util"
	"github.com/huangc28/go-darkpanda-backend/internal/app/models"
	darkfirestore "github.com/huangc28/go-darkpanda-backend/internal/app/pkg/dark_firestore"
	"github.com/huangc28/go-darkpanda-backend/internal/app/pkg/requestbinder"
	"github.com/jmoiron/sqlx"
	"github.com/looplab/fsm"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"github.com/teris-io/shortid"
)

// @TODO
//   - budget received from client should be type float instead of string.
//   - budget should be converted to type string before stored in DB.
//   - Body should include "appointment time"
//   - Body should include "Service duration"
type EmitInquiryBody struct {
	Budget          float64   `form:"budget" uri:"budget" json:"budget" binding:"required"`
	ServiceType     string    `form:"service_type" uri:"service_type" json:"service_type" binding:"required"`
	AppointmentTime time.Time `form:"appointment_time" binding:"required"`
	ServiceDuration int       `form:"service_duration" binding:"required"`
}

func EmitInquiryHandler(c *gin.Context) {
	body := &EmitInquiryBody{}
	ctx := context.Background()

	if err := requestbinder.Bind(c, body); err != nil {
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

	// Check if the user already has an active inquiry
	// If active inquiry exists but expired, change the
	// inquiry status to `expired`. If exists but has not
	// expired, respond with error.
	dao := NewInquiryDAO(db.GetDB())
	activeIqExists, err := dao.CheckHasActiveInquiryByID(usr.ID)

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToCheckActiveInquiry,
				err.Error(),
			),
		)

		return
	}

	if activeIqExists {
		resIq, err := q.GetInquiryByInquirerID(ctx, models.GetInquiryByInquirerIDParams{
			InquirerID: sql.NullInt32{
				Int32: int32(usr.ID),
				Valid: true,
			},
			InquiryStatus: models.InquiryStatusInquiring,
		})

		if err != nil {
			c.AbortWithError(
				http.StatusInternalServerError,
				apperr.NewErr(apperr.FailedToGetInquiryByInquirerID),
			)

			return
		}

		if !util.IsExpired(resIq.CreatedAt) {
			c.AbortWithError(
				http.StatusBadRequest,
				apperr.NewErr(apperr.UserAlreadyHasActiveInquiry),
			)

			return
		}

		// @TODO also makesure records in the firestore is marked expired.
		if err := q.PatchInquiryStatus(ctx, models.PatchInquiryStatusParams{
			ID:            resIq.ID,
			InquiryStatus: models.InquiryStatusExpired,
		}); err != nil {
			c.AbortWithError(
				http.StatusInternalServerError,
				apperr.NewErr(
					apperr.FailedToPatchInquiryStatus,
					err.Error(),
				),
			)

			return
		}
	}

	// ------------------- create a new inquiry -------------------
	sid, _ := shortid.Generate()
	iq, err := q.CreateInquiry(ctx, models.CreateInquiryParams{
		Uuid: sid,
		InquirerID: sql.NullInt32{
			Int32: int32(usr.ID),
			Valid: true,
		},
		Budget:        decimal.NewFromFloat(body.Budget).String(),
		ServiceType:   models.ServiceType(body.ServiceType),
		InquiryStatus: models.InquiryStatusInquiring,
		ExpiredAt: sql.NullTime{
			Time:  time.Now().Add(InquiryDuration),
			Valid: true,
		},
	})

	if err != nil {
		log.WithFields(log.Fields{
			"uuid":  usr.Uuid,
			"error": err.Error(),
		}).Debug("User has active inquiry.")

		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToCreateInquiry,
				err.Error(),
			),
		)

		return
	}

	df := darkfirestore.Get()
	df.CreateInquiringUser(ctx, darkfirestore.CreateInquiringUserParams{
		InquiryUUID:   iq.Uuid,
		InquiryStatus: iq.InquiryStatus.ToString(),
	})

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToJoinLobby,
				err.Error(),
			),
		)

		return
	}

	trf, err := NewTransform().TransformEmitInquiry(iq)

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToTransformResponse,
				err.Error(),
			),
		)
		return
	}

	c.JSON(http.StatusOK, trf)
}

// Fetch nearby(?) inquiries information. Only female user can fetch inquiries info.
// Each inquiry should also contains inquier's base information.
// @TODO
//   - Offset should be passed from client via query param.
//   - If no record exists, `has_more` indicator should set to false. Client request should be based on this indicator
//   - wrap `GetInquiries` and `HasMore` should be wrapped in a transaction.
type GetInquiriesBody struct {
	Offset  int `form:"offset,default=0"`
	PerPage int `form:"perpage,default=7"`
}

func GetInquiriesHandler(c *gin.Context) {
	body := &GetInquiriesBody{}

	if err := requestbinder.Bind(c, &body); err != nil {
		c.AbortWithError(
			http.StatusBadRequest,
			apperr.NewErr(
				apperr.FailedToValidateGetInquiryListParams,
				err.Error(),
			),
		)

		return
	}

	inquiryDao := NewInquiryDAO(db.GetDB())

	// offset should be passed from client
	inquiries, err := inquiryDao.GetInquiries(
		models.InquiryStatusInquiring,
		body.Offset,
		body.PerPage,
	)

	if err != nil {
		c.AbortWithError(
			http.StatusOK,
			apperr.NewErr(
				apperr.FailedToGetInquiryList,
				err.Error(),
			),
		)

		return
	}

	// DB has no more records if number of retrieved records is less then the value of `perPage`.
	// In which case, we should set `has_more` indicator to `false`
	hasMoreRecord, err := inquiryDao.HasMoreInquiries(
		body.Offset,
		body.PerPage,
	)

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToCheckHasMoreInquiry,
				err.Error(),
			),
		)

		return
	}

	tres, err := NewTransform().TransformInquiryList(
		inquiries,
		hasMoreRecord,
	)

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToTransformGetInquiriesResponse,
				err.Error(),
			),
		)

		return
	}

	c.JSON(http.StatusOK, tres)
}

func GetInquiryHandler(c *gin.Context) {
	var inquiryUuid string = c.Param("uuid")

	// Retrieve inquiry along with inquirer by inquiry uuid
	dao := NewInquiryDAO(db.GetDB())
	iq, err := dao.GetInquiryByUuid(inquiryUuid)

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToGetInquiryByUuid,
				err.Error(),
			),
		)

		return
	}

	trfIq, err := NewTransform().TransformGetInquiry(*iq)

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToTransformGetInquiry,
				err.Error(),
			),
		)

		return
	}

	c.JSON(http.StatusOK, trfIq)
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

	trf, err := NewTransform().TransformInquiry(uiq)

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToTransformResponse,
				err.Error(),
			),
		)
		return

	}

	c.JSON(http.StatusOK, trf)
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

	trf, err := NewTransform().TransformInquiry(uiq)

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToTransformResponse,
				err.Error(),
			),
		)
		return

	}

	c.JSON(http.StatusOK, trf)
}

type PickupInquiryHandlerParams struct {
	InquiryUuid string `uri:"inquiry_uuid" binding:"required"`
}

func PickupInquiryHandler(c *gin.Context, depCon container.Container) {
	var params PickupInquiryHandlerParams

	if err := c.ShouldBindUri(&params); err != nil {
		c.AbortWithError(
			http.StatusBadRequest,
			apperr.NewErr(
				apperr.FailedToValidateCancelInquiryParams,
				err.Error(),
			),
		)

		return
	}

	ctx := context.Background()
	q := models.New(db.GetDB())

	// Retrieve inquiry information
	iq, err := q.GetInquiryByUuid(ctx, params.InquiryUuid)

	if err != nil {
		c.AbortWithError(
			http.StatusBadRequest,
			apperr.NewErr(
				apperr.FailedToGetInquiryByUuid,
				err.Error(),
			),
		)

		return

	}

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToGetUserByID,
				err.Error(),
			),
		)

		return
	}

	// ------------------- Check inquiry status is inquiring -------------------
	if iq.InquiryStatus != models.InquiryStatusInquiring {
		c.AbortWithError(
			http.StatusBadRequest,
			apperr.NewErr(
				apperr.FailedToPickupStatusNotInquiring,
				err.Error(),
			),
		)

		return
	}

	err, errCode := db.Transact(db.GetDB(), func(tx *sqlx.Tx) (error, interface{}) {
		fsm, _ := NewInquiryFSM(iq.InquiryStatus)

		if err := fsm.Event(Pickup.ToString()); err != nil {
			return err, apperr.InquiryFSMTransitionFailed
		}

		// Patch inquiry status in DB to be `asking`
		q := models.New(tx)
		_, err := q.UpdateInquiryByUuid(
			ctx,
			models.UpdateInquiryByUuidParams{
				PickerID: sql.NullInt32{
					Int32: iq.PickerID.Int32,
					Valid: true,
				},
				InquiryStatus: models.InquiryStatus(fsm.Current()),
				Uuid:          iq.Uuid,
			},
		)

		if err != nil {
			return err, apperr.FailedToUpdateInquiryContent
		}

		// Patch inquiry status in firestore to be `asking`
		df := darkfirestore.Get()

		err = df.AskingInquiringUser(
			ctx,
			darkfirestore.AskingInquiringUserParams{
				InquiryUUID: iq.Uuid,
			},
		)

		if err != nil {
			return err, apperr.FailedToAskInquiringUser
		}

		return nil, nil
	})

	if err != nil {
		c.AbortWithError(
			http.StatusBadRequest,
			apperr.NewErr(
				fmt.Sprintf("%v", errCode),
				err.Error(),
			),
		)

		return
	}

	// @TODO
	//   should return inquiry uuid for female user to subscribe to firestore document.
	c.JSON(
		http.StatusOK,
		NewTransform().TransformPickupInquiry(iq),
	)

}

type AgreePickupInquiryHandlerParams struct {
	InquiryUuid string `uri:"inquiry_uuid" binding:"required"`
}

// AgreePickupInquiryHandler Male user agree to have a chat with the male user.
// Perform following operations when male user agrees to chat.
//   - Check inquiry status can be transitioned to `chatting`
//   - Change inquiry status to `chatting` on DB
//   - Change inquiry status to `chatting` on firestore
func AgreeToChatInquiryHandler(c *gin.Context, depCon container.Container) {
	var params AgreePickupInquiryHandlerParams

	if err := c.ShouldBindUri(&params); err != nil {
		c.AbortWithError(
			http.StatusBadRequest,
			apperr.NewErr(
				apperr.FailedToBindInquiryUriParams,
				err.Error(),
			),
		)

		return
	}

	// Retrieve inquiry by inquiry uuid
	ctx := context.Background()

	q := models.New(db.GetDB())
	iq, err := q.GetInquiryByUuid(ctx, params.InquiryUuid)

	if err != nil {
		c.AbortWithError(
			http.StatusBadRequest,
			apperr.NewErr(
				apperr.FailedToGetInquiryByUuid,
				err.Error(),
			),
		)

		return
	}

	fsm, _ := NewInquiryFSM(iq.InquiryStatus)

	if err := fsm.Event(AgreePickup.ToString()); err != nil {
		c.AbortWithError(
			http.StatusBadRequest,
			apperr.NewErr(
				apperr.InquiryFSMTransitionFailed,
				err.Error(),
			),
		)

		return
	}

	// Update inquiry status in DB.
	inquiryDao := NewInquiryDAO(db.GetDB())
	if err := inquiryDao.PatchInquiryStatusByUUID(
		contracts.PatchInquiryStatusByUUIDParams{
			UUID:          iq.Uuid,
			InquiryStatus: models.InquiryStatusChatting,
		},
	); err != nil {
		c.AbortWithError(
			http.StatusBadRequest,
			apperr.NewErr(
				apperr.FailedToPatchInquiryStatus,
				err.Error(),
			),
		)

		return
	}

	// Update inquiry status in firestore.
	df := darkfirestore.Get()
	if err := df.ChatInquiringUser(ctx, darkfirestore.ChatInquiringUserParams{
		InquiryUUID: iq.Uuid,
	}); err != nil {
		c.AbortWithError(
			http.StatusBadRequest,
			apperr.NewErr(
				apperr.FailedToChangeFirestoreInquiryStatus,
				err.Error(),
			),
		)

		return
	}

	// Create private chatroom using chat service.
	var chat contracts.ChatServicer
	depCon.Make(&chat)

	chat.CreateAndJoinChatroom(
		iq.ID,
		int64(iq.InquirerID.Int32),
	)

	// Respoonse:
	//   - service provider's info
	//   - inquiry info
	//   - private chat uuid in firestore
	//trf := NewTransform()

	//c.JSON(http.StatusOK, trf.TransformPickupInquiry(iq))

	c.JSON(http.StatusOK, struct{}{})
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
	Price               float64   `json:"price"`
	Duration            int       `json:"duration"`
	AppointmentTime     time.Time `json:"appointment_time"`
	Lng                 float64   `json:"lng"`
	Lat                 float64   `json:"lat"`
	ServiceType         string    `json:"service_type"`
	ServiceProviderUuid string    `json:"service_provider_uuid"`
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

		tx.Rollback()
		return
	}

	srvProvider, err := q.GetUserByUuid(ctx, body.ServiceProviderUuid)

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToGetUserIDByUuid,
				err.Error(),
			),
		)

		tx.Rollback()
		return
	}

	srv, err := q.CreateService(
		ctx,
		models.CreateServiceParams{
			Price: sql.NullString{
				String: decimal.NewFromFloat(body.Price).String(),
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
				String: decimal.NewFromFloat(body.Lng).String(),

				Valid: true,
			},
			Lat: sql.NullString{
				String: decimal.NewFromFloat(body.Lat).String(),
				Valid:  true,
			},
			ServiceStatus: models.ServiceStatusUnpaid,
			ServiceType:   iq.ServiceType,
			InquiryID:     int32(iq.ID),
			CustomerID: sql.NullInt32{
				Int32: iq.InquirerID.Int32,
				Valid: true,
			},
			ServiceProviderID: sql.NullInt32{
				Int32: int32(srvProvider.ID),
				Valid: true,
			},
		},
	)

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToCreateService,
				err.Error(),
			),
		)

		tx.Rollback()
		return
	}

	tx.Commit()

	c.JSON(http.StatusOK, NewTransform().TransformBookedService(srv, srvProvider))
}

// @TODO
//   - Emit both chat participants that the chat is closed.
//   - All chat participants should be removed from the chat.
func RevertChatHandler(c *gin.Context, depCon container.Container) {
	eiiq, exists := c.Get("inquiry")

	if !exists {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(apperr.FSMNotSetInMiddleware),
		)
		return
	}

	iq := (eiiq).(*models.ServiceInquiry)

	var (
		chatDao contracts.ChatDaoer
		userDao contracts.UserDAOer
	)

	depCon.Make(&chatDao)
	depCon.Make(&userDao)

	// Find chatroom by inquiry_id, find inquiry_id by inquiry_uuid
	chatroom, err := chatDao.GetChatRoomByInquiryID(
		iq.ID,
		"id",
		"channel_uuid",
	)

	if err != nil {
		c.AbortWithError(
			http.StatusBadRequest,
			apperr.NewErr(
				apperr.FailedToGetChatRoomByInquiryID,
				err.Error(),
			),
		)
		return

	}

	// Leave chat.
	ctx := context.Background()
	tx := db.GetDB().MustBegin()

	removedUsers, err := chatDao.WithTx(tx).LeaveAllMemebers(chatroom.ID)

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToLeaveAllMembers,
				err.Error(),
			),
		)

		tx.Rollback()
		return
	}

	// Soft delete chatroom
	if err := chatDao.
		WithTx(tx).
		DeleteChatRoom(chatroom.ID); err != nil {
		c.AbortWithError(
			http.StatusBadRequest,
			apperr.NewErr(
				apperr.FailedToDeleteChat,
				err.Error(),
			),
		)

		tx.Rollback()
		return
	}

	// Change inquiry status to `inquiring` if inquiry has not expired.
	q := models.New(tx)
	var lobbyChannelID *string
	if IsInquiryExpired(iq.ExpiredAt.Time) {
		*iq, err = q.PatchInquiryStatusByUuid(
			ctx,
			models.PatchInquiryStatusByUuidParams{
				Uuid:          iq.Uuid,
				InquiryStatus: models.InquiryStatusExpired,
			},
		)

		if err != nil {
			c.AbortWithError(
				http.StatusInternalServerError,
				apperr.NewErr(
					apperr.FailedToPatchInquiryStatus,
					err.Error(),
				),
			)

			tx.Rollback()
			return
		}
	} else {
		*iq, err = q.PatchInquiryStatusByUuid(
			ctx,
			models.PatchInquiryStatusByUuidParams{
				Uuid:          iq.Uuid,
				InquiryStatus: models.InquiryStatusInquiring,
			},
		)

		if err != nil {
			c.AbortWithError(
				http.StatusInternalServerError,
				apperr.NewErr(
					apperr.FailedToPatchInquiryStatus,
					err.Error(),
				),
			)

			tx.Rollback()
			return
		}

		// If requester is male user. Rejoin the user to lobby
		isMale, err := userDao.
			WithTx(tx).
			CheckIsMaleByUuid(c.GetString("uuid"))

		if err != nil {
			c.AbortWithError(
				http.StatusInternalServerError,
				apperr.NewErr(
					apperr.FailedToCheckGender,
					err.Error(),
				),
			)

			tx.Rollback()
			return
		}

		if isMale {
			lobbyService := NewLobbyService(NewLobbyDao(db.GetDB()))
			df := darkfirestore.Get()

			*lobbyChannelID, err = lobbyService.
				WithTx(tx).
				JoinLobby(iq.ID, df)

			if err != nil {
				c.AbortWithError(
					http.StatusInternalServerError,
					apperr.NewErr(
						apperr.FailedToJoinLobby,
						err.Error(),
					),
				)

				tx.Rollback()
				return
			}
		}
	}

	tx.Commit()

	c.JSON(http.StatusOK, NewTransform().TransformRevertChatting(
		removedUsers,
		*iq,
		*chatroom,
		lobbyChannelID,
	))
}

func GetServiceByInquiryUUID(c *gin.Context, depCon container.Container) {
	iqUUID := c.Param("uuid")

	var (
		serviceDao contracts.ServiceDAOer
	)

	depCon.Make(&serviceDao)

	// Retrieve service by inquiry uuid given
	srvModel, err := serviceDao.GetServiceByInquiryUUID(
		iqUUID,
		"services.uuid",
		"services.service_type",
		"services.price",
		"services.duration",
		"services.appointment_time",
	)

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToGetServiceByInquiryUUID,
				err.Error(),
			),
		)
		return
	}

	trfed, err := NewTransform().TransformGetServiceByInquiryUUID(*srvModel)

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToTransformServiceModel,
				err.Error(),
			),
		)

		return
	}

	c.JSON(http.StatusOK, trfed)
}

func GetInquirerInfo(c *gin.Context, depCon container.Container) {
	iqUUID := c.Param("uuid")

	// Retrieve inquiry info by UUID
	var (
		inquiryDAO contracts.InquiryDAOer
		imageDAO   contracts.ImageDAOer
	)

	depCon.Make(&inquiryDAO)
	depCon.Make(&imageDAO)

	inquirer, err := inquiryDAO.GetInquirerByInquiryUUID(iqUUID)

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToGetInquirerByInquiryUUID,
				err.Error(),
			),
		)

		return
	}

	images, err := imageDAO.GetImagesByUserID(int(inquirer.ID))

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToGetImagesByUserID,
				err.Error(),
			),
		)

		return
	}

	trfm, err := NewTransform().TransformGetInquirerInfo(
		*inquirer,
		images,
	)

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToTransformInquirerResponse,
				err.Error(),
			),
		)

		return
	}

	c.JSON(http.StatusOK, trfm)
}
