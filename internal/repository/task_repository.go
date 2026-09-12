package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/Yoshikrit/observability-test/internal/model"
)

var ErrNotFound = errors.New("task not found")

type TaskRepository interface {
	Create(ctx context.Context, task *model.Task) error
	GetByID(ctx context.Context, id uint) (*model.Task, error)
	List(ctx context.Context) ([]model.Task, error)
	Update(ctx context.Context, task *model.Task) error
	Delete(ctx context.Context, id uint) error
}

type taskRepository struct {
	db *gorm.DB
}

func NewTaskRepository(db *gorm.DB) TaskRepository {
	return &taskRepository{db: db}
}

func (r *taskRepository) Create(ctx context.Context, task *model.Task) error {
	return r.db.WithContext(ctx).Create(task).Error
}

func (r *taskRepository) GetByID(ctx context.Context, id uint) (*model.Task, error) {
	var task model.Task
	if err := r.db.WithContext(ctx).First(&task, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &task, nil
}

func (r *taskRepository) List(ctx context.Context) ([]model.Task, error) {
	var tasks []model.Task
	if err := r.db.WithContext(ctx).Order("id asc").Find(&tasks).Error; err != nil {
		return nil, err
	}
	return tasks, nil
}

func (r *taskRepository) Update(ctx context.Context, task *model.Task) error {
	return r.db.WithContext(ctx).Save(task).Error
}

func (r *taskRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&model.Task{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
