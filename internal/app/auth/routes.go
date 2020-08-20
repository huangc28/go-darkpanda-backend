package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/huangc28/go-darkpanda-backend/config"
	"github.com/huangc28/go-darkpanda-backend/internal/app/pkg/jwtactor"
)

func Routes(r *gin.RouterGroup) {
	r.POST("/register", RegisterHandler)
	r.POST("/send-verify-code", SendVerifyCodeHandler)
	r.POST("/verify-phone", VerifyPhoneHandler)
	r.POST(
		"/revoke-jwt",
		jwtactor.JwtValidator(jwtactor.JwtMiddlewareOptions{
			Secret: config.GetAppConf().JwtSecret,
		}),
		RevokeJwtHandler,
	)
}
