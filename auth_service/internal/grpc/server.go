package grpc

import (
	"context"

	"auth_service/internal/repository"
	"auth_service/pkg/jwt"
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

	return &pb.ValidateTokenResponse{
		Valid:  true,
		UserId: claims.UserID.String(),
	}, nil
}

func (s *AuthServer) GetUserEmail(ctx context.Context, req *pb.GetUserEmailRequest) (*pb.GetUserEmailResponse, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return &pb.GetUserEmailResponse{
			Error: "invalid user_id format",
		}, nil
	}

	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return &pb.GetUserEmailResponse{
			Error: "user not found",
		}, nil
	}

	return &pb.GetUserEmailResponse{
		Email: user.Email,
	}, nil
}
