package app

import (
	"github.com/gorilla/mux"
	"github.com/huangc28/go-darkpanda-backend/internal/app/auth"
)

// initialize IoC container
// importing router from domains

func StartApp(r *mux.Router) {
	auth.Routes(r)
	//user.Routes(r)
}
