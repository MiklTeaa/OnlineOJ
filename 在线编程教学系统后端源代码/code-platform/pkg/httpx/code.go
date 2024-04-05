package httpx

import "fmt"

const (
	ErrCodeFailToGetParams = -20000 + iota
	ErrCodeBadParams
	ErrCodeInvalidLength
	ErrCodeUnsupportFileType

	ErrCodeFailToAuth = 10002

	ErrCodeNotFound = 10003
)

type errCode struct {
	Msg  string `json:"message,omitempty"`
	Code int    `json:"code"`
}

func NewErrCode(code int, format string, values ...interface{}) *errCode {
	msg := format
	if len(values) > 0 {
		msg = fmt.Sprintf(format, values...)
	}
	return &errCode{
		Code: code,
		Msg:  msg,
	}
}

type jsonResponse struct {
	Data interface{} `json:"data,omitempty"`
}

func NewJSONResponse(data interface{}) *jsonResponse {
	return &jsonResponse{
		Data: data,
	}
}
