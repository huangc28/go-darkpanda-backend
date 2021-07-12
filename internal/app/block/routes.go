package block

import (
	"github.com/gin-gonic/gin"
	"github.com/golobby/container/pkg/container"
	"github.com/huangc28/go-darkpanda-backend/config"
	"github.com/huangc28/go-darkpanda-backend/internal/app/contracts"
	"github.com/huangc28/go-darkpanda-backend/internal/app/pkg/jwtactor"
)

func Routes(r *gin.RouterGroup, depCon container.Container) {
	var authDaoer contracts.AuthDaoer
	depCon.Make(&authDaoer)

	g := r.Group(
		"/block",
		jwtactor.JwtValidator(jwtactor.JwtMiddlewareOptions{
			Secret: config.GetAppConf().JwtSecret,
		}, authDaoer),
	)

	g.GET("", func(c *gin.Context) {
		GetBlockedUsersHandler(c, depCon)
	})

	g.POST("", func(c *gin.Context) {
		BlockUserHandler(c, depCon)
	})

	g.DELETE("", func(c *gin.Context) {
		UnblockHandler(c, depCon)
	})
}
