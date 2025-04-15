package auth

import (
	"auth/pkg/token"
	"context"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/render"
)

type contextKey string

const UserKey contextKey = "user"

func AuthMiddleware(tm *token.TokenManager, log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")

			if authHeader == "" {
				log.Warn("Authorization header missing")
				render.Status(r, http.StatusUnauthorized)
				render.JSON(w, r, map[string]string{"error": "token required"})
				return
			}

			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

			claims, err := tm.ParseToken(tokenStr)
			if err != nil {
				log.Warn("Token validation falied", "error", err)
				render.Status(r, http.StatusUnauthorized)
				render.JSON(w, r, map[string]string{"error": "invalid token"})
				return
			}

			ctx := context.WithValue(r.Context(), UserKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		}

		return http.HandlerFunc(fn)
	}
}
