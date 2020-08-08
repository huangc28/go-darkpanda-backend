package auth

import (
	"net/http"

	"github.com/gorilla/mux"
)

func LoginHandler(w http.ResponseWriter, r *http.Request) {}

// - /v1/register
// - /v1/login
// - /v1/logout
func Routes(r *mux.Router) {
	r.HandleFunc("/register", RegisterHandler)
	r.HandleFunc("/login", LoginHandler)
}
