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
	var (
		userDAO contracts.UserDAOer
		authDao contracts.AuthDaoer
	)
	container.Make(&userDAO)
	container.Make(&authDao)
	g := r.Group(
		"/services",
		jwtactor.JwtValidator(jwtactor.JwtMiddlewareOptions{
			Secret: config.GetAppConf().JwtSecret,
		}, authDao),
	)

	g.GET("/:seg", func(c *gin.Context) {
		uriSeg := c.Param("seg")

		switch uriSeg {
		default:
			GetServiceDetailHandler(c, container)

		case "incoming":

			// User can retrieve a list of all `unpaid` and `to_be_fulfilled` services.
			// If the gender of the requester is `male`, use `customer_id` as the matching
			// criteria. If is `female`, use `service_provider_id` as the matching criteria.
			GetIncomingServicesHandler(c, container)

		case "overdue":

			// List of all overdued services whether they are failed or completed
			//   - canceled
			//   - failed_due_to_both
			//   - completed
			//   - failed_due_to_man
			//   - failed_due_to_girl
			GetOverduedServicesHandlers(c, container)

		case "service-list":

			// Get list of services available.
			GetAvailableServices(c)
		}
	})

	g.GET(
		"/:seg/qrcode",
		func(c *gin.Context) {
			GetServiceQRCode(c, container)
		},
	)

	g.GET(
		"/:seg/payment-details",
		func(c *gin.Context) {
			GetServicePaymentDetails(c, container)
		},
	)

	g.GET(
		"/:seg/rating",
		func(c *gin.Context) {
			GetServiceRating(c, container)
		},
	)

	g.POST(
		"/:seg",
		func(c *gin.Context) {
			seg := c.Param("seg")

			switch seg {
			case "scan-service-qrcode":
				ScanServiceQrCode(c, container)
			}
		},
	)

	g.POST(
		"/:seg/rating",
		func(c *gin.Context) {
			CreateServiceRating(c, container)
		},
	)

}
