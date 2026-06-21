package grpc

import (
	"context"
	"fmt"
	"time"

	"notifier_service/pkg/log"
	"notifier_service/pkg/pb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/status"
)

type AuthClient struct {
	client pb.AuthServiceClient
	conn   *grpc.ClientConn
}

func NewAuthClient(addr string) (*AuthClient, error) {
	kacp := keepalive.ClientParameters{
		Time:                30 * time.Second,
		Timeout:             10 * time.Second,
		PermitWithoutStream: true,
	}

	serviceConfig := `{
		"methodConfig": [{
			"name": [{"service": "auth.AuthService"}],
			"retryPolicy": {
				"MaxAttempts": 3,
				"InitialBackoff": "0.2s",
				"MaxBackoff": "1s",
				"BackoffMultiplier": 2,
				"RetryableStatusCodes": ["UNAVAILABLE", "DEADLINE_EXCEEDED"]
			}
		}]
	}`

	conn, err := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithKeepaliveParams(kacp),
		grpc.WithDefaultServiceConfig(serviceConfig),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to auth_service: %w", err)
	}

	client := pb.NewAuthServiceClient(conn)

	if err := warmupConnection(client); err != nil {
		conn.Close()
		return nil, err
	}

	log.Logger.Info().Str("addr", addr).Msg("Connected to auth_service via gRPC")

	return &AuthClient{
		client: client,
		conn:   conn,
	}, nil
}

func warmupConnection(client pb.AuthServiceClient) error {
	for attempts := 0; attempts < 10; attempts++ {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		_, err := client.ValidateToken(ctx, &pb.ValidateTokenRequest{Token: "warmup"})
		cancel()

		if err == nil {
			return nil
		}

		code := status.Code(err)
		if code != codes.Unavailable && code != codes.DeadlineExceeded {
			return nil
		}

		log.Logger.Warn().Int("attempt", attempts+1).Err(err).Msg("gRPC warmup failed, retrying...")
		time.Sleep(2 * time.Second)
	}

	return fmt.Errorf("failed to establish gRPC connection after 10 attempts")
}

func (c *AuthClient) ActivateTelegram(ctx context.Context, code string, chatID int64) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	resp, err := c.client.ActivateTelegram(ctx, &pb.ActivateTelegramRequest{
		Code:   code,
		ChatId: chatID,
	})
	if err != nil {
		return false, fmt.Errorf("gRPC ActivateTelegram error: %w", err)
	}

	return resp.Success, nil
}

func (c *AuthClient) GetUserContacts(ctx context.Context, userID string) (string, int64, bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	resp, err := c.client.GetUserContacts(ctx, &pb.GetUserContactsRequest{UserId: userID})
	if err != nil {
		return "", 0, false, fmt.Errorf("gRPC GetUserContacts error: %w", err)
	}

	if resp.Error != "" {
		return "", 0, false, fmt.Errorf("contacts lookup failed: %s", resp.Error)
	}

	return resp.Email, resp.TelegramChatId, resp.HasTelegram, nil
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
