package mailx_test

import (
	"testing"

	"code-platform/config"
	"code-platform/pkg/mailx"

	"github.com/stretchr/testify/require"
)

func TestSendEmail(t *testing.T) {
	var m mailx.MailConfig
	err := config.Mail.Unmarshal(&m)
	require.NoError(t, err)
	err = mailx.SendEmail(m.Email, "试一下先", []byte("成功没？"))
	require.NoError(t, err)
}
