package routes

import (
	"log/slog"
	"net/http"
	"todo-app/internal/handlers/create"
	"todo-app/internal/handlers/delete"
	"todo-app/internal/handlers/read"
	"todo-app/internal/handlers/update"
	kafka "todo-app/internal/kafka/producer"
	mwAuth "todo-app/internal/middleware/auth"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"todo-app/internal/lib/logger/sl"
	"todo-app/internal/token"
	"os"
	"github.com/rail52/myprojects/dbpb"
)

func NewRouter(log *slog.Logger,client dbpb.PostgresClient, kafkaProducer *kafka.Producer) http.Handler {
	// tokenManager (public key)
	TokenMn, err := token.NewTokenManagerRSA(os.Getenv("JWT_PUBLIC_KEY_PATH"))
	if err != nil {
		log.Error("error created a new token manager", sl.Err(err))
		os.Exit(1)
	}
	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Logger)

	router.Route("/tasks", func(r chi.Router) {
		r.Use(mwAuth.AuthMiddleware(TokenMn, log))
		r.Post("/", create.CreateTask(log, client, kafkaProducer))
		r.Get("/", read.GetTasks(log, client, kafkaProducer))
		r.Get("/{id}", read.GetTask(log, client, kafkaProducer))
		r.Put("/{id}", update.UpdateTask(log, client, kafkaProducer))
		r.Patch("/{id}", update.MarkAsDone(log, client, kafkaProducer))
		r.Delete("/{id}", delete.DeleteTask(log, client, kafkaProducer))
	})

	return router
}
