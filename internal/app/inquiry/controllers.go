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
	"github.com/huangc28/go-darkpanda-backend/internal/app/contracts"
	"github.com/huangc28/go-darkpanda-backend/internal/app/inquiry/util"
	"github.com/huangc28/go-darkpanda-backend/internal/app/models"
	"github.com/huangc28/go-darkpanda-backend/internal/app/pkg/requestbinder"
	"github.com/jmoiron/sqlx"
	"github.com/looplab/fsm"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"github.com/teris-io/shortid"
)

type InquiryHandlers struct {
	UserDao       contracts.UserDAOer
	InquiryDao    InquiryDAOer
	LobbyServices LobbyServicer
	ChatServices  contracts.ChatServicer
	ChatDao       contracts.ChatDaoer
}

// @TODO budget received from client should be type float instead of string.
//       budget should be converted to type string before stored in DB.
type EmitInquiryBody struct {
	Budget      float64 `form:"budget" uri:"budget" json:"budget" binding:"required"`
	ServiceType string  `form:"service_type" uri:"service_type" json:"service_type" binding:"required"`
}

func (h *InquiryHandlers) EmitInquiryHandler(c *gin.Context) {
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

	// Joins the lobby and returns lobby channel id
	channelID, err := h.LobbyServices.JoinLobby(iq.ID)

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

	trf, err := NewTransform().TransformEmitInquiry(iq, channelID)

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

func (h *InquiryHandlers) GetInquiriesHandler(c *gin.Context) {
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

	// offset should be passed from client
	inquiries, err := h.InquiryDao.GetInquiries(
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
	hasMoreRecord, err := h.InquiryDao.HasMoreInquiries(
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

func (h *InquiryHandlers) PickupInquiryHandler(c *gin.Context) {
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
	ctx := context.Background()

	// retrieve inquirier information
	q := models.New(db.GetDB())
	inquirer, err := q.GetUserByID(ctx, int64(iq.InquirerID.Int32))

	if err != nil {
		log.Fatal(err)
	}

	// Check if user in the lobby has already expired
	expired, err := h.InquiryDao.IsInquiryExpired(iq.ID)

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToCheckLobbyExpiry,
				err.Error(),
			),
		)

		return
	}

	if expired {
		// Remove user from lobby
		if err := h.LobbyServices.LeaveLobby(iq.ID); err != nil {
			c.AbortWithError(
				http.StatusInternalServerError,
				apperr.NewErr(
					apperr.FailedToLeaveLobby,
					err.Error(),
				),
			)

			return
		}

		// @TODO Notify clients via socket event that the inquiry is expired and should be removed from inquiry list.

		c.AbortWithError(
			http.StatusBadRequest,
			apperr.NewErr(apperr.CanNotPickupExpiredInquiry),
		)

		return
	}

	var (
		uiq          models.ServiceInquiry
		chatroomInfo *models.ChatInfo
	)

	err, errCode := db.Transact(db.GetDB(), func(tx *sqlx.Tx) (error, interface{}) {
		dao := NewInquiryDAO(tx)
		lastVerIq, err := dao.GetInquiryByUuid(
			uriParams.InquiryUuid,
			"updated_at",
		)

		if err != nil {
			return err, apperr.FailedToGetInquiryByUuid
		}

		// Before patching the inquiry status, we apply optimistic lock strategy which checks `updated_at` column of that inquiry again
		// makesure it hasn't been modified by other transactions / processes. If `updated_at` at has been modified, abort the transaction.
		if !lastVerIq.UpdatedAt.Time.Equal(iq.UpdatedAt.Time) {
			return apperr.NewErr(apperr.FailedToPickupInquiryDueToDirtyVersion), apperr.FailedToPickupInquiryDueToDirtyVersion
		}

		uiq, err = q.PatchInquiryStatusByUuid(ctx, models.PatchInquiryStatusByUuidParams{
			InquiryStatus: models.InquiryStatus(fsm.Current()),
			Uuid:          uriParams.InquiryUuid,
		})

		if err != nil {
			return err, apperr.FailedToPatchInquiryStatus
		}

		// @TODO
		//   - We would need to notify the male user waiting in the lobby to enter the chatroom that the female user has created for him.
		//   - Male user should leave lobby
		if err := h.LobbyServices.
			WithTx(tx).
			LeaveLobby(uiq.ID); err != nil {
			return err, apperr.FailedToLeaveLobby
		}

		servicePicker, err := h.UserDao.
			WithTx(tx).
			GetUserByUuid(c.GetString("uuid"))

		if err != nil {
			return err, apperr.FailedToGetUserByUuid
		}

		// Both male and Female user should also join private chatroom.
		chatroomInfo, err = h.ChatServices.
			WithTx(tx).
			CreateAndJoinChatroom(uiq.ID, inquirer.ID, servicePicker.ID)

		if err != nil {
			return err, apperr.FailedToCreateAndJoinLobby
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

	trf, err := NewTransform().TransformPickupInquiry(
		uiq,
		inquirer,
		chatroomInfo.ChanelUuid,
	)

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
func (h *InquiryHandlers) RevertChat(c *gin.Context) {
	eiiq, exists := c.Get("inquiry")

	if !exists {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(apperr.FSMNotSetInMiddleware),
		)
		return
	}

	iq := (eiiq).(*models.ServiceInquiry)

	// Find chatroom by inquiry_id, find inquiry_id by inquiry_uuid
	chatroom, err := h.ChatDao.GetChatRoomByInquiryID(
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

	removedUsers, err := h.ChatDao.WithTx(tx).LeaveAllMemebers(chatroom.ID)

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
	if err := h.ChatDao.
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
		isMale, err := h.UserDao.
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
			*lobbyChannelID, err = h.LobbyServices.
				WithTx(tx).
				JoinLobby(iq.ID)

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

	// Returns:
	//   - [x]Removed users
	//   - [x]Inquiry info
	//   - [x]Deleted chatroom
	c.JSON(http.StatusOK, NewTransform().TransformRevertChatting(
		removedUsers,
		*iq,
		*chatroom,
		lobbyChannelID,
	))
}
