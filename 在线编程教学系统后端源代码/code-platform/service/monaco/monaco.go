package monaco

import (
	"context"
	"time"

	"code-platform/api/grpc/monaco/pb"
	"code-platform/config"
	"code-platform/log"
	"code-platform/pkg/errorx"
	"code-platform/repository"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func NewMonacoClient() pb.MonacoServerServiceClient {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	port := config.MonacoServer.GetString("port")
	conn, err := grpc.DialContext(ctx, "localhost:"+port, grpc.WithInsecure(), grpc.WithBlock())

	switch err {
	case nil:
	case ctx.Err():
		log.Sub("grpc.checkCode").Errorf(err, "dial to grpc server timeout")
		fallthrough
	default:
		panic(err)
	}

	return pb.NewMonacoServerServiceClient(conn)
}

type MonacoService struct {
	Dao          *repository.Dao
	Logger       *log.Logger
	MonacoClient pb.MonacoServerServiceClient
}

func NewMonacoService(dao *repository.Dao, logger *log.Logger, monacoClient pb.MonacoServerServiceClient) *MonacoService {
	return &MonacoService{
		Dao:          dao,
		Logger:       logger,
		MonacoClient: monacoClient,
	}
}

func (m *MonacoService) ExecCode(ctx context.Context, language int8, code string) (string, error) {
	resp, err := m.MonacoClient.ExecCode(ctx, &pb.ExecCodeRequest{
		Language: uint32(language),
		Code:     code,
	})

	if err == nil {
		if !resp.Success {
			return resp.Tip, errorx.ErrWrongCode
		}
		return resp.Tip, nil
	}

	switch status.Code(err) {
	case codes.Canceled, codes.DeadlineExceeded:
		return "", errorx.ErrContextCancel
	case codes.OutOfRange:
		return "", errorx.ErrOOMKilled
	default:
		return "", errorx.InternalErr(err)
	}
}
