package main

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"code-platform/api/grpc/monaco/pb"
	"code-platform/config"
	"code-platform/pkg/errorx"
	"code-platform/pkg/osx"
	"code-platform/pkg/parallelx"
	"code-platform/pkg/randx"
	"code-platform/pkg/strconvx"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (m *MonacoServer) ExecCode(ctx context.Context, req *pb.ExecCodeRequest) (*pb.ExecCodeResponse, error) {
	language := int8(req.Language)
	code := req.Code
	imageName := getImageName(language)

	// rand code 加以混淆，防止同一时刻同时产生多个容器名
	uuid, err := randx.NewRandCode(6)
	if err != nil {
		m.Logger.Error(err, "generate new code failed")
		return nil, status.Error(codes.Internal, err.Error())
	}
	containerName := fmt.Sprintf("mymonaco-%d-%d-%s", language, time.Now().UnixNano(), uuid)

	commandInContainer := getCommand(language, code)

	/*
		单容器最高 35% CPU 占用率
		默认最大内存 100M
		交换内存后最多使用 300M
	*/
	commandInHost := fmt.Sprintf(`docker run --rm=true --cpus=0.35 --memory=100m --memory-swap=300m --name=%s %s sh -c '%s'`, containerName, imageName, commandInContainer)

	dockerRunCommand := exec.CommandContext(ctx, "sh", "-c", commandInHost)
	stdout, stderr, err := osx.CommandOutput(ctx, dockerRunCommand)
	switch err {
	case nil:
	case errorx.ErrOOMKilled:
		return nil, status.Error(codes.OutOfRange, "OOM")
	case ctx.Err(), errorx.ErrContextCancel:
		m.Logger.Debugf("command %q is canceled or deadline", dockerRunCommand)
		// 异步开启任务停止
		parallelx.DoAsyncWithTimeOut(context.TODO(), 30*time.Second, m.Logger, func(ctx context.Context) (err error) {
			// 重试5次
			for i := 0; i < 5; i++ {
				stopCmd := exec.CommandContext(ctx, "docker", "stop", "-t", "3", containerName)
				if err = stopCmd.Run(); err != nil {
					// 如是 exit status 1
					if _, ok := err.(*exec.ExitError); ok {
						err = nil
					} else {
						// 继续重试
						continue
					}
					break
				}
			}
			if err != nil {
				m.Logger.Debugf("docker stop container %q failed for %v", containerName, err)
				return err
			}
			return nil
		})
		return nil, status.Error(codes.Canceled, err.Error())
	case errorx.ErrWrongCode:
		m.Logger.Debugf("docker run container[%q] failed with wrong code %q", containerName, code)
		return &pb.ExecCodeResponse{Tip: strconvx.BytesToString(stderr), Success: false}, nil
	default:
		m.Logger.Errorf(err, "docker run container[%q] failed with code %q with stderr[%s]", containerName, code, string(stderr))
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.ExecCodeResponse{Tip: strconvx.BytesToString(stdout), Success: true}, nil
}

func getImageName(language int8) string {
	languageMap := config.Monaco.GetStringMapString("imageName")
	switch language {
	case 0:
		return languageMap["python3"]
	case 1:
		return languageMap["cpp"]
	default:
		return languageMap["java"]
	}
}

func getCommand(language int8, code string) string {

	codeEscaped := fmt.Sprintf("%q", strings.ReplaceAll(code, "'", `'\''`))
	codeEscaped = strings.ReplaceAll(codeEscaped, `\\'`, `\'`)
	var commandFormat string
	switch language {
	case 0:
		commandFormat = `echo -e %s > solution.py; python3 solution.py;`
	case 1:
		commandFormat = `echo -e %s > solution.cpp; g++ -o result.out solution.cpp && ./result.out;`
	default:
		commandFormat = `echo -e %s > Solution.java; javac Solution.java && java Solution;`
	}

	return fmt.Sprintf(commandFormat, codeEscaped)
}
