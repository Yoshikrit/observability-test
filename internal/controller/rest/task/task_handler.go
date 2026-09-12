package task

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog/log"

	"github.com/Yoshikrit/observability-test/internal/repository"
	"github.com/Yoshikrit/observability-test/internal/service"
)

type TaskHandler struct {
	service service.TaskService
}

func NewTaskHandler(s service.TaskService) *TaskHandler {
	return &TaskHandler{service: s}
}

type taskRequest struct {
	Title       string `json:"title" validate:"required,max=200"`
	Description string `json:"description" validate:"max=1000"`
	Done        bool   `json:"done"`
}

func (h *TaskHandler) Create(c fiber.Ctx) error {
	var req taskRequest
	if err := c.Bind().Body(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	task, err := h.service.CreateTask(c.Context(), service.CreateTaskInput{
		Title:       req.Title,
		Description: req.Description,
	})
	if err != nil {
		log.Error().Err(err).Str("title", req.Title).Msg("task: failed to create")
		return fiber.NewError(fiber.StatusInternalServerError, "failed to create task")
	}

	return c.Status(fiber.StatusCreated).JSON(task)
}

func (h *TaskHandler) List(c fiber.Ctx) error {
	tasks, err := h.service.ListTasks(c.Context())
	if err != nil {
		log.Error().Err(err).Msg("task: failed to list")
		return fiber.NewError(fiber.StatusInternalServerError, "failed to list tasks")
	}
	return c.JSON(tasks)
}

func (h *TaskHandler) Get(c fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return err
	}

	task, err := h.service.GetTask(c.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "task not found")
		}
		log.Error().Err(err).Uint("id", id).Msg("task: failed to get")
		return fiber.NewError(fiber.StatusInternalServerError, "failed to get task")
	}
	return c.JSON(task)
}

func (h *TaskHandler) Update(c fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return err
	}

	var req taskRequest
	if err := c.Bind().Body(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	task, err := h.service.UpdateTask(c.Context(), id, service.UpdateTaskInput{
		Title:       req.Title,
		Description: req.Description,
		Done:        req.Done,
	})
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "task not found")
		}
		log.Error().Err(err).Uint("id", id).Msg("task: failed to update")
		return fiber.NewError(fiber.StatusInternalServerError, "failed to update task")
	}
	return c.JSON(task)
}

func (h *TaskHandler) Delete(c fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return err
	}

	if err := h.service.DeleteTask(c.Context(), id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "task not found")
		}
		log.Error().Err(err).Uint("id", id).Msg("task: failed to delete")
		return fiber.NewError(fiber.StatusInternalServerError, "failed to delete task")
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func parseID(c fiber.Ctx) (uint, error) {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return 0, fiber.NewError(fiber.StatusBadRequest, "invalid id")
	}
	return uint(id), nil
}
