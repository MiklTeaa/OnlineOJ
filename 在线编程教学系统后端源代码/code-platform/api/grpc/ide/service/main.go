package main

import (
	"net"

	"code-platform/api/grpc/ide/pb"
	"code-platform/config"
	"code-platform/log"

	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"google.golang.org/grpc"
)

type IDEServer struct {
	Logger *log.Logger
}

func NewIDEServer(logger *log.Logger) *IDEServer {
	return &IDEServer{
		Logger: logger,
	}
}

var _ pb.IDEServerServiceServer = (*IDEServer)(nil)

func main() {
	server := grpc.NewServer(grpc.UnaryInterceptor(
		grpc_recovery.UnaryServerInterceptor(),
	))
	ideServer := NewIDEServer(log.Sub("ide_server"))
	pb.RegisterIDEServerServiceServer(server, ideServer)

	port := config.IDEServer.GetString("port")
	address := "localhost:" + port
	conn, err := net.Listen("tcp", address)
	if err != nil {
		panic(err)
	}

	if err := server.Serve(conn); err != nil {
		panic(err)
	}
}
