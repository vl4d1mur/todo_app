package grpc

import (
	"context"
	"fmt"
	"time"

	"task_service/pkg/log"
	"task_service/pkg/pb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type AuthClient struct {
	client pb.AuthServiceClient
	conn   *grpc.ClientConn
}

func NewAuthClient(addr string) (*AuthClient, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to auth_service: %w", err)
	}

	log.Logger.Info().Str("addr", addr).Msg("Connected to auth_service via gRPC")

	return &AuthClient{
		client: pb.NewAuthServiceClient(conn),
		conn:   conn,
	}, nil
}

func (c *AuthClient) ValidateToken(ctx context.Context, token string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	resp, err := c.client.ValidateToken(ctx, &pb.ValidateTokenRequest{Token: token})
	if err != nil {
		return "", fmt.Errorf("gRPC ValidateToken error: %w", err)
	}

	if !resp.Valid {
		return "", fmt.Errorf("invalid token %s", resp.Error)
	}

	return resp.UserId, nil
}

func (c *AuthClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
