package app

import (
	"github.com/gin-gonic/gin"
	"github.com/huangc28/go-darkpanda-backend/internal/app/apperr"
	"github.com/huangc28/go-darkpanda-backend/internal/app/auth"
)

func StartApp(e *gin.Engine) *gin.Engine {
	e.Use(gin.Logger())
	e.Use(gin.Recovery())
	e.Use(apperr.HandleError())

	rv1 := e.Group("/v1")

	auth.Routes(rv1)
	//user.Routes(r)

	return e
}
