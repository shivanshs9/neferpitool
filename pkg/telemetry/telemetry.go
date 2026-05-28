package telemetry

import (
	"context"
	"os"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

var enabled bool

// Init configures OTLP trace export when OTEL_EXPORTER_OTLP_ENDPOINT is set.
func Init(ctx context.Context) (func(context.Context) error, error) {
	endpoint := strings.TrimSpace(os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"))
	if endpoint == "" {
		return func(context.Context) error { return nil }, nil
	}

	opts := []otlptracehttp.Option{
		otlptracehttp.WithEndpoint(trimEndpointHost(endpoint)),
	}
	if os.Getenv("OTEL_EXPORTER_OTLP_INSECURE") == "true" || strings.HasPrefix(endpoint, "http://") {
		opts = append(opts, otlptracehttp.WithInsecure())
	}

	exporter, err := otlptracehttp.New(ctx, opts...)
	if err != nil {
		return nil, err
	}

	serviceName := serviceNameFromEnv()
	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			"",
			attribute.String("service.name", serviceName),
		),
	)
	if err != nil {
		return nil, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)
	enabled = true
	return tp.Shutdown, nil
}

// Enabled reports whether OTLP trace export is active.
func Enabled() bool {
	return enabled
}

// RecordEvent emits an OTEL span for a domain monitoring event (shows in SigNoz Traces).
func RecordEvent(ctx context.Context, name string, attrs map[string]string) {
	if !enabled {
		return
	}
	tracer := otel.Tracer("github.com/moorada/neferpitool")
	ctx, span := tracer.Start(ctx, name)
	defer span.End()

	span.SetAttributes(attribute.String("event.name", name))
	for k, v := range attrs {
		span.SetAttributes(attribute.String(k, v))
	}
}

func serviceNameFromEnv() string {
	if v := os.Getenv("OTEL_SERVICE_NAME"); v != "" {
		return v
	}
	raw := os.Getenv("OTEL_RESOURCE_ATTRIBUTES")
	for _, part := range strings.Split(raw, ",") {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "service.name=") {
			return strings.TrimPrefix(part, "service.name=")
		}
	}
	return "neferpitool"
}

func trimEndpointHost(endpoint string) string {
	endpoint = strings.TrimPrefix(endpoint, "https://")
	endpoint = strings.TrimPrefix(endpoint, "http://")
	if i := strings.Index(endpoint, "/"); i >= 0 {
		endpoint = endpoint[:i]
	}
	return endpoint
}

// BackgroundContext returns a root context for background work spans.
func BackgroundContext() context.Context {
	return context.Background()
}

// MonitorCycleSpan wraps a full background monitoring cycle.
func MonitorCycleSpan(ctx context.Context) (context.Context, func()) {
	if !enabled {
		return ctx, func() {}
	}
	tracer := otel.Tracer("github.com/moorada/neferpitool")
	ctx, span := tracer.Start(ctx, "neferpitool.monitor.cycle")
	return ctx, func() {
		span.End()
		time.Sleep(0) // allow batch flush ordering
	}
}
