package auth

import "github.com/gin-gonic/gin"

type JwtValidatorParams struct {
	Jwt string `json:"jwt"`
}

func JwtValidator() gin.HandlerFunc {
	return func(c *gin.Context) {
		// retrieve jwt token from either header or payload
	}
}
