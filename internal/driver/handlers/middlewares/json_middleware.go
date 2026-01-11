package middlewares

import "net/http"

func JsonMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Body != nil && r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "incorrect content type", http.StatusBadRequest)
			return
		}

		next.ServeHTTP(w, r)
	}
}
