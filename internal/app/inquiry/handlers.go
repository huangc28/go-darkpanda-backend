package inquiry

import (
	"context"
	"database/sql"
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
	AppointmentTime time.Time `form:"appointment_time" json:"appointment_time" binding:"required"`
	ServiceDuration int       `form:"service_duration" json:"service_duration" binding:"required"`
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

	q := models.New(db.GetDB())
	usr, err := q.GetUserByUuid(ctx, c.GetString("uuid"))

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
	iq, err := q.CreateInquiry(
		ctx,
		models.CreateInquiryParams{
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
			AppointmentTime: sql.NullTime{
				Valid: true,

				// Convert appointment time to UTC to be consistent.
				Time: body.AppointmentTime.UTC(),
			},
			Duration: sql.NullInt32{
				Valid: true,
				Int32: int32(body.ServiceDuration),
			},
		},
	)

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
			InquiryUuid:  iq.Uuid,
			InquirerUuid: usr.Uuid,
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
	PerPage int `form:"per_page,default=7"`
}

func GetInquiriesHandler(c *gin.Context, depCon container.Container) {
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
	var userDao contracts.UserDAOer
	depCon.Make(&userDao)
	user, err := userDao.GetUserByUuid(c.GetString("uuid"), "id")

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

	// offset should be passed from client. The query should exclude
	// those ids that is in the block list.
	inquiries, err := inquiryDao.GetInquiries(
		int(user.ID),
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

type CancelInquiryParam struct {
	InquiryUuid string `json:"inquiry_uuid" form:"inquiry_uuid" binding:"required"`
}

func CancelInquiryHandler(c *gin.Context) {
	body := CancelInquiryParam{}

	if err := requestbinder.Bind(c, &body); err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToBindBodyParams,
				err.Error(),
			),
		)

		return
	}

	// Check if requester is the inquiry owner.
	ctx := context.Background()
	q := models.New(db.GetDB())
	err := q.CheckUserOwnsInquiry(
		ctx, models.CheckUserOwnsInquiryParams{
			Uuid:   c.GetString("uuid"),
			Uuid_2: body.InquiryUuid,
		},
	)

	if err == sql.ErrNoRows {
		c.AbortWithError(
			http.StatusBadRequest,
			apperr.NewErr(apperr.UserNotOwnInquiry),
		)

		return
	}

	// Emit fsm state transition.
	iq, err := q.GetInquiryByUuid(ctx, body.InquiryUuid)

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
	if err := fsm.Event(Cancel.ToString()); err != nil {
		c.AbortWithError(
			http.StatusBadRequest,
			apperr.NewErr(
				apperr.InquiryFSMTransitionFailed,
				err.Error(),
			),
		)

		return
	}

	// ------------------- Update inquiry status to cancel  -------------------
	uiq, err := q.PatchInquiryStatusByUuid(
		ctx, models.PatchInquiryStatusByUuidParams{
			InquiryStatus: models.InquiryStatus(fsm.Current()),
			Uuid:          body.InquiryUuid,
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

		return
	}

	df := darkfirestore.Get()
	_, err = df.UpdateInquiryStatus(
		ctx,
		darkfirestore.UpdateInquiryStatusParams{
			InquiryUuid: uiq.Uuid,
			Status:      models.InquiryStatusCanceled,
		},
	)

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToChangeFirestoreInquiryStatus,
				err.Error(),
			),
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
	InquiryUuid string `form:"inquiry_uuid" json:"inquiry_uuid" binding:"required"`
}

func PickupInquiryHandler(c *gin.Context, depCon container.Container) {
	var params PickupInquiryHandlerParams

	if err := requestbinder.Bind(c, &params); err != nil {
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
	// pickerID, err := q.GetUserIDByUuid(ctx, c.GetString("uuid"))
	picker, err := q.GetUserByUuid(ctx, c.GetString("uuid"))

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

	fsm, _ := NewInquiryFSM(iq.InquiryStatus)

	if err := fsm.Event(Pickup.ToString()); err != nil {
		c.AbortWithError(
			http.StatusBadRequest,
			apperr.NewErr(
				apperr.InquiryFSMTransitionFailed,
				err.Error(),
			),
		)

		return
	}

	// Patch inquiry status in DB to be `asking`.
	iqDao := NewInquiryDAO(db.GetDB())
	if _, err := iqDao.AskingInquiry(
		picker.ID,
		iq.ID,
	); err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToUpdateInquiryContent,
				err.Error(),
			),
		)

		return
	}

	// Patch inquiry status in firestore to be `asking`
	df := darkfirestore.Get()
	if err = df.AskingInquiringUser(
		ctx,
		darkfirestore.AskingInquiringUserParams{
			InquiryUuid:    iq.Uuid,
			PickerUuid:     c.GetString("uuid"),
			PickerUsername: picker.Username,
		},
	); err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToAskInquiringUser,
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

