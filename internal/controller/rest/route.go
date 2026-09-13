package rest

import (
	"github.com/gofiber/fiber/v3"

	"github.com/Yoshikrit/observability-test/internal/controller/rest/task"
)

func RegisterRoutes(app *fiber.App, taskHandler *task.TaskHandler) {
	api := app.Group("/api/v1")

	tasks := api.Group("/tasks")
	tasks.Post("/", taskHandler.Create)
	tasks.Get("/", taskHandler.List)
	tasks.Get("/:id", taskHandler.Get)
	tasks.Put("/:id", taskHandler.Update)
	tasks.Delete("/:id", taskHandler.Delete)
}
