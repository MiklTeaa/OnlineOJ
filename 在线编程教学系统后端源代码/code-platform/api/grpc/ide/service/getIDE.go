package main

import (
	"context"
	"fmt"
	"net"

	"code-platform/api/grpc/ide/pb"
	"code-platform/service/ide/define"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (i *IDEServer) GetIDEForStudent(ctx context.Context, req *pb.GetIDEForStudentRequest) (*pb.GetIDEResponse, error) {
	containerName := define.GetContainerNameForStudent(req.LabId, req.StudentId)
	mountWorkSpace := getMountWorkSpace(req.LabId, req.StudentId)
	return i.getIDE(ctx, containerName, mountWorkSpace, req.CanEdit, int8(req.Language))
}

func (i *IDEServer) GetIDEForTeacher(ctx context.Context, req *pb.GetIDEForTeacherRequest) (*pb.GetIDEResponse, error) {
	containerName := define.GetContainerNameForTeacher(req.LabId, req.StudentId, req.TeacherId)
	mountWorkSpace := getMountWorkSpace(req.LabId, req.StudentId)
	return i.getIDE(ctx, containerName, mountWorkSpace, false, int8(req.Language))
}

func (i *IDEServer) getIDE(ctx context.Context, containerName, mountWorkSpace string, canEdit bool, language int8) (*pb.GetIDEResponse, error) {
	if isContainerAlive(ctx, containerName) {
		port, err := getContainerPort(ctx, containerName)
		if err != nil {
			i.Logger.Errorf(err, "getContainerPort failed")
			return nil, status.Error(codes.Internal, err.Error())
		}
		token, err := getContainerToken(ctx, containerName)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		i.Logger.Debugf("return port[%d] directly for active container name %q", port, containerName)
		return &pb.GetIDEResponse{Port: uint32(port), IsReused: true, Token: token}, nil
	} else if isContainerStop(ctx, containerName) {
		// 容器已停止
		port, err := getContainerPort(ctx, containerName)
		if err != nil {
			i.Logger.Errorf(err, "getContainerPort failed")
			return nil, status.Error(codes.Internal, err.Error())
		}

		// 端口是否已被占用
		listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err == nil {
			listener.Close()
			if err = startContainer(ctx, containerName); err == nil {
				i.Logger.Debugf("return port[%d] directly for active container name %q", port, containerName)
				token, err := getContainerToken(ctx, containerName)
				if err != nil {
					return nil, status.Error(codes.Internal, err.Error())
				}
				return &pb.GetIDEResponse{Port: uint32(port), Token: token}, nil
			}
			i.Logger.Errorf(err, "start container failed for %q", containerName)
		}
		if err := removeContainer(ctx, containerName); err != nil {
			i.Logger.Errorf(err, "remove container failed for %q", containerName)
			// return nil, status.Error(codes.Internal, err.Error())
		}
	}

	imageName := define.GetImageName(language)
	port, token, err := i.runTheiaContainer(ctx, imageName, containerName, mountWorkSpace, canEdit)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.GetIDEResponse{Port: uint32(port), Token: token}, nil
}
