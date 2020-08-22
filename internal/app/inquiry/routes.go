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

	// create inquiry
	g.POST(
		"",
		IsMale(userDao),
		EmitInquiryHandler,
	)

	// cancel inquiry
	g.PATCH(
		"/:inquiry_uuid/cancel",
		IsMale(userDao),
		ValidateBeforeAlterInquiryStatus(),
		CancelInquiry,
	)

	// expire an inquiry
	g.PATCH(
		"/:inquiry_uuid/expire",
		IsMale(userDao),
		ValidateBeforeAlterInquiryStatus(),
		ExpireInquiry,
	)
}
