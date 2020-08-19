package auth

import (
	"github.com/gin-gonic/gin"
)

// - /v1/register
// - /v1/send-verify-code
// - /v1/verify-phone
// - /v1/logout
func Routes(r *gin.RouterGroup) {
	r.POST("/register", RegisterHandler)
	r.POST("/send-verify-code", SendVerifyCodeHandler)
	r.POST("/verify-phone", VerifyPhoneHandler)
	r.POST("/revoke-jwt", RevokeJwtHandler)
}
