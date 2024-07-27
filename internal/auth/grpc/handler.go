package grpcapp

import (
	"context"
	"errors"

	authv1 "github.com/SmoothWay/gophkeeper/api/gen"
	"github.com/SmoothWay/gophkeeper/internal/auth/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Auth interface {
	Login(ctx context.Context, email string, password string, appID int) (token string, err error)
	Register(ctx context.Context, email string, password string) (userID int64, err error)
	Close()
}

type Server struct {
	authv1.UnimplementedAuthServer
	auth Auth
}

func Register(gRPC *grpc.Server, auth Auth) {
	authv1.RegisterAuthServer(gRPC, &Server{auth: auth})
}

func (s *Server) Register(ctx context.Context, in *authv1.RegisterRequest) (*authv1.RegisterResponse, error) {
	uid, err := s.auth.Register(ctx, in.GetEmail(), in.GetPassword())
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidData):
			return nil, status.Error(codes.InvalidArgument, err.Error())
		case errors.Is(err, service.ErrUserExists):
			return nil, status.Error(codes.AlreadyExists, "user already exists")
		default:
			return nil, status.Error(codes.Internal, "failed to register user")
		}
	}
	return &authv1.RegisterResponse{UserId: uid}, nil
}

func (s *Server) Login(ctx context.Context, in *authv1.LoginRequest) (*authv1.LoginResponse, error) {
	token, err := s.auth.Login(ctx, in.GetEmail(), in.GetPassword(), int(in.GetAppId()))
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			return nil, status.Error(codes.InvalidArgument, "invalid")
		}
		return nil, status.Error(codes.Internal, "failed to login")
	}
	return &authv1.LoginResponse{
		Token: token,
	}, nil
}
