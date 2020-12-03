package user

import (
	"github.com/gin-gonic/gin"
	"github.com/golobby/container/pkg/container"
	"github.com/huangc28/go-darkpanda-backend/config"
	"github.com/huangc28/go-darkpanda-backend/internal/app/pkg/jwtactor"
)

func Routes(r *gin.RouterGroup, depCon container.Container) {
	g := r.Group(
		"/users",
		jwtactor.JwtValidator(
			jwtactor.JwtMiddlewareOptions{
				Secret: config.GetAppConf().JwtSecret,
			},
		),
	)

	handlers := UserHandlers{
		Container: depCon,
	}

	g.GET("/:uuid/services", handlers.GetUserServiceHistory)

	g.GET("/:uuid/payments", handlers.GetUserPayments)

	g.GET("/:uuid/images", handlers.GetUserImagesHandler)

	// issue: https://github.com/gin-gonic/gin/issues/205
	// issue: https://github.com/julienschmidt/httprouter/issues/12
	g.GET("/:uuid", func(c *gin.Context) {
		switch c.Param("uuid") {
		case "me":
			handlers.GetMyProfileHandler(c)
		default:
			handlers.GetUserProfileHandler(c)
		}

	})

	g.PUT("/:uuid", handlers.PutUserInfo)
}
