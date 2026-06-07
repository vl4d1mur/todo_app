package jwt

import (
	"errors"
	"time"

	"auth_service/internal/config"
	"auth_service/internal/models"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token expired")
)

const AccessTokenTTL = 15 * time.Minute
const RefreshTokenTTL = 30 * 24 * time.Hour

type Claims struct {
	UserID uuid.UUID `json:"user_id"`
	jwt.RegisteredClaims
}

func GenerateAccess(user models.User) (string, error) {
	return generateToken(user.ID, AccessTokenTTL, config.JwtSecret)
}

func GenerateRefresh(user models.User) (string, error) {
	return generateToken(user.ID, RefreshTokenTTL, config.JwtRefreshSecret)
}

func ParseAccess(tokenString string) (*Claims, error) {
	return parseToken(tokenString, config.JwtSecret)
}

func ParseRefresh(tokenString string) (*Claims, error) {
	return parseToken(tokenString, config.JwtRefreshSecret)
}

func generateToken(userID uuid.UUID, ttl time.Duration, secret []byte) (string, error) {
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.New().String(),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

func parseToken(tokenString string, secret []byte) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return secret, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}
	return claims, nil
}