type AgreeToChatParams struct {
	InquiryUuid string `form:"inquiry_uuid" json:"inquiry_uuid" binding:"required,gt=0"`
}

// AgreePickupInquiryHandler Male user agree to have a chat with the male user.
// Perform following operations when male user agrees to chat.
//   - Check inquiry status can be transitioned to `chatting`
//   - Change inquiry status to `chatting` on DB
//   - Change inquiry status to `chatting` on firestore
func AgreeToChatInquiryHandler(c *gin.Context, depCon container.Container) {
	var params AgreeToChatParams

	if err := requestbinder.Bind(c, &params); err != nil {
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
			apperr.NewErr(apperr.InquiryHasNoPicker),
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

	// Update inquiry status in firestore.
	df := darkfirestore.Get()
	if err := df.ChatInquirer(
		ctx,
		darkfirestore.ChatInquirerParams{
			InquiryUUID: iq.Uuid,
		},
	); err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToChangeFirestoreInquiryStatus,
				err.Error(),
			),
		)

		return
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
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToCreatePrivateChatroomInFirestore,
				err.Error(),
			),
		)

		return
	}

	// Retrieve chatroom relative information.
	var chatDao contracts.ChatDaoer
	depCon.Make(&chatDao)

	chatInfoModel, err := chatDao.GetCompleteChatroomInfoById(int(chatroom.ID))

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToGetChatroomById,
				err.Error(),
			),
		)

		return
	}

	inquirer, err := userDao.GetUserByID(int64(iq.InquirerID.Int32))

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

	// Respoonse:
	//   - service provider's info
	//   - private chat uuid in firestore for inquirer to subscribe
	trf := NewTransform().TransformAgreePickupInquiry(
		*picker,
		*inquirer,
		chatInfoModel,
	)

	c.JSON(http.StatusOK, trf)
}

type SkipPickupHandlerBody struct {
	InquiryUuid string `form:"inquiry_uuid" json:"inquiry_uuid" binding:"required"`
}

func SkipPickupHandler(c *gin.Context, container container.Container) {
	body := SkipPickupHandlerBody{}

	if err := requestbinder.Bind(c, &body); err != nil {
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
	q := models.New(db.GetDB())
	if _, err := q.UpdateInquiryByUuid(
		ctx,
		models.UpdateInquiryByUuidParams{
			Uuid:          iq.Uuid,
			InquiryStatus: models.InquiryStatus(fsm.Current()),
		},
	); err != nil {

		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToUpdateInquiry,
				err.Error(),
			),
		)

		return
	}

	var df darkfirestore.DarkFireStorer
	container.Make(&df)

	_, err = df.UpdateInquiryStatus(
		ctx,
		darkfirestore.UpdateInquiryStatusParams{
			InquiryUuid:    iq.Uuid,
			Status:         models.InquiryStatus(fsm.Current()),
			PickerUuid:     "",
			PickerUsername: "",
		},
	)

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToChangeFirestoreInquiryStatus,
				err.Error(),
			),
		)

		return
	}

	c.JSON(http.StatusOK, struct{}{})
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

type PatchInquiryBody struct {
	Uuid            string     `uri:"inquiry_uuid" form:"uuid" json:"uuid"`
	AppointmentTime *time.Time `form:"appointment_time" json:"appointment_time"`
	Price           *float32   `form:"price" json:"price"`
	Budget          *float32   `form:"budget" json:"budget"`
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

	dao := NewInquiryDAO(db.GetDB())
	inquiry, err := dao.PatchInquiryByInquiryUUID(
		contracts.PatchInquiryParams{
			Uuid:            c.Param("inquiry_uuid"),
			Budget:          body.Budget,
			AppointmentTime: body.AppointmentTime,
			Price:           body.Price,
			Duration:        body.Duration,
			ServiceType:     body.ServiceType,
			Address:         body.Address,
		},
	)

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

func GetActiveInquiry(c *gin.Context, depCon container.Container) {
	userUuid := c.GetString("uuid")

	var userDao contracts.UserDAOer
	depCon.Make(&userDao)

	user, err := userDao.GetUserByUuid(userUuid, "id")

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

	iqDao := NewInquiryDAO(db.GetDB())
	iq, err := iqDao.GetActiveInquiry(int(user.ID))

	if err == sql.ErrNoRows {
		c.AbortWithError(
			http.StatusNotFound,
			apperr.NewErr(
				apperr.NoActiveInquiry,
				err.Error(),
			),
		)

		return
	}

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToActiveInquiry,
				err.Error(),
			),
		)

		return
	}

	trf, err := NewTransform().TransformActiveInquiry(iq)

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToTransformActiveInquiry,
				err.Error(),
			),
		)

		return

	}

	c.JSON(http.StatusOK, trf)
}
