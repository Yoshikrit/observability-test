package middleware

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/healthcheck"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// Metrics records RED metrics (rate, errors, duration) per request, labeled by route pattern (not raw path) to avoid cardinality explosions.
func Metrics(serviceName string) fiber.Handler {
	meter := otel.Meter(serviceName)

	requestCounter, err := meter.Int64Counter(
		"http_requests_total",
		metric.WithDescription("Total number of HTTP requests"),
	)
	if err != nil {
		panic(err) // hardcoded config, so a failure here means a setup bug
	}

	duration, err := meter.Float64Histogram(
		"http_request_duration_seconds",
		metric.WithDescription("HTTP request duration in seconds"),
		metric.WithUnit("s"),
		// Overrides the default buckets (tuned for large values) so second-scale durations distribute across buckets instead of collapsing into one.
		metric.WithExplicitBucketBoundaries(
			.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10,
		),
	)
	if err != nil {
		panic(err)
	}

	inflight, err := meter.Int64UpDownCounter(
		"http_requests_inflight",
		metric.WithDescription("Number of HTTP requests currently being processed"),
	)
	if err != nil {
		panic(err)
	}

	return func(c fiber.Ctx) error {
		path := c.Path()
		if path == healthcheck.LivenessEndpoint || path == healthcheck.ReadinessEndpoint || path == "/metrics" {
			return c.Next()
		}

		inflight.Add(c.Context(), 1)
		defer inflight.Add(c.Context(), -1)

		start := time.Now()
		err := c.Next()

		attrs := metric.WithAttributes(
			attribute.String("method", c.Method()),
			attribute.String("route", c.Route().Path),
			attribute.Int("status_code", c.Response().StatusCode()),
		)

		requestCounter.Add(c.Context(), 1, attrs)
		duration.Record(c.Context(), time.Since(start).Seconds(), attrs)

		return err
	}
}
