package mailx

import (
	"strings"

	"code-platform/pkg/errorx"
)

func ToEmailError(err error) error {
	errString := err.Error()
	if strings.Contains(errString, "550") {
		return errorx.ErrMailUserNotFound
	}
	return err
}
