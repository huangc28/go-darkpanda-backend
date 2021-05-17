package coin

import (
	"github.com/gin-gonic/gin"
	"github.com/golobby/container/pkg/container"
	"github.com/huangc28/go-darkpanda-backend/config"
	"github.com/huangc28/go-darkpanda-backend/internal/app/pkg/jwtactor"
)

func Routes(r *gin.RouterGroup, depCon container.Container) {
	g := r.Group(
		"/coin",
		jwtactor.JwtValidator(jwtactor.JwtMiddlewareOptions{
			Secret: config.GetAppConf().JwtSecret,
		}),
	)

	// Get my coin current coin balance.
	g.GET(
		"",
		func(c *gin.Context) {
			GetCoinBalance(c, depCon)
		},
	)

	g.POST("", func(c *gin.Context) {
		BuyCoin(c, depCon)
	})
}
