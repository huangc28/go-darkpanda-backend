package apperr

import (
	"github.com/gin-gonic/gin"
)

type Error struct {
	ErrCode string `json:"err_code"`
	ErrMsg  string `json:"err_msg"`
}

func (e *Error) Error() string {
	return e.ErrMsg
}

func NewErr(errCode string, args ...interface{}) *Error {
	errMsg := GetErrorMessage(errCode)

	if len(args) == 1 {
		errMsg = args[0].(string)
	}

	return &Error{
		errCode,
		errMsg,
	}
}

func GetErrorMessage(code string) string {
	message := ""

	if msg, exists := ErrCodeMsgMap[code]; exists {
		message = msg
	}

	return message
}

// Format and return failure response to following format:
//
//   {
//     status: 500,
//     err_code: "1000001",
//     err_msg : "some shit happened"
//   }
func AbortWithResponse(c *gin.Context) func(statusCode int, errCode string, args ...interface{}) *gin.Error {
	return func(statusCode int, errCode string, args ...interface{}) *gin.Error {
		errMsg := GetErrorMessage(errCode)

		if len(args) == 1 {
			errMsg = args[0].(string)
		}

		return c.AbortWithError(
			statusCode,
			NewErr(
				errCode,
				errMsg,
			),
		)
	}
}
