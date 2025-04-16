package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"
	"todo-app/internal/config"

	"errors"
	"todo-app/internal/kafka/producer"
	"todo-app/internal/routes"

	"os/signal"
	"syscall"
	"github.com/rail52/myprojects/dbpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

var ctxTime = 10 * time.Second

func main() {
	//preparing
	cfg := config.MustLoad()
	log := setupLogger(cfg.Env)
	log.Info("starting todo-app", slog.String("env", cfg.Env))
	// kafka
	kafkaProducer, err := kafka.NewProducer(
		cfg.Brokers,
		cfg.Topic,
		log,
	)
	if err != nil {
		log.Error("failed to initialize Kafka producer", slog.String("error", err.Error()))
	}
	defer kafkaProducer.Close()
	
	// grpc client
	conn, err := grpc.NewClient(cfg.DBServiceAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Error("did not connect:", "error", err)
	}
	defer conn.Close()
	client := dbpb.NewPostgresClient(conn)
	// router
	router := routes.NewRouter(log, client, kafkaProducer)
	// server
	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

	srv := &http.Server{
		Addr:         cfg.Address,
		Handler:      router,
		ReadTimeout:  cfg.Timeout,
		WriteTimeout: cfg.Timeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("failed to start server", "error", err)
		}
	}()

	log.Info("server started")

	<-done
	log.Info("stopping server")

	ctx, cancel := context.WithTimeout(context.Background(), ctxTime)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("failed to stop server", "error", err)
		return
	}

	log.Info("server stopped")
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
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
