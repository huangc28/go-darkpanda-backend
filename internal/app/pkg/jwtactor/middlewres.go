package jwtactor

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	apperr "github.com/huangc28/go-darkpanda-backend/internal/app/apperr"
	log "github.com/sirupsen/logrus"
)

type JwtValidatorParams struct {
	Jwt string `json:"jwt"`
}

func ExtractTokenFromHeader(req *http.Request) string {
	bearToken := req.Header.Get("Authorization")
	strArr := strings.Split(bearToken, " ")
	if len(strArr) == 2 {
		return strArr[1]
	}

	return ""
}

func ExtractTokenFromBody(req *http.Request) (string, error) {
	bodyStruct := struct {
		Jwt string `json:"jwt"`
	}{}

	dec := json.NewDecoder(req.Body)

	if err := dec.Decode(&bodyStruct); err != nil {
		return "", err
	}

	return bodyStruct.Jwt, nil
}

func ExtractTokenFromRequest(req *http.Request) (string, error) {
	token := ExtractTokenFromHeader(req)

	if token != "" && len(token) > 0 {
		return token, nil
	}

	token, err := ExtractTokenFromBody(req)

	if err != nil {
		return "", nil
	}

	return token, nil
}

type JwtMiddlewareOptions struct {
	Secret string
}

func JwtValidator(opt JwtMiddlewareOptions) gin.HandlerFunc {
	return func(c *gin.Context) {
		// retrieve jwt token from either header or url
		// header:
		//   Authorization: Bearer ${JWT}
		// url:
		token, err := ExtractTokenFromRequest(c.Request)

		if err != nil {
			c.AbortWithError(
				http.StatusUnauthorized,
				apperr.NewErr(apperr.JWTNotProvided),
			)

			return
		}

		claims := &Claim{}
		tkn, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(opt.Secret), nil
		})

		if err != nil {
			if err == jwt.ErrSignatureInvalid {
				c.AbortWithError(
					http.StatusUnauthorized,
					apperr.NewErr(
						apperr.InvalidSignature,
						err.Error(),
					),
				)

				return
			}

			c.AbortWithError(
				http.StatusBadRequest,
				apperr.NewErr(
					apperr.FailedToParseSignature,
					err.Error(),
				),
			)

			return
		}

		if !tkn.Valid {
			c.AbortWithError(
				http.StatusUnauthorized,
				apperr.NewErr(
					apperr.InvalidSigature,
					err.Error(),
				),
			)

			return
		}

		// ------------------- set uuid and jwt -------------------
		c.Set("uuid", claims.Uuid)
		c.Set("jwt", token)

		log.Infoln("jwt given passes the validation!")

		c.Next()
	}
}
