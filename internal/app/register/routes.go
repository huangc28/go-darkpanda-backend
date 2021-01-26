package register

import (
	"github.com/gin-gonic/gin"
	"github.com/golobby/container/pkg/container"
)

func Routes(r *gin.RouterGroup, depCon container.Container) {
	g := r.Group("/register")

	g.POST("", HandleRegister)

	g.POST("/send-verify-code", func(c *gin.Context) {
		HandleSendVerifyCode(c, depCon)
	})

	// Verify username.
	g.POST("/verify-username", func(c *gin.Context) {
		HandleVerifyUsername(c, depCon)
	})
}
