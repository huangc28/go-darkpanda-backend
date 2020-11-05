package chat

import (
	"context"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin"
	"github.com/huangc28/go-darkpanda-backend/config"
	"github.com/huangc28/go-darkpanda-backend/internal/app/apperr"
	"github.com/huangc28/go-darkpanda-backend/internal/app/contracts"
	"github.com/huangc28/go-darkpanda-backend/internal/app/models"
	darkfirestore "github.com/huangc28/go-darkpanda-backend/internal/app/pkg/dark_firestore"
	"github.com/huangc28/go-darkpanda-backend/internal/app/pkg/darkpubnub"
	"github.com/huangc28/go-darkpanda-backend/internal/app/pkg/requestbinder"
	"github.com/huangc28/go-darkpanda-backend/internal/app/util"
)

type ChatHandlers struct {
	ChatDao contracts.ChatDaoer
	UserDao contracts.UserDAOer
}

type EmitTextMessageBody struct {
	Content   string `form:"content" binding:"required"`
	ChannelID string `form:"channel_id" binding:"required"`
}

// @TODOs
//   - Check if chatroom has expired
//   - Check if message count is still within valid range
func (h *ChatHandlers) EmitTextMessage(c *gin.Context) {
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

	// check if chatroom has expired
	channel, err := h.ChatDao.GetChatRoomByChannelID(
		body.ChannelID,
		"expired_at",
		"message_count",
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

	if IsChatroomExpired(channel.ExpiredAt) {
		c.AbortWithError(
			http.StatusBadRequest,
			apperr.NewErr(apperr.MessageExceedMaximumCount),
		)

		return
	}

	if IsExceedMaxMessageCount(int(channel.MessageCount.Int32)) {
		c.AbortWithError(
			http.StatusBadRequest,
			apperr.NewErr(apperr.MessageExceedMaximumCount),
		)

		return
	}

	pubnubConf := config.GetAppConf().PubnubCredentials
	uuid := c.GetString("uuid")

	dp := darkpubnub.NewDarkPubNub(
		darkpubnub.Config{
			PublishKey:   pubnubConf.PublishKey,
			SubscribeKey: pubnubConf.SubscribeKey,
			SecretKey:    pubnubConf.SecretKey,
			UUID:         uuid,
		},
	)

	sentTime, err := dp.SendTextMessage(body.ChannelID, darkpubnub.TextMessage{
		Content: body.Content,
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
	c.JSON(http.StatusOK, NewTransformer().TransformEmitTextMessage(
		TransformEmitTextMessageParams{
			Timestamp: sentTime,
			Content:   body.Content,
		},
	))

}

// If the requester is female find all chatrooms that qualify the following conditions:
//   - Those chatrooms's related inquiry status is chatting
//   - Those chatrooms's related inquiry picker_id equals requester's id
func (h *ChatHandlers) GetInquiryChatRooms(c *gin.Context) {
	// Recognize the gender of the requester
	userUUID := c.GetString("uuid")
	user, err := h.UserDao.GetUserByUuid(userUUID, "id", "gender")

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
		chatrooms, err = h.ChatDao.GetFemaleInquiryChatRooms(user.ID)

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

	c.JSON(http.StatusOK, NewTransformer().TransformInquiryChats(chatrooms, channelUUIDMessageMap))
}

// Add pagination for firestore. We have to get `page` and `limit` from client.
//   - page
//   - perpage
// Calculate offset from page number and perpage .
type GetMessagesBody struct {
	PerPage int `form:"perpage,default=10"`
	Page    int `form:"page,default=0"`
}

func (h *ChatHandlers) GetHistoricalMessages(c *gin.Context) {
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
