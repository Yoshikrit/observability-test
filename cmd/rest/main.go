package main

import (
	"context"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/gofiber/fiber/v3/middleware/requestid"
	"github.com/rs/zerolog/log"

	"github.com/Yoshikrit/observability-test/config"
	"github.com/Yoshikrit/observability-test/internal/controller/rest"
	"github.com/Yoshikrit/observability-test/internal/controller/rest/task"
	"github.com/Yoshikrit/observability-test/internal/middleware"
	"github.com/Yoshikrit/observability-test/internal/pkg/logger"
	"github.com/Yoshikrit/observability-test/internal/pkg/validate"
	"github.com/Yoshikrit/observability-test/internal/repository"
	"github.com/Yoshikrit/observability-test/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("observability-api: failed to load config")
	}

	logger.Init(cfg.Log.Level)
	logger.InitAccess(cfg.Log.AccessLogEnabled)

	db, err := config.InitDatabase(cfg.Database.DatabaseUrl)
	if err != nil {
		log.Fatal().Err(err).Msg("observability-api: failed to connect database")
	}
	if err := config.MigrateDatabase(db); err != nil {
		log.Fatal().Err(err).Msg("observability-api: failed to migrate database")
	}
	config.SeedDatabase(db)

	taskRepo := repository.NewTaskRepository(db)
	taskService := service.NewTaskService(taskRepo)
	taskHandler := task.NewTaskHandler(taskService)

	app := fiber.New(fiber.Config{
		AppName:         "observability-test",
		StructValidator: validate.FiberValidator{},
	})

	app.Use(recover.New())
	app.Use(requestid.New())
	app.Use(middleware.RequestLogger())

	rest.RegisterRoutes(app, taskHandler)

	go func() {
		if err := app.Listen(":" + cfg.App.Port); err != nil {
			log.Fatal().Err(err).Msg("observability-api: server error")
		}
	}()
	log.Info().Str("port", cfg.App.Port).Str("env", cfg.App.Env).Msg("observability-api: started")

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()

	log.Info().Msg("observability-api: shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := app.ShutdownWithContext(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("observability-api: forced shutdown")
	}
}
