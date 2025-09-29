package middleware

import (
	"fmt"
	"net/http"
)

func WithAPIKey(apiKey string) func(http.Handler) http.Handler {

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			key := r.Header.Get("X-API-KEY")
			if key != apiKey {
				fmt.Printf("Authentication failed: invalid API key from %s", r.RemoteAddr)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
