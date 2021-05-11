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
	df.CreateInquiringUser(
		ctx, darkfirestore.CreateInquiringUserParams{
			InquiryUUID: iq.Uuid,
		},
	)

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
		body.Offset,
		body.PerPage,
		models.InquiryStatusInquiring,
		models.InquiryStatusAsking,
	)

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
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
	transResp := db.TransactWithFormatStruct(
		db.GetDB(),
		func(tx *sqlx.Tx) db.FormatResp {
			ctx := context.Background()
			q := models.New(tx)

			uiq, err := q.PatchInquiryStatusByUuid(
				ctx, models.PatchInquiryStatusByUuidParams{
					InquiryStatus: models.InquiryStatus(fsm.Current()),
					Uuid:          uriParams.InquiryUuid,
				},
			)

			if err != nil {
				return db.FormatResp{
					HttpStatusCode: http.StatusInternalServerError,
					Err:            err,
					ErrCode:        apperr.FailedToPatchInquiryStatus,
				}
			}

			df := darkfirestore.Get()
			err = df.UpdateInquiryStatus(
				ctx,
				darkfirestore.UpdateInquiryStatusParams{
					InquiryUUID: uiq.Uuid,
					Status:      models.InquiryStatusCanceled,
				},
			)

			if err != nil {
				return db.FormatResp{
					HttpStatusCode: http.StatusInternalServerError,
					Err:            err,
					ErrCode:        apperr.FailedToChangeFirestoreInquiryStatus,
				}
			}

			return db.FormatResp{
				Response: uiq,
			}
		},
	)

	if transResp.Err != nil {
		c.AbortWithError(
			transResp.HttpStatusCode,
			apperr.NewErr(
				transResp.ErrCode,
				transResp.Err.Error(),
			),
		)

		return
	}

	uiq := transResp.Response.(models.ServiceInquiry)

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

	// Retrieve inquiry picker's ID which is the ID of the current requester.
	pickerID, err := q.GetUserIDByUuid(ctx, c.GetString("uuid"))

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToGetUserIDByUuid,
				err.Error(),
			),
		)

		return
	}

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
			apperr.NewErr(apperr.FailedToPickupStatusNotInquiring),
		)

		return
	}

	err, errCode := db.Transact(db.GetDB(), func(tx *sqlx.Tx) (error, interface{}) {
		fsm, _ := NewInquiryFSM(iq.InquiryStatus)

		if err := fsm.Event(Pickup.ToString()); err != nil {
			return err, apperr.InquiryFSMTransitionFailed
		}

		// Patch inquiry status in DB to be `asking`.
		iqDao := NewInquiryDAO(tx)
		_, err := iqDao.AskingInquiry(
			pickerID,
			iq.ID,
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
	InquiryUuid string `uri:"uuid" binding:"required"`
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

	// Retrieve picker info by uuid
	var userDao contracts.UserDAOer
	depCon.Make(&userDao)

	picker, err := userDao.GetUserByID(int64(iq.PickerID.Int32))

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

	fsm, err := NewInquiryFSM(iq.InquiryStatus)

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToCreateFSM,
				err.Error(),
			),
		)

		return
	}

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

	// Wrap the following actions in transaction
	//   - Update inquiry status in DB
	//   - Update inquiry status in firestore
	//   - Create and join inquirer and picker in a chatroom
	//   - Create a chatroom for inquirer and picker in firestore
	// @TODO wrap the following actions in to a service
	tranResp := db.TransactWithFormatStruct(db.GetDB(), func(tx *sqlx.Tx) db.FormatResp {

		// Update inquiry status in DB.
		inquiryDao := NewInquiryDAO(tx)
		if err := inquiryDao.PatchInquiryStatusByUUID(
			contracts.PatchInquiryStatusByUUIDParams{
				UUID:          iq.Uuid,
				InquiryStatus: models.InquiryStatusChatting,
			},
		); err != nil {
			return db.FormatResp{
				HttpStatusCode: http.StatusBadRequest,
				Err:            err,
				ErrCode:        apperr.FailedToPatchInquiryStatus,
			}

		}

		// Update inquiry status in firestore.
		df := darkfirestore.Get()
		if err := df.ChatInquiringUser(
			ctx,
			darkfirestore.ChatInquiringUserParams{
				InquiryUUID: iq.Uuid,
			},
		); err != nil {

			return db.FormatResp{
				HttpStatusCode: http.StatusBadRequest,
				Err:            err,
				ErrCode:        apperr.FailedToChangeFirestoreInquiryStatus,
			}
		}

		// Create private chatroom record in DB Join both inquirer and picker into the chatroom.
		var chat contracts.ChatServicer
		depCon.Make(&chat)

		chatroom, err := chat.WithTx(tx).CreateAndJoinChatroom(
			iq.ID,
			int64(iq.InquirerID.Int32),
			int64(iq.PickerID.Int32),
		)

		if err != nil {
			return db.FormatResp{
				HttpStatusCode: http.StatusBadRequest,
				Err:            err,
				ErrCode:        apperr.FailedToCreatePrivateChatRoom,
			}
		}

		// Create private chatroom in firestore
		if err := df.CreatePrivateChatRoom(
			ctx,
			darkfirestore.CreatePrivateChatRoomParams{
				ChatRoomName: chatroom.ChannelUuid.String,
				Data: darkfirestore.ChatMessage{
					Type: darkfirestore.Text,
					From: c.GetString("uuid"),
				},
			},
		); err != nil {
			return db.FormatResp{
				HttpStatusCode: http.StatusInternalServerError,
				Err:            err,
				ErrCode:        apperr.FailedToCreatePrivateChatroomInFirestore,
			}
		}

		return db.FormatResp{
			Response: chatroom,
		}
	})

	if tranResp.Err != nil {
		c.AbortWithError(
			tranResp.HttpStatusCode,
			apperr.NewErr(
				tranResp.ErrCode,
				tranResp.Err.Error(),
			),
		)

		return
	}

	chatroom := tranResp.Response.(*models.Chatroom)

	// Respoonse:
	//   - service provider's info
	//   - private chat uuid in firestore for inquirer to subscribe
	trf := NewTransform().TransformAgreePickupInquiry(
		*picker,
		chatroom.ChannelUuid.String,
	)

	c.JSON(http.StatusOK, trf)
}

