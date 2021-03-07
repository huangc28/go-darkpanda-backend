package service

import (
	"github.com/gin-gonic/gin"
	cintrnal "github.com/golobby/container/pkg/container"
	"github.com/huangc28/go-darkpanda-backend/config"
	"github.com/huangc28/go-darkpanda-backend/internal/app/contracts"
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

	// User can retrieve a list of all `unpaid` and `to be fulfilled` services.
	// If the gender of the requester is `male`, use `customer_id` as the matching
	// criteria. If is `female`, use `service_provider_id` as the matching criteria.
	g.GET(
		"/incoming",
		func(c *gin.Context) {
			GetListOfCurrentServicesHandler(c, container)
		},
	)

	// List of all overdued services whether they are failed or completed
	//   - canceled
	//   - failed_due_to_both
	//   - completed
	//   - failed_due_to_man
	//   - failed_due_to_girl
	g.GET(
		"/overdue",
		func(c *gin.Context) {
			GetOverduedServicesHandlers(c, container)
		},
	)
}
