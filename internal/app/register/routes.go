package register

import (
	"github.com/gin-gonic/gin"
	"github.com/golobby/container/pkg/container"
)

func Routes(r *gin.RouterGroup, depCon container.Container) {
	g := r.Group("/register")

	g.POST("", func(c *gin.Context) {
		RegisterHandler(c, depCon)
	})

	// Check username availability.
	g.POST("/verify-username", func(c *gin.Context) {
		VerifyUsernameHandler(c, depCon)
	})

	g.POST("/send-mobile-verify-code", func(c *gin.Context) {
		SendMobileVerifyCodeHandler(c, depCon)
	})

	g.POST("/verify-mobile", func(c *gin.Context) {
		VerifyMobileHandler(c, depCon)
	})
}
