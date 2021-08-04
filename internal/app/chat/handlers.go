package chat

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/golobby/container/pkg/container"
	"github.com/jmoiron/sqlx"
	"github.com/skip2/go-qrcode"
	"github.com/teris-io/shortid"

	"github.com/gin-gonic/gin"
	"github.com/huangc28/go-darkpanda-backend/config"
	"github.com/huangc28/go-darkpanda-backend/db"
	"github.com/huangc28/go-darkpanda-backend/internal/app/apperr"
	"github.com/huangc28/go-darkpanda-backend/internal/app/contracts"
	"github.com/huangc28/go-darkpanda-backend/internal/app/models"
	darkfirestore "github.com/huangc28/go-darkpanda-backend/internal/app/pkg/dark_firestore"
	gcsenhancer "github.com/huangc28/go-darkpanda-backend/internal/app/pkg/gcs_enhancer"
	"github.com/huangc28/go-darkpanda-backend/internal/app/pkg/requestbinder"
	"github.com/huangc28/go-darkpanda-backend/internal/app/util"

	convertnullsql "github.com/huangc28/go-darkpanda-backend/internal/app/pkg/convert_null_sql"
)

type ChatHandlers struct {
	ChatDao    contracts.ChatDaoer
	UserDao    contracts.UserDAOer
	ServiceDao contracts.ServiceDAOer
	InquiryDao contracts.InquiryDAOer
}

type EmitTextMessageBody struct {
	Content     string `form:"content" binding:"required"`
	ChannelUUID string `form:"channel_uuid" binding:"required"`
}

// @TODOs
//   - Check if chatroom has expired
func EmitTextMessage(c *gin.Context, depCon container.Container) {
	body := EmitTextMessageBody{}

	if err := requestbinder.Bind(c, &body); err != nil {
		c.AbortWithError(
			http.StatusBadRequest,
			apperr.NewErr(
				apperr.FailedToValidateEmitTextMessageParams,
				err.Error(),
			),
		)

		return
	}

	var chatDao contracts.ChatDaoer

	depCon.Make(&chatDao)

	// Emit message to firestore.
	ctx := context.Background()
	df := darkfirestore.Get()
	message, err := df.SendTextMessageToChatroom(ctx, darkfirestore.SendTextMessageParams{
		ChannelUuid: body.ChannelUUID,
		Data: darkfirestore.ChatMessage{
			Content: body.Content,
			From:    c.GetString("uuid"),
		},
	})

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToSendTextMessage,
				err.Error(),
			),
		)

		return
	}

	// Emit message to channel
	c.JSON(
		http.StatusOK,
		NewTransformer().TransformEmitTextMessage(
			body.ChannelUUID,
			message,
		),
	)
}

// EmitServiceSettingMessage if the female user edited service details and saved the service settings,
// the chatroom would emit a service setting message. Male user would be notified with the service message.
// Male user can click on the service message and would show the service detail set by the female user.
type EmitServiceSettingMessage struct {
	Price           float64   `form:"price" binding:"required"`
	ChannelUUID     string    `form:"channel_uuid" binding:"required"`
	InquiryUUID     string    `form:"inquiry_uuid" binding:"required"`
	ServiceTime     time.Time `form:"service_time" binding:"required"`
	ServiceDuration int       `form:"service_duration" binding:"required"`
	ServiceType     string    `form:"service_type" binding:"required"`
}

