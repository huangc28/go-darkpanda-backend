package jwtactor

import (
	"context"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	apperr "github.com/huangc28/go-darkpanda-backend/internal/app/apperr"
	"github.com/huangc28/go-darkpanda-backend/internal/app/contracts"
)

func ExtractTokenFromRequest(c *gin.Context) (string, error) {
	headerJwt := JwtToken{}

	if err := c.ShouldBindHeader(&headerJwt); err != nil {
		return "", err
	}

	if len(headerJwt.AuthJwt) > 0 {
		strArr := strings.Split(headerJwt.AuthJwt, " ")

		if len(strArr) >= 2 {
			return strArr[1], nil
		}

		return headerJwt.AuthJwt, nil
	}

	if err := c.ShouldBindQuery(&headerJwt); err != nil {
		return "", err
	}

	return headerJwt.AuthJwt, nil
}

type JwtMiddlewareOptions struct {
	Secret string
}

type JwtToken struct {
	AuthJwt string `header:"Authorization" form:"jwt"`
}

func JwtValidator(opt JwtMiddlewareOptions, authDaoer contracts.AuthDaoer) gin.HandlerFunc {
	return func(c *gin.Context) {
		// retrieve jwt token from either header or url
		// header:
		//   Authorization: Bearer ${JWT}
		token, err := ExtractTokenFromRequest(c)

		if err != nil {
			c.AbortWithError(
				http.StatusInternalServerError,
				apperr.NewErr(
					apperr.FailedToBindJwtInHeader,
					err.Error(),
				),
			)

			return
		}

		if len(token) <= 0 {
			c.AbortWithError(
				http.StatusBadRequest,
				apperr.NewErr(apperr.MissingAuthToken),
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

		// Check redis to makesure the jwt token is valid.
		ctx := context.Background()
		isInvalid, err := authDaoer.IsTokenInvalid(ctx, token)

		if err != nil {
			c.AbortWithError(
				http.StatusInternalServerError,
				apperr.NewErr(
					apperr.FailedToValidateToken,
					err.Error(),
				),
			)

			return
		}

		if isInvalid {
			c.AbortWithError(
				http.StatusForbidden,
				apperr.NewErr(
					apperr.TokenIsInvalidated,
					err.Error(),
				),
			)

			return
		}

		// ------------------- set uuid and jwt -------------------
		c.Set("uuid", claims.Uuid)
		c.Set("jwt", token)

		c.Next()
	}
}
