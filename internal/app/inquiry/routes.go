package inquiry

import (
	"github.com/gin-gonic/gin"
	"github.com/huangc28/go-darkpanda-backend/config"
	"github.com/huangc28/go-darkpanda-backend/db"
	"github.com/huangc28/go-darkpanda-backend/internal/app/contracts"
	"github.com/huangc28/go-darkpanda-backend/internal/app/middlewares"
	"github.com/huangc28/go-darkpanda-backend/internal/app/pkg/jwtactor"
)

type InquiryRoutesParams struct {
	UserDAO      contracts.UserDAOer
	ChatServicer contracts.ChatServicer
	ChatDAO      contracts.ChatDaoer
	ServiceDAO   contracts.ServiceDAOer
}

// userDao contracts.UserDAOer, chatServices contracts.ChatServicer, chatDao contracts.ChatDaoer)
func Routes(r *gin.RouterGroup, params *InquiryRoutesParams) {
	g := r.Group(
		"/inquiries",
		jwtactor.JwtValidator(jwtactor.JwtMiddlewareOptions{
			Secret: config.GetAppConf().JwtSecret,
		}),
	)

	handlers := &InquiryHandlers{
		InquiryDao: NewInquiryDAO(db.GetDB()),
		UserDao:    params.UserDAO,
		LobbyServices: &LobbyServices{
			LobbyDao: &LobbyDao{
				DB: db.GetDB(),
			},
		},
		ChatServices: params.ChatServicer,
		ChatDao:      params.ChatDAO,
		ServiceDAO:   params.ServiceDAO,
	}

	g.GET(
		"",
		middlewares.IsFemale(params.UserDAO),
		handlers.GetInquiriesHandler,
	)

	g.GET(
		"/:uuid",
		middlewares.IsFemale(params.UserDAO),
		GetInquiryHandler,
	)

	g.GET(
		"/:uuid/service",
		handlers.GetServiceByInquiryUUID,
	)

	// Emit a new inquiry by male user.
	g.POST(
		"",
		middlewares.IsMale(params.UserDAO),
		handlers.EmitInquiryHandler,
	)

	// Cancel a inquiry.
	g.PATCH(
		"/:inquiry_uuid/cancel",
		ValidateInqiuryURIParams(),
		ValidateBeforeAlterInquiryStatus(Cancel),
		CancelInquiryHandler,
	)

	// If either user leaves the chat, we should perform soft delete on both the user and the chatroom.
	g.PATCH(
		"/:inquiry_uuid/revert-chat",
		ValidateInqiuryURIParams(),
		middlewares.IsMale(params.UserDAO),
		ValidateBeforeAlterInquiryStatus(RevertChat),
		handlers.RevertChat,
	)

	// expire an inquiry
	g.PATCH(
		"/:inquiry_uuid/expire",
		ValidateInqiuryURIParams(),
		middlewares.IsMale(params.UserDAO),
		ValidateBeforeAlterInquiryStatus(Expire),
		ExpireInquiryHandler,
	)

	// pickup an inquiry
	g.POST(
		"/:inquiry_uuid/pickup",
		ValidateInqiuryURIParams(),
		middlewares.IsFemale(params.UserDAO),
		ValidateBeforeAlterInquiryStatus(Pickup),
		handlers.PickupInquiryHandler,
	)

	// After chatting, inquiry can be approved by girl
	g.POST(
		"/:inquiry_uuid/girl-approve",
		ValidateInqiuryURIParams(),
		middlewares.IsFemale(params.UserDAO),
		ValidateBeforeAlterInquiryStatus(GirlApprove),
		GirlApproveInquiryHandler,
	)

	// Man book the inquiry
	g.POST(
		"/:inquiry_uuid/book",
		ValidateInqiuryURIParams(),
		middlewares.IsMale(params.UserDAO),
		ValidateBeforeAlterInquiryStatus(Book),
		ManApproveInquiry,
	)
}
