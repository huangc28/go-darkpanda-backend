package inquiry

import (
	"github.com/gin-gonic/gin"
	"github.com/huangc28/go-darkpanda-backend/config"
	"github.com/huangc28/go-darkpanda-backend/internal/app/pkg/jwtactor"
)

func Routes(r *gin.RouterGroup) {
	g := r.Group("/inquiries", jwtactor.JwtValidator(jwtactor.JwtMiddlewareOptions{
		Secret: config.GetAppConf().JwtSecret,
	}))

	g.POST("", EmitInquiryHandler)
}
