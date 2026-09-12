package service

import (
	"context"

	"github.com/Yoshikrit/observability-test/internal/model"
	"github.com/Yoshikrit/observability-test/internal/repository"
)

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
	task := &model.Task{
		Title:       input.Title,
		Description: input.Description,
	}
	if err := s.repo.Create(ctx, task); err != nil {
		return nil, err
	}
	return task, nil
}

func (s *taskService) GetTask(ctx context.Context, id uint) (*model.Task, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *taskService) ListTasks(ctx context.Context) ([]model.Task, error) {
	return s.repo.List(ctx)
}

func (s *taskService) UpdateTask(ctx context.Context, id uint, input UpdateTaskInput) (*model.Task, error) {
	task, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	task.Title = input.Title
	task.Description = input.Description
	task.Done = input.Done

	if err := s.repo.Update(ctx, task); err != nil {
		return nil, err
	}
	return task, nil
}

func (s *taskService) DeleteTask(ctx context.Context, id uint) error {
	return s.repo.Delete(ctx, id)
}
