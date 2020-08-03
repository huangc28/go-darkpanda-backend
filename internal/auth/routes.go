package auth

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "from register route")
}

// - /v1/register
// - /v1/login
// - /v1/logout
func Routes(r *mux.Router) {
	r.HandleFunc("/register", RegisterHandler)
	//r.HandleFunc("/login")
}