type SkipPickupHandlerBody struct {
	InquiryUuid string `uri:"inquiry_uuid" binding:"required"`
}

func SkipPickupHandler(c *gin.Context, container container.Container) {
	body := SkipPickupHandlerBody{}

	if err := c.ShouldBindUri(&body); err != nil {
		c.AbortWithError(
			http.StatusBadRequest,
			apperr.NewErr(
				apperr.FailedToBindInquiryUriParams,
				err.Error(),
			),
		)

		return
	}

	iqDao := NewInquiryDAO(db.GetDB())
	iq, err := iqDao.GetInquiryByUuid(body.InquiryUuid)

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

	fsm, err := NewInquiryFSM(iq.InquiryStatus)

	if err != nil {
		c.AbortWithError(
			http.StatusBadRequest,
			apperr.NewErr(
				apperr.FailedToCreateFSM,
				err.Error(),
			),
		)

		return
	}

	if err := fsm.Event(Skip.ToString()); err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.InquiryFSMTransitionFailed,
				err.Error(),
			),
		)

		return
	}

	ctx := context.Background()

	// Change inquiry status from `asking` to `inquiring` in DB.
	// Change inquiry status from `asking` to `inquiring` in firestore. We use
	// inquiry uuid retrieved from DB to find the document in firestore.
	transResp := db.TransactWithFormatStruct(db.GetDB(), func(tx *sqlx.Tx) db.FormatResp {
		q := models.New(tx)
		iq, err := q.UpdateInquiryByUuid(
			ctx,
			models.UpdateInquiryByUuidParams{
				Uuid:          iq.Uuid,
				InquiryStatus: models.InquiryStatus(fsm.Current()),
			},
		)

		if err != nil {
			return db.FormatResp{
				Err:     err,
				ErrCode: apperr.FailedToUpdateInquiry,
			}
		}

		var df darkfirestore.DarkFireStorer
		container.Make(&df)

		err = df.UpdateInquiryStatus(
			ctx,
			darkfirestore.UpdateInquiryStatusParams{
				InquiryUUID: iq.Uuid,
				Status:      models.InquiryStatus(fsm.Current()),
			},
		)

		if err != nil {
			return db.FormatResp{
				Err:     err,
				ErrCode: apperr.FailedToChangeFirestoreInquiryStatus,
			}
		}

		return db.FormatResp{
			Response: iq,
		}
	})

	if transResp.Err != nil {
		c.AbortWithError(
			transResp.HttpStatusCode,
			apperr.NewErr(
				transResp.Err.Error(),
				err.Error(),
			),
		)

		return
	}

	c.JSON(http.StatusOK, struct{}{})
}

