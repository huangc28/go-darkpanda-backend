package release

import (
	"github.com/gin-gonic/gin"
	"github.com/huangc28/go-darkpanda-backend/internal/app/middlewares"
)

func Routes(r *gin.RouterGroup) {
	g := r.Group("/release")

	g.GET("/android/latest", middlewares.CORSMiddlewares(), AndroidLatestDLLinkHandler)
}
