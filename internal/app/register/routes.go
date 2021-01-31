package register

import (
	"github.com/gin-gonic/gin"
	"github.com/golobby/container/pkg/container"
)

func Routes(r *gin.RouterGroup, depCon container.Container) {
	g := r.Group("/register")

	g.POST("", RegisterHandler)

	// This API receives user uuid and mobile number to send SMS verify code.
	// This API is used in pair with `/register` API letting newly registered user
	// to verify their SMS.
	//g.POST("/send-verify-code", func(c *gin.Context) {
	//SendVerifyCodeHandler(c, depCon)
	//})

	// Check username availability.
	g.POST("/verify-username", func(c *gin.Context) {
		HandleVerifyUsername(c, depCon)
	})

	// Check referral code availabiity.

}
