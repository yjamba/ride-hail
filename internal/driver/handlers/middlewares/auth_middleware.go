package middlewares

import (
	"net/http"

	"github.com/golang-jwt/jwt/v5"
)

var secretKey string = "secret-key"

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("Authorizations")
		if tokenString == "" {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte("Missing authorization header"))
			return
		}
		tokenString = tokenString[len("Bearer "):]

		token, err := jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
			return secretKey, nil
		})
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(err.Error()))
			return
		}

		if !token.Valid {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte("Invalid toke"))
		}

		next.ServeHTTP(w, r)
	}
}
