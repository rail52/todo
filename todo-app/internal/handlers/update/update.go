package update

import (
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"todo-app/internal/domain/requests"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/rail52/myprojects/dbpb"
)

type KafkaProducer interface {
	SendApiEvent(apiRequest *requests.ApiRequest) error
}

func UpdateTask(log *slog.Logger, storage dbpb.PostgresClient, kafkaProducer KafkaProducer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fn := "internal/http-server/handlers/handlers.go|UpdateTask()"
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
		req := requests.UpdateTaskRequest{}
		err = render.DecodeJSON(r.Body, &req)
		if errors.Is(err, io.EOF) {
			Err := "request body is empty"
			log.Info(Err)
			render.JSON(w, r, Err)
			return
		}

		task, err := storage.UpdateTask(r.Context(),
			&dbpb.UpdateTaskRequest{
				Id:      int64(id),
				Title:   &req.Title,
				Content: &req.Content,
				IsDone:  &req.IsDone,
			},
		)
		if err != nil {
			Err := "Failed to fetch task with ID: " + idStr
			log.Error(Err)
			http.Error(w, Err, http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		render.JSON(w, r, &task)
		event := &requests.ApiRequest{
			Action: "updated",
		}

		if err := kafkaProducer.SendApiEvent(event); err != nil {
			log.Error("failed to send kafka even", (slog.String("error", err.Error())))
		}
	}
}
func MarkAsDone(log *slog.Logger, storage dbpb.PostgresClient, kafkaProducer KafkaProducer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fn := "internal/http-server/handlers/handlers.go|MarkAsDone()"
		log = log.With(
			slog.String("fn", fn),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)
		idStr := chi.URLParam(r, "id")

		task, err := storage.MarkAsDone(r.Context(),
			&dbpb.MarkAsDoneRequest{
				Id: idStr,
			},
		)

		if err != nil {
			Err := "Failed to MarkAsDone task with ID: " + idStr
			log.Error(Err)
			http.Error(w, Err, http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		render.JSON(w, r, &task)
		event := &requests.ApiRequest{
			Action: "marked",
		}

		if err := kafkaProducer.SendApiEvent(event); err != nil {
			log.Error("failed to send kafka even", (slog.String("error", err.Error())))
		}
	}
}
