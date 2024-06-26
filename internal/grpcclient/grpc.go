package grpcclient

import (
	"context"
	"fmt"

	authv1 "github.com/SmoothWay/gophkeeper/api/gen"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type GRPCClient struct {
	conn   *grpc.ClientConn
	client authv1.AuthClient
}

func NewGRPCClient(address string, caCertFile string) (*GRPCClient, error) {
	// tlsCredentials, err := tls.LoadTLSCredentials(caCertFile)
	// if err != nil {
	// 	log.Fatal("cannot load TLS credentials ", err)
	// }

	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &GRPCClient{
		conn:   conn,
		client: authv1.NewAuthClient(conn),
	}, nil
}

func (c *GRPCClient) Register(ctx context.Context, login, password string) error {
	req := authv1.RegisterRequest{Email: login, Password: password}
	_, err := c.client.Register(ctx, &req)
	if err != nil {
		if e, ok := status.FromError(err); ok {
			switch e.Code() {
			case codes.AlreadyExists:
				return fmt.Errorf("user with email %s already registered", login)
			case codes.InvalidArgument:
				return fmt.Errorf("invalid login or password")
			default:
				return fmt.Errorf("something went wrong, please try again later")
			}
		}
		return fmt.Errorf("something went wrong, please try again later")
	}
	return nil
}

func (c *GRPCClient) Login(ctx context.Context, login, password string) (string, error) {
	req := authv1.LoginRequest{Email: login, Password: password, AppId: 1}

	res, err := c.client.Login(ctx, &req)
	if err != nil {
		if e, ok := status.FromError(err); ok {
			switch e.Code() {
			case codes.InvalidArgument:
				return "", fmt.Errorf("invalid login or argument")
			default:
				return "", fmt.Errorf("something went wrong")
			}
		}
		return "", fmt.Errorf("something went wrong")
	}
	return res.Token, nil
}

func (c *GRPCClient) Stop() {
	_ = c.conn.Close()
}
