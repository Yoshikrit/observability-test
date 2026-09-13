package middleware

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/healthcheck"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// fiberHeaderCarrier lets OTel read/write Fiber's request headers.
type fiberHeaderCarrier struct {
	c fiber.Ctx
}

func (h fiberHeaderCarrier) Get(key string) string {
	return h.c.Get(key)
}

func (h fiberHeaderCarrier) Set(key, value string) {
	h.c.Set(key, value)
}

func (h fiberHeaderCarrier) Keys() []string {
	var keys []string
	h.c.Request().Header.VisitAll(func(k, _ []byte) {
		keys = append(keys, string(k))
	})
	return keys
}

// Tracing creates one span per request; register it before RequestLogger so the response is already final by the time this middleware reads it.
func Tracing(serviceName string) fiber.Handler {
	tracer := otel.Tracer(serviceName)

	return func(c fiber.Ctx) error {
		path := c.Path()
		if path == healthcheck.LivenessEndpoint || path == healthcheck.ReadinessEndpoint {
			return c.Next()
		}

		// Reuses the caller's trace if a traceparent header is present, otherwise starts a new one.
		ctx := otel.GetTextMapPropagator().Extract(c.Context(), fiberHeaderCarrier{c: c})

		// Name is a placeholder method and path until routing resolves further down.
		ctx, span := tracer.Start(ctx, c.Method(),
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(
				attribute.String("http.method", c.Method()),
				attribute.String("http.target", c.Path()),
			),
		)
		defer span.End()

		// Downstream code reads the span via c.Context(), so it must be attached here.
		c.SetContext(ctx)

		err := c.Next()

		// Route is only known now that routing has resolved.
		route := c.Route().Path
		span.SetName(c.Method() + " " + route)
		span.SetAttributes(
			attribute.String("http.route", route),
			attribute.Int("http.status_code", c.Response().StatusCode()),
		)

		// Only real server failures (>=500) count as span errors, not 4xx.
		if c.Response().StatusCode() >= 500 {
			span.SetStatus(codes.Error, "server error")
		}

		return err
	}
}
