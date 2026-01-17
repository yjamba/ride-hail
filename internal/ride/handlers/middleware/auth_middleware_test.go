package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type errorResponse struct {
	StatusCode int    `json:"status_code"`
	Message    string `json:"message"`
}

func newToken(t *testing.T, userID, role string, key string) string {
	t.Helper()
	claims := UserClaims{
		UserId: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(5 * time.Minute)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	str, err := token.SignedString([]byte(key))
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}
	return str
}

func decodeError(t *testing.T, rr *httptest.ResponseRecorder) errorResponse {
	t.Helper()
	var resp errorResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}
	return resp
}

func TestAuthMiddleware_MissingHeader(t *testing.T) {
	h := AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {})
	req := httptest.NewRequest(http.MethodPost, "/drivers/123/online", nil)
	req.SetPathValue("driver_id", "123")
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rr.Code)
	}
	resp := decodeError(t, rr)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected status_code %d, got %d", http.StatusUnauthorized, resp.StatusCode)
	}
}

func TestAuthMiddleware_BadPrefix(t *testing.T) {
	h := AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {})
	req := httptest.NewRequest(http.MethodPost, "/drivers/123/online", nil)
	req.SetPathValue("driver_id", "123")
	req.Header.Set("Authorization", "Token abc")
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestAuthMiddleware_InvalidSignature(t *testing.T) {
	h := AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {})
	req := httptest.NewRequest(http.MethodPost, "/drivers/123/online", nil)
	req.SetPathValue("driver_id", "123")
	token := newToken(t, "123", "DRIVER", "wrongkey")
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rr.Code)
	}
}

func TestAuthMiddleware_WrongRole(t *testing.T) {
	h := AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {})
	req := httptest.NewRequest(http.MethodPost, "/drivers/123/online", nil)
	req.SetPathValue("driver_id", "123")
	token := newToken(t, "123", "PASSENGER", secretKey)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, rr.Code)
	}
}

func TestAuthMiddleware_MismatchedDriverID(t *testing.T) {
	h := AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {})
	req := httptest.NewRequest(http.MethodPost, "/drivers/123/online", nil)
	req.SetPathValue("driver_id", "123")
	token := newToken(t, "999", "DRIVER", secretKey)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, rr.Code)
	}
}

func TestAuthMiddleware_Success(t *testing.T) {
	called := false
	h := AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	req := httptest.NewRequest(http.MethodPost, "/drivers/123/online", strings.NewReader("{}"))
	req.SetPathValue("driver_id", "123")
	token := newToken(t, "123", "DRIVER", secretKey)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
	if !called {
		t.Fatal("expected next handler to be called")
	}
}

func TestAuthMiddleware_MalformedToken(t *testing.T) {
	h := AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {})
	req := httptest.NewRequest(http.MethodPost, "/drivers/123/online", nil)
	req.SetPathValue("driver_id", "123")
	req.Header.Set("Authorization", "Bearer not-a-token")
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestAuthMiddleware_ExpiredToken(t *testing.T) {
	claims := UserClaims{
		UserId: "123",
		Role:   "DRIVER",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Minute)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	str, err := token.SignedString([]byte(secretKey))
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	h := AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {})
	req := httptest.NewRequest(http.MethodPost, "/drivers/123/online", nil)
	req.SetPathValue("driver_id", "123")
	req.Header.Set("Authorization", "Bearer "+str)
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestAuthMiddleware_WrongSigningMethod(t *testing.T) {
	claims := UserClaims{
		UserId: "123",
		Role:   "DRIVER",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(5 * time.Minute)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	str, err := token.SignedString([]byte(secretKey))
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	h := AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {})
	req := httptest.NewRequest(http.MethodPost, "/drivers/123/online", nil)
	req.SetPathValue("driver_id", "123")
	req.Header.Set("Authorization", "Bearer "+str)
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}
