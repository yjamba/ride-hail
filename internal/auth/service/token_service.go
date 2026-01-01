package service

import (
	"log/slog"
	"time"

	"ride-hail/internal/auth/domain/models"

	"github.com/golang-jwt/jwt/v5"
)

type TokenService struct {
	secretKey []byte
}

func NewTokenService(secretKey []byte) *TokenService {
	return &TokenService{
		secretKey: secretKey,
	}
}

func (s *TokenService) GenerateTokenPair(userID, role string) (*models.TokenPair, error) {
	now := time.Now()
	accessClaims := jwt.MapClaims{
		"iat":     now.Unix(),
		"exp":     now.Add(time.Minute * 15).Unix(),
		"user_id": userID,
		"role":    role,
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	signedAccess, err := accessToken.SignedString(s.secretKey)
	if err != nil {
		slog.Error("failed to sign access token", "error", err.Error())
		return nil, err
	}

	refreshClaims := jwt.MapClaims{
		"iat":     now.Unix(),
		"exp":     now.Add(time.Hour * 24 * 7).Unix(),
		"user_id": userID,
		"type":    "refresh",
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	signedRefresh, err := refreshToken.SignedString(s.secretKey)
	if err != nil {
		slog.Error("failed to sign refresh token", "error", err.Error())
		return nil, err
	}

	return &models.TokenPair{
		AccessToken:  signedAccess,
		RefreshToken: signedRefresh,
	}, nil
}
