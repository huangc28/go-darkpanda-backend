package jwtactor

import (
	"time"

	"github.com/dgrijalva/jwt-go"
)

type Claim struct {
	Uuid       string `json:"uuid"`
	Authorized bool   `json:"authorized"`
	jwt.StandardClaims
}

// TODO: extend jwt token valid time. Extract the time variable to config.
func CreateToken(uUuid string, jwtSecret string) (string, error) {
	claim := &Claim{
		Uuid: uUuid,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 24 * 30).Unix(),
		},
	}

	at := jwt.NewWithClaims(jwt.SigningMethodHS256, claim)
	token, err := at.SignedString([]byte(jwtSecret))

	if err != nil {
		return "", err
	}

	return token, nil
}
