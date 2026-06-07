package grpc

import (
	"context"

	"auth_service/pkg/jwt"
	"auth_service/pkg/pb"
)

type AuthServer struct {
	pb.UnimplementedAuthServiceServer
}

func NewAuthServer() *AuthServer { 
	return &AuthServer{}
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
		Valid: true,
		UserId: claims.UserID.String(),
	}, nil
}