// Girl has approved the inquiry, thus, update the inquiry content.
//   - price
//   - duration
//   - appointment time
//   - lng
//   - lat
type GirlApproveInquiryBody struct {
	InquiryUuid     string    `uri:"inquiry_uuid" binding:"required"`
	Price           float64   `json:"price"`
	Duration        int       `json:"duration"`
	AppointmentTime time.Time `json:"appointment_time"`
	Lat             float64   `json:"lat"`
	Lng             float64   `json:"lng"`
}

func GirlApproveInquiryHandler(c *gin.Context, depCon container.Container) {
	body := GirlApproveInquiryBody{}

	if err := c.ShouldBindUri(&body); err != nil {
		c.AbortWithError(
			http.StatusBadRequest,
			apperr.NewErr(
				apperr.FailedToBindInquiryUriParams,
				err.Error(),
			),
		)

		return
	}

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
	var iqDaoer contracts.InquiryDAOer
	depCon.Make(&iqDaoer)
	iq, err := iqDaoer.GetInquiryByUuid(body.InquiryUuid)

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

	// Perform inquiry status transition
	fsm, err := NewInquiryFSM(iq.InquiryStatus)

	if err != nil {
		c.AbortWithError(
			http.StatusBadRequest,
			apperr.NewErr(
				apperr.FailedToCreateFSM,
				err.Error(),
			),
		)

		return
	}

	if err := fsm.Event(GirlApprove.ToString()); err != nil {
		c.AbortWithError(
			http.StatusBadRequest,
			apperr.NewErr(
				apperr.InquiryFSMTransitionFailed,
				err.Error(),
			),
		)

		return
	}

	// Wrap the following actions in a transaction:
	//   - Update inquiry status in DB
	//   - Update inquiry status in firestore
	transResp := db.TransactWithFormatStruct(db.GetDB(), func(tx *sqlx.Tx) db.FormatResp {
		ctx := context.Background()

		q := models.New(tx)

		// Update inqiury status in DB
		latDec := decimal.NewFromFloat(body.Lng)
		lngDec := decimal.NewFromFloat(body.Lat)

		uiq, err := q.UpdateInquiryByUuid(
			ctx,
			models.UpdateInquiryByUuidParams{
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

				Uuid:          body.InquiryUuid,
				InquiryStatus: models.InquiryStatus(fsm.Current()),
			},
		)

		if err != nil {
			return db.FormatResp{
				Err:     err,
				ErrCode: apperr.FailedToUpdateInquiry,
			}
		}

		// Update inquiry status in firestore
		df := darkfirestore.Get()
		if err := df.UpdateInquiryStatus(
			ctx,
			darkfirestore.UpdateInquiryStatusParams{
				InquiryUUID: iq.Uuid,
				Status:      models.InquiryStatus(fsm.Current()),
			},
		); err != nil {
			return db.FormatResp{
				Err:     err,
				ErrCode: apperr.FailedToChangeFirestoreInquiryStatus,
			}
		}

		return db.FormatResp{
			Response: &uiq,
		}
	})

	if transResp.Err != nil {
		c.AbortWithError(
			transResp.HttpStatusCode,
			apperr.NewErr(
				transResp.ErrCode,
				transResp.Err,
			),
		)

		return
	}

	uiq := transResp.Response.(*models.ServiceInquiry)

	trf, err := NewTransform().TransformGirlApproveInquiry(*uiq)

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

	c.JSON(http.StatusOK, trf)
}

