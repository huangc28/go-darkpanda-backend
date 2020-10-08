package user

import (
	"github.com/gin-gonic/gin"
	"github.com/huangc28/go-darkpanda-backend/config"
	"github.com/huangc28/go-darkpanda-backend/internal/app/pkg/jwtactor"
)

func Routes(r *gin.RouterGroup) {
	g := r.Group("/users", jwtactor.JwtValidator(jwtactor.JwtMiddlewareOptions{
		Secret: config.GetAppConf().JwtSecret,
	}))

	g.GET("/:uuid/images", GetUserImagesHandler)

	// issue: https://github.com/gin-gonic/gin/issues/205
	// issue: https://github.com/julienschmidt/httprouter/issues/12
	g.GET("/:uuid", func(c *gin.Context) {
		switch c.Param("uuid") {
		case "me":
			GetMyProfileHandler(c)
		default:
			GetUserProfileHandler(c)
		}

	})

	g.PUT("/:uuid", PutUserInfo)
}
