package rate

import (
	"github.com/gin-gonic/gin"
	"github.com/golobby/container/pkg/container"
	"github.com/huangc28/go-darkpanda-backend/config"
	"github.com/huangc28/go-darkpanda-backend/internal/app/pkg/jwtactor"
)

func Routes(r *gin.RouterGroup, depCon container.Container) {
	g := r.Group(
		"/rate",
		jwtactor.JwtValidator(jwtactor.JwtMiddlewareOptions{
			Secret: config.GetAppConf().JwtSecret,
		}),
	)

	g.GET("/:uuid", func(c *gin.Context) {
		GetUserRating(c, depCon)
	})
}
