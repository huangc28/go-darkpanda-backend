package chat

import (
	"github.com/gin-gonic/gin"
	"github.com/golobby/container/pkg/container"
	"github.com/huangc28/go-darkpanda-backend/config"
	"github.com/huangc28/go-darkpanda-backend/internal/app/contracts"
	"github.com/huangc28/go-darkpanda-backend/internal/app/middlewares"
	"github.com/huangc28/go-darkpanda-backend/internal/app/pkg/jwtactor"
)

type ChatRoutesParams struct {
	UserDao    contracts.UserDAOer
	ServiceDao contracts.ServiceDAOer
	InquiryDao contracts.InquiryDAOer
}

func Routes(r *gin.RouterGroup, depCon container.Container) {
	g := r.Group(
		"/chat",
		jwtactor.JwtValidator(jwtactor.JwtMiddlewareOptions{
			Secret: config.GetAppConf().JwtSecret,
		}),
	)

	var userDao contracts.UserDAOer

	depCon.Make(&userDao)

	g.GET("", func(c *gin.Context) {
		GetChatrooms(c, depCon)
	})

	g.POST(
		"/emit-text-message",
		func(c *gin.Context) {
			EmitTextMessage(c, depCon)
		},
	)

	g.POST(
		"/emit-service-message",
		func(c *gin.Context) {
			EmitServiceSettingMessageHandler(c, depCon)
		},
	)

	// Male user can agree on service detail set by female user. Once agreed, female user would receive
	// a message saying that the service has been established, the chatroom should be suspended both party should
	// leave the current inquiry chatroom.
	g.POST(
		"/emit-service-confirmed-message",
		middlewares.IsMale(userDao),
		func(c *gin.Context) {
			EmitServiceConfirmedMessage(c, depCon)
		},
	)

	g.GET(
		"/:channel_uuid/messages",
		GetHistoricalMessages,
	)
}