func EmitServiceSettingMessageHandler(c *gin.Context, depCon container.Container) {
	// Check if the corresponding service has already been created with given inquiry
	// if service has been created, update the service detail,
	// if not, create the serice with given service detail.
	// After DB operation, emit service updated chatroom message.
	body := EmitServiceSettingMessage{}
	if err := requestbinder.Bind(c, &body); err != nil {
		c.AbortWithError(
			http.StatusBadRequest,
			apperr.NewErr(
				apperr.FailedToValidateEmitServiceSettingMessageParams,
				err.Error(),
			),
		)

		return
	}

	var (
		userDao    contracts.UserDAOer
		inquiryDao contracts.InquiryDAOer
		serviceDao contracts.ServiceDAOer
	)

	depCon.Make(&userDao)
	depCon.Make(&inquiryDao)
	depCon.Make(&serviceDao)

	user, err := userDao.GetUserByUuid(
		c.GetString("uuid"),
		"id",
		"username",
	)

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

	inquiry, err := inquiryDao.GetInquiryByUuid(body.InquiryUUID, "")

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

	service, err := serviceDao.GetServiceByInquiryUUID(body.InquiryUUID, "services.*")

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

	ctx := context.Background()

	if errors.Is(err, sql.ErrNoRows) {
		q := models.New(db.GetDB())
		sid, _ := shortid.Generate()
		*service, err = q.CreateService(ctx, models.CreateServiceParams{
			Uuid: sql.NullString{
				Valid:  true,
				String: sid,
			},
			CustomerID: sql.NullInt32{
				Valid: true,
				Int32: inquiry.InquirerID.Int32,
			},
			ServiceProviderID: sql.NullInt32{
				Int32: int32(user.ID),
				Valid: true,
			},
			Price: sql.NullString{
				String: fmt.Sprintf("%f", body.Price),
				Valid:  true,
			},
			Duration: sql.NullInt32{
				Int32: int32(body.ServiceDuration),
				Valid: true,
			},
			AppointmentTime: sql.NullTime{
				Time:  body.ServiceTime,
				Valid: true,
			},
			InquiryID:     int32(inquiry.ID),
			ServiceStatus: models.ServiceStatusUnpaid,
			ServiceType:   models.ServiceType(body.ServiceType),
		})

		if err != nil {
			c.AbortWithError(
				http.StatusInternalServerError,
				apperr.NewErr(
					apperr.FailedToCreateService,
					err.Error(),
				),
			)

			return
		}
	}

	if err != nil {
		c.AbortWithError(
			http.StatusBadRequest,
			apperr.NewErr(
				apperr.FailedToGetServiceByInquiryUUID,
				err.Error(),
			),
		)

		return
	}

	// Corresponding service exists, update detail of the service.
	srvType := models.ServiceType(body.ServiceType)
	service, err = serviceDao.UpdateServiceByID(contracts.UpdateServiceByIDParams{
		ID:          service.ID,
		Price:       &body.Price,
		Duration:    &body.ServiceDuration,
		Appointment: &body.ServiceTime,
		ServiceType: &srvType,
	})

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToUpdateService,
				err.Error(),
			),
		)

		return
	}

	// Emit service setting message to chatroom.
	var coinPkgDao contracts.CoinPackageDAOer
	depCon.Make(&coinPkgDao)

	matchingFee, err := coinPkgDao.GetMatchingFee()

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToGetMatchingFee,
				err.Error(),
			),
		)

		return
	}

	df := darkfirestore.Get()
	message, err := df.SendServiceDetailMessageToChatroom(ctx, darkfirestore.SendServiceDetailMessageParams{
		ChannelUuid: body.ChannelUUID,
		Data: darkfirestore.ServiceDetailMessage{
			ChatMessage: darkfirestore.ChatMessage{
				Content:  "",
				From:     c.GetString("uuid"),
				Username: user.Username,
			},
			Price:       body.Price,
			Duration:    int(service.Duration.Int32),
			ServiceTime: service.AppointmentTime.Time.UnixNano() / int64(time.Microsecond),
			ServiceType: body.ServiceType,
			ServiceUUID: service.Uuid.String,
			MatchingFee: int(matchingFee.Cost.Int32),
		},
	})

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToSendServiceDetailMsg,
				err.Error(),
			),
		)

		return
	}

	c.JSON(http.StatusOK, message)
}

type GetChatroomsBody struct {
	Offset  int `form:"offset,default=0"`
	PerPage int `form:"per_page,default=7"`
}

