package service

import (
	"github.com/gin-gonic/gin"
	cintrnal "github.com/golobby/container/pkg/container"
	"github.com/huangc28/go-darkpanda-backend/config"
	"github.com/huangc28/go-darkpanda-backend/internal/app/contracts"
	"github.com/huangc28/go-darkpanda-backend/internal/app/middlewares"
	"github.com/huangc28/go-darkpanda-backend/internal/app/pkg/jwtactor"
)

func Routes(r *gin.RouterGroup, container cintrnal.Container) {
	g := r.Group(
		"/services",
		jwtactor.JwtValidator(jwtactor.JwtMiddlewareOptions{
			Secret: config.GetAppConf().JwtSecret,
		}),
	)

	var userDAO contracts.UserDAOer
	container.Make(&userDAO)

	// Female user can retrieve a list of all `unpaid` and `to be fulfilled` services
	g.GET(
		"/incoming",
		middlewares.IsFemale(userDAO),
		func(c *gin.Context) {
			GetListOfCurrentServicesHandler(c, container)
		},
	)

	// List of all historical services whether they are failed or completed
	//   - canceled
	//   - failed_due_to_both
	//   - completed
	//   - failed_due_to_man
	//   - failed_due_to_girl
	//g.GET(
	//"/"

	//)

	// Retrieve service by uuid uuid

}
