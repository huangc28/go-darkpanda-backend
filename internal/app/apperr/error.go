package apperr

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