// If the requester is female find all chatrooms that qualify the following conditions:
//   - Chatrooms related inquiry status is chatting
//   - Chatrooms related inquiry picker_id equals requester's id
func GetChatrooms(c *gin.Context, depCon container.Container) {
	body := GetChatroomsBody{}

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

	// Recognize the gender of the requester
	var (
		userDao  contracts.UserDAOer
		chatDao  contracts.ChatDaoer
		userUUID string = c.GetString("uuid")
	)

	depCon.Make(&userDao)
	depCon.Make(&chatDao)

	user, err := userDao.GetUserByUuid(userUUID, "id", "gender")

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

	var chatrooms []models.InquiryChatRoom

	chatrooms, err = chatDao.GetFemaleInquiryChatRooms(
		user.ID,
		int64(body.Offset),
		int64(body.PerPage),
	)

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToGetFemaleChatRooms,
				err.Error(),
			),
		)

		return
	}

	// Retrieve first message of each chatroom from firestore
	channelUUIDs := []string{}
	for _, chatroom := range chatrooms {
		channelUUIDs = append(channelUUIDs, chatroom.ChannelUUID)
	}

	ctx := context.Background()
	channelUUIDMessageMap, err := darkfirestore.
		Get().
		GetLatestMessageForEachChatroom(ctx, channelUUIDs)

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToGetMessageFromFireStore,
				err.Error(),
			),
		)

		return
	}

	c.JSON(
		http.StatusOK,
		NewTransformer().TransformInquiryChats(
			chatrooms,
			channelUUIDMessageMap,
		),
	)

}

// Add pagination for firestore. We have to get `page` and `limit` from client.
//   - page
//   - perpage
// Calculate offset from page number and perpage .
type GetMessagesBody struct {
	PerPage int `form:"perpage,default=10"`
	Page    int `form:"page,default=0"`
}

func GetHistoricalMessages(c *gin.Context) {
	channelUUID := c.Param("channel_uuid")

	body := GetMessagesBody{}

	if err := requestbinder.Bind(c, &body); err != nil {
		c.AbortWithError(
			http.StatusBadRequest,
			apperr.NewErr(
				apperr.FailedGetHistoricalMessagesFromFireStore,
				err.Error(),
			),
		)

		return
	}

	// Calculate offset from `perpage` and `page`.
	offset := util.CalcPaginateOffset(
		body.Page,
		body.PerPage,
	)

	// Retrieve the last record of the previous page.
	ctx := context.Background()
	msgs, err := darkfirestore.
		Get().
		GetHistoricalMessages(ctx, darkfirestore.GetHistoricalMessagesParams{
			Offset:      offset,
			Limit:       body.PerPage,
			ChannelUUID: channelUUID,
		})

	if err != nil {
		c.AbortWithError(
			http.StatusBadRequest,
			apperr.NewErr(
				apperr.FailedGetHistoricalMessagesFromFireStore,
				err.Error(),
			),
		)

		return
	}

	c.JSON(http.StatusOK, NewTransformer().TransformGetHistoricalMessages(msgs))
}

type EmitServiceUpdateMessageBody struct {
	ServiceUuid     string    `json:"service_uuid" form:"service_uuid" binding:"required"`
	ServiceType     string    `json:"service_type" form:"service_type" binding:"required"`
	Price           float64   `json:"price" form:"price" binding:"required"`
	AppointmentTime time.Time `json:"appointment_time" form:"appointment_time" binding:"required"`
	Duration        int       `json:"duration" form:"duration" binding:"required"`
	Address         string    `json:"address" form:"address" binding:"required,gt=0"`
}

