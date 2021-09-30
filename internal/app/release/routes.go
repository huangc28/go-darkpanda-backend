package release

import (
	"github.com/gin-gonic/gin"
)

func Routes(r *gin.RouterGroup) {
	g := r.Group("/release")

	g.GET("/android/latest", AndroidLatestDLLinkHandler)
}
