package middleware

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/healthcheck"
	"go.opentelemetry.io/otel/trace"

	"github.com/Yoshikrit/observability-test/internal/pkg/logger"
)

const maxLoggedBodyBytes = 8 * 1024

func AppLogger() fiber.Handler {
	return func(c fiber.Ctx) error {
		reqLogger := logger.AppLogger.With().
			Str("request_id", c.GetRespHeader(fiber.HeaderXRequestID)).
			Str("trace_id", trace.SpanContextFromContext(c.Context()).TraceID().String()).
			Logger()

		c.SetContext(logger.WithRequestLogger(c.Context(), reqLogger))

		return c.Next()
	}
}

// isHealthCheckPath excludes high-frequency orchestrator probes from the
// access log - they'd otherwise fire every few seconds and drown out real
// traffic without adding any diagnostic value.
func isHealthCheckPath(path string) bool {
	return path == healthcheck.LivenessEndpoint || path == healthcheck.ReadinessEndpoint
}

func AccessLogger() fiber.Handler {
	return func(c fiber.Ctx) error {
		if isHealthCheckPath(c.Path()) {
			return c.Next()
		}

		start := time.Now()
		reqBody := truncateBody(c.Body())

		err := c.Next()

		event := logger.AccessLogger.Info()
		if err != nil {
			if handleErr := c.App().ErrorHandler(c, err); handleErr != nil {
				_ = c.SendStatus(fiber.StatusInternalServerError)
			}
			event = logger.AccessLogger.Error().Err(err)
		}

		end := time.Now()
		event.
			Str("method", c.Method()).
			Str("path", c.Path()).
			Int("status_code", c.Response().StatusCode()).
			Time("request_time", start).
			Time("response_time", end).
			Dur("duration_ms", end.Sub(start)).
			Str("ip", c.IP()).
			Str("request_id", c.GetRespHeader(fiber.HeaderXRequestID)).
			Str("trace_id", trace.SpanContextFromContext(c.Context()).TraceID().String()).
			Str("request_body", reqBody).
			Str("response_body", truncateBody(c.Response().Body())).
			Msg("request")

		return nil
	}
}

func truncateBody(body []byte) string {
	if len(body) > maxLoggedBodyBytes {
		return string(body[:maxLoggedBodyBytes]) + "...[truncated]"
	}
	return string(body)
}
