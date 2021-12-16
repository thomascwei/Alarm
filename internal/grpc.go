package internal

import (
	"context"
	"errors"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
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

// gPRC的middleware
func unaryInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	Trace.Println("--> unary interceptor: ", info.FullMethod)
	md, _ := metadata.FromIncomingContext(ctx)
	authStringSlice := md.Get("Authorization")
	// FIXME若有需要驗證token, error response格式須修正
	if false {
		if len(authStringSlice) == 0 {
			return nil, errors.New("auth fail")
		}
		if md.Get("Authorization")[0] != "thomas" {
			return nil, errors.New("auth fail")
		}
	}

	return handler(ctx, req)
}

type Server struct{}

func (s *Server) InitAlarmRules(ctx context.Context, in *proto.Empty) (*proto.SQLresponse, error) {
	info := "initial success"
	err := queries.TruncateRules(ctx)
	if err != nil {
		info = "sql truncate fail"
	}
	err = InitSQLAlarmRulesFromCSV("./alarm.csv")
	if err != nil {
		info = "err"
	}
	if err := contextError(ctx); err != nil {
		return nil, err
	}
	return &proto.SQLresponse{Info: info}, nil
}

// TODO接收hotdata
func (s *Server) Insert(ctx context.Context, input *proto.HotDataRequest) (*proto.HotDataResponse, error) {
	Trace.Println(input.ObjectID, input.Value)
	return &proto.HotDataResponse{StatusOK: true, Message: "ok"}, nil
}

func GrpcServer() {
	// Starts a TCP server listening on port 55555 and handles any errors.
	l, err := net.Listen("tcp", ":55555")
	// The gRPC server will use it.
	if err != nil {
		log.Fatalf("failed to listen for tcp: %s", err)
	}
	// Creates a gRPC server and handles requests over the TCP connection
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(unaryInterceptor),
	)
	// 多個server可再此註冊
	proto.RegisterAlarmRulesManagerServer(grpcServer, &Server{})
	proto.RegisterHotDataReceiverServer(grpcServer, &Server{})
	err = grpcServer.Serve(l)
	if err != nil {
		log.Fatalf("failed to create gRPC server")
	}
}
