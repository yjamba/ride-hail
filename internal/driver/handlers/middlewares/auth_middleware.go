package middlewares

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
)

var secretKey string = "supersecretkey"

type UserClaims struct {
	UserId string `json:"user_id"`
	Role   string `json:"role"`

	jwt.RegisteredClaims
}

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("Authorization")
		if tokenString == "" {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte("Missing authorization header"))
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
			slog.Error("Error 1", err.Error())

			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(err.Error()))
			return
		}

		if !token.Valid {
			slog.Error("Error 2", "Invalid token")
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte("Invalid toke"))
		}

		if claims.Role != "driver" {
			slog.Error("Invalid role")
			return
		}

		next.ServeHTTP(w, r)
	}
}
