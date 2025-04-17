package handlers

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rail52/myprojects/dbpb"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	// "log/slog"
	// "net"
	// "os"
	"time"
	// "todo-app/internal/config"
	// "todo-app/internal/lib/logger/handlers/slogpretty"
)

type Server struct {
	dbpb.UnimplementedPostgresServer
	DB *pgxpool.Pool
}

func (s *Server) CreateTask(ctx context.Context, req *dbpb.CreateTaskRequest) (*dbpb.Task, error) {
	const op = "db/internal/handlers|CreateTask()"
	task := dbpb.Task{}
	var createdAt, updatedAt time.Time
	query := `INSERT INTO task (title, content) 
			  VALUES ($1, $2)
			  RETURNING id, title, content, is_done, created_at, updated_at`
	err := s.DB.QueryRow(ctx, query, req.GetTitle(), req.GetContent()).Scan(
		&task.Id,
		&task.Title,
		&task.Content,
		&task.IsDone,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to execute query: %w", op, err)
	}
	task.CreatedAt = timestamppb.New(createdAt)
	task.UpdatedAt = timestamppb.New(updatedAt)

	return &task, nil
}
func (s *Server) GetTasks(_ *emptypb.Empty, stream grpc.ServerStreamingServer[dbpb.Task]) error {
	const op = "db/internal/handlers|GetTasks()"
	var task dbpb.Task
	var createdAt, updatedAt time.Time

	query := "SELECT * from task"
	rows, err := s.DB.Query(context.Background(), query)
	if err != nil {
		return fmt.Errorf("%s: failed to execute query: %w", op, err)
	}
	for rows.Next() {
		err := rows.Scan(
			&task.Id,
			&task.Title,
			&task.Content,
			&task.IsDone,
			&createdAt,
			&updatedAt,
		)
		if err != nil {
			return fmt.Errorf("%s: failed to Scan() the part of query: %w", op, err)
		}
		task.CreatedAt = timestamppb.New(createdAt)
		task.UpdatedAt = timestamppb.New(updatedAt)
		if err := stream.Send(&task); err != nil {
			return err
		}
	}
	return nil
}

func (s *Server) GetTask(ctx context.Context, req *dbpb.GetTaskRequest) (*dbpb.Task, error) {
	const op = "db/internal/handlers|CreateTask()"
	task := dbpb.Task{}
	var createdAt, updatedAt time.Time

	// query := "select * from task WHERE id ="+string(id)
	query := `select * 
			  from task 
			  WHERE id = $1`
	err := s.DB.QueryRow(ctx, query, req.GetId()).Scan(
		&task.Id,
		&task.Title,
		&task.Content,
		&task.IsDone,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to Scan(): %w", op, err)
	}
	task.CreatedAt = timestamppb.New(createdAt)
	task.UpdatedAt = timestamppb.New(updatedAt)
	return &task, nil
}

type UpdateQueryParams struct {
	Title   string
	Content string
	IsDone  bool
}

func (s *Server) UpdateTask(ctx context.Context, req *dbpb.UpdateTaskRequest) (*dbpb.Task, error) {
	const op = "db/internal/handlers|UpdateTask()"
	task := dbpb.Task{}
	var createdAt, updatedAt time.Time
	var updateQueryParams UpdateQueryParams
	if req.Title == nil {
		updateQueryParams.Title = "title"
	} else {
		updateQueryParams.Title = req.GetTitle()

	}
	if req.Content == nil {
		updateQueryParams.Content = "content"
	} else {
		updateQueryParams.Content = req.GetContent()

	}

	query := `update task 
			  set title = $2, content = $3, is_done = $4
			  where id = $1;`
	_, err := s.DB.Exec(ctx, query, req.GetId(), updateQueryParams.Title, updateQueryParams.Content, req.GetIsDone())
	if err != nil {
		return nil, fmt.Errorf("%s: failed to execute query: %w", op, err)
	}
	query = `select * 
			 from task 
			 where id = $1`
	err = s.DB.QueryRow(ctx, query, req.GetId()).Scan(
		&task.Id,
		&task.Title,
		&task.Content,
		&task.IsDone,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to execute query: %w", op, err)
	}
	task.CreatedAt = timestamppb.New(createdAt)
	task.UpdatedAt = timestamppb.New(updatedAt)
	return &task, nil

}
func (s *Server) MarkAsDone(ctx context.Context, req *dbpb.MarkAsDoneRequest) (*dbpb.Task, error) {
	const op = "db/internal/handlers|MarkAsDone()"

	task := dbpb.Task{}
	var createdAt, updatedAt time.Time

	query := `update task 
			  set is_done = True
			  where id = $1;`
	_, err := s.DB.Exec(ctx, query, req.GetId())
	if err != nil {
		return nil, fmt.Errorf("%s: failed to execute query: %w", op, err)
	}

	query = `select * 
			from task 
			where id = $1`
	err = s.DB.QueryRow(ctx, query, req.GetId()).Scan(
		&task.Id,
		&task.Title,
		&task.Content,
		&task.IsDone,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to execute query: %w", op, err)
	}
	task.CreatedAt = timestamppb.New(createdAt)
	task.UpdatedAt = timestamppb.New(updatedAt)
	return &task, nil
}
func (s *Server) DeleteTask(ctx context.Context, req *dbpb.DeleteTaskRequest) (*emptypb.Empty, error) {
	const op = "db/internal/handlers|DeleteTask()"
	query := `delete from task
			  where id = $1`
	_, err := s.DB.Exec(ctx, query, req.GetId())
	if err != nil {
		return nil, fmt.Errorf("%s: failed to execute query: %w", op, err)
	}
	return nil, nil

}
