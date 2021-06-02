package service

import (
	"github.com/gin-gonic/gin"
	cintrnal "github.com/golobby/container/pkg/container"
	"github.com/huangc28/go-darkpanda-backend/config"
	"github.com/huangc28/go-darkpanda-backend/internal/app/contracts"
	"github.com/huangc28/go-darkpanda-backend/internal/app/pkg/jwtactor"
)

// TODO: modify the get API format to restful.
func Routes(r *gin.RouterGroup, container cintrnal.Container) {
	g := r.Group(
		"/services",
		jwtactor.JwtValidator(jwtactor.JwtMiddlewareOptions{
			Secret: config.GetAppConf().JwtSecret,
		}),
	)

	var userDAO contracts.UserDAOer
	container.Make(&userDAO)

	g.GET("/:seg", func(c *gin.Context) {
		uriSeg := c.Param("seg")

		switch uriSeg {
		case "incoming":

			// User can retrieve a list of all `unpaid` and `to_be_fulfilled` services.
			// If the gender of the requester is `male`, use `customer_id` as the matching
			// criteria. If is `female`, use `service_provider_id` as the matching criteria.
			GetListOfCurrentServicesHandler(c, container)

		case "overdue":

			// List of all overdued services whether they are failed or completed
			//   - canceled
			//   - failed_due_to_both
			//   - completed
			//   - failed_due_to_man
			//   - failed_due_to_girl
			GetOverduedServicesHandlers(c, container)
		}
	})

	g.GET(
		"/:seg/qrcode",
		func(c *gin.Context) {
			GetServiceQRCode(c, container)
		},
	)

	g.POST(
		"/scan-service-qrcode",
		func(c *gin.Context) {
			ScanServiceQrCode(c, container)
		},
	)

}
