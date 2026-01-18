package handlers

import (
	"net/http"

	"ride-hail/internal/auth/handlers/middlewares"
)

func RegisterRoutes(handler *AuthHandler) http.Handler {
	mux := http.NewServeMux()

	mux.Handle("POST /auth/sing_up/passenger", middlewares.JsonMiddleware(handler.HandleSingupPassenger))
	mux.Handle("POST /auth/sing_up/driver", middlewares.JsonMiddleware(handler.HandleSingupDriver))
	mux.Handle("POST /auth/sing_up/admin", middlewares.JsonMiddleware(handler.HandleSingupAdmin))

	return mux
}
