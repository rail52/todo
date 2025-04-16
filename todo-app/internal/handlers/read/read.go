package read

import (
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"todo-app/internal/domain/requests"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/rail52/myprojects/dbpb"
	"google.golang.org/protobuf/types/known/emptypb"
)

type KafkaProducer interface {
	SendApiEvent(apiRequest *requests.ApiRequest) error
}

func GetTask(log *slog.Logger, storage dbpb.PostgresClient, kafkaProducer KafkaProducer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fn := "internal/http-server/handlers/handlers.go|GetTask()"
		log = log.With(
			slog.String("fn", fn),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)
		idStr := chi.URLParam(r, "id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			Err := "Invalid task ID"
			log.Info(Err)
			http.Error(w, Err, http.StatusBadRequest)
			return
		}

		task, err := storage.GetTask(r.Context(), &dbpb.GetTaskRequest{
			Id: int64(id),
		})
		if err != nil {
			Err := "Failed to fetch task with ID: " + idStr
			log.Error(Err)
			http.Error(w, Err, http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		render.JSON(w, r, &task)

		event := &requests.ApiRequest{
			Action: "fetched",
		}

		if err := kafkaProducer.SendApiEvent(event); err != nil {
			log.Error("failed to send kafka even", (slog.String("error", err.Error())))
		}
	}
}
func GetTasks(log *slog.Logger, storage dbpb.PostgresClient, kafkaProducer KafkaProducer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fn := "internal/http-server/handlers/handlers.go|GetTasks()"
		log = log.With(
			slog.String("fn", fn),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)
		stream, err := storage.GetTasks(r.Context(), &emptypb.Empty{})
		if err != nil {
			Err := "GetTasks failed:"
			log.Error(Err, slog.String("err", err.Error()))
			http.Error(w, Err, http.StatusBadRequest)
		}

		w.Header().Set("Content-Type", "application/json")
		for {
			task, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				Err := "receiving task Error:"
				log.Error(Err, slog.String("err", err.Error()))
				http.Error(w, Err, http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			render.JSON(w, r, &task)
		}

		event := &requests.ApiRequest{
			Action: "allfetched",
		}

		if err := kafkaProducer.SendApiEvent(event); err != nil {
			log.Error("failed to send kafka even", (slog.String("error", err.Error())))
		}
	}
}
