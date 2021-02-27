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
		middlewares.IsFemale(userDAO),
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

	// A Female user can pickup an inquiry.
	g.POST(
		"/:inquiry_uuid/pickup",
		middlewares.IsFemale(userDAO),
		func(c *gin.Context) {
			PickupInquiryHandler(c, container)
		},
	)

	// A Male user agreed to chat with the female. Both parties
	// would enter
	g.POST(
		"/:inquiry/agree-to-chat",
		middlewares.IsMale(userDAO),
		func(c *gin.Context) {
			AgreeToChatInquiryHandler(c, container)
		},
	)

	// A male user is not interested in chatting the the female
	// who picked up the inquiry. He can `skip` the pickup request
	// and proceed to the next girl.
	g.POST(
		"/:inquiry/skip",
		middlewares.IsMale(userDAO),
		func(c *gin.Context) {
			SkipPickupHandler(c, container)
		},
	)

	// Inquiry can cancel an inquiry via this API. Only workable when inquiry
	// status is `inquiring`.
	g.PATCH(
		"/:inquiry_uuid/cancel",
		middlewares.IsMale(userDAO),
		ValidateInqiuryURIParams(),
		ValidateBeforeAlterInquiryStatus(Cancel),
		CancelInquiryHandler,
	)

	// If either user leaves the chat, we should perform soft delete on both the user and the chatroom.
	// Moreover, notify both user in the firestore that the other party has left.
	g.PATCH(
		"/:inquiry_uuid/revert-chat",
		ValidateInqiuryURIParams(),
		middlewares.IsMale(userDAO),
		ValidateBeforeAlterInquiryStatus(RevertChat),
		func(c *gin.Context) {
			RevertChatHandler(c, container)
		},
	)

	// expire an inquiry
	//g.PATCH(
	//"/:inquiry_uuid/expire",
	//ValidateInqiuryURIParams(),
	//middlewares.IsMale(userDAO),
	//ValidateBeforeAlterInquiryStatus(Expire),
	//ExpireInquiryHandler,
	//)

	// Man book the inquiry
	//g.POST(
	//"/:inquiry_uuid/book",
	//ValidateInqiuryURIParams(),
	//middlewares.IsMale(userDAO),
	//ValidateBeforeAlterInquiryStatus(Book),
	//ManApproveInquiry,
	//)

	// After chatting, inquiry can be approved by girl
	g.POST(
		"/:inquiry_uuid/girl-approve",
		middlewares.IsFemale(userDAO),
		func(c *gin.Context) {
			GirlApproveInquiryHandler(c, container)
		},
	)
}
