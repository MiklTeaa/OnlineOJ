package mailx

import (
	"net/smtp"

	"code-platform/pkg/strconvx"
)

type LoginAuth struct {
	userName string
	password string
}

func (l *LoginAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	return "LOGIN", strconvx.StringToBytes(l.userName), nil
}

func (l *LoginAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	switch strconvx.BytesToString(fromServer) {
	case "Username:":
		return strconvx.StringToBytes(l.userName), nil
	case "Password:":
		return strconvx.StringToBytes(l.password), nil
	}
	return nil, nil
}

func newLoginAuth(userName, password string) *LoginAuth {
	return &LoginAuth{
		userName: userName,
		password: password,
	}
}
