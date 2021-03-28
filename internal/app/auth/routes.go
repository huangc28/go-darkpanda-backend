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

	g.POST(
		"/send-verify-code",
		func(c *gin.Context) {
			authController.SendVerifyCodeHandler(c, depCon)
		},
	)

	// Client attempt to verify login code he / she received via SMS from `/send-verify-code`. If the
	// If code is verified, grants user auth permission.
	g.POST("/verify-code", func(c *gin.Context) {
		authController.VerifyLoginCode(c, depCon)
	})

	g.POST(
		"/revoke-jwt",
		jwtactor.JwtValidator(jwtactor.JwtMiddlewareOptions{
			Secret: config.GetAppConf().JwtSecret,
		}),
		authController.RevokeJwtHandler,
	)
}
