package inquiry

import (
	"github.com/gin-gonic/gin"
	"github.com/huangc28/go-darkpanda-backend/config"
	"github.com/huangc28/go-darkpanda-backend/db"
	"github.com/huangc28/go-darkpanda-backend/internal/app/middlewares"
	"github.com/huangc28/go-darkpanda-backend/internal/app/pkg/jwtactor"
)

func Routes(r *gin.RouterGroup, userDao UserDaoer) {
	g := r.Group(
		"/inquiries",
		jwtactor.JwtValidator(jwtactor.JwtMiddlewareOptions{
			Secret: config.GetAppConf().JwtSecret,
		}),
	)

	handlers := &InquiryHandlers{
		UserDao: userDao,
		LobbyServices: &LobbyServices{
			LobbyDao: &LobbyDao{
				DB: db.GetDB(),
			},
		},
		ChatServices: &ChatServices{
			ChatDao: &ChatDao{
				DB: db.GetDB(),
			},
		},
	}

	g.GET(
		"",
		middlewares.IsFemale(userDao),
		GetInquiriesHandler,
	)

	g.GET(
		"/:uuid",
		middlewares.IsFemale(userDao),
		GetInquiryHandler,
	)

	// emit a new inquiry
	g.POST(
		"",
		middlewares.IsMale(userDao),
		handlers.EmitInquiryHandler,
	)

	// cancel inquiry
	g.PATCH(
		"/:inquiry_uuid/cancel",
		ValidateInqiuryURIParams(),
		middlewares.IsMale(userDao),
		ValidateBeforeAlterInquiryStatus(Cancel),
		CancelInquiryHandler,
	)

	// expire an inquiry
	g.PATCH(
		"/:inquiry_uuid/expire",
		ValidateInqiuryURIParams(),
		middlewares.IsMale(userDao),
		ValidateBeforeAlterInquiryStatus(Expire),
		ExpireInquiryHandler,
	)

	// pickup an inquiry
	g.POST(
		"/:inquiry_uuid/pickup",
		ValidateInqiuryURIParams(),
		middlewares.IsFemale(userDao),
		ValidateBeforeAlterInquiryStatus(Pickup),
		handlers.PickupInquiryHandler,
	)

	// After chatting, inquiry can be approved by girl
	g.POST(
		"/:inquiry_uuid/girl-approve",
		ValidateInqiuryURIParams(),
		middlewares.IsFemale(userDao),
		ValidateBeforeAlterInquiryStatus(GirlApprove),
		GirlApproveInquiryHandler,
	)

	// Man book the inquiry
	g.POST(
		"/:inquiry_uuid/book",
		ValidateInqiuryURIParams(),
		middlewares.IsMale(userDao),
		ValidateBeforeAlterInquiryStatus(Book),
		ManApproveInquiry,
	)
}
