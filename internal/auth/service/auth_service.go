package service

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"

	"ride-hail/internal/auth/domain/models"
	"ride-hail/internal/auth/domain/ports"

	"github.com/golang-jwt/jwt/v5"
)

var ErrInvalidEmail = errors.New("invalid email")

type AuthService struct {
	userRepo   ports.UserRepository
	driverRepo ports.DriverRepository

	secretKey []byte
}

func NewAuthService(userRepo ports.UserRepository, driverRepo ports.DriverRepository, secretKey []byte) *AuthService {
	return &AuthService{
		userRepo:   userRepo,
		driverRepo: driverRepo,
		secretKey:  secretKey,
	}
}

func (s *AuthService) Signup(ctx context.Context, email, password, role string) (string, *models.TokenPair, error) {
	if role == "" || email == "" || !strings.Contains(email, "@") {
		return "", nil, ErrInvalidEmail
	}

	user := &models.User{
		Role:     role,
		Email:    email,
		Password: password,
	}

	id, err := s.userRepo.Save(ctx, user)
	if err != nil {
		slog.Error("failed to save user for signup", "error", err.Error())
		return "", nil, err
	}

	tokenPair, err := s.generateTokenPair(id, user.Role)
	if err != nil {
		slog.Error("failed to generate token pair for signup", "error", err.Error())
		return "", nil, err
	}

	return id, tokenPair, nil
}

func (s *AuthService) SignupDriver(ctx context.Context, email, password string, driver *models.Driver) (string, *models.TokenPair, error) {
	if email == "" || !strings.Contains(email, "@") {
		return "", nil, ErrInvalidEmail
	}

	user := &models.User{
		Role:     "DRIVER",
		Email:    email,
		Password: password,
	}

	id, err := s.userRepo.Save(ctx, user)
	if err != nil {
		slog.Error("failed to save user for driver signup", "error", err.Error())
		return "", nil, err
	}

	driver.UserID = id
	driverID, err := s.driverRepo.Save(ctx, driver)
	if err != nil {
		slog.Error("failed to save driver for driver signup", "error", err.Error())
		return "", nil, err
	}

	tokerPair, err := s.generateTokenPair(driverID, user.Role)
	if err != nil {
		return "", nil, err
	}
	return driverID, tokerPair, nil
}

func (s *AuthService) RefreshAuthToken(oldToken string) (string, error) {
	return "", nil
}

func (s *AuthService) generateTokenPair(userID, role string) (*models.TokenPair, error) {
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
