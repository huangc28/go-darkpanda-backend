package app

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/huangc28/go-darkpanda-backend/db"
	"github.com/huangc28/go-darkpanda-backend/internal/app/apperr"
	"github.com/huangc28/go-darkpanda-backend/internal/app/auth"
	bankAccount "github.com/huangc28/go-darkpanda-backend/internal/app/bank_account"
	"github.com/huangc28/go-darkpanda-backend/internal/app/block"
	"github.com/huangc28/go-darkpanda-backend/internal/app/chat"
	"github.com/huangc28/go-darkpanda-backend/internal/app/coin"
	"github.com/huangc28/go-darkpanda-backend/internal/app/deps"
	"github.com/huangc28/go-darkpanda-backend/internal/app/image"
	"github.com/huangc28/go-darkpanda-backend/internal/app/inquiry"
	"github.com/huangc28/go-darkpanda-backend/internal/app/middlewares"
	"github.com/huangc28/go-darkpanda-backend/internal/app/payment"
	"github.com/huangc28/go-darkpanda-backend/internal/app/referral"
	"github.com/huangc28/go-darkpanda-backend/internal/app/register"
	"github.com/huangc28/go-darkpanda-backend/internal/app/release"
	"github.com/huangc28/go-darkpanda-backend/internal/app/service"
	"github.com/huangc28/go-darkpanda-backend/internal/app/user"
)

func StartApp(e *gin.Engine) *gin.Engine {
	e.Use(gin.Logger())

	// Log the response so frontend can better normalize the result.
	e.Use(middlewares.ResponseLogger)
	e.Use(gin.Recovery())
	e.Use(apperr.HandleError())

	e.GET("/health", func(c *gin.Context) {
		dbConn := db.GetDB()

		if err := dbConn.Ping(); err != nil {
			c.AbortWithError(
				http.StatusInternalServerError,
				apperr.NewErr(
					apperr.DBConnectionError,
					err.Error(),
				),
			)
			return
		}

		ctx := context.Background()
		redisConn := db.GetRedis()

		if err := redisConn.Ping(ctx).Err(); err != nil {
			c.AbortWithError(
				http.StatusInternalServerError,
				apperr.NewErr(
					apperr.RedisConnectionError,
					err.Error(),
				),
			)
			return

		}

		c.JSON(http.StatusOK, struct {
			Health string `json:"health"`
		}{
			"OK",
		})
	})

	rv1 := e.Group("/v1")

	// Resolve dependencies from different domains from IOC container. We'll inject the dependencies
	// to each domain rotues.
	referral.Routes(
		rv1,
		deps.Get().Container,
	)

	register.Routes(
		rv1,
		deps.Get().Container,
	)

	auth.Routes(
		rv1,
		deps.Get().Container,
	)

	user.Routes(
		rv1,
		deps.Get().Container,
	)

	inquiry.Routes(
		rv1,
		deps.Get().Container,
	)

	service.Routes(
		rv1,
		deps.Get().Container,
	)

	image.Routes(
		rv1,
		deps.Get().Container,
	)

	chat.Routes(
		rv1,
		deps.Get().Container,
	)

	bankAccount.Routes(
		rv1,
		deps.Get().Container,
	)

	coin.Routes(
		rv1,
		deps.Get().Container,
	)

	block.Routes(
		rv1,
		deps.Get().Container,
	)

	payment.Routes(
		rv1,
		deps.Get().Container,
	)

	release.Routes(
		rv1,
	)

	return e
}
