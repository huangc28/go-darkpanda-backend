package app

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/huangc28/go-darkpanda-backend/db"
	"github.com/huangc28/go-darkpanda-backend/internal/app/apperr"
	"github.com/huangc28/go-darkpanda-backend/internal/app/auth"
	"github.com/huangc28/go-darkpanda-backend/internal/app/chat"
	"github.com/huangc28/go-darkpanda-backend/internal/app/contracts"
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
	var (
		serviceDao contracts.ServiceDAOer
		userDao    contracts.UserDAOer
		inquiryDao contracts.InquiryDAOer
	)

	deps.Get().Container.Make(&userDao)
	deps.Get().Container.Make(&serviceDao)
	deps.Get().Container.Make(&inquiryDao)

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
		&inquiry.InquiryRoutesParams{
			UserDAO:      userDao,
			ChatServicer: chat.NewChatServices(chat.NewChatDao(db.GetDB())),
			ChatDAO:      chat.NewChatDao(db.GetDB()),
			ServiceDAO:   serviceDao,
		},
	)

	image.Routes(rv1)

	chat.Routes(rv1, &chat.ChatRoutesParams{
		UserDao:    userDao,
		ServiceDao: serviceDao,
		InquiryDao: inquiryDao,
	})

	return e
}
