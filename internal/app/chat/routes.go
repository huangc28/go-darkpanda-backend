package chat

import (
	"github.com/gin-gonic/gin"
	"github.com/huangc28/go-darkpanda-backend/config"
	"github.com/huangc28/go-darkpanda-backend/db"
	"github.com/huangc28/go-darkpanda-backend/internal/app/contracts"
	"github.com/huangc28/go-darkpanda-backend/internal/app/middlewares"
	"github.com/huangc28/go-darkpanda-backend/internal/app/pkg/jwtactor"
)

type ChatRoutesParams struct {
	UserDao    contracts.UserDAOer
	ServiceDao contracts.ServiceDAOer
	InquiryDao contracts.InquiryDAOer
}

func Routes(r *gin.RouterGroup, params *ChatRoutesParams) {
	g := r.Group(
		"/chat",
		jwtactor.JwtValidator(jwtactor.JwtMiddlewareOptions{
			Secret: config.GetAppConf().JwtSecret,
		}),
	)

	handlers := ChatHandlers{
		ChatDao: &ChatDao{
			db.GetDB(),
		},
		UserDao:    params.UserDao,
		ServiceDao: params.ServiceDao,
		InquiryDao: params.InquiryDao,
	}

	g.GET("", handlers.GetChatrooms)

	g.POST("/emit-text-message", handlers.EmitTextMessage)

	g.POST(
		"/emit-service-message",
		handlers.EmitServiceSettingMessage,
	)

	// Male user can agree on service detail set by female user. Once agreed, female user would receive
	// a message saying that the service has been established, the chatroom should be suspended both party should
	// leave the current inquiry chatroom.
	g.POST(
		"/emit-service-confirmed-message",
		middlewares.IsMale(params.UserDao),
		handlers.EmitServiceConfirmedMessage,
	)

	g.GET("/:channel_uuid/messages", handlers.GetHistoricalMessages)
}
