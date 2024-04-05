package mailx

import (
	"bytes"
	"code-platform/config"
	"fmt"
	"net/smtp"
	"strings"
)

var defaultLoginAuth *LoginAuth

var (
	userName             string
	defaultFromEmail     string
	defaultFromEmailInfo string
	authCode             string
	smtpAddr             string
)

type MailConfig struct {
	UserName  string `yaml:"userName"`
	Email     string `yaml:"email"`
	EmailInfo string `yaml:"emailInfo"`
	AuthCode  string `yaml:"authCode"`
	SMTPAddr  string `yaml:"smtpAddr"`
}

func init() {
	var m MailConfig
	if err := config.Mail.Unmarshal(&m); err != nil {
		panic(err)
	}
	userName, defaultFromEmail, defaultFromEmailInfo, authCode, smtpAddr = m.UserName, m.Email, m.EmailInfo, m.AuthCode, m.SMTPAddr

	defaultLoginAuth = newLoginAuth(userName, authCode)
}

var (
	headersLayout = []string{"FROM: %s", "TO: %s", "SUBJECT: %s", "MIME-VERSION: 1.0", "Content-Type: text/html; CHARSET=UTF-8"}
	headerLayout  = strings.Join(headersLayout, "\r\n")
)

func SendEmail(to string, subject string, content []byte) error {
	header := fmt.Sprintf(headerLayout, defaultFromEmailInfo, to, subject)
	buf := bytes.NewBuffer(nil)
	buf.Grow(len(header) + len(content) + 4)

	buf.WriteString(header)
	buf.WriteString("\r\n\r\n")
	buf.Write(content)

	if err := smtp.SendMail(smtpAddr, defaultLoginAuth, defaultFromEmail, []string{to}, buf.Bytes()); err != nil {
		return ToEmailError(err)
	}
	return nil
}
