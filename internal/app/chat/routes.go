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

	var (
		userDao    contracts.UserDAOer
		serviceDao contracts.ServiceDAOer
		inquiryDao contracts.InquiryDAOer
	)

	deps.Get().Container.Make(&userDao)
	deps.Get().Container.Make(&serviceDao)
	deps.Get().Container.Make(&inquiryDao)

	handlers := ChatHandlers{
		ChatDao: &ChatDao{
			db.GetDB(),
		},
		UserDao:    userDao,
		ServiceDao: serviceDao,
		InquiryDao: inquiryDao,
	}

	g.GET("", handlers.GetChatrooms)

	g.POST("/emit-text-message", handlers.EmitTextMessage)

	g.POST("/emit-service-message", handlers.EmitServiceSettingMessage)

	g.GET("/:channel_uuid/messages", handlers.GetHistoricalMessages)
}
