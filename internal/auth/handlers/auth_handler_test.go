package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"ride-hail/internal/auth/domain/models"
	"ride-hail/internal/auth/service"
)

func TestDecodeUserRequest_InvalidJSON(t *testing.T) {
	h := NewAuthHandler(nil)
	req := httptest.NewRequest(http.MethodPost, "/signup", strings.NewReader("{"))

	_, err := h.decodeUserRequest(req)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestDecodeUserRequest_Success(t *testing.T) {
	h := NewAuthHandler(nil)
	body := `{"email":"a@b.com","password":"secret"}`
	req := httptest.NewRequest(http.MethodPost, "/signup", strings.NewReader(body))

	got, err := h.decodeUserRequest(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Email != "a@b.com" {
		t.Fatalf("expected email a@b.com, got %s", got.Email)
	}
	if got.Password != "secret" {
		t.Fatalf("expected password secret, got %s", got.Password)
	}
}

func TestSendResponse_SetsCookieAndBody(t *testing.T) {
	h := NewAuthHandler(nil)
	rr := httptest.NewRecorder()

	pairs := &models.TokenPair{AccessToken: "access", RefreshToken: "refresh"}
	h.sendResponse(rr, "id-1", pairs)

	res := rr.Result()
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, res.StatusCode)
	}

	cookies := res.Cookies()
	if len(cookies) == 0 {
		t.Fatal("expected cookie to be set")
	}
	if cookies[0].Name != "session_token" || cookies[0].Value != "refresh" {
		t.Fatalf("unexpected cookie: %v", cookies[0])
	}

	var payload struct {
		ID           string `json:"user_id"`
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if payload.ID != "id-1" || payload.AccessToken != "access" || payload.RefreshToken != "refresh" {
		t.Fatalf("unexpected payload: %+v", payload)
	}
}

type fakeUserRepo struct {
	id string
}

func (f *fakeUserRepo) Save(ctx context.Context, user *models.User) (string, error) {
	return f.id, nil
}

type fakeDriverRepo struct {
	id string
}

func (f *fakeDriverRepo) Save(ctx context.Context, driver *models.Driver) (string, error) {
	return f.id, nil
}

type fakeTxManager struct{}

func (f *fakeTxManager) WithTx(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

func TestHandleSignupPassenger_InvalidJSON(t *testing.T) {
	h := NewAuthHandler(nil)
	req := httptest.NewRequest(http.MethodPost, "/signup/passenger", strings.NewReader("{"))
	rr := httptest.NewRecorder()

	h.HandleSingupPassenger(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestHandleSignupPassenger_Success(t *testing.T) {
	svc := service.NewAuthService(service.NewTokenService([]byte("secret")), &fakeUserRepo{id: "u1"}, &fakeDriverRepo{}, &fakeTxManager{})
	h := NewAuthHandler(svc)

	body := `{"email":"a@b.com","password":"pass"}`
	req := httptest.NewRequest(http.MethodPost, "/signup/passenger", strings.NewReader(body))
	rr := httptest.NewRecorder()

	h.HandleSingupPassenger(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, rr.Code)
	}
}

func TestHandleSignupAdmin_Success(t *testing.T) {
	svc := service.NewAuthService(service.NewTokenService([]byte("secret")), &fakeUserRepo{id: "u2"}, &fakeDriverRepo{}, &fakeTxManager{})
	h := NewAuthHandler(svc)

	body := `{"email":"admin@b.com","password":"pass"}`
	req := httptest.NewRequest(http.MethodPost, "/signup/admin", strings.NewReader(body))
	rr := httptest.NewRecorder()

	h.HandleSingupAdmin(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, rr.Code)
	}
}

func TestHandleSignupDriver_InvalidJSON(t *testing.T) {
	h := NewAuthHandler(nil)
	req := httptest.NewRequest(http.MethodPost, "/signup/driver", strings.NewReader("{"))
	rr := httptest.NewRecorder()

	h.HandleSingupDriver(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestHandleSignupDriver_Success(t *testing.T) {
	svc := service.NewAuthService(service.NewTokenService([]byte("secret")), &fakeUserRepo{id: "u3"}, &fakeDriverRepo{id: "d1"}, &fakeTxManager{})
	h := NewAuthHandler(svc)

	body := `{"email":"driver@b.com","password":"pass","license_number":"LIC","vehicle_type":"ECONOMY","vehicle_attrs":{"vehicle_make":"Toyota","vehicle_model":"Camry","vehicle_color":"White","vehicle_plate":"KZ","vehicle_year":2020}}`
	req := httptest.NewRequest(http.MethodPost, "/signup/driver", strings.NewReader(body))
	rr := httptest.NewRecorder()

	h.HandleSingupDriver(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, rr.Code)
	}
}
