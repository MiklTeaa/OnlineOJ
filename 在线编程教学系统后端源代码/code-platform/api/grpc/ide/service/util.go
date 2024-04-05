package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"code-platform/pkg/osx"
	"code-platform/pkg/randx"
	"code-platform/pkg/strconvx"
	"code-platform/service/ide/define"

	"github.com/bytedance/sonic"
)

func (i *IDEServer) runTheiaContainer(ctx context.Context, imageName string, containerName string, mountWorkSpace string, canEdit bool) (uint16, string, error) {
	port := getAvailablePort()

	var readOnlyOpt string
	if !canEdit {
		readOnlyOpt = "ro"
	} else {
		readOnlyOpt = "rw"
	}

	token, err := randx.NewRandCode(8)
	if err != nil {
		return 0, "", err
	}

	/*
		单容器最高 15% CPU 占用率
		默认最大内存 500M
		交换内存后最多使用 900M
	*/

	dockerRunCommand := fmt.Sprintf(
		`run -d -u root --restart=always --cpus=0.38 --memory=500m --memory-swap=900m -e token=%s -p %d:10443 -v %s:/home/project:%s --name=%s %s`,
		token,
		port,
		mountWorkSpace,
		readOnlyOpt,
		containerName,
		imageName,
	)

	cmd := exec.CommandContext(ctx, "docker", strings.Fields(dockerRunCommand)...)

	if _, stderr, err := osx.CommandOutput(ctx, cmd); err != nil {
		i.Logger.Errorf(err, "Failed to exec command %q for %s", cmd.String(), string(stderr))
		return 0, "", err
	}

	time.Sleep(2 * time.Second)
	return port, token, nil
}

func getAvailablePort() uint16 {
	rand.Seed(time.Now().UnixNano())
	for {
		port := rand.Intn(2000) + 30000
		listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err != nil {
			continue
		}

		listener.Close()
		return uint16(port)
	}
}

func startContainer(ctx context.Context, containerName string) error {
	cmd := exec.CommandContext(ctx, "docker", "start", containerName)
	if _, stderr, err := osx.CommandOutput(ctx, cmd); err != nil {
		return errors.New(err.Error() + "\n" + string(stderr))
	}
	return nil
}

func removeContainer(ctx context.Context, containerName string) error {
	cmd := exec.CommandContext(ctx, "docker", "rm", "-f", containerName)
	if _, stderr, err := osx.CommandOutput(ctx, cmd); err != nil {
		return errors.New(err.Error() + "\n" + string(stderr))
	}
	return nil
}

func isContainerAlive(ctx context.Context, containerName string) bool {
	cmd := exec.CommandContext(ctx, "docker", "ps", "--filter", fmt.Sprintf("name=%s", containerName))

	stdout, _, err := osx.CommandOutput(ctx, cmd)
	if err != nil {
		return false
	}
	return bytes.Contains(stdout, strconvx.StringToBytes(containerName))
}

func isContainerStop(ctx context.Context, containerName string) bool {
	cmd := exec.CommandContext(ctx, "docker", "ps", "-a", "--filter", "status=exited", "--filter", "name="+containerName)
	stdout, _, err := osx.CommandOutput(ctx, cmd)
	if err != nil {
		return false
	}
	return bytes.Contains(stdout, strconvx.StringToBytes(containerName))
}

func getContainerPort(ctx context.Context, containerName string) (int, error) {
	cmd := exec.CommandContext(ctx, "docker", "inspect", "-f", "'{{json .HostConfig.PortBindings}}'", containerName)
	stdout, stderr, err := osx.CommandOutput(ctx, cmd)
	if err != nil {
		return 0, fmt.Errorf("failed to find port for container %q because of error %s", containerName, err.Error()+"\n"+string(stderr))
	}

	stdout = bytes.TrimSuffix(stdout, []byte{'\n'})
	stdout = bytes.TrimPrefix(stdout, []byte{'\''})
	stdout = bytes.TrimSuffix(stdout, []byte{'\''})

	type hostInfo struct {
		HostIP   string `json:"HostIP"`
		HostPort string `json:"HostPort"`
	}

	m := make(map[string][]*hostInfo)

	if err := sonic.Unmarshal(stdout, &m); err != nil {
		return 0, fmt.Errorf("unable to unmarshel for stdout %s because of error %s", string(stdout), err.Error())
	}

	for _, v := range m {
		if len(v) > 0 {
			hostPort, err := strconv.Atoi(v[0].HostPort)
			if err != nil {
				return 0, fmt.Errorf("unable to atoi for hostPort %s", v[0].HostPort)
			}
			return hostPort, nil
		}
	}

	return 0, fmt.Errorf("failed to find hostPort in map %v", m)
}

func getMountWorkSpace(labID, studentID uint64) string {
	return filepath.Join(define.InitBasePath, "codespaces", fmt.Sprintf("workspace-%d", labID), strconv.FormatUint(studentID, 10))
}

var tokenRegexpCompiled = regexp.MustCompile(`token: (\w+)`)

func getContainerToken(ctx context.Context, containerName string) (string, error) {
	cmd := exec.CommandContext(ctx, "docker", "logs", containerName)
	stdout, stderr, err := osx.CommandOutput(ctx, cmd)
	if err != nil {
		return "", fmt.Errorf("failed to find port for container %q because of error %s", containerName, err.Error()+"\n"+string(stderr))
	}

	data := tokenRegexpCompiled.FindSubmatch(stdout)
	return strconvx.BytesToString(data[1]), nil
}
