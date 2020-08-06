package auth

import (
	"fmt"
	"net/http"
)

// We need the following to register new user
//   - reference code
//   - username
//   - phone verification code
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "from register route")
}
