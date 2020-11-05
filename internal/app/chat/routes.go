package chat

import (
	"github.com/gin-gonic/gin"
	"github.com/huangc28/go-darkpanda-backend/config"
	"github.com/huangc28/go-darkpanda-backend/db"
	"github.com/huangc28/go-darkpanda-backend/internal/app/contracts"
	"github.com/huangc28/go-darkpanda-backend/internal/app/deps"
	"github.com/huangc28/go-darkpanda-backend/internal/app/pkg/jwtactor"
)

func Routes(r *gin.RouterGroup) {
	g := r.Group(
		"/chat",
		jwtactor.JwtValidator(jwtactor.JwtMiddlewareOptions{
			Secret: config.GetAppConf().JwtSecret,
		}),
	)

	var userDao contracts.UserDAOer
	deps.Get().Container.Make(&userDao)

	handlers := ChatHandlers{
		ChatDao: &ChatDao{
			db.GetDB(),
		},
		UserDao: userDao,
	}

	g.GET("/:channel_uuid", func(c *gin.Context) {
		switch c.Param(":channel_uuid") {

		// Get list of inquiry chatrooms
		case "inquiry-chatrooms":
			handlers.GetInquiryChatRooms(c)

		// Fetch message from inquiry chat
		default:
			handlers.GetHistoricalMessages(c)
		}
	})

	g.POST("/emit-text-message", handlers.EmitTextMessage)
}
