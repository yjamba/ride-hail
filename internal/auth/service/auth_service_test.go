package service

import (
	"context"
	"errors"
	"testing"

	"ride-hail/internal/auth/domain/models"
	"ride-hail/internal/auth/domain/ports"
)

type fakeUserRepo struct {
	id     string
	err    error
	called int
	last   *models.User
}

func (f *fakeUserRepo) Save(ctx context.Context, user *models.User) (string, error) {
	f.called++
	f.last = user
	return f.id, f.err
}

type fakeDriverRepo struct {
	id     string
	err    error
	called int
	last   *models.Driver
}

func (f *fakeDriverRepo) Save(ctx context.Context, driver *models.Driver) (string, error) {
	f.called++
	f.last = driver
	return f.id, f.err
}

type fakeTxManager struct {
	called int
	err    error
}

func (f *fakeTxManager) WithTx(ctx context.Context, fn func(context.Context) error) error {
	f.called++
	if f.err != nil {
		return f.err
	}
	return fn(ctx)
}

var _ ports.UserRepository = (*fakeUserRepo)(nil)
var _ ports.DriverRepository = (*fakeDriverRepo)(nil)
var _ ports.TransactionManager = (*fakeTxManager)(nil)

func TestAuthService_Signup_InvalidEmail(t *testing.T) {
	svc := NewAuthService(NewTokenService([]byte("secret")), &fakeUserRepo{}, &fakeDriverRepo{}, &fakeTxManager{})
	_, _, err := svc.Signup(context.Background(), "bad", "pass", "PASSENGER")
	if !errors.Is(err, ErrInvalidEmail) {
		t.Fatalf("expected ErrInvalidEmail, got %v", err)
	}
}

func TestAuthService_Signup_Success(t *testing.T) {
	userRepo := &fakeUserRepo{id: "u1"}
	svc := NewAuthService(NewTokenService([]byte("secret")), userRepo, &fakeDriverRepo{}, &fakeTxManager{})
	id, tokens, err := svc.Signup(context.Background(), "a@b.com", "pass", "PASSENGER")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != "u1" {
		t.Fatalf("expected id u1, got %s", id)
	}
	if tokens == nil || tokens.AccessToken == "" || tokens.RefreshToken == "" {
		t.Fatal("expected token pair to be generated")
	}
	if userRepo.last == nil || userRepo.last.Role != "PASSENGER" {
		t.Fatalf("unexpected user saved: %+v", userRepo.last)
	}
}

func TestAuthService_SignupDriver_InvalidEmail(t *testing.T) {
	svc := NewAuthService(NewTokenService([]byte("secret")), &fakeUserRepo{}, &fakeDriverRepo{}, &fakeTxManager{})
	_, _, err := svc.SignupDriver(context.Background(), "bad", "pass", &models.Driver{})
	if !errors.Is(err, ErrInvalidEmail) {
		t.Fatalf("expected ErrInvalidEmail, got %v", err)
	}
}

func TestAuthService_SignupDriver_TxError(t *testing.T) {
	svc := NewAuthService(NewTokenService([]byte("secret")), &fakeUserRepo{}, &fakeDriverRepo{}, &fakeTxManager{err: errors.New("tx fail")})
	_, _, err := svc.SignupDriver(context.Background(), "a@b.com", "pass", &models.Driver{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestAuthService_SignupDriver_Success(t *testing.T) {
	userRepo := &fakeUserRepo{id: "u1"}
	driverRepo := &fakeDriverRepo{id: "d1"}
	tx := &fakeTxManager{}
	svc := NewAuthService(NewTokenService([]byte("secret")), userRepo, driverRepo, tx)

	driver := &models.Driver{LicenseNumber: "LIC", VehicleType: "ECONOMY"}
	id, tokens, err := svc.SignupDriver(context.Background(), "a@b.com", "pass", driver)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != "d1" {
		t.Fatalf("expected driver id d1, got %s", id)
	}
	if tokens == nil || tokens.AccessToken == "" || tokens.RefreshToken == "" {
		t.Fatal("expected token pair to be generated")
	}
	if tx.called != 1 {
		t.Fatalf("expected tx to be called once, got %d", tx.called)
	}
	if driverRepo.last == nil || driverRepo.last.UserID != "u1" {
		t.Fatalf("expected driver user_id u1, got %+v", driverRepo.last)
	}
}