// Emit event to girl for the purpose of notifying them the iquiry is booked by the man
type ManBookInquiryBody struct {
	Price               float64   `json:"price" binding:"required"`
	Duration            int       `json:"duration" binding:"required"`
	AppointmentTime     time.Time `json:"appointment_time" binding:"required"`
	Lng                 float64   `json:"lng" binding:"required"`
	Lat                 float64   `json:"lat" binding:"required"`
	ServiceType         string    `json:"service_type" binding:"required"`
	ServiceProviderUuid string    `json:"service_provider_uuid" binding:"required"`
	ChannelUuid         string    `json:"channel_uuid" binding:required`
}

// BookInquiryTransResp stores the data of the transaction that performs
// booking inquiry logic.
type BookInquiryTransResp struct {
	Service  models.Service
	Chatroom models.Chatroom
}

func ManBookInquiry(c *gin.Context, depCon container.Container) {
	inquiryUuid := c.Param("inquiry_uuid")

	if inquiryUuid == "" {
		c.AbortWithError(
			http.StatusBadRequest,
			apperr.NewErr(apperr.InquiryUUIDNotInParams),
		)

		return

	}

	body := ManBookInquiryBody{}

	if err := requestbinder.Bind(c, &body); err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToValidateBookInquiryParams,
				err.Error(),
			),
		)

		return
	}

	q := models.New(db.GetDB())
	ctx := context.Background()

	srvProvider, err := q.GetUserByUuid(ctx, body.ServiceProviderUuid)

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

	// Wrap db actions of performing booking inquiry in a transaction
	transResp := db.TransactWithFormatStruct(db.GetDB(), func(tx *sqlx.Tx) db.FormatResp {
		transq := models.New(tx)

		uInquiry, err := transq.UpdateInquiryByUuid(
			ctx,
			models.UpdateInquiryByUuidParams{
				Uuid:          inquiryUuid,
				InquiryStatus: models.InquiryStatusBooked,
			},
		)

		if err != nil {
			return db.FormatResp{
				Err:     err,
				ErrCode: apperr.FailedToGetUserIDByUuid,
			}
		}

		// Change the inquiry status in firestore to booked
		df := darkfirestore.Get()
		if err = df.UpdateInquiryStatus(
			ctx,
			darkfirestore.UpdateInquiryStatusParams{
				InquiryUUID: uInquiry.Uuid,
				Status:      models.InquiryStatusBooked,
			},
		); err != nil {
			return db.FormatResp{
				Err:     err,
				ErrCode: apperr.FailedToChangeFirestoreInquiryStatus,
			}
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
				ServiceType:   uInquiry.ServiceType,
				InquiryID:     int32(uInquiry.ID),
				CustomerID: sql.NullInt32{
					Int32: uInquiry.InquirerID.Int32,
					Valid: true,
				},

				ServiceProviderID: sql.NullInt32{
					Int32: int32(srvProvider.ID),
					Valid: true,
				},
			},
		)

		if err != nil {
			return db.FormatResp{
				Err:     err,
				ErrCode: apperr.FailedToCreateService,
			}
		}

		// We need to change the chatroom type of the given inquiry to service_chat in DB
		var chatroomDao contracts.ChatDaoer
		depCon.Make(&chatroomDao)

		chatroom, err := chatroomDao.
			WithTx(tx).
			UpdateChatByUuid(
				contracts.UpdateChatByUuidParams{
					ChatroomType: models.ChatroomTypeServiceChat,
					ChannelUuid:  body.ChannelUuid,
				},
			)

		if err != nil {
			return db.FormatResp{
				Err:     err,
				ErrCode: apperr.FailedToUpdateChatroom,
			}
		}

		return db.FormatResp{
			Response: &BookInquiryTransResp{
				Service:  srv,
				Chatroom: *chatroom,
			},
		}
	})

	if transResp.Err != nil {
		c.AbortWithError(
			transResp.HttpStatusCode,
			apperr.NewErr(
				transResp.ErrCode,
				err.Error(),
			),
		)

		return
	}

	transResult := transResp.Response.(*BookInquiryTransResp)

	trf := NewTransform().TransformBookedService(
		transResult.Service,
		srvProvider,
	)

	c.JSON(http.StatusOK, trf)
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

	// Leave chat for both inquirer and picker
	type TransResult struct {
		RemovedUsers []models.User
		Inquiry      models.ServiceInquiry
	}

	transResp := db.TransactWithFormatStruct(
		db.GetDB(),
		func(tx *sqlx.Tx) db.FormatResp {
			ctx := context.Background()

			removedUsers, err := chatDao.WithTx(tx).LeaveAllMemebers(chatroom.ID)

			if err != nil {
				return db.FormatResp{
					Err:     err,
					ErrCode: apperr.FailedToLeaveAllMembers,
				}
			}

			// Soft delete chatroom
			if chatDao.
				WithTx(tx).
				DeleteChatRoom(chatroom.ID); err != nil {
				return db.FormatResp{
					Err:            err,
					ErrCode:        apperr.FailedToDeleteChat,
					HttpStatusCode: http.StatusBadRequest,
				}
			}

			// Change inquiry status to `inquiring`
			q := models.New(tx)
			iq, err := q.UpdateInquiryByUuid(
				ctx,
				models.UpdateInquiryByUuidParams{
					Uuid:          iq.Uuid,
					InquiryStatus: models.InquiryStatusInquiring,
				},
			)

			if err != nil {
				return db.FormatResp{
					Err:     err,
					ErrCode: apperr.FailedToPatchInquiryStatus,
				}
			}

			// Emit new inquiry status to firestore `inquiring` so that the other
			// party knows to quit the chatroom.
			df := darkfirestore.Get()
			if err := df.UpdateInquiryStatus(
				ctx,
				darkfirestore.UpdateInquiryStatusParams{
					InquiryUUID: iq.Uuid,
					Status:      models.InquiryStatusInquiring,
				},
			); err != nil {
				return db.FormatResp{
					HttpStatusCode: http.StatusBadRequest,
					Err:            err,
					ErrCode:        apperr.FailedToChangeFirestoreInquiryStatus,
				}
			}

			return db.FormatResp{
				Response: &TransResult{
					RemovedUsers: removedUsers,
					Inquiry:      iq,
				},
			}
		},
	)

	if transResp.Err != nil {
		c.AbortWithError(
			transResp.HttpStatusCode,
			apperr.NewErr(
				transResp.ErrCode,
				transResp.Err,
			),
		)

		return
	}

	transResult := transResp.Response.(*TransResult)

	c.JSON(http.StatusOK, NewTransform().TransformRevertChatting(
		transResult.RemovedUsers,
		transResult.Inquiry,
		*chatroom,
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

type PatchInquiryUriUuid struct {
	Uuid string `uri:"inquiry_uuid"`
}

type PatchInquiryBody struct {
	Uuid            string     `uri:"inquiry_uuid" form:"uuid" json:"uuid"`
	AppointmentTime *time.Time `form:"appointment_time" json:"appointment_time"`
	Price           *float32   `form:"price" json:"price"`
	Duration        *int       `form:"duration" json:"duration"`
	ServiceType     *string    `form:"service_type" json:"service_type"`
	Address         *string    `form:"address" json:"address"`
}

func PatchInquiryHandler(c *gin.Context, depCon container.Container) {
	body := PatchInquiryBody{}

	if err := requestbinder.Bind(c, &body); err != nil {
		c.AbortWithError(
			http.StatusBadRequest,
			apperr.NewErr(
				apperr.FailedToValidatePatchInquiryParams,
				err.Error(),
			),
		)

		return
	}

	iqUuidUri := PatchInquiryUriUuid{}

	if err := c.BindUri(&iqUuidUri); err != nil {
		return
	}

	dao := NewInquiryDAO(db.GetDB())
	inquiry, err := dao.PatchInquiryByInquiryUUID(contracts.PatchInquiryParams{
		Uuid:            iqUuidUri.Uuid,
		AppointmentTime: body.AppointmentTime,
		Price:           body.Price,
		Duration:        body.Duration,
		ServiceType:     body.ServiceType,
		Address:         body.Address,
	})

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToPatchInquiry,
				err.Error(),
			),
		)

		return
	}

	trf, err := NewTransform().TransformUpdateInquiry(inquiry)

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToTransformUpdateInquiry,
				err.Error(),
			),
		)

		return
	}

	c.JSON(http.StatusOK, trf)
}
