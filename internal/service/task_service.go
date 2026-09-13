package service

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/Yoshikrit/observability-test/internal/model"
	"github.com/Yoshikrit/observability-test/internal/repository"
)

// tracer is scoped to this package so every span it creates is attributed to
// "task-service" in Tempo, distinct from the repository's own GORM spans.
var tracer = otel.Tracer("task-service")

type CreateTaskInput struct {
	Title       string
	Description string
}

type UpdateTaskInput struct {
	Title       string
	Description string
	Done        bool
}

type TaskService interface {
	CreateTask(ctx context.Context, input CreateTaskInput) (*model.Task, error)
	GetTask(ctx context.Context, id uint) (*model.Task, error)
	ListTasks(ctx context.Context) ([]model.Task, error)
	UpdateTask(ctx context.Context, id uint, input UpdateTaskInput) (*model.Task, error)
	DeleteTask(ctx context.Context, id uint) error
}

type taskService struct {
	repo repository.TaskRepository
}

func NewTaskService(repo repository.TaskRepository) TaskService {
	return &taskService{repo: repo}
}

func (s *taskService) CreateTask(ctx context.Context, input CreateTaskInput) (*model.Task, error) {
	ctx, span := tracer.Start(ctx, "task_service.CreateTask")
	defer span.End()

	task := &model.Task{
		Title:       input.Title,
		Description: input.Description,
	}
	if err := s.repo.Create(ctx, task); err != nil {
		recordError(span, err)
		return nil, err
	}
	return task, nil
}

func (s *taskService) GetTask(ctx context.Context, id uint) (*model.Task, error) {
	ctx, span := tracer.Start(ctx, "task_service.GetTask")
	defer span.End()

	task, err := s.repo.GetByID(ctx, id)
	if err != nil {
		recordError(span, err)
		return nil, err
	}
	return task, nil
}

func (s *taskService) ListTasks(ctx context.Context) ([]model.Task, error) {
	ctx, span := tracer.Start(ctx, "task_service.ListTasks")
	defer span.End()

	tasks, err := s.repo.List(ctx)
	if err != nil {
		recordError(span, err)
		return nil, err
	}
	return tasks, nil
}

func (s *taskService) UpdateTask(ctx context.Context, id uint, input UpdateTaskInput) (*model.Task, error) {
	ctx, span := tracer.Start(ctx, "task_service.UpdateTask")
	defer span.End()

	task, err := s.repo.GetByID(ctx, id)
	if err != nil {
		recordError(span, err)
		return nil, err
	}

	task.Title = input.Title
	task.Description = input.Description
	task.Done = input.Done

	if err := s.repo.Update(ctx, task); err != nil {
		recordError(span, err)
		return nil, err
	}
	return task, nil
}

func (s *taskService) DeleteTask(ctx context.Context, id uint) error {
	ctx, span := tracer.Start(ctx, "task_service.DeleteTask")
	defer span.End()

	if err := s.repo.Delete(ctx, id); err != nil {
		recordError(span, err)
		return err
	}
	return nil
}

func recordError(span trace.Span, err error) {
	span.RecordError(err)
	span.SetStatus(codes.Error, err.Error())
}
