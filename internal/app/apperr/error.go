package apperr

func mergeMaps(maps ...map[string]string) map[string]string {
	mm := make(map[string]string)

	for _, m := range maps {
		for k, v := range m {
			mm[k] = v
		}
	}

	return mm
}

type Error struct {
	ErrCode string `json:"err_code"`
	ErrMsg  string `json:"err_msg"`
}

var MasterErrorMessageMap map[string]string = make(map[string]string)

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
	if len(MasterErrorMessageMap) == 0 {
		MasterErrorMessageMap = mergeMaps(
			AuthErrCodeMsgMap,
			InquiryErrCodeMsgMap,
			ChatErrorMessageMap,
			ImageErrCodeMap,
			ServiceErrorMessageMap,
		)
	}

	message := ""

	if msg, exists := MasterErrorMessageMap[code]; exists {
		message = msg
	}

	return message
}
