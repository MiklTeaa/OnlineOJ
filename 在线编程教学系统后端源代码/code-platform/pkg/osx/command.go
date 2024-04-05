package osx

import (
	"bytes"
	"context"
	"os"
	"os/exec"

	"code-platform/pkg/errorx"
)

var killErrString = "signal: " + os.Kill.String()

func CommandOutput(ctx context.Context, c *exec.Cmd) (output []byte, errOutput []byte, err error) {
	var (
		stdout bytes.Buffer
		stderr bytes.Buffer
	)
	c.Stdout = &stdout
	c.Stderr = &stderr

	err = c.Run()
	if err == nil {
		return stdout.Bytes(), stderr.Bytes(), nil
	}

	if exitError, ok := err.(*exec.ExitError); ok {
		switch exitError.ExitCode() {
		// fallback
		case 136, 139:
			fallthrough
		// exit status 1 即代码执行错误
		case 1:
			return nil, stderr.Bytes(), errorx.ErrWrongCode
		// exit status 137 即容器内存超限
		case 137:
			return nil, nil, errorx.ErrOOMKilled
		}
	}

	switch {
	case err.Error() == killErrString:
		return nil, nil, errorx.ErrContextCancel
	case err == ctx.Err():
		return nil, nil, err
	default:
		return nil, stderr.Bytes(), err
	}
}
