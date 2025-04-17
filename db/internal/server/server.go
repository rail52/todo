package server

import (
	"github.com/rail52/myprojects/dbpb"
	"google.golang.org/grpc"
	// "google.golang.org/protobuf/types/known/timestamppb"
	"db/internal/storage/db/postgres"
	"log/slog"
	"net"
	"os"
	// "time"
	"db/internal/config"
	"db/internal/handlers"
)

func Run(log *slog.Logger, cfg *config.Config){
	const op = "db/internal/server/|New()"
	lis, err := net.Listen("tcp", cfg.Address)
	if err != nil {
		log.Error("Ошибка при запуске сервера: ", "error", err)
		os.Exit(52)
	}
	opts := []grpc.ServerOption{}
	s := grpc.NewServer(opts...)

	db, err := postgres.NewStorage(cfg.Postgres)
	if err != nil {
		log.Error("Ошибка при запуске сервера: ", "error", err)
		os.Exit(42)
	}
	dbpb.RegisterPostgresServer(s, &handlers.Server{DB: db})
	log.Info("Сервер запущен на: " + cfg.Address)
	if err := s.Serve(lis); err != nil {
		log.Error("Ошибка сервера:", "error", err)
		os.Exit(52)
	}
}
