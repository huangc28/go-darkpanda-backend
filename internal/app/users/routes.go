package user

import (
	"net/http"

	"github.com/gorilla/mux"
)

func GetUserInfo(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("get user info"))
}

func Routes(r *mux.Router) {
	r.HandleFunc("/me", GetUserInfo)
}
