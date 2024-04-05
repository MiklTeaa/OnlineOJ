package main

import (
	"context"
	"io"
	"os/exec"

	"code-platform/api/grpc/ide/pb"
)

// StopAllIDE : close all theia containers
func (i *IDEServer) StopAllIDE(ctx context.Context, empty *pb.Empty) (*pb.Empty, error) {
	containerIDs := exec.CommandContext(ctx, "docker", "ps", "-f", "name=mytheia*", "-q")
	stoper := exec.CommandContext(ctx, "xargs", "docker", "stop", "-t", "3")

	r, err := stoper.StdinPipe()
	if err != nil {
		return nil, err
	}

	w, err := containerIDs.StdoutPipe()
	if err != nil {
		return nil, err
	}

	err = containerIDs.Start()
	if err != nil {
		return nil, err
	}

	data, err := io.ReadAll(w)
	if err != nil {
		return nil, err
	}

	err = containerIDs.Wait()
	if err != nil {
		return nil, err
	}

	err = stoper.Start()
	if err != nil {
		return nil, err
	}

	_, err = r.Write(data)
	if err != nil {
		return nil, err
	}

	err = r.Close()
	if err != nil {
		return nil, err
	}

	err = stoper.Wait()
	if err != nil {
		return nil, err
	}

	return &pb.Empty{}, nil
}
