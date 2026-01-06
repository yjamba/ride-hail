package service

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"ride-hail/internal/auth/domain/models"
	"ride-hail/internal/auth/domain/ports"
)

var ErrInvalidEmail = errors.New("invalid email")

type AuthService struct {
	tokenService *TokenService

	userRepo   ports.UserRepository
	driverRepo ports.DriverRepository
	txManager  ports.TransactionManager

	secretKey []byte
}

func NewAuthService(tokenService *TokenService, userRepo ports.UserRepository, driverRepo ports.DriverRepository, txManager ports.TransactionManager) *AuthService {
	return &AuthService{
		tokenService: tokenService,
		userRepo:     userRepo,
		driverRepo:   driverRepo,
		txManager:    txManager,
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

	tokenPair, err := s.tokenService.GenerateTokenPair(id, user.Role)
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

	var driverID string
	var tokerPair *models.TokenPair

	err := s.txManager.WithTx(ctx, func(txCtx context.Context) error {
		user := &models.User{
			Role:     "DRIVER",
			Email:    email,
			Password: password,
		}

		id, err := s.userRepo.Save(txCtx, user)
		if err != nil {
			slog.Error("failed to save user for driver signup", "error", err.Error())
			return err
		}

		driver.UserID = id
		driverID, err = s.driverRepo.Save(txCtx, driver)
		if err != nil {
			slog.Error("failed to save driver for driver signup", "error", err.Error())
			return err
		}

		tokerPair, err = s.tokenService.GenerateTokenPair(driverID, user.Role)
		return err
	})
	if err != nil {
		return "", nil, err
	}

	return driverID, tokerPair, nil
}

func (s *AuthService) RefreshAuthToken(refreshToken string) (string, error) {
	return "", nil
}
