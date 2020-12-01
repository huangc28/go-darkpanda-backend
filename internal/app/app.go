package app

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/huangc28/go-darkpanda-backend/db"
	"github.com/huangc28/go-darkpanda-backend/internal/app/apperr"
	"github.com/huangc28/go-darkpanda-backend/internal/app/auth"
	"github.com/huangc28/go-darkpanda-backend/internal/app/chat"
	"github.com/huangc28/go-darkpanda-backend/internal/app/deps"
	"github.com/huangc28/go-darkpanda-backend/internal/app/image"
	"github.com/huangc28/go-darkpanda-backend/internal/app/inquiry"
	"github.com/huangc28/go-darkpanda-backend/internal/app/middlewares"
	"github.com/huangc28/go-darkpanda-backend/internal/app/payment"
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
		c.JSON(http.StatusOK, struct {
			Health string `json:"health"`
		}{
			"OK",
		})
	})

	rv1 := e.Group("/v1")

	// Resolve dependencies from different domains from IOC container. We'll inject the dependencies
	// to each domain rotues.

	auth.Routes(
		rv1,
		user.NewUserDAO(db.GetDB()),
	)

	user.Routes(
		rv1,
		&payment.PaymentDAO{
			DB: db.GetDB(),
		},
		&service.ServiceDAO{
			DB: db.GetDB(),
		},
	)

	inquiry.Routes(
		rv1,
		deps.Get().Container,
	)

	image.Routes(rv1)

	chat.Routes(
		rv1,
		deps.Get().Container,
	)

	return e
}
