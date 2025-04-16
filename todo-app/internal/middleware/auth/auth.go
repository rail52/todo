package mwAuth

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"todo-app/internal/token"
	"strings"

	"github.com/go-chi/render"
)

type contextKey string

const UserKey contextKey = "user"

func AuthMiddleware(tokenMn *token.TokenManager, log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")

			if authHeader == "" {
				log.Warn("Authorization header missing")
				render.Status(r, http.StatusUnauthorized)
				render.JSON(w, r, map[string]string{"error": "token required"})
				return
			}

			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

			claims, err := tokenMn.ParseToken(tokenStr)
			if err != nil {
				if errors.Is(err, token.ErrTokenExpired) {
					log.Warn("Token expired", "error", err)
					render.Status(r, http.StatusUnauthorized)
					render.JSON(w, r, map[string]string{"error": "token expired"})
				} else {
					log.Warn("Token validation failed", "error", err)
					render.Status(r, http.StatusUnauthorized)
					render.JSON(w, r, map[string]string{"error": "invalid token"})
				}
				return

			}

			if claims.TokenType != "access" {
				log.Warn("Invalid token type", "type", claims.TokenType)
				render.Status(r, http.StatusUnauthorized)
				render.JSON(w, r, map[string]string{"error": "invalid token type"})
				return
			}

			ctx := context.WithValue(r.Context(), UserKey, claims)

			next.ServeHTTP(w, r.WithContext(ctx))
		})

	}
}
