package internal

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	// "google.golang.org/protobuf/types/known/timestamppb"
	"alarm/pkg/proto"
	"log"
	"net"
)

// 對client設置之timeout與canceled正確回應
func contextError(ctx context.Context) error {
	switch ctx.Err() {
	case context.Canceled:
		Info.Println("request is canceled")
		return status.Error(codes.Canceled, "request is canceled")
	case context.DeadlineExceeded:
		Info.Println("deadline is exceeded")
		return status.Error(codes.DeadlineExceeded, "deadline is exceeded")
	default:
		return nil
	}
}

type Server struct{}

func (s *Server) InitAlarmRules(ctx context.Context, in *proto.Empty) (*proto.SQLresponse, error) {
	err := queries.TruncateRules(ctx)
	info := "initial success"
	if err != nil {
		info = "sql truncate fail"
	}
	err = InitSQLAlarmRulesFromCSV("./alarm.csv")
	if err != nil {
		info = "err"
	}
	return &proto.SQLresponse{Info: info}, nil
}

func GrpcServer() {
	// Starts a TCP server listening on port 55555 and handles any errors.
	l, err := net.Listen("tcp", ":55555")
	// The gRPC server will use it.
	if err != nil {
		log.Fatalf("failed to listen for tcp: %s", err)
	}
	grpcServer := grpc.NewServer() // Creates a gRPC server and handles requests over the TCP connection
	// TODO後續hot data在這邊再註冊一個即可
	proto.RegisterAlarmRulesManagerServer(grpcServer, &Server{})
	err = grpcServer.Serve(l)
	if err != nil {
		log.Fatalf("failed to create gRPC server")
	}
}
