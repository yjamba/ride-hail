package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

var secretKey string = "supersecretkey"

type UserClaims struct {
	UserId string `json:"user_id"`
	Role   string `json:"role"`

	jwt.RegisteredClaims
}

type errorMessage struct {
	StatusCode int    `json:"status_code"`
	Message    string `json:"message"`
}

func sendError(w http.ResponseWriter, msg errorMessage) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(msg.StatusCode)
	_ = json.NewEncoder(w).Encode(msg)
}

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("Authorization")
		if tokenString == "" {
			sendError(w, errorMessage{
				StatusCode: http.StatusUnauthorized,
				Message:    "Authorization header is required",
			})
			return
		}

		if !strings.HasPrefix(tokenString, "Bearer ") {
			sendError(w, errorMessage{
				StatusCode: http.StatusBadRequest,
				Message:    "Authorization header must be in the format 'Bearer <token>'",
			})
			return
		}

		tokenString = tokenString[len("Bearer "):]

		claims := &UserClaims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
			if t.Method != jwt.SigningMethodHS256 {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}

			return []byte(secretKey), nil
		})
		if err != nil {
			if strings.Contains(err.Error(), "signature") {
				sendError(w, errorMessage{
					StatusCode: http.StatusUnauthorized,
					Message:    "Invalid token signature",
				})
			} else {
				sendError(w, errorMessage{
					StatusCode: http.StatusBadRequest,
					Message:    "Malformed token",
				})
			}
			return
		}

		if !token.Valid {
			sendError(w, errorMessage{
				StatusCode: http.StatusUnauthorized,
				Message:    "Invalid or expired token",
			})
			return
		}

		if claims.Role != "DRIVER" {
			sendError(w, errorMessage{
				StatusCode: http.StatusForbidden,
				Message:    "Access denied: driver role required",
			})
			return
		}

		id := r.PathValue("driver_id")

		if id != claims.UserId {
			sendError(w, errorMessage{
				StatusCode: http.StatusForbidden,
				Message:    "Access denied: you can only access your own driver profile",
			})
			return
		}

		next.ServeHTTP(w, r)
	}
}

// PassengerAuthMiddleware validates JWT tokens for passengers
func PassengerAuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("Authorization")
		if tokenString == "" {
			sendError(w, errorMessage{
				StatusCode: http.StatusUnauthorized,
				Message:    "Authorization header is required",
			})
			return
		}

		if !strings.HasPrefix(tokenString, "Bearer ") {
			sendError(w, errorMessage{
				StatusCode: http.StatusBadRequest,
				Message:    "Authorization header must be in the format 'Bearer <token>'",
			})
			return
		}

		tokenString = tokenString[len("Bearer "):]

		claims := &UserClaims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
			if t.Method != jwt.SigningMethodHS256 {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}

			return []byte(secretKey), nil
		})
		if err != nil {
			if strings.Contains(err.Error(), "signature") {
				sendError(w, errorMessage{
					StatusCode: http.StatusUnauthorized,
					Message:    "Invalid token signature",
				})
			} else {
				sendError(w, errorMessage{
					StatusCode: http.StatusBadRequest,
					Message:    "Malformed token",
				})
			}
			return
		}

		if !token.Valid {
			sendError(w, errorMessage{
				StatusCode: http.StatusUnauthorized,
				Message:    "Invalid or expired token",
			})
			return
		}

		if claims.Role != "PASSENGER" {
			sendError(w, errorMessage{
				StatusCode: http.StatusForbidden,
				Message:    "Access denied: passenger role required",
			})
			return
		}

		next.ServeHTTP(w, r)
	}
}
