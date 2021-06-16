package rate

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
		"/rate",
		jwtactor.JwtValidator(jwtactor.JwtMiddlewareOptions{
			Secret: config.GetAppConf().JwtSecret,
		}, authDao),
	)

	g.GET("/:service_uuid", func(c *gin.Context) {
		GetServiceRating(c, depCon)
	})

	g.POST("", func(c *gin.Context) {
		CreateServiceRating(c)
	})
}
