package errorx

type myError struct {
	msg  string
	code Code
}

func (m *myError) Error() string {
	return m.msg
}

var _ error = (*myError)(nil)

var (
	ErrFailToAuth              = New(CodeNoAuth, "Fail to auth")
	ErrIsNotFound              = New(CodeNotFound, "Record is not found")
	ErrPersonalInfoNotComplete = New(CodeForbidden, "Personal info is not complete")
	ErrPersonalInfoInvalid     = New(CodeForbidden, "Personal info is invalid")
	ErrUnsupportFileType       = New(CodeForbidden, "Unsupported file type")
	ErrRedisKeyNil             = New(CodeNotFound, "The redis key is not found")
	ErrMySQLDuplicateKey       = New(CodeForbidden, "mysql unique key duplicated")
	ErrContextCancel           = New(CodeForbidden, "context canceled")
	ErrWrongCode               = New(CodeInternal, "code is too wrong to exec")
	ErrNotExpire               = New(CodeForbidden, "session not expire")
	// ErrMailUserNotFound 发送的邮箱用户并不存在
	ErrMailUserNotFound = New(CodeNotFound, "email user is not found")
	ErrOOMKilled        = New(CodeForbidden, "OOM")
)

func New(code Code, msg string) error {
	return &myError{
		msg:  msg,
		code: code,
	}
}

func InternalErr(err error) error {
	if err == nil {
		return nil
	}
	return &myError{
		msg:  err.Error(),
		code: CodeInternal,
	}
}
