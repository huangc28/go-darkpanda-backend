package auth

import (
	"github.com/gin-gonic/gin"
)

//func LoginHandler(w http.ResponseWriter, r *http.Request) {}

// - /v1/register
// - /v1/login
// - /v1/logout
func Routes(r *gin.RouterGroup) {
	r.POST("/register", RegisterHandler)
	r.POST("/send-verify-code", SendVerifyCodeHandler)
	//r.POST("/verify-phone")

	//r.HandleFunc("/login", LoginHandler).Methods("POST")
}
