package app

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/huangc28/go-darkpanda-backend/db"
	"github.com/huangc28/go-darkpanda-backend/internal/app/apperr"
	"github.com/huangc28/go-darkpanda-backend/internal/app/auth"
	"github.com/huangc28/go-darkpanda-backend/internal/app/chat"
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

	userDao := user.NewUserDAO(db.GetDB())

	inquiry.Routes(
		rv1,
		userDao,
		chat.NewChatServices(chat.NewChatDao(db.GetDB())),
		chat.NewChatDao(db.GetDB()),
	)

	image.Routes(rv1)

	chat.Routes(rv1)

	return e
}