func EmitServiceUpdateMessage(c *gin.Context, depCon container.Container) {
	body := EmitServiceUpdateMessageBody{}

	if err := requestbinder.Bind(c, &body); err != nil {
		c.AbortWithError(
			http.StatusBadRequest,
			apperr.NewErr(
				apperr.FailedToValidateRequestBody,
				err.Error(),
			),
		)

		return
	}

	var (
		iqDao      contracts.InquiryDAOer
		srvDao     contracts.ServiceDAOer
		chatDao    contracts.ChatDaoer
		coinPkgDao contracts.CoinPackageDAOer
		userDao    contracts.UserDAOer
	)

	depCon.Make(&iqDao)
	depCon.Make(&srvDao)
	depCon.Make(&chatDao)
	depCon.Make(&coinPkgDao)
	depCon.Make(&userDao)

	sender, err := userDao.GetUserByUuid(c.GetString("uuid"), "username", "id")

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

	srv, err := srvDao.GetServiceByUuid(body.ServiceUuid)

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToGetServiceByUuid,
				err.Error(),
			),
		)

		return
	}

	// Check if the editor is the service_provider.
	if int64(srv.ServiceProviderID.Int32) != sender.ID {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.ServiceEditorIsNotServiceProvider,
				err.Error(),
			),
		)

		return
	}

	chatroom, err := chatDao.GetChatroomByServiceId(int(srv.ID))

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToGetChatroomByServiceId,
				err.Error(),
			),
		)

		return
	}

	// Get inquiry by service uuid
	iq, err := srvDao.GetInquiryByServiceUuid(body.ServiceUuid)

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToGetInquiryByServiceUuid,
				err.Error(),
			),
		)

		return

	}

	matchingFee, err := coinPkgDao.GetMatchingFee()

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToGetMatchingFee,
				err.Error(),
			),
		)
		return
	}

	// Update service detail by service ID.
	txResp := db.TransactWithFormatStruct(db.GetDB(), func(tx *sqlx.Tx) db.FormatResp {
		srvDao.WithTx(tx)

		usrv, err := srvDao.UpdateServiceByID(contracts.UpdateServiceByIDParams{
			ID:            srv.ID,
			Price:         &body.Price,
			ServiceStatus: &srv.ServiceStatus,
			ServiceType:   (*models.ServiceType)(&body.ServiceType),
			Appointment:   &body.AppointmentTime,
			Duration:      &body.Duration,
			Address:       &body.Address,
		})

		if err != nil {
			return db.FormatResp{
				HttpStatusCode: http.StatusInternalServerError,
				ErrCode:        apperr.FailedToUpdateService,
				Err:            err,
			}
		}

		if err := iqDao.PatchInquiryStatusByUUID(
			contracts.PatchInquiryStatusByUUIDParams{
				InquiryStatus: models.InquiryStatusWaitForInquirerApprove,
				UUID:          iq.Uuid,
			},
		); err != nil {
			return db.FormatResp{
				HttpStatusCode: http.StatusInternalServerError,
				ErrCode:        apperr.FailedToPatchInquiryStatus,
				Err:            err,
			}
		}

		return db.FormatResp{
			Response: usrv,
		}
	})

	if txResp.Err != nil {
		c.AbortWithError(
			txResp.HttpStatusCode,
			apperr.NewErr(
				txResp.ErrCode,
				txResp.Err.Error(),
			),
		)

		return
	}

	df := darkfirestore.Get()
	ctx := context.Background()

	msg, err := df.UpdateInquiryDetail(
		ctx,
		darkfirestore.UpdateInquiryDetailParams{
			InquiryUuid: iq.Uuid,
			ChannelUuid: chatroom.ChannelUuid.String,
			Status:      models.InquiryStatusWaitForInquirerApprove,
			Data: darkfirestore.InquiryDetailMessage{
				ChatMessage: darkfirestore.ChatMessage{
					Content:  "",
					From:     c.GetString("uuid"),
					Username: sender.Username,
				},
				Price:           body.Price,
				Duration:        body.Duration,
				AppointmentTime: body.AppointmentTime.UnixNano() / int64(time.Microsecond),
				ServiceType:     body.ServiceType,
				Address:         body.Address,
				MatchingFee:     int(matchingFee.Cost.Int32),
				InquiryUuid:     iq.Uuid,
			},
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

	c.JSON(http.StatusOK, msg)
}

type EmitInquiryUpdateMessage struct {
	Price           float64   `json:"price" form:"price" binding:"required"`
	ChannelUUID     string    `json:"channel_uuid" form:"channel_uuid" binding:"required"`
	AppointmentTime time.Time `json:"appointment_time" form:"appointment_time" binding:"required"`
	Duration        int       `json:"duration" form:"duration" binding:"required"`
	ServiceType     string    `json:"service_type" form:"service_type" binding:"required"`
	Address         string    `json:"address" form:"address" binding:"required,gt=0"`
}

// EmitInquiryUpdatedMessage emits inquiry updated message to the chatroom.
// This message notifies the male user to confirm the inquiry detail by clicking
// on the message bubble.
// @TODO extract firestore message emission away from transaction.
func EmitInquiryUpdatedMessage(c *gin.Context, depCon container.Container) {
	body := EmitInquiryUpdateMessage{}

	if err := requestbinder.Bind(c, &body); err != nil {
		c.AbortWithError(
			http.StatusBadRequest,
			apperr.NewErr(
				apperr.FailedToValidateRequestBody,
				err.Error(),
			),
		)

		return
	}

	ctx := context.Background()
	df := darkfirestore.Get()

	// - Send update inquiry message
	// - Change inquiry status in firestore
	// - Update inquiry status in DB
	var (
		iqDao      contracts.InquiryDAOer
		chatDao    contracts.ChatDaoer
		coinPkgDao contracts.CoinPackageDAOer
	)

	depCon.Make(&iqDao)
	depCon.Make(&chatDao)
	depCon.Make(&coinPkgDao)

	iq, err := chatDao.GetInquiryByChannelUuid(body.ChannelUUID)

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToGetInquiryByChannelUuid,
				err.Error(),
			),
		)

		return
	}

	matchingFee, err := coinPkgDao.GetMatchingFee()

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToGetMatchingFee,
				err.Error(),
			),
		)
		return
	}

	if err := iqDao.PatchInquiryStatusByUUID(
		contracts.PatchInquiryStatusByUUIDParams{
			InquiryStatus: models.InquiryStatusWaitForInquirerApprove,
			UUID:          iq.Uuid,
		},
	); err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToPatchInquiryStatus,
				err.Error(),
			),
		)
		return
	}

	var userDao contracts.UserDAOer
	depCon.Make(&userDao)
	sender, err := userDao.GetUserByUuid(c.GetString("uuid"), "username")

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

	msg, err := df.UpdateInquiryDetail(
		ctx,
		darkfirestore.UpdateInquiryDetailParams{
			InquiryUuid: iq.Uuid,
			ChannelUuid: body.ChannelUUID,
			Status:      models.InquiryStatusWaitForInquirerApprove,
			Data: darkfirestore.InquiryDetailMessage{
				ChatMessage: darkfirestore.ChatMessage{
					Content:  "",
					From:     c.GetString("uuid"),
					Username: sender.Username,
				},
				Price:           body.Price,
				Duration:        body.Duration,
				AppointmentTime: body.AppointmentTime.UnixNano() / int64(time.Microsecond),
				ServiceType:     body.ServiceType,
				Address:         body.Address,
				MatchingFee:     int(matchingFee.Cost.Int32),
			},
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

	c.JSON(http.StatusOK, msg)
}

// @TODO the client shouldn't need be needing to provide channel uuid. We should get channel uuid by inquiry uuid.
type EmitServiceConfirmedMessageBody struct {
	InquiryUUID string `json:"inquiry_uuid" form:"inquiry_uuid" binding:"required"`
}

func EmitServiceConfirmedMessage(c *gin.Context, depCon container.Container) {
	ctx := context.Background()
	body := EmitServiceConfirmedMessageBody{}

	if err := requestbinder.Bind(c, &body); err != nil {
		c.AbortWithError(
			http.StatusBadRequest,
			apperr.NewErr(
				apperr.FailedToValidateEmitConfirmedServiceParams,
				err.Error(),
			),
		)

		return
	}

	var (
		userDao     contracts.UserDAOer
		serviceDao  contracts.ServiceDAOer
		inquiryDao  contracts.InquiryDAOer
		chatDao     contracts.ChatDaoer
		gcsEnhancer gcsenhancer.GCSEnhancerInterface
	)

	depCon.Make(&userDao)
	depCon.Make(&serviceDao)
	depCon.Make(&inquiryDao)
	depCon.Make(&chatDao)
	depCon.Make(&gcsEnhancer)

	// Get user by uuid.
	sender, err := userDao.GetUserByUuid(c.GetString("uuid"), "username", "id")

	// Retrieve inquiry by inquiry uuid.
	iqRes, err := inquiryDao.GetInquiryByUuid(body.InquiryUUID)

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

	// Are there any existing services has overlapped in time interval for the female user?
	olSrv, err := serviceDao.GetOverlappedServices(contracts.GetOverlappedServicesParams{
		UserId:                 sender.ID,
		InquiryAppointmentTime: iqRes.AppointmentTime.Time,
	})

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToGetOverlappedServices,
				err.Error(),
			),
		)

		return
	}

	if len(olSrv) > 0 {
		c.AbortWithError(
			http.StatusBadRequest,
			apperr.NewErr(apperr.OverlappingService),
		)

		return
	}

	// Retrieve chatroom by inquiry id.
	chatroom, err := chatDao.GetChatRoomByInquiryID(iqRes.ID, "channel_uuid")

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToGetChatRoomByInquiryID,
				err.Error(),
			),
		)

		return
	}

	// Change inquiry status from `chatting` to `booked` and create a new service with status `unpaid`
	transResp := db.TransactWithFormatStruct(db.GetDB(), func(tx *sqlx.Tx) db.FormatResp {

		statusUnpaid := models.ServiceStatusUnpaid

		if err != nil {
			return db.FormatResp{
				Err:            err,
				ErrCode:        apperr.FailedToGenerateShortId,
				HttpStatusCode: http.StatusInternalServerError,
			}
		}

		srvModel, err := serviceDao.UpdateServiceByInquiryId(
			contracts.UpdateServiceByInquiryIdParams{
				InquiryId:     int64(iqRes.ID),
				ServiceStatus: &statusUnpaid,
			},
		)

		if err != nil {
			return db.FormatResp{
				HttpStatusCode: http.StatusInternalServerError,
				Err:            err,
				ErrCode:        apperr.FailedToUpdateService,
			}
		}

		// Update service status from `negotiating` to unpaid
		df := darkfirestore.Get()
		if err := df.UpdateService(ctx, darkfirestore.UpdateServiceParams{
			ServiceUuid:   srvModel.Uuid.String,
			ServiceStatus: statusUnpaid.ToString(),
		}); err != nil {
			return db.FormatResp{
				Err:            err,
				ErrCode:        apperr.FirestoreFailedToUpdateService,
				HttpStatusCode: http.StatusInternalServerError,
			}
		}

		err = inquiryDao.WithTx(tx).PatchInquiryStatusByUUID(contracts.PatchInquiryStatusByUUIDParams{
			InquiryStatus: models.InquiryStatusBooked,
			UUID:          body.InquiryUUID,
		})

		if err != nil {
			return db.FormatResp{
				Err:            err,
				ErrCode:        apperr.FailedToPatchInquiryStatus,
				HttpStatusCode: http.StatusInternalServerError,
			}
		}

		return db.FormatResp{
			Response: &srvModel,
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

	// We need to create QR code file. Upload it to cloud storage and store qrcode public url.
	// Encode the following info:
	//  - qrcode_uuid
	//  - qrcode secret
	srvUuid, err := shortid.Generate()

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToGenQRCodeUuid,
				err.Error(),
			),
		)

		return
	}

	qrcodeContentByte, err := json.Marshal(
		map[string]interface{}{
			"qrcode_secret": config.GetAppConf().ServiceQrCodeSecret,
			"qrcode_uuid":   srvUuid,
		},
	)

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToMarshQRCodeInfo,
				err.Error(),
			),
		)

		return
	}

	qrcodePngByte, err := qrcode.Encode(
		string(qrcodeContentByte),
		qrcode.Medium,
		256,
	)

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToMarshQRCodeContent,
				err.Error(),
			),
		)

		return
	}

	qrcodeUrl, err := gcsEnhancer.Upload(
		ctx,
		bytes.NewReader(qrcodePngByte),
		fmt.Sprintf("s_qrcode_%s.png", srvUuid),
	)

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToUploadQRCode,
				err.Error(),
			),
		)

		return
	}

	service := transResp.Response.(*models.Service)

	// Create service qrcode record.
	_, err = serviceDao.CreateServiceQRCode(contracts.CreateServiceQRCodeParams{
		Uuid:      srvUuid,
		Url:       qrcodeUrl,
		ServiceId: int(service.ID),
	})

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToCreateServiceQRCodeRecord,
				err.Error(),
			),
		)

		return
	}

	price, err := convertnullsql.ConvertSqlNullStringToFloat32(service.Price)

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToConvertNullSQLStringToFloat,
				err.Error(),
			),
		)

		return
	}

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

	_, msg, err := darkfirestore.Get().SendServiceConfirmedMessage(
		ctx,
		darkfirestore.SendServiceConfirmedMessageParams{
			ChannelUUID: chatroom.ChannelUuid.String,
			Data: darkfirestore.ServiceDetailMessage{
				ChatMessage: darkfirestore.ChatMessage{
					Content:   "",
					From:      c.GetString("uuid"),
					Username:  sender.Username,
					CreatedAt: time.Now(),
				},
				Price:       float64(*price),
				Duration:    int(service.Duration.Int32),
				ServiceUUID: service.Uuid.String,
				ServiceType: service.ServiceType.ToString(),

				// Convert unix nano to unix micro so that the flutter can parse it using flutter DateTime.
				ServiceTime: service.AppointmentTime.Time.UnixNano() / int64(time.Microsecond),
			},
		},
	)

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToSendServiceConfirmedMsg,
				err.Error(),
			),
		)

		return
	}

	resp := struct {
		Message            interface{} `json:"message"`
		ChannelID          string      `json:"channel_uuid"`
		ServiceChannelUuid string      `json:"service_channel_uuid"`
		QrCodeUrl          string      `json:"qrcode_url"`
	}{
		Message:            msg,
		ServiceChannelUuid: service.Uuid.String,
		ChannelID:          chatroom.ChannelUuid.String,
		QrCodeUrl:          qrcodeUrl,
	}

	c.JSON(http.StatusOK, resp)
}

