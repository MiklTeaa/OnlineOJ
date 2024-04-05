package main

import (
	"context"
	"os/exec"
	"strings"

	"code-platform/api/grpc/ide/pb"
	"code-platform/pkg/errorx"
	"code-platform/pkg/osx"
	"code-platform/pkg/strconvx"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const listContainersNamesFormat = `{{.Names}}`

func (i *IDEServer) GetContainerNames(ctx context.Context, _ *pb.Empty) (*pb.GetContainerNamesResponse, error) {
	cmd := exec.CommandContext(ctx, "docker", "ps", "--format", listContainersNamesFormat)
	stdout, stderr, err := osx.CommandOutput(ctx, cmd)
	if err != nil {
		i.Logger.Errorf(err, "docker ps --format by name failed with stderr %s", string(stderr))
		return nil, status.Error(codes.Internal, err.Error())
	}

	names := strings.Split(strconvx.BytesToString(stdout), "\n")
	infos := make([]*pb.GetContainerNamesResponse_ContainerNameInfo, 0, len(names))
	for _, name := range names {
		if name == "" {
			continue
		}

		labID, studentID, teacherInfo, err := containerNameToIDs(name)
		if err != nil {
			i.Logger.Errorf(err, "convert container name %q failed", name)
			return nil, status.Error(codes.Internal, err.Error())
		}

		infos = append(infos, &pb.GetContainerNamesResponse_ContainerNameInfo{
			LabId:       labID,
			StudentId:   studentID,
			TeacherInfo: teacherInfo,
		})
	}
	return &pb.GetContainerNamesResponse{Infos: infos[:len(infos):len(infos)]}, nil
}

func (i *IDEServer) RemoveContainer(ctx context.Context, req *pb.RemoveContainerRequest) (*pb.Empty, error) {
	if len(req.ContainerNames) == 0 {
		return &pb.Empty{}, nil
	}
	args := append(append(make([]string, 0, len(req.ContainerNames)+2), "rm", "-f"), req.ContainerNames...)
	cmd := exec.CommandContext(ctx, "docker", args...)
	_, _, err := osx.CommandOutput(ctx, cmd)
	switch err {
	case nil:
	case errorx.ErrWrongCode:
		// 容器不存在？应该不太可能出现
		i.Logger.Debugf("sweater remove containers %v but received exit status 1", req.ContainerNames)
		return nil, status.Errorf(codes.NotFound, "container %v is not found", req.ContainerNames)
	default:
		i.Logger.Errorf(err, "sweater remove container %v failed", req.ContainerNames)
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.Empty{}, nil
}
