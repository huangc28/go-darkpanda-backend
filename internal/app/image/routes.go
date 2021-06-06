package image

import (
	"github.com/gin-gonic/gin"
	"github.com/golobby/container/pkg/container"
	"github.com/huangc28/go-darkpanda-backend/config"
	"github.com/huangc28/go-darkpanda-backend/internal/app/contracts"
	"github.com/huangc28/go-darkpanda-backend/internal/app/pkg/jwtactor"
)

func Routes(r *gin.RouterGroup, depCon container.Container) {
	var authDao contracts.AuthDaoer
	depCon.Make(&authDao)

	g := r.Group(
		"/images",
		jwtactor.JwtValidator(jwtactor.JwtMiddlewareOptions{
			Secret: config.GetAppConf().JwtSecret,
		}, authDao),
	)

	g.POST("", func(c *gin.Context) {
		UploadImagesHandler(c, depCon)
	})

	g.POST("/avatar", UploadAvatarHandler)
}
