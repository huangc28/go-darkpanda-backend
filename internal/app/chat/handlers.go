package chat

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/golobby/container/pkg/container"
	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin"
	"github.com/huangc28/go-darkpanda-backend/db"
	"github.com/huangc28/go-darkpanda-backend/internal/app/apperr"
	"github.com/huangc28/go-darkpanda-backend/internal/app/contracts"
	"github.com/huangc28/go-darkpanda-backend/internal/app/models"
	darkfirestore "github.com/huangc28/go-darkpanda-backend/internal/app/pkg/dark_firestore"
	"github.com/huangc28/go-darkpanda-backend/internal/app/pkg/requestbinder"
	"github.com/huangc28/go-darkpanda-backend/internal/app/util"
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
//   - Check if message count is still within valid range
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

	// check if chatroom has expired
	// channel, err := chatDao.GetChatRoomByChannelUUID(
	// 	body.ChannelUUID,
	// 	"expired_at",
	// 	"message_count",
	// )

	// if err != nil {
	// 	c.AbortWithError(
	// 		http.StatusInternalServerError,
	// 		apperr.NewErr(
	// 			apperr.FailedToGetChatRoomByChannelUuid,
	// 			err.Error(),
	// 		),
	// 	)

	// 	return
	// }

	// if IsChatroomExpired(channel.ExpiredAt) {
	// 	c.AbortWithError(
	// 		http.StatusBadRequest,
	// 		apperr.NewErr(apperr.ChatRoomHasExpired),
	// 	)

	// 	return
	// }

	// if IsExceedMaxMessageCount(int(channel.MessageCount.Int32)) {
	// 	c.AbortWithError(
	// 		http.StatusBadRequest,
	// 		apperr.NewErr(apperr.MessageExceedMaximumCount),
	// 	)

	// 	return
	// }

	// Emit message to firestore.
	ctx := context.Background()
	df := darkfirestore.Get()
	message, err := df.SendTextMessageToChatroom(ctx, darkfirestore.SendTextMessageParams{
		ChatroomName: body.ChannelUUID,
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

	ctx := context.Background()
	if err != nil {
		if err == sql.ErrNoRows {
			// Create a service for that inquiry.
			q := models.New(db.GetDB())
			*service, err = q.CreateService(ctx, models.CreateServiceParams{
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
				ServiceStatus: models.ServiceStatusNegotiating,
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
		} else {
			c.AbortWithError(
				http.StatusBadRequest,
				apperr.NewErr(
					apperr.FailedToGetServiceByInquiryUUID,
					err.Error(),
				),
			)

			return
		}
	} else {
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
	}

	// Emit service setting message to chatroom.
	df := darkfirestore.Get()
	message, err := df.SendServiceDetailMessageToChatroom(ctx, darkfirestore.SendServiceDetailMessageParams{
		ChatroomName: body.ChannelUUID,
		Data: darkfirestore.ServiceDetailMessage{
			ChatMessage: darkfirestore.ChatMessage{
				Content: "",
				From:    c.GetString("uuid"),
			},
			Price:       body.Price,
			Duration:    int(service.Duration.Int32),
			ServiceTime: service.AppointmentTime.Time.UnixNano() / 1000,
			ServiceType: body.ServiceType,
			ServiceUUID: service.Uuid.String(),
		},
	})

	c.JSON(http.StatusOK, message)
}

// If the requester is female find all chatrooms that qualify the following conditions:
//   - Those chatrooms's related inquiry status is chatting
//   - Those chatrooms's related inquiry picker_id equals requester's id
func GetInquiryChatRooms(c *gin.Context, depCon container.Container) {
	// Recognize the gender of the requester
	var (
		userDao contracts.UserDAOer
		chatDao contracts.ChatDaoer
	)

	depCon.Make(&userDao)
	depCon.Make(&chatDao)

	userUUID := c.GetString("uuid")
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

	if user.Gender == models.GenderFemale {
		chatrooms, err = chatDao.GetFemaleInquiryChatRooms(user.ID)
	} else {
		// Retrieve inquiry chatrooms for male user.
		log.Println("DEBUG * 3")
	}

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

// GetChatrooms gets list of chatrooms based on chatroom type (service / inquiry). If chatroom type
// is not given in the query params, the default type is inquiry.
type QueryChatroomType string

const (
	Service QueryChatroomType = "service"
	Inquiry QueryChatroomType = "inquiry"
)

type GetChatroomsBody struct {
	ChatroomType QueryChatroomType `form:"chatroom_type,default='inquiry'" json:"chatroom_type,default='inquiry'"`
}

func GetChatrooms(c *gin.Context, depCon container.Container) {
	body := GetChatroomsBody{}

	if err := requestbinder.Bind(c, &body); err != nil {
		c.AbortWithError(
			http.StatusBadRequest,
			apperr.NewErr(
				apperr.FailedToValidateGetChatroomsParams,
				err.Error(),
			),
		)
		return
	}

	switch body.ChatroomType {
	case Inquiry:
		GetInquiryChatRooms(c, depCon)
	case Service:
		c.JSON(http.StatusOK, struct{}{})
	default:
		GetInquiryChatRooms(c, depCon)
	}
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

type EmitInquiryUpdateMessage struct {
	Price           float64   `json:"price" form:"price" binding:"required"`
	ChannelUUID     string    `json:"channel_uuid" form:"channel_uuid" binding:"required"`
	InquiryUUID     string    `json:"inquiry_uuid" form:"inquiry_uuid" binding:"required"`
	AppointmentTime time.Time `json:"appointment_time" form:"appointment_time" binding:"required"`
	Duration        int       `json:"duration" form:"duration" binding:"required"`
	ServiceType     string    `json:"service_type" form:"service_type" binding:"required"`
}

// EmitInquiryUpdatedMessage emits inquiry updated message to the chatroom.
// This message notifies the male user to confirm the inquiry detail by clicking
// on the message bubble.
func EmitInquiryUpdatedMessage(c *gin.Context, container container.Container) {
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
	_, msg, err := df.SendUpdateInquiryMessage(
		ctx,
		darkfirestore.UpdateInquiryMessage{
			ChannelUuid: body.ChannelUUID,
			Data: darkfirestore.InquiryDetailMessage{
				ChatMessage: darkfirestore.ChatMessage{
					Content:   "",
					From:      c.GetString("uuid"),
					CreatedAt: time.Now(),
				},
				Price:           body.Price,
				Duration:        body.Duration,
				AppointmentTime: body.AppointmentTime.UnixNano() / 1000,
				ServiceType:     body.ServiceType,
			},
		},
	)

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToSendUpdateInquiryMessage,
				err.Error(),
			),
		)

		return

	}

	c.JSON(http.StatusOK, msg)
}

type EmitServiceConfirmedMessageBody struct {
	Price           float64   `json:"price" form:"price" binding:"required"`
	ChannelUUID     string    `json:"channel_uuid" form:"channel_uuid" binding:"required"`
	InquiryUUID     string    `json:"inquiry_uuid" form:"inquiry_uuid" binding:"required"`
	ServiceTime     time.Time `json:"service_time" form:"service_time" binding:"required"`
	ServiceDuration int       `json:"service_duration" form:"service_duration" binding:"required"`
	ServiceType     string    `json:"service_type" form:"service_type" binding:"required"`
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
		serviceDao contracts.ServiceDAOer
		inquiryDao contracts.InquiryDAOer
	)

	depCon.Make(&serviceDao)
	depCon.Make(&inquiryDao)

	// Retrieve service by inquiry uuid
	service, err := serviceDao.GetServiceByInquiryUUID(
		body.InquiryUUID,
		"services.id",
		"services.uuid",
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

	// Wrap service and inquiry update in a transaction.
	// Change inquiry status from `chatting` to `booked`
	// Change service status from `negotiating` to `unpaid`
	err, _ = db.Transact(db.GetDB(), func(tx *sqlx.Tx) (error, interface{}) {
		serviceDao.WithTx(tx)

		statusUnpaid := models.ServiceStatusUnpaid

		_, err := serviceDao.UpdateServiceByID(contracts.UpdateServiceByIDParams{
			ID:            service.ID,
			ServiceStatus: &statusUnpaid,
		})

		if err != nil {
			return err, nil
		}

		err = inquiryDao.PatchInquiryStatusByUUID(contracts.PatchInquiryStatusByUUIDParams{
			InquiryStatus: models.InquiryStatusBooked,
			UUID:          body.InquiryUUID,
		})

		if err != nil {
			return err, nil
		}

		return nil, nil
	})

	if err != nil {
		c.AbortWithError(
			http.StatusInternalServerError,
			apperr.NewErr(
				apperr.FailedToChangeStatusWhenEmittingServiceConfirmMessage,
				err.Error(),
			),
		)

		return
	}

	docRef, msg, err := darkfirestore.Get().SendServiceConfirmedMessage(
		ctx,
		darkfirestore.SendServiceConfirmedMessageParams{
			ChannelUUID: body.ChannelUUID,
			Data: darkfirestore.ServiceDetailMessage{
				ChatMessage: darkfirestore.ChatMessage{
					Content:   "",
					From:      c.GetString("uuid"),
					CreatedAt: time.Now(),
				},
				Price:       body.Price,
				Duration:    body.ServiceDuration,
				ServiceUUID: service.Uuid.String(),
				ServiceType: body.ServiceType,
				// Convert unix nano to unix micro so that the flutter can parse it using DateTime.
				ServiceTime: body.ServiceTime.UnixNano() / 1000,
			},
		},
	)

	resp := struct {
		Message   interface{} `json:"message"`
		MessageID string      `json:"message_id"`
	}{
		Message:   msg,
		MessageID: docRef.ID,
	}

	c.JSON(http.StatusOK, resp)
}
