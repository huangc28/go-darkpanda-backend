package chat

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/huangc28/go-darkpanda-backend/config"
	"github.com/huangc28/go-darkpanda-backend/internal/app/apperr"
	"github.com/huangc28/go-darkpanda-backend/internal/app/contracts"
	"github.com/huangc28/go-darkpanda-backend/internal/app/models"
	"github.com/huangc28/go-darkpanda-backend/internal/app/pkg/darkpubnub"
	"github.com/huangc28/go-darkpanda-backend/internal/app/pkg/requestbinder"
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

	c.JSON(http.StatusOK, NewTransformer().TransformInquiryChats(chatrooms))
}
