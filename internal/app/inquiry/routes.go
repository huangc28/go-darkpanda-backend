package inquiry

import (
	"log"

	"github.com/gin-gonic/gin"
	cintrnal "github.com/golobby/container/pkg/container"
	"github.com/huangc28/go-darkpanda-backend/config"
	"github.com/huangc28/go-darkpanda-backend/internal/app/contracts"
	"github.com/huangc28/go-darkpanda-backend/internal/app/middlewares"
	"github.com/huangc28/go-darkpanda-backend/internal/app/pkg/jwtactor"
)

func Routes(r *gin.RouterGroup, container cintrnal.Container) {
	var (
		userDAO contracts.UserDAOer
		authDao contracts.AuthDaoer
	)

	container.Make(&userDAO)
	container.Make(&authDao)

	g := r.Group(
		"/inquiries",
		jwtactor.JwtValidator(jwtactor.JwtMiddlewareOptions{
			Secret: config.GetAppConf().JwtSecret,
		}, authDao),
	)

	// Female user gets list of available inquiries.
	g.GET(
		"",
		middlewares.IsFemale(userDAO),
		func(ctx *gin.Context) {
			log.Println("wrong API !!!!")
			GetInquiriesHandler(ctx, container)
		},
	)

	g.GET("/:uuid", func(c *gin.Context) {
		seg := c.Param("uuid")

		switch seg {
		case "active-inquiry":
			GetActiveInquiry(c, container)
		case "requests":
			// Retrieve list of direct inquiry requests send from male users. The status
			// of these inquiries are `asking`.
			middlewares.IsFemale(userDAO)(c)
			GetDirectInquiryRequests(c, container)
		default:
			GetInquiryHandler(c)
		}
	})

	g.GET(
		"/:uuid/service",
		func(c *gin.Context) {
			GetServiceByInquiryUUID(c, container)
		},
	)

	// Emit a new inquiry by male user.
	g.POST(
		"",
		middlewares.IsMale(userDAO),
		func(c *gin.Context) {
			EmitInquiryHandler(c, container)
		},
	)

	// Patch inquiry detail.
	g.PATCH(
		"/:inquiry_uuid",
		func(c *gin.Context) {
			PatchInquiryHandler(c, container)
		},
	)

	// ------------------- Change inquiry status  -------------------
	// A Female user pickups an inquiry.
	g.POST(
		"/pickup",
		middlewares.IsFemale(userDAO),
		func(c *gin.Context) {
			PickupInquiryHandler(c, container)
		},
	)

	// Both female and male user can accept chatting request via
	// this API to agree to chat with the counter party on request.
	g.POST(
		"/agree-to-chat",
		func(c *gin.Context) {
			AgreeToChatInquiryHandler(c, container)
		},
	)

	// A male user is not interested in chatting the the female
	// who picked up the inquiry. He can `skip` the pickup request
	// and proceed to the next girl.
	g.POST(
		"/skip",
		middlewares.IsMale(userDAO),
		func(c *gin.Context) {
			SkipPickupHandler(c, container)
		},
	)

	// Inquiry can cancel an inquiry via this API. Only workable when inquiry status is `inquiring`.
	g.POST(
		"/cancel",
		func(c *gin.Context) {
			CancelInquiryHandler(c, container)
		},
	)
}
