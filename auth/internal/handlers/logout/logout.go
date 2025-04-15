package logout

import (
	"auth/internal/domain/requests"
	"auth/internal/domain/response"
	"auth/internal/lib/logger/sl"
	"auth/internal/lib/validate"
	"auth/pkg/token"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

//go:generate go run github.com/vektra/mockery/v2@latest --name=RedisRepo
type RedisRepo interface {
	AddToBlackList(token string, exp int64) error
	IsBlackListed(token string) (bool, error)
}

type TokenMn interface {
	GenerateToken(userID int64, ttl time.Duration, tokenType string) (string, error)
	ParseToken(tokenStr string) (*token.Claims, error)
}

func Logout(log *slog.Logger, redisRepo RedisRepo, tokenMn TokenMn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.logout.Logout"
		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req requests.RFToken
		if err := render.DecodeJSON(r.Body, &req); err != nil {
			log.Error("failed to decode request body", sl.Err(err))
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, response.Error("invalid request"))
			return
		}

		if err := validate.IsValid(req); err != nil {
			log.Warn("request is not valid", slog.String("valid", "false"))
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, response.Error("invalid request"))
			return
		}

		isBlackListed, err := redisRepo.IsBlackListed(req.RefreshToken)
		if err != nil {
			log.Error("failed to check token blacklist", sl.Err(err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, response.Error("server error"))
			return
		}
		if isBlackListed {
			log.Warn("token already blacklisted")
			render.Status(r, http.StatusConflict)
			render.JSON(w, r, response.Error("token revoked"))
			return
		}

		claims, err := tokenMn.ParseToken(req.RefreshToken)
		if err != nil {
			log.Error("failed to parse redresh token", sl.Err(err))
			render.Status(r, http.StatusUnauthorized)
			render.JSON(w, r, response.Error("invalid token"))
			return
		}

		if claims.TokenType != "refresh" {
			log.Error("invalid token type", slog.String("type", claims.TokenType))
			render.Status(r, http.StatusUnauthorized)
			render.JSON(w, r, response.Error("invalid token type"))
			return
		}

		if err := redisRepo.AddToBlackList(req.RefreshToken, claims.ExpiresAt.Unix()); err != nil {
			log.Error("failed to blacklist token", sl.Err(err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, response.Error("server error"))
			return
		}

		render.Status(r, http.StatusOK)
		render.JSON(w, r, map[string]any{
			"status": "logged_out",
		})
	}
}
