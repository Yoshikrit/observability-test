package metrics

import (
	"context"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	otelprom "go.opentelemetry.io/otel/exporters/prometheus"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.43.0"
)

var meterProvider *sdkmetric.MeterProvider

// Http Handler for metric to call by fiber adaptor
// exposes recorded metrics in Prometheus format for Alloy to scrape.
var Handler http.Handler = promhttp.Handler()

// Init wires up a global MeterProvider backed by the Prometheus exporter; no on/off flag needed since metrics are cheap and inert until scraped.
func Init(serviceName, env string) error {
	res, err := resource.Merge(
		resource.Default(),
		resource.NewSchemaless(
			semconv.ServiceName(serviceName),
			attribute.String("deployment.environment", env),
		),
	)
	if err != nil {
		return err
	}

	// export otel to prometheus format
	exporter, err := otelprom.New()
	if err != nil {
		return err
	}

	// create sdk for collect counter/histrogram with exporter and resource
	meterProvider = sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(exporter),
		sdkmetric.WithResource(res),
	)
	// plug with global provider with these sdk setting
	otel.SetMeterProvider(meterProvider)

	return nil
}

func Shutdown(ctx context.Context) error {
	if meterProvider == nil {
		return nil
	}
	return meterProvider.Shutdown(ctx)
}
