package delete

import (
	"log/slog"
	"net/http"
	"todo-app/internal/domain/requests"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/rail52/myprojects/dbpb"
)

type KafkaProducer interface {
	SendApiEvent(apiRequest *requests.ApiRequest) error
}

func DeleteTask(log *slog.Logger, storage dbpb.PostgresClient, kafkaProducer KafkaProducer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fn := "internal/http-server/handlers/handlers.go|DeleteTask()"
		log = log.With(
			slog.String("fn", fn),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)
		idStr := chi.URLParam(r, "id")
		_, err := storage.DeleteTask(
			r.Context(),
			&dbpb.DeleteTaskRequest{
				Id: idStr,
			},
		)
		if err != nil {
			Err := "Failed to Delete task with ID: " + idStr
			log.Error(Err)
			http.Error(w, Err, http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		render.JSON(w, r, "task deleted")
		event := &requests.ApiRequest{
			Action: "deleted",
		}

		if err := kafkaProducer.SendApiEvent(event); err != nil {
			log.Error("failed to send kafka even", (slog.String("error", err.Error())))
		}
	}
}
