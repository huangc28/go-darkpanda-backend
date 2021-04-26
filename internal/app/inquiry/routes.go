package inquiry

import (
	"net/http"

	"github.com/gin-gonic/gin"
	cintrnal "github.com/golobby/container/pkg/container"
	"github.com/huangc28/go-darkpanda-backend/config"
	"github.com/huangc28/go-darkpanda-backend/internal/app/contracts"
	"github.com/huangc28/go-darkpanda-backend/internal/app/middlewares"
	"github.com/huangc28/go-darkpanda-backend/internal/app/pkg/jwtactor"
)

func Routes(r *gin.RouterGroup, container cintrnal.Container) {
	g := r.Group(
		"/inquiries",
		jwtactor.JwtValidator(jwtactor.JwtMiddlewareOptions{
			Secret: config.GetAppConf().JwtSecret,
		}),
	)

	var userDAO contracts.UserDAOer

	container.Make(&userDAO)

	g.GET(
		"",
		middlewares.IsFemale(userDAO),
		GetInquiriesHandler,
	)

	g.GET(
		"/:uuid",
		GetInquiryHandler,
	)

	g.GET(
		"/:uuid/service",
		func(c *gin.Context) {
			GetServiceByInquiryUUID(c, container)
		},
	)

	g.GET(
		"/:uuid/inquirer",
		ValidateInqiuryURIParams(),
		middlewares.IsFemale(userDAO),
		func(c *gin.Context) {
			GetInquirerInfo(c, container)
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
		ValidateInqiuryURIParams(),
		func(c *gin.Context) {
			PatchInquiryHandler(c, container)
		},
	)

	g.POST(
		"/:inquiry_uuid/:sub_route",
		func(c *gin.Context) {
			subRoute := c.Param("sub_route")

			switch subRoute {

			// A Female user can pickup an inquiry.
			case "pickup":
				middlewares.IsFemale(userDAO)(c)
				PickupInquiryHandler(c, container)

			// A Male user agreed to chat with the female. Both parties
			// would enter
			case "agree-to-chat":
				middlewares.IsMale(userDAO)(c)
				AgreeToChatInquiryHandler(c, container)

			// A male user is not interested in chatting the the female
			// who picked up the inquiry. He can `skip` the pickup request
			// and proceed to the next girl.
			case "skip":
				middlewares.IsMale(userDAO)(c)
				SkipPickupHandler(c, container)

			// Inquiry can cancel an inquiry via this API. Only workable when inquiry
			// status is `inquiring`.
			case "cancel":
				middlewares.IsMale(userDAO)(c)
				ValidateInqiuryURIParams()(c)
				ValidateBeforeAlterInquiryStatus(Cancel)(c)
				CancelInquiryHandler(c)

			// If either user leaves the chat, we should perform soft delete on both the user and the chatroom.
			// Moreover, notify both user in the firestore that the other party has left.
			case "revert-chat":
				ValidateInqiuryURIParams()(c)
				middlewares.IsMale(userDAO)(c)
				ValidateBeforeAlterInquiryStatus(RevertChat)(c)
				RevertChatHandler(c, container)

			// Man book the inquiry
			case "book":
				middlewares.IsMale(userDAO)(c)
				ManBookInquiry(c, container)

			// After inquiry chatting, inquiry can be approved by girl
			case "girl-approve":
				middlewares.IsFemale(userDAO)(c)
				GirlApproveInquiryHandler(c, container)
			default:
				c.String(http.StatusNotFound, "page not found")
			}

		},
	)

}
