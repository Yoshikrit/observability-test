package main

import (
	"context"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/adaptor"
	"github.com/gofiber/fiber/v3/middleware/healthcheck"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/gofiber/fiber/v3/middleware/requestid"

	"github.com/Yoshikrit/observability-test/config"
	"github.com/Yoshikrit/observability-test/internal/controller/rest"
	"github.com/Yoshikrit/observability-test/internal/controller/rest/task"
	"github.com/Yoshikrit/observability-test/internal/middleware"
	"github.com/Yoshikrit/observability-test/internal/pkg/logger"
	"github.com/Yoshikrit/observability-test/internal/pkg/metrics"
	"github.com/Yoshikrit/observability-test/internal/pkg/tracing"
	"github.com/Yoshikrit/observability-test/internal/pkg/validate"
	"github.com/Yoshikrit/observability-test/internal/repository"
	"github.com/Yoshikrit/observability-test/internal/service"
)

const serviceName = "observability-test"

func main() {
	logger.Init("info") // safe default so a config-load failure can still be logged

	cfg, err := config.Load()
	if err != nil {
		logger.AppLogger.Fatal().Err(err).Msg("observability-api: failed to load config")
	}

	logger.Init(cfg.Log.Level) // reconfigure with the actual configured level
	logger.InitAccess(cfg.Log.AccessLogEnabled)

	if err := tracing.Init(serviceName, cfg.App.Env, cfg.Tracing.ConsoleExportEnabled); err != nil {
		logger.AppLogger.Fatal().Err(err).Msg("observability-api: failed to init tracing")
	}

	if err := metrics.Init(serviceName, cfg.App.Env); err != nil {
		logger.AppLogger.Fatal().Err(err).Msg("observability-api: failed to init metrics")
	}

	db, err := config.InitDatabase(cfg.Database.DatabaseUrl)
	if err != nil {
		logger.AppLogger.Fatal().Err(err).Msg("observability-api: failed to connect database")
	}
	if err := config.MigrateDatabase(db); err != nil {
		logger.AppLogger.Fatal().Err(err).Msg("observability-api: failed to migrate database")
	}
	config.SeedDatabase(db)

	taskRepo := repository.NewTaskRepository(db)
	taskService := service.NewTaskService(taskRepo)
	taskHandler := task.NewTaskHandler(taskService)

	app := fiber.New(fiber.Config{
		AppName:         serviceName,
		StructValidator: validate.FiberValidator{},
	})

	app.Use(recover.New())
	app.Use(requestid.New())
	app.Use(middleware.Tracing(serviceName))
	app.Use(middleware.Metrics(serviceName))
	app.Use(middleware.AppLogger())
	app.Use(middleware.AccessLogger())

	// Prometheus scrape endpoint - the permanent shape Alloy will read from later.
	app.Get("/metrics", adaptor.HTTPHandler(metrics.Handler))

	// Liveness: is the process itself alive? No external dependency checks
	app.Get(healthcheck.LivenessEndpoint, healthcheck.New())

	// Readiness: can this instance actually serve traffic right now (database or external dependency check)
	app.Get(healthcheck.ReadinessEndpoint, healthcheck.New(healthcheck.Config{
		Probe: func(c fiber.Ctx) bool {
			sqlDB, err := db.DB()
			if err != nil {
				return false
			}
			return sqlDB.PingContext(c.Context()) == nil
		},
	}))

	rest.RegisterRoutes(app, taskHandler)

	go func() {
		if err := app.Listen(":" + cfg.App.Port); err != nil {
			logger.AppLogger.Fatal().Err(err).Msg("observability-api: server error")
		}
	}()
	logger.AppLogger.Info().Str("port", cfg.App.Port).Str("env", cfg.App.Env).Msg("observability-api: started")

	// context to geting shutdown signal
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()

	// context for clearing works before shutdown
	logger.AppLogger.Info().Msg("observability-api: shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := app.ShutdownWithContext(shutdownCtx); err != nil {
		logger.AppLogger.Error().Err(err).Msg("observability-api: forced shutdown")
	}
	if err := tracing.Shutdown(shutdownCtx); err != nil {
		logger.AppLogger.Error().Err(err).Msg("observability-api: failed to shut down tracing")
	}
	if err := metrics.Shutdown(shutdownCtx); err != nil {
		logger.AppLogger.Error().Err(err).Msg("observability-api: failed to shut down metrics")
	}
}
