package inquiry

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/huangc28/go-darkpanda-backend/config"
	"github.com/huangc28/go-darkpanda-backend/db"
	"github.com/huangc28/go-darkpanda-backend/internal/app/contracts"
	"github.com/huangc28/go-darkpanda-backend/internal/app/middlewares"
	"github.com/huangc28/go-darkpanda-backend/internal/app/pkg/jwtactor"
)

func Routes(r *gin.RouterGroup, userDao contracts.UserDAOer, chatServices contracts.ChatServicer, chatDao contracts.ChatDaoer) {

	log.Printf("DEBUG 001 userDao addr %p ", userDao)

	g := r.Group(
		"/inquiries",
		jwtactor.JwtValidator(jwtactor.JwtMiddlewareOptions{
			Secret: config.GetAppConf().JwtSecret,
		}),
	)

	handlers := &InquiryHandlers{
		InquiryDao: NewInquiryDAO(db.GetDB()),
		UserDao:    userDao,
		LobbyServices: &LobbyServices{
			LobbyDao: &LobbyDao{
				DB: db.GetDB(),
			},
		},
		ChatServices: chatServices,
		ChatDao:      chatDao,
	}

	g.GET(
		"",
		middlewares.IsFemale(),
		handlers.GetInquiriesHandler,
	)

	g.GET(
		"/:uuid",
		middlewares.IsFemale(),
		GetInquiryHandler,
	)

	// Emit a new inquiry by male user.
	g.POST(
		"",
		middlewares.IsMale(),
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
		middlewares.IsMale(),
		ValidateBeforeAlterInquiryStatus(RevertChat),
		handlers.RevertChat,
	)

	// expire an inquiry
	g.PATCH(
		"/:inquiry_uuid/expire",
		ValidateInqiuryURIParams(),
		middlewares.IsMale(),
		ValidateBeforeAlterInquiryStatus(Expire),
		ExpireInquiryHandler,
	)

	// pickup an inquiry
	g.POST(
		"/:inquiry_uuid/pickup",
		ValidateInqiuryURIParams(),
		middlewares.IsFemale(),
		ValidateBeforeAlterInquiryStatus(Pickup),
		handlers.PickupInquiryHandler,
	)

	// After chatting, inquiry can be approved by girl
	g.POST(
		"/:inquiry_uuid/girl-approve",
		ValidateInqiuryURIParams(),
		middlewares.IsFemale(),
		ValidateBeforeAlterInquiryStatus(GirlApprove),
		GirlApproveInquiryHandler,
	)

	// Man book the inquiry
	g.POST(
		"/:inquiry_uuid/book",
		ValidateInqiuryURIParams(),
		middlewares.IsMale(),
		ValidateBeforeAlterInquiryStatus(Book),
		ManApproveInquiry,
	)
}
