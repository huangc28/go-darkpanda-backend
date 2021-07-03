package referral

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
		"/referral_code",
		jwtactor.JwtValidator(jwtactor.JwtMiddlewareOptions{
			Secret: config.GetAppConf().JwtSecret,
		},
			authDaoer,
		),
	)

	// Gets currently active referral token that the user can use.
	// If the latest token is occupied, generate a fresh referral
	// token.
	g.GET("", func(c *gin.Context) {
		GetReferralCodeHandler(c, depCon)
	})

	r.POST("/verify", func(c *gin.Context) {
		HandleVerifyReferralCode(c, depCon)
	})
}
