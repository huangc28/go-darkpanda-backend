package register

import (
	"github.com/gin-gonic/gin"
	"github.com/golobby/container/pkg/container"
)

func Routes(r *gin.RouterGroup, depCon container.Container) {
	g := r.Group("/register")

	g.POST("", RegisterHandler)

	// Check username availability.
	g.POST("/verify-username", func(c *gin.Context) {
		VerifyUsernameHandler(c, depCon)
	})

	// Check referral code availabiity.
	g.POST("/verify-referral-code", func(c *gin.Context) {
		VerifyReferralCodeHandler(c, depCon)
	})

	g.POST("/send-mobile-verify-code", func(c *gin.Context) {
		SendMobileVerifyCodeHandler(c, depCon)
	})

	g.POST("/verify-mobile", func(c *gin.Context) {
		VerifyMobileHandler(c, depCon)
	})
}
