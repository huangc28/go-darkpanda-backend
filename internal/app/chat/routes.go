package chat

import (
	"github.com/gin-gonic/gin"
	"github.com/huangc28/go-darkpanda-backend/config"
	"github.com/huangc28/go-darkpanda-backend/db"
	"github.com/huangc28/go-darkpanda-backend/internal/app/pkg/jwtactor"
)

func Routes(r *gin.RouterGroup) {
	g := r.Group("/chat", jwtactor.JwtValidator(jwtactor.JwtMiddlewareOptions{
		Secret: config.GetAppConf().JwtSecret,
	}))

	handlers := ChatHandlers{
		ChatDao: &ChatDao{
			db.GetDB(),
		},
	}

	g.POST("/emit-text-message", handlers.EmitTextMessage)
}
