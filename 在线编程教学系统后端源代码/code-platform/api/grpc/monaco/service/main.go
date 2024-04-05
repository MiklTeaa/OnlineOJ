package main

import (
	"net"

	"code-platform/api/grpc/monaco/pb"
	"code-platform/config"
	"code-platform/log"

	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"google.golang.org/grpc"
)

type MonacoServer struct {
	Logger *log.Logger
}

func NewMonacoServer(logger *log.Logger) *MonacoServer {
	return &MonacoServer{
		Logger: logger,
	}
}

var _ pb.MonacoServerServiceServer = (*MonacoServer)(nil)

func main() {
	server := grpc.NewServer(grpc.UnaryInterceptor(
		grpc_recovery.UnaryServerInterceptor(),
	))
	monacoServer := NewMonacoServer(log.Sub("ide_server"))
	pb.RegisterMonacoServerServiceServer(server, monacoServer)

	port := config.MonacoServer.GetString("port")
	address := ":" + port
	conn, err := net.Listen("tcp", address)
	if err != nil {
		panic(err)
	}

	if err := server.Serve(conn); err != nil {
		panic(err)
	}
}
