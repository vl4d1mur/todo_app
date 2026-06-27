package grpc

import (
	"context"

	"auth_service/internal/db/redisConn"
	"auth_service/internal/repository"
	"auth_service/pkg/jwt"
	"auth_service/pkg/log"
	"auth_service/pkg/metrics"
	"auth_service/pkg/pb"

	"github.com/google/uuid"
)

type AuthServer struct {
	pb.UnimplementedAuthServiceServer
	userRepo repository.UserRepository
}

func NewAuthServer(userRepo repository.UserRepository) *AuthServer {
	return &AuthServer{userRepo: userRepo}
}

func (s *AuthServer) ValidateToken(ctx context.Context, req *pb.ValidateTokenRequest) (*pb.ValidateTokenResponse, error) {
	claims, err := jwt.ParseAccess(req.Token)
	if err != nil {
		return &pb.ValidateTokenResponse{
			Valid: false,
			Error: err.Error(),
		}, nil
	}

    blacklisted, err := redisConn.IsBlacklisted(claims.ID)
    if err != nil {
        log.Logger.Warn().Err(err).Msg("Failed to check blacklist")
    } else if blacklisted {
        return &pb.ValidateTokenResponse{
            Valid: false,
            Error: "token is blacklisted",
        }, nil
    }

	return &pb.ValidateTokenResponse{
		Valid:  true,
		UserId: claims.UserID.String(),
	}, nil
}

func (s *AuthServer) ActivateTelegram(ctx context.Context, req *pb.ActivateTelegramRequest) (*pb.ActivateTelegramResponse, error) {
	userIDStr, err := redisConn.GetUserIDByCode(req.Code)
	if err != nil {
		return &pb.ActivateTelegramResponse{
			Success: false,
			Error:   "invalid or expired code",
		}, nil
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return &pb.ActivateTelegramResponse{
			Success: false,
			Error:   "invalid user_id in cache",
		}, nil
	}

	if err := s.userRepo.UpdateTelegramChatID(ctx, userID, req.ChatId); err != nil {
		return &pb.ActivateTelegramResponse{
			Success: false,
			Error:   "failed to save chat_id",
		}, nil
	}

	_ = redisConn.DeleteTelegramCode(req.Code)

	metrics.TelegramActivations.Inc()
	return &pb.ActivateTelegramResponse{Success: true}, nil
}

func (s *AuthServer) GetUserContacts(ctx context.Context, req *pb.GetUserContactsRequest) (*pb.GetUserContactsResponse, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return &pb.GetUserContactsResponse{Error: "invalid user_id"}, nil
	}

	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return &pb.GetUserContactsResponse{Error: "user not found"}, nil
	}

	resp := &pb.GetUserContactsResponse{
		Email: user.Email,
	}
	if user.TelegramChatID != nil {
		resp.TelegramChatId = *user.TelegramChatID
		resp.HasTelegram = true
	}
	return resp, nil
}
