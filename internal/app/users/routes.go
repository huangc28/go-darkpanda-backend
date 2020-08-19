package user

import (
	"net/http"

	"github.com/gorilla/mux"
)

// Get the following information from the user:
//   - Gender
//   - Username
func GetUserInfo(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("get user info"))
}

func Routes(r *mux.Router) {
	r.HandleFunc("/me", GetUserInfo)
}
