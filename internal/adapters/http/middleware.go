package http

import (
	"context"
	"github.com/go-chi/chi/v5"
	"net/http"
	"os"
)

type contextKey string

const (
	ownerKey    contextKey = "owner"
	repoNameKey contextKey = "repo_name"
)

func Authenticate() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract owner and repo_name from URL parameters (example with chi router)
			owner := chi.URLParam(r, "owner")
			repoName := chi.URLParam(r, "repo_name")

			// Store these values in the context
			ctx := context.WithValue(r.Context(), ownerKey, owner)
			ctx = context.WithValue(ctx, repoNameKey, repoName)

			// Pass the request with the updated context to the next handler
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func APIKeyAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get("X-API-Key")
		if apiKey == "" || apiKey != os.Getenv("API_KEY") {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}
