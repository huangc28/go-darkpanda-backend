package register

import (
	"github.com/gin-gonic/gin"
	"github.com/golobby/container/pkg/container"
)

func Routes(r *gin.RouterGroup, depCon container.Container) {
	r.POST("/verify-referral-code", func(c *gin.Context) {
		HandleVerifyReferralCode(c, depCon)
	})
}
