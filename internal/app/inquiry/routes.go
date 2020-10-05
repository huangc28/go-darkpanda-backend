package inquiry

import (
	"github.com/gin-gonic/gin"
	"github.com/huangc28/go-darkpanda-backend/config"
	"github.com/huangc28/go-darkpanda-backend/internal/app/pkg/jwtactor"
)

func Routes(r *gin.RouterGroup, userDao UserDaoer) {
	g := r.Group(
		"/inquiries",
		jwtactor.JwtValidator(jwtactor.JwtMiddlewareOptions{
			Secret: config.GetAppConf().JwtSecret,
		}),
	)

	g.GET(
		"",
		IsFemale(userDao),
		GetInquiriesHandler,
	)

	g.GET(
		"/:uuid",
		IsFemale(userDao),
		GetInquiryHandler,
	)

	// create inquiry
	g.POST(
		"",
		IsMale(userDao),
		EmitInquiryHandler,
	)

	// cancel inquiry
	g.PATCH(
		"/:inquiry_uuid/cancel",
		ValidateInqiuryURIParams(),
		IsMale(userDao),
		ValidateBeforeAlterInquiryStatus(Cancel),
		CancelInquiryHandler,
	)

	// expire an inquiry
	g.PATCH(
		"/:inquiry_uuid/expire",
		ValidateInqiuryURIParams(),
		IsMale(userDao),
		ValidateBeforeAlterInquiryStatus(Expire),
		ExpireInquiryHandler,
	)

	// pickup an inquiry
	g.POST(
		"/:inquiry_uuid/pickup",
		ValidateInqiuryURIParams(),
		IsFemale(userDao),
		ValidateBeforeAlterInquiryStatus(Pickup),
		PickupInquiryHandler,
	)

	// After chatting, inquiry can be approved by girl
	g.POST(
		"/:inquiry_uuid/girl-approve",
		ValidateInqiuryURIParams(),
		IsFemale(userDao),
		ValidateBeforeAlterInquiryStatus(GirlApprove),
		GirlApproveInquiryHandler,
	)

	// Man book the inquiry
	g.POST(
		"/:inquiry_uuid/book",
		ValidateInqiuryURIParams(),
		IsMale(userDao),
		ValidateBeforeAlterInquiryStatus(Book),
		ManApproveInquiry,
	)
}