type QuitChatroomBody struct {
	ChannelUuid string `json:"channel_uuid" form:"channel_uuid" binding:"required,gt=0"`
}

// QuitChatroomHandler either party can choose to euit the chatroom.
// Both parties will notified. The inquiry status will be changed to
// "inquiring".
func QuitChatroomHandler(c *gin.Context, depCon container.Container) {
	body := QuitChatroomBody{}

	if err := requestbinder.Bind(c, &body); err != nil {
		c.AbortWithError(
			http.StatusBadRequest,
			apperr.NewErr(
				apperr.FailedToBindBodyParams,
				err.Error(),
			),
		)

		return

	}

	// Retrieve inqiury by channel uuid.
	chatDao := NewChatDao(db.GetDB())
	iq, err := chatDao.GetInquiryByChannelUuid(body.ChannelUuid)

	if err != nil {
		c.AbortWithError(
			http.StatusBadRequest,
			apperr.NewErr(
				apperr.FailedToGetInquiryByChannelUuid,
			),
		)

		return
	}

	// Check if user is in the chatroom.
	exists, err := chatDao.IsUserInChatroom(
		c.GetString("uuid"),
		body.ChannelUuid,
	)

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToCheckIsUserInChatroom,
				err.Error(),
			),
		)
	}

	if !exists {
		c.AbortWithError(
			http.StatusBadRequest,
			apperr.NewErr(
				apperr.UserIsNotInTheChatroom,
			),
		)

		return
	}

	chatroom, err := chatDao.GetChatRoomByChannelUUID(
		body.ChannelUuid,
		"id",
		"channel_uuid",
	)

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToGetChatRoomByChannelUuid,
				err.Error(),
			),
		)

		return
	}

	// Both inquirer and picker leave the chatroom.
	type TransResult struct {
		RemovedUsers []models.User
		Inquiry      *models.ServiceInquiry
	}

	transResp := db.TransactWithFormatStruct(
		db.GetDB(),
		func(tx *sqlx.Tx) db.FormatResp {
			removedUsers, err := chatDao.WithTx(tx).LeaveAllMemebers(chatroom.ID)

			if err != nil {
				return db.FormatResp{
					Err:     err,
					ErrCode: apperr.FailedToLeaveAllMembers,
				}
			}

			// Soft delete chatroom
			if err := chatDao.
				WithTx(tx).
				DeleteChatRoom(chatroom.ID); err != nil {
				return db.FormatResp{
					Err:            err,
					ErrCode:        apperr.FailedToDeleteChat,
					HttpStatusCode: http.StatusBadRequest,
				}
			}

			// Change inquiry status to `inquiring`
			var iqDao contracts.InquiryDAOer
			depCon.Make(&iqDao)

			newStatus := models.InquiryStatusInquiring
			iq, err := iqDao.PatchInquiryByInquiryUUID(contracts.PatchInquiryParams{
				Uuid:          iq.Uuid,
				InquiryStatus: &newStatus,
			})

			if err != nil {
				return db.FormatResp{
					Err:     err,
					ErrCode: apperr.FailedToPatchInquiryStatus,
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
			http.StatusInternalServerError,
			apperr.NewErr(
				transResp.ErrCode,
				transResp.Err.Error(),
			),
		)

		return
	}

	df := darkfirestore.Get()
	ctx := context.Background()

	// Emit quit chatroom messsage to firestore `inquiring` so that the other
	// party knows it's time to quit the chatroom.
	var userDao contracts.UserDAOer
	depCon.Make(&userDao)
	sender, err := userDao.GetUserByUuid(c.GetString("uuid"), "username")

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

	_, err = df.QuitChatroom(
		ctx,
		darkfirestore.QuitChatroomMessageParams{
			ChannelUuid: chatroom.ChannelUuid.String,
			InquiryUuid: iq.Uuid,
			Data: darkfirestore.ChatMessage{
				Content:  "",
				From:     c.GetString("uuid"),
				Username: sender.Username,
			},
		},
	)

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToSendQuitChatroomMsg,
				err.Error(),
			),
		)
		return
	}

	transResult := transResp.Response.(*TransResult)

	c.JSON(http.StatusOK, TransformRevertChatting(
		transResult.RemovedUsers,
		*transResult.Inquiry,
		*chatroom,
	))
}

