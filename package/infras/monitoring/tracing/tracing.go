package tracing

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	metric2 "go.opentelemetry.io/otel/metric"
	metricNoop "go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
	tracerNoop "go.opentelemetry.io/otel/trace/noop"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/gianglt2198/federation-go/package/config"
	"github.com/gianglt2198/federation-go/package/infras/monitoring/logging"
)

type tracingClient struct {
	shutdown func(context.Context) error
}

type TracingParams struct {
	fx.In

	Log           *logging.Logger
	AppConfig     config.AppConfig
	TracingConfig config.TracingConfig
}

func NewTracing(params TracingParams) *tracingClient {
	if !params.TracingConfig.Enabled {
		tracerProvider := tracerNoop.NewTracerProvider()
		otel.SetTracerProvider(tracerProvider)
		metricProvider := metricNoop.NewMeterProvider()
		otel.SetMeterProvider(metricProvider)
		return &tracingClient{
			shutdown: func(ctx context.Context) error {
				return nil
			},
		}
	}

	shutdown, err := setupOTelSDK(context.Background(), params.Log, params.AppConfig, params.TracingConfig)
	if err != nil {
		panic(err)
	}

	params.Log.GetLogger().Info("Tracing OTelemetry initialized successfully",
		zap.String("endpoint", params.TracingConfig.Endpoint),
		zap.String("name", params.AppConfig.Name),
	)

	return &tracingClient{
		shutdown: shutdown,
	}
}

func Tracer(name string) trace.Tracer {
	return otel.GetTracerProvider().Tracer(name)
}

func Meter(name string) metric2.Meter {
	return otel.GetMeterProvider().Meter(name)
}

func Propagation() propagation.TextMapPropagator {
	return otel.GetTextMapPropagator()
}

func newPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}

func newTraceProvider(ctx context.Context, appCfg config.AppConfig, tracingCfg config.TracingConfig) (*sdktrace.TracerProvider, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	conn, err := grpc.DialContext(ctx, tracingCfg.Endpoint,
		// Note the use of insecure transport here. TLS is recommended in production.
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, err
	}

	traceExporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithGRPCConn(conn),
		otlptracegrpc.WithCompressor("gzip"),
	)
	if err != nil {
		return nil, err
	}

	// Create a resource with service information
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(appCfg.Name),
			semconv.DeploymentEnvironment(appCfg.Environment),
			semconv.ServiceVersion("1.0.0"),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("creating trace resource: %w", err)
	}

	// Create trace provider
	traceProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	return traceProvider, nil
}

func newMeterProvider(ctx context.Context, appCfg config.AppConfig, tracingCfg config.TracingConfig) (*metric.MeterProvider, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	conn, err := grpc.DialContext(ctx, tracingCfg.Endpoint,
		// Note the use of insecure transport here. TLS is recommended in production.
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, err
	}

	metricExporter, err := otlpmetricgrpc.New(ctx,
		otlpmetricgrpc.WithGRPCConn(conn),
		otlpmetricgrpc.WithCompressor("gzip"),
	)
	if err != nil {
		return nil, err
	}

	// Create a resource with service information
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(appCfg.Name),
			semconv.DeploymentEnvironment(appCfg.Environment),
			semconv.ServiceVersion("1.0.0"),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("creating metric resource: %w", err)
	}

	meterProvider := metric.NewMeterProvider(
		metric.WithReader(metric.NewPeriodicReader(metricExporter, metric.WithInterval(5*time.Second))),
		metric.WithResource(res),
	)
	return meterProvider, nil
}

// setupOTelSDK bootstraps the OpenTelemetry pipeline.
// If it does not return an error, make sure to call shutdown for proper cleanup.
func setupOTelSDK(ctx context.Context, log *logging.Logger, appCfg config.AppConfig, tracingCfg config.TracingConfig) (shutdown func(context.Context) error, err error) {
	log.Info("Initializing OpenTelemetry SDK")

	// // Verbose error handling
	// handleErr := func(inErr error) {
	// 	log.Info("OpenTelemetry setup error", zap.Error(inErr))
	// 	err = errors.Join(inErr, shutdown(ctx))
	// }

	var shutdownFuncs []func(context.Context) error

	// shutdown calls cleanup functions registered via shutdownFuncs.
	// The errors from the calls are joined.
	// Each registered cleanup will be invoked once.
	shutdown = func(ctx context.Context) error {
		var err error
		for _, fn := range shutdownFuncs {
			err = errors.Join(err, fn(ctx))
		}
		shutdownFuncs = nil
		return err
	}

	// Set up propagator.
	prop := newPropagator()
	otel.SetTextMapPropagator(prop)

	// Set up trace provider.
	tracerProvider, err := newTraceProvider(ctx, appCfg, tracingCfg)
	if err != nil {
		panic(err)
	}
	shutdownFuncs = append(shutdownFuncs, tracerProvider.Shutdown)
	otel.SetTracerProvider(tracerProvider)

	// Set up meter provider.
	meterProvider, err := newMeterProvider(ctx, appCfg, tracingCfg)
	if err != nil {
		panic(err)
	}
	shutdownFuncs = append(shutdownFuncs, meterProvider.Shutdown)
	otel.SetMeterProvider(meterProvider)

	return
}
