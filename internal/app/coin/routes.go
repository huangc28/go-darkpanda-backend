package coin

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
		"/coin",
		jwtactor.JwtValidator(jwtactor.JwtMiddlewareOptions{
			Secret: config.GetAppConf().JwtSecret,
		}, authDao),
	)

	// Get my coin current coin balance.
	g.GET(
		"",
		func(c *gin.Context) {
			GetCoinBalance(c, depCon)
		},
	)

	// Deposit coin to user balance.
	g.POST("", func(c *gin.Context) {
		BuyCoin(c, depCon)
	})

	// Get list of coin packages for purchasing.
	g.GET(
		"/packages",
		func(c *gin.Context) {
			GetCoinPackages(c)
		},
	)
}
