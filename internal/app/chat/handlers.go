package chat

import (
	"net/http"
	"time"

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

	pubnub := darkpubnub.NewPubnub(darkpubnub.Config{
		PublishKey:   pubnubConf.PublishKey,
		SubscribeKey: pubnubConf.SubscribeKey,
		SecretKey:    pubnubConf.SecretKey,
		UUID:         uuid,
	})

	msg := darkpubnub.FormatTextMessage(darkpubnub.TextMessage{
		Content: body.Content,
	})

	pubResp, _, err := pubnub.Publish().
		Channel(body.ChannelID).
		Message(msg).
		Execute()

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
			Timestamp: time.Unix(darkpubnub.PubnubTimestampToUnix(pubResp.Timestamp), 0),
			Content:   body.Content,
		},
	))
}
