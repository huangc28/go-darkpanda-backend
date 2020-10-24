package chat

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/huangc28/go-darkpanda-backend/config"
	"github.com/huangc28/go-darkpanda-backend/internal/app/apperr"
	"github.com/huangc28/go-darkpanda-backend/internal/app/contracts"
	"github.com/huangc28/go-darkpanda-backend/internal/app/pkg/darkpubnub"
	"github.com/huangc28/go-darkpanda-backend/internal/app/pkg/requestbinder"
)

type ChatHandlers struct {
	ChatDao contracts.ChatDaoer
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
