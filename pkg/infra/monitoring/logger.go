package monitoring

import (
	"context"
	"os"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/gianglt2198/graphql-gateway-go/pkg/common"
	"github.com/gianglt2198/graphql-gateway-go/pkg/utils"
)

type LoggerConfig struct {
	Env string `yaml:"env,omitempty" envDefault:"dev"`
}

type AppLogger struct {
	serviceName string
	*zap.Logger
}

func NewLogger(cfg LoggerConfig, serviceName string) *AppLogger {
	var coreArr []zapcore.Core

	// Log levels
	highPriority := zap.LevelEnablerFunc(func(lev zapcore.Level) bool { // Error level
		return lev >= zap.ErrorLevel
	})
	lowPriority := zap.LevelEnablerFunc(func(lev zapcore.Level) bool { // Info and debug levels, debug level is the lowest
		return lev < zap.ErrorLevel && lev >= zap.DebugLevel
	})

	if cfg.Env == "prod" {
		encoderConfig := zap.NewProductionEncoderConfig()
		encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		encoder := zapcore.NewJSONEncoder(encoderConfig)

		infoLogCore := zapcore.NewCore(encoder, zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout)), lowPriority)   // The third and subsequent parameters are the log levels for writing to the file. In ErrorLevel mode, only error - level logs are recorded.
		errorLogCore := zapcore.NewCore(encoder, zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout)), highPriority) // The third and subsequent parameters are the log levels for writing to the file. In ErrorLevel mode, only error - level logs are recorded.

		coreArr = append(coreArr, infoLogCore, errorLogCore)

	} else {
		encoderConfig := zap.NewProductionEncoderConfig()
		encoder := zapcore.NewJSONEncoder(encoderConfig)
		consoleCore := zapcore.NewCore(encoder, zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdin)), zapcore.InfoLevel) // The third and subsequent parameters are the log levels for writing to the file. In ErrorLevel mode, only error - level logs are recorded.

		coreArr = append(coreArr, consoleCore)
	}

	log := zap.New(zapcore.NewTee(coreArr...), zap.AddCaller()) // zap.AddCaller() is used to display the file name and line number and can be omitted.
	// defer log.Sync()

	log.WithOptions()
	return &AppLogger{
		serviceName: serviceName,
		Logger:      log,
	}
}

func (l *AppLogger) GetLogger() *zap.Logger { return l.Logger }

func (l *AppLogger) Debug(msg string, fields ...zap.Field) {
	l.Logger.Debug(msg, fields...)
}

func (l *AppLogger) extractContext(ctx context.Context) []zap.Field {
	span := trace.SpanFromContext(ctx)
	spanContext := span.SpanContext()
	requestID := utils.GetRequestIDFromCtx(ctx)
	userID := utils.GetUserIDFromCtx(ctx)
	return []zap.Field{
		zap.String("service_name", l.serviceName),
		zap.String(string(common.KEY_REQUEST_ID), requestID),
		zap.String(string(common.KEY_AUTH_USER_ID), userID),
		zap.Int("pid", os.Getpid()),
		zap.String("trace_id", spanContext.TraceID().String()),
		zap.String("span_id", spanContext.SpanID().String()),
	}
}

func (l *AppLogger) InfoC(ctx context.Context, msg string, fields ...zap.Field) {
	l.Logger.Info(msg, append(fields, l.extractContext(ctx)...)...)
}

func (l *AppLogger) DebugC(ctx context.Context, msg string, fields ...zap.Field) {
	l.Logger.Debug(msg, append(fields, l.extractContext(ctx)...)...)
}

func (l *AppLogger) WarnC(ctx context.Context, msg string, fields ...zap.Field) {
	l.Logger.Warn(msg, append(fields, l.extractContext(ctx)...)...)
}

func (l *AppLogger) ErrorC(ctx context.Context, msg string, fields ...zap.Field) {
	l.Logger.Error(msg, append(fields, l.extractContext(ctx)...)...)
}
