package monitoring

import (
	"context"
	"errors"

	"github.com/gianglt2198/graphql-gateway-go/pkg/config"
	"go.opentelemetry.io/otel"
	metric2 "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type MonitoringConfig struct {
	Enabled  bool   `yaml:"enabled,omitempty" envDefault:"false"`
	Endpoint string `yaml:"endpoint,omitempty" envDefault:"localhost:4317"`
}

type (
	monitoring struct {
		log  *AppLogger
		mcfg MonitoringConfig
		ccfg config.Config

		tracerProvider trace.TracerProvider
		meterProvider  metric2.MeterProvider
		shutdownFuncs  []func(context.Context) error
	}

	Monitoring interface {
		Tracer() trace.Tracer
		Meter() metric2.Meter
		Shutdown(ctx context.Context) error
	}
)

func (m *monitoring) Tracer() trace.Tracer {
	return otel.GetTracerProvider().Tracer(m.ccfg.Name)
}

func (m *monitoring) Meter() metric2.Meter {
	return otel.GetMeterProvider().Meter(m.ccfg.Name)
}

func (m *monitoring) Shutdown(ctx context.Context) error {
	var err error
	for _, fn := range m.shutdownFuncs {
		err = errors.Join(err, fn(ctx))
	}
	return err
}

type MonitoringParams struct {
	fx.In

	Log  *AppLogger
	MCfg MonitoringConfig
	CCfg config.Config
}

type MonitoringResult struct {
	fx.Out

	MonitoringClient Monitoring
}

// setupOTelSDK bootstraps the OpenTelemetry pipeline.
// If it does not return an error, make sure to call shutdown for proper cleanup.
func SetupOTelSDK(params MonitoringParams) MonitoringResult {
	provider := newOTel(params.Log, params.CCfg, params.MCfg)
	return MonitoringResult{
		MonitoringClient: provider,
	}
}

func newOTel(log *AppLogger, ccfg config.Config, cfg MonitoringConfig) *monitoring {
	if !cfg.Enabled {
		return nil
	}

	var (
		err error
	)
	ctx := context.Background()

	var shutdownFuncs []func(context.Context) error

	// shutdown calls cleanup functions registered via shutdownFuncs.
	// The errors from the calls are joined.
	// Each registered cleanup will be invoked once.
	shutdown := func(ctx context.Context) error {
		var err error
		for _, fn := range shutdownFuncs {
			err = errors.Join(err, fn(ctx))
		}
		shutdownFuncs = nil
		return err
	}

	// Verbose error handling
	handleErr := func(inErr error) {
		log.Error("OpenTelemetry setup error: %v", zap.Error(inErr))
		err = errors.Join(inErr, shutdown(ctx))
	}

	// Set up propagator.
	prop := newPropagator()
	otel.SetTextMapPropagator(prop)

	// Set up trace provider.
	tracerProvider, err := NewTraceProvider(ctx, ccfg.Name, cfg.Endpoint)
	if err != nil {
		handleErr(err)
		log.Fatal("Monitoring Failed!", zap.Error(err))
	}
	shutdownFuncs = append(shutdownFuncs, tracerProvider.Shutdown)
	otel.SetTracerProvider(tracerProvider)

	// Set up meter provider.
	meterProvider, err := NewMeterProvider(ctx, ccfg.Name, cfg.Endpoint)
	if err != nil {
		handleErr(err)
		log.Fatal("Monitoring Failed!", zap.Error(err))
	}
	shutdownFuncs = append(shutdownFuncs, meterProvider.Shutdown)
	otel.SetMeterProvider(meterProvider)

	return &monitoring{
		log:            log,
		mcfg:           cfg,
		ccfg:           ccfg,
		tracerProvider: tracerProvider,
		meterProvider:  meterProvider,
		shutdownFuncs:  shutdownFuncs,
	}
}

func newPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}
