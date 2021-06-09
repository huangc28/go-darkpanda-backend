package inquiry

import (
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

	g.GET(
		"",
		middlewares.IsFemale(userDAO),
		GetInquiriesHandler,
	)

	g.GET("/:uuid", func(c *gin.Context) {
		seg := c.Param("uuid")

		switch seg {
		case "active-inquiry":
			GetActiveInquiry(c, container)
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
		EmitInquiryHandler,
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

	g.POST(
		"/agree-to-chat",
		middlewares.IsMale(userDAO),
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
		middlewares.IsMale(userDAO),
		func(c *gin.Context) {
			CancelInquiryHandler(c)
		},
	)

	g.POST(
		"/revert-chat",
		func(c *gin.Context) {
			RevertChatHandler(c, container)
		},
	)

	// ------------------- APIs to be fixed -------------------
	// Man book the inquiry.
	//g.POST(
	//"/book",
	//middlewares.IsMale(userDAO),
	//func(c *gin.Context) {
	//ManBookInquiry(c, container)
	//},
	//)

	// ------------------- APIs to be fixed end -------------------
}
