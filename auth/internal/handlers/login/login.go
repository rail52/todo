package login

import (
	"auth/internal/domain/model"
	"auth/internal/domain/requests"
	"auth/internal/domain/response"
	"auth/internal/lib/logger/sl"
	"auth/internal/lib/validate"
	"auth/internal/storage/db"
	"auth/pkg/token"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"golang.org/x/crypto/bcrypt"
)

var (
	AccessTokenTTL  int64 = 1800
	RefreshTokenTTL int64 = 604800
)

//go:generate go run github.com/vektra/mockery/v2@latest --name=Auth
type Auth interface {
	CreateUser(u *model.User) error
	GetByEmail(email string) (*model.User, error)
}

//go:generate go run github.com/vektra/mockery/v2@latest --name=TokenMn
type TokenMn interface {
	GenerateToken(userID int64, ttl time.Duration, tokenType string) (string, error)
	ParseToken(tokenStr string) (*token.Claims, error)
}

func Login(log *slog.Logger, auth Auth, tokenMn TokenMn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.login.Login"
		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req requests.Login
		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			log.Error("failed to decode request body", sl.Err(err))
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, response.Error("invalid request body"))
			return
		}

		if err := validate.IsValid(req); err != nil {
			log.Warn("request is not valid", slog.String("valid", "false"))
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, response.Error("invalid request"))
			return
		}

		user, err := auth.GetByEmail(req.Email)
		if err != nil {
			if errors.Is(err, db.ErrUserNotFound) {
				log.Warn("user not found", slog.String("email", req.Email))
				render.Status(r, http.StatusUnauthorized)
				render.JSON(w, r, response.Error("status invalid credentials"))
				return
			}
			log.Error("failed to get user", sl.Err(err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, response.Error("internal error"))
			return
		}

		err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
		if err != nil {
			log.Warn("invalid password", slog.String("email", req.Email))
			render.Status(r, http.StatusUnauthorized)
			render.JSON(w, r, response.Error("invalid credentials"))
			return
		}

		access, err := tokenMn.GenerateToken(
			user.ID,
			time.Duration(AccessTokenTTL)*time.Second,
			"access",
		)
		if err != nil {
			log.Error("failed to generate access token", sl.Err(err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, response.Error("failed to generate access token"))
			return
		}

		refresh, err := tokenMn.GenerateToken(
			user.ID,
			time.Duration(RefreshTokenTTL)*time.Second,
			"refresh",
		)
		if err != nil {
			log.Error("failed to generate refresh token", sl.Err(err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, response.Error("failed to generate refresh token"))
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "refresh_token",
			Value:    refresh,
			Path:     "/",
			MaxAge:   int(RefreshTokenTTL),
			HttpOnly: true,
			Secure:   false,
			SameSite: http.SameSiteDefaultMode,
		})

		render.Status(r, http.StatusOK)
		render.JSON(w, r, map[string]any{
			"access_token":  access,
			"refresh_token": refresh,
		})

	}
}
