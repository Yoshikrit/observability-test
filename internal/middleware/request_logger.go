package middleware

import (
	"time"

	"github.com/gofiber/fiber/v3"

	"github.com/Yoshikrit/observability-test/internal/pkg/logger"
)

// Bodies larger than this are truncated before logging, as a guard against a
// single oversized payload (e.g. an upload) blowing up one log line.
const maxLoggedBodyBytes = 8 * 1024

// RequestLogger logs one structured access log line per request, including
// request/response bodies unconditionally. This is a deliberate trade-off for
// this project (fake seed data, short Loki retention as a compensating
// control) - don't carry unredacted body logging into a service handling
// real user data without adding field-level redaction first.
func RequestLogger() fiber.Handler {
	return func(c fiber.Ctx) error {
		start := time.Now()
		reqBody := truncateBody(c.Body())

		err := c.Next()

		event := logger.AccessLogger.Info()
		if err != nil {
			// The router only invokes the app's ErrorHandler (which actually
			// writes the error response body) *after* this whole middleware
			// chain returns - so at this point c.Response() is still empty.
			// Invoke it ourselves so status/body are populated before we log,
			// then swallow err so the router doesn't invoke it a second time.
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