// Emit disapprove message when male user disagree with the inquiry detail set by the female user
type EmitDisagreeInquiry struct {
	ChannelUuid string `json:"channel_uuid" form:"channel_uuid" binding:"required,gt=0"`
}

func EmitDisagreeInquiryHandler(c *gin.Context, depCon container.Container) {
	body := EmitDisagreeInquiry{}

	if err := requestbinder.Bind(c, &body); err != nil {
		c.AbortWithError(
			http.StatusBadRequest,
			apperr.NewErr(
				apperr.FailedToBindBodyParams,
				err.Error(),
			),
		)

		return
	}

	// - Change inquiry status in DB
	// - Change inquiry status in firestore
	// - Emit disapprove message
	var (
		iqDao   contracts.InquiryDAOer
		userDao contracts.UserDAOer
	)

	depCon.Make(&iqDao)
	depCon.Make(&userDao)

	sender, err := userDao.GetUserByUuid(c.GetString("uuid"), "username")

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

	iq, err := iqDao.GetInquiryByChannelUuid(body.ChannelUuid)

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToGetInquiryByChannelUuid,
				err.Error(),
			),
		)

		return
	}

	if err := iqDao.PatchInquiryStatusByUUID(contracts.PatchInquiryStatusByUUIDParams{
		UUID:          iq.Uuid,
		InquiryStatus: models.InquiryStatusChatting,
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

	df := darkfirestore.Get()
	ctx := context.Background()
	msg, err := df.DisagreeInquiry(
		ctx,
		darkfirestore.DisagreeInquiryParams{
			InquiryUuid: iq.Uuid,
			ChannelUuid: body.ChannelUuid,
			Data: darkfirestore.ChatMessage{
				Content:   "",
				From:      c.GetString("uuid"),
				Username:  sender.Username,
				CreatedAt: time.Now(),
			},
		},
	)

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToSendTextMessage,
				err.Error(),
			),
		)

		return
	}

	c.JSON(http.StatusOK, msg)
}

type EmitImageMessageBody struct {
	ImageUrl    string `json:"image_url" form:"image_url" binding:"required,gt=0"`
	ChannelUuid string `json:"channel_uuid" form:"channel_uuid" binding:"required,gt=0"`
}

func EmitImageMessage(c *gin.Context, depCon container.Container) {
	body := EmitImageMessageBody{}

	if err := requestbinder.Bind(c, &body); err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToBindApiBodyParams,
				err.Error(),
			),
		)

		return
	}

	ctx := context.Background()
	df := darkfirestore.Get()
	err := df.SendImageMessage(ctx, darkfirestore.SendImageMessageParams{
		ChannelUuid: body.ChannelUuid,
		ImageUrls: []string{
			body.ImageUrl,
		},
		From: c.GetString("uuid"),
	})

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToSendImageMessage,
				err.Error(),
			),
		)

		return
	}

	c.JSON(http.StatusOK, struct{}{})
}
