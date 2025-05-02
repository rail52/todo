package create

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"todo-app/internal/domain/requests"
	"google.golang.org/grpc"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/rail52/myprojects/dbpb"
)
//go:generate go run github.com/vektra/mockery/v2@latest --name=KafkaProducer
type KafkaProducer interface {
	SendApiEvent(apiRequest *requests.ApiRequest) error
}

//go:generate go run github.com/vektra/mockery/v2@latest --name=Creator
type Creator interface {
	CreateTask(ctx context.Context, in *dbpb.CreateTaskRequest, opts ...grpc.CallOption) (*dbpb.Task, error)
}

func CreateTask(log *slog.Logger, client Creator, kafkaProducer KafkaProducer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fn := "internal/http-server/handlers/handlers.go|CreateTask()"
		log = log.With(
			slog.String("fn", fn),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req requests.CreateTaskRequest
		err := render.DecodeJSON(r.Body, &req)
		if errors.Is(err, io.EOF) {
			Err := "request body is empty"
			log.Info(Err)			
			http.Error(w, Err, http.StatusBadRequest)
			return
		}
		if req.Title == "" || req.Content == "" {
			Err := "Title or Content in request Body is empty or invalid"
			log.Info(Err)
			http.Error(w, Err, http.StatusBadRequest)
			return
		}
		
		task, err := client.CreateTask(r.Context(),
		&dbpb.CreateTaskRequest{
			Title:   req.Title,
			Content: req.Content},
		)
		if err != nil {
			log.Error("Failed to take tasks: ", slog.String("err", err.Error()))
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		
		render.JSON(w, r, &task)
		event := &requests.ApiRequest{
			Action: "created",
		}
		if err := kafkaProducer.SendApiEvent(event); err != nil {
			log.Error("failed to send kafka even", (slog.String("error", err.Error())))
		}
	}
}
