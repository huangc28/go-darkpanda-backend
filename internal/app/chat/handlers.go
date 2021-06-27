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
	log "github.com/sirupsen/logrus"
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
			ServiceUUID: service.Uuid.String,
			MatchingFee: int(matchingFee.Cost.Int32),
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

	msg, err := df.UpdateInquiryStatus(
		ctx,
		darkfirestore.UpdateInquiryStatusParams{
			InquiryUuid: iq.Uuid,
			Status:      models.InquiryStatusWaitForInquirerApprove,
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
		serviceDao  contracts.ServiceDAOer
		inquiryDao  contracts.InquiryDAOer
		chatDao     contracts.ChatDaoer
		gcsEnhancer gcsenhancer.GCSEnhancerInterface
	)

	depCon.Make(&serviceDao)
	depCon.Make(&inquiryDao)
	depCon.Make(&chatDao)
	depCon.Make(&gcsEnhancer)

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
		serviceDBCli := models.New(tx)

		statusUnpaid := models.ServiceStatusUnpaid
		sid, err := shortid.Generate()

		if err != nil {
			return db.FormatResp{
				Err:            err,
				ErrCode:        apperr.FailedToGenerateShortId,
				HttpStatusCode: http.StatusInternalServerError,
			}
		}

		service, err := serviceDBCli.CreateService(
			ctx,
			models.CreateServiceParams{
				Uuid: sql.NullString{
					Valid:  true,
					String: sid,
				},
				CustomerID:        iqRes.InquirerID,
				ServiceProviderID: iqRes.PickerID,
				Price:             iqRes.Price,
				Duration:          iqRes.Duration,
				AppointmentTime:   iqRes.AppointmentTime,
				InquiryID:         int32(iqRes.ID),
				ServiceStatus:     statusUnpaid,
				ServiceType:       iqRes.ServiceType,
				Address:           iqRes.Address,
			},
		)

		if err != nil {
			return db.FormatResp{
				Err:     err,
				ErrCode: apperr.FailedToCreateService,
			}
		}

		// Create a new service record in firestore
		df := darkfirestore.Get()
		err = df.CreateService(
			ctx,
			darkfirestore.CreateServiceParams{
				ServiceUuid:   service.Uuid.String,
				ServiceStatus: service.ServiceStatus.ToString(),
			},
		)

		if err != nil {
			return db.FormatResp{
				Err:            err,
				ErrCode:        apperr.FirestoreFailedToCreateService,
				HttpStatusCode: http.StatusInternalServerError,
			}
		}

		err = inquiryDao.PatchInquiryStatusByUUID(contracts.PatchInquiryStatusByUUIDParams{
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
			Response: &service,
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

	qrcodeEncodeContent := map[string]interface{}{
		"qrcode_secret": config.GetAppConf().ServiceQrCodeSecret,
		"qrcode_uuid":   srvUuid,
	}

	qrcodeContentByte, err := json.Marshal(qrcodeEncodeContent)

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

	_, msg, err := darkfirestore.Get().SendServiceConfirmedMessage(
		ctx,
		darkfirestore.SendServiceConfirmedMessageParams{
			ChannelUUID: chatroom.ChannelUuid.String,
			Data: darkfirestore.ServiceDetailMessage{
				ChatMessage: darkfirestore.ChatMessage{
					Content:   "",
					From:      c.GetString("uuid"),
					CreatedAt: time.Now(),
				},
				Price:       float64(*price),
				Duration:    int(service.Duration.Int32),
				ServiceUUID: service.Uuid.String,
				ServiceType: service.ServiceType.ToString(),
				// Convert unix nano to unix micro so that the flutter can parse it using DateTime.
				ServiceTime: service.AppointmentTime.Time.UnixNano() / 1000,
			},
		},
	)
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

			df := darkfirestore.Get()
			if _, err := df.UpdateInquiryStatus(
				ctx,
				darkfirestore.UpdateInquiryStatusParams{
					InquiryUuid: iq.Uuid,
					Status:      models.InquiryStatusInquiring,
				},
			); err != nil {
				return db.FormatResp{
					HttpStatusCode: http.StatusBadRequest,
					Err:            err,
					ErrCode:        apperr.FailedToChangeFirestoreInquiryStatus,
				}
			}

			// Emit quit chatroom messsage to firestore `inquiring` so that the other
			// party knows it's time to quit the chatroom.
			_, _, err = df.SendQuitChatroomMessage(
				ctx,
				darkfirestore.QuitChatroomMessageParams{
					ChannelUuid: chatroom.ChannelUuid.String,
					Data: darkfirestore.ChatMessage{
						Content: "",
						From:    c.GetString("uuid"),
					},
				},
			)

			if err != nil {
				return db.FormatResp{
					Err:            err,
					ErrCode:        apperr.FailedToSendQuitChatroomMsg,
					HttpStatusCode: http.StatusInternalServerError,
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

	transResult := transResp.Response.(*TransResult)

	c.JSON(http.StatusOK, TransformRevertChatting(
		transResult.RemovedUsers,
		transResult.Inquiry,
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
	var iqDao contracts.InquiryDAOer
	depCon.Make(&iqDao)
	ctx := context.Background()

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

	txResp := db.TransactWithFormatStruct(
		db.GetDB(),
		func(tx *sqlx.Tx) db.FormatResp {
			iqDao.WithTx(tx)
			if err := iqDao.PatchInquiryStatusByUUID(contracts.PatchInquiryStatusByUUIDParams{
				UUID:          iq.Uuid,
				InquiryStatus: models.InquiryStatusChatting,
			}); err != nil {
				return db.FormatResp{
					Err:            err,
					ErrCode:        apperr.FailedToPatchInquiryStatus,
					HttpStatusCode: http.StatusInternalServerError,
				}
			}

			df := darkfirestore.Get()

			if _, err := df.UpdateInquiryStatus(
				ctx,
				darkfirestore.UpdateInquiryStatusParams{
					InquiryUuid: iq.Uuid,
					Status:      models.InquiryStatusChatting,
				},
			); err != nil {
				return db.FormatResp{
					Err:            err,
					ErrCode:        apperr.FailedToChangeFirestoreInquiryStatus,
					HttpStatusCode: http.StatusInternalServerError,
				}
			}

			_, msg, err := df.SendDisagreeInquiryMessage(
				ctx,
				darkfirestore.SendDisagreeInquiryMessageParams{
					ChannelUuid: body.ChannelUuid,
					Data: darkfirestore.ChatMessage{
						Content:   "",
						From:      c.GetString("uuid"),
						CreatedAt: time.Now(),
					},
				},
			)

			if err != nil {
				return db.FormatResp{
					ErrCode:        apperr.FailedToSendTextMessage,
					Err:            err,
					HttpStatusCode: http.StatusInternalServerError,
				}
			}

			return db.FormatResp{
				Response: &msg,
			}
		},
	)

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

	msg := txResp.Response.(*darkfirestore.ChatMessage)

	c.JSON(http.StatusOK, msg)
}
