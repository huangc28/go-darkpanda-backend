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
	g := r.Group("/chat", jwtactor.JwtValidator(jwtactor.JwtMiddlewareOptions{
		Secret: config.GetAppConf().JwtSecret,
	}))

	var userDao contracts.UserDAOer
	deps.Get().Container.Make(&userDao)

	handlers := ChatHandlers{
		ChatDao: &ChatDao{
			db.GetDB(),
		},
		UserDao: userDao,
	}

	// Get list of inquiry chatrooms
	g.GET("/inquiry-chatrooms", handlers.GetInquiryChatRooms)

	// Fetch message from inquiry chat
	g.GET("/:channel_uuid/messages", handlers.GetHistoricalMessages)

	g.POST("/emit-text-message", handlers.EmitTextMessage)
}
