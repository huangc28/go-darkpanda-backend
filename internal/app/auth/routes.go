package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/huangc28/go-darkpanda-backend/config"
	"github.com/huangc28/go-darkpanda-backend/db"
	"github.com/huangc28/go-darkpanda-backend/internal/app/auth/internal/twilio"
	"github.com/huangc28/go-darkpanda-backend/internal/app/pkg/jwtactor"
	"github.com/spf13/viper"
)

func Routes(r *gin.RouterGroup, userDAO UserDAO) {
	authController := AuthController{
		TwilioClient: twilio.New(twilio.TwilioConf{
			AccountSID:   viper.GetString("twilio.account_id"),
			AccountToken: viper.GetString("twilio.auth_token"),
		}),
		DB:    db.GetDB(),
		Redis: db.GetRedis(),
	}

	// Provides user information to register a new user.
	// Newly registered user still has to verify their mobile number
	// to proceed usage.
	r.POST("/register", authController.RegisterHandler)

	// This API receives user uuid and mobile number to send SMS verify code.
	// This API is used in pair with `/register` API letting newly registered user
	// to verify their SMS.
	r.POST("/send-verify-code", authController.SendVerifyCodeHandler)

	// Send verify code to user that attempts to login to the application
	r.POST("/send-login-verify-code", authController.SendLoginVerifyCode)

	// // Client attempt to verify login code he / she received via SMS from `/send-verify-code`. If the
	// // verify code matches then grants user auth permission.
	r.POST("/verify-login-code", authController.VerifyLoginCode)

	r.POST("/verify-phone", authController.VerifyPhoneHandler)
	r.POST(
		"/revoke-jwt",
		jwtactor.JwtValidator(jwtactor.JwtMiddlewareOptions{
			Secret: config.GetAppConf().JwtSecret,
		}),
		authController.RevokeJwtHandler,
	)
}
