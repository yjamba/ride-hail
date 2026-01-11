package middlewares

import (
	"fmt"
	"net/http"
	"strings"

	"ride-hail/internal/driver/handlers/utils"

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
			utils.SendError(w, utils.ErrorMessage{
				StatusCode: http.StatusUnauthorized,
				Message:    "Authorization header is required",
			})
			return
		}

		if !strings.HasPrefix(tokenString, "Bearer ") {
			utils.SendError(w, utils.ErrorMessage{
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
				utils.SendError(w, utils.ErrorMessage{
					StatusCode: http.StatusUnauthorized,
					Message:    "Invalid token signature",
				})
			} else {
				utils.SendError(w, utils.ErrorMessage{
					StatusCode: http.StatusBadRequest,
					Message:    "Malformed token",
				})
			}
			return
		}

		if !token.Valid {
			utils.SendError(w, utils.ErrorMessage{
				StatusCode: http.StatusUnauthorized,
				Message:    "Invalid or expired token",
			})
			return
		}

		if claims.Role != "DRIVER" {
			utils.SendError(w, utils.ErrorMessage{
				StatusCode: http.StatusForbidden,
				Message:    "Access denied: driver role required",
			})
			return
		}

		next.ServeHTTP(w, r)
	}
}
