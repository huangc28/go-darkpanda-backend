package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/golobby/container/pkg/container"
	"github.com/huangc28/go-darkpanda-backend/config"
	"github.com/huangc28/go-darkpanda-backend/internal/app/pkg/jwtactor"
)

func Routes(r *gin.RouterGroup, depCon container.Container) {
	authController := AuthController{
		Container: depCon,
	}

	g := r.Group("/auth")

	// Send verify code to user that attempts to login to the application
	// authController.SendVerifyCodeHandler
	g.POST("/send-verify-code", authController.SendVerifyCodeHandler)

	// Client attempt to verify login code he / she received via SMS from `/send-verify-code`. If the
	// If code is verified, grants user auth permission.
	g.POST("/verify-code", authController.VerifyLoginCode)

	g.POST(
		"/revoke-jwt",
		jwtactor.JwtValidator(jwtactor.JwtMiddlewareOptions{
			Secret: config.GetAppConf().JwtSecret,
		}),
		authController.RevokeJwtHandler,
	)
}
