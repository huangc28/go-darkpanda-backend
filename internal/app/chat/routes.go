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
	var (
		userDao contracts.UserDAOer
		authDao contracts.AuthDaoer
	)

	depCon.Make(&userDao)
	depCon.Make(&authDao)

	g := r.Group(
		"/chat",
		jwtactor.JwtValidator(jwtactor.JwtMiddlewareOptions{
			Secret: config.GetAppConf().JwtSecret,
		}, authDao),
	)

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
		"/emit-inquiry-updated-message",
		func(c *gin.Context) {
			EmitInquiryUpdatedMessage(c, depCon)
		},
	)

	// The female user edited service details and saved the service settings, the chatroom would emit a service setting message.
	// Male user would be notified with the service message.  Male user sees the popup contains service detail set by the female user.
	g.POST(
		"/emit-service-message",
		func(c *gin.Context) {
			EmitServiceSettingMessageHandler(c, depCon)
		},
	)

	// Male user can agree on service detail set by female user. Once agreed, female user would receive
	// a message saying that the service has been established, the chatroom will be forwarded
	g.POST(
		"/emit-service-confirmed-message",
		middlewares.IsMale(userDao),
		func(c *gin.Context) {
			EmitServiceConfirmedMessage(c, depCon)
		},
	)

	// Get historical messages of a specific chatroom.
	g.GET(
		"/:channel_uuid/messages",
		GetHistoricalMessages,
	)

	// If male user disagree with the inquiry detail set by the female user in the inquiry chatroom.
	g.POST(
		"/disagree",
		func(c *gin.Context) {
			EmitDisagreeInquiryHandler(c, depCon)
		},
	)

	// If either user leaves the chat, we should perform soft delete on both the user and the chatroom.
	// Moreover, notify both user in the firestore that the other party has left.
	g.POST(
		"/quit-chatroom",
		func(c *gin.Context) {
			QuitChatroomHandler(c, depCon)
		},
	)
}
