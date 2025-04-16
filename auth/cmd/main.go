package main

import (
	"auth/internal/config"
	"auth/internal/handlers/login"
	"auth/internal/handlers/logout"
	"auth/internal/handlers/refresh"
	"auth/internal/handlers/register"
	mwLogger "auth/internal/middleware/logger"
	"context"
	"log/slog"
	"os/signal"
	"syscall"
	"time"

	"auth/internal/lib/logger/sl"
	"auth/internal/storage/cache"
	"auth/internal/storage/db"
	"auth/internal/token"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {

	cfg := config.LoadConfig()

	log := setupLogger(cfg.Env)

	log.Info(
		"starting auth-service",
		slog.String("env", cfg.Env),
	)

	log.Debug("debug messages are enabled")

	redisClient := cache.New(cfg.RedisConfig)

	redisRepository := cache.NewRedisRepository(redisClient)

	storage, err := db.NewStorage(cfg.Config)
	if err != nil {
		log.Error("error creation storage", sl.Err(err))
		os.Exit(1)
	}

	tokemMn, err := token.NewTokenmanagerRSA(cfg.PrivateKeyPath, cfg.PublicKeyPath)
	if err != nil {
		log.Error("error with token manager", sl.Err(err))
		os.Exit(1)
	}


	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(mwLogger.New(log))
	router.Use(middleware.Logger)
	router.Use(middleware.URLFormat)

	router.Post("/auth/register", register.Register(context.Background(), log, storage))
	router.Post("/auth/login", login.Login(log, storage, tokemMn))
	router.Post("/auth/logout", logout.Logout(log, redisRepository, tokemMn))
	router.Post("/auth/refresh", refresh.RefreshTokens(log, redisRepository, tokemMn))

	log.Info("starting server", slog.String("address", cfg.Address))

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	srv := &http.Server{
		Addr:         cfg.Address,
		Handler:      router,
		ReadTimeout:  cfg.Timeout,
		WriteTimeout: cfg.Timeout,
		IdleTimeout:  cfg.IdleTimeout,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("failed to start server", sl.Err(err))
		}
	}()

	log.Info("server started")

	<-done
	log.Info("stopping server")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("failed to stop server", sl.Err(err))

		return
	}

	log.Info("server stopped")

}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	default:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}
	return log
}
