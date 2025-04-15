package register

import (
	"auth/internal/domain/model"
	"auth/internal/domain/requests"
	"auth/internal/domain/response"
	"auth/internal/lib/logger/sl"
	"auth/internal/lib/validate"
	"auth/internal/storage/db"
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"golang.org/x/crypto/bcrypt"
)

//go:generate go run github.com/vektra/mockery/v2@latest --name=UserCreater
type UserCreater interface {
	CreateUser(u *model.User) error
	GetByEmail(email string) (*model.User, error)
}

type NewMentor interface {
	NewMentor(ctx context.Context, mentorEmail, contact string) error
}

func Register(ctx context.Context, log *slog.Logger, userCreater UserCreater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.register.Register"
		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req requests.Register
		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			log.Error("failed to decode request body", sl.Err(err))
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, response.Error("invalid request body"))
			return
		}

		if err := validate.IsValid(req); err != nil {
			log.Error("validation error", sl.Err(err))
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, response.Error("server error"))
			return
		}

		existing, err := userCreater.GetByEmail(req.Email)
		if err != nil && !errors.Is(err, db.ErrUserNotFound) {
			log.Error("failed to check user existence", sl.Err(err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, response.Error("internal error"))
			return
		}

		if existing != nil {
			log.Error("user already exists", slog.String("email", req.Email))
			render.Status(r, http.StatusConflict)
			render.JSON(w, r, response.Error("user already exists"))
			return
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			log.Error("failed to hash password", sl.Err(err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, "failed to process password")
			return
		}

		user := &model.User{
			Email:    req.Email,
			Password: string(hash),
		}

		if err = userCreater.CreateUser(user); err != nil {
			log.Error("failed to create user", sl.Err(err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, response.Error("failed to create user"))
			return
		}

		

		render.Status(r, http.StatusCreated)
		render.JSON(w, r, map[string]any{
			"id":    user.ID,
			"email": user.Email,
		})

	}
}
