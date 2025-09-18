package logging

import (
	"context"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/gianglt2198/federation-go/package/common"
	"github.com/gianglt2198/federation-go/package/config"
	"github.com/gianglt2198/federation-go/package/utils"
)

// Logger wraps zap.Logger with additional functionality
type Logger struct {
	*zap.Logger
	serviceName string
}

// NewLogger creates a new logger instance
func NewLogger(config config.AppConfig, natsConfg config.NATSConfig) *Logger {
	var coreArr []zapcore.Core

	if config.Environment == "production" {
		encoderConfig := zap.NewProductionEncoderConfig()
		encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		encoder := zapcore.NewJSONEncoder(encoderConfig)

		if natsConfg.Enabled {
			natsCore := NewNatsCore(natsConfg) // Replace with your NATS logging subject
			natsLogCore := zapcore.NewCore(encoder, natsCore, zapcore.ErrorLevel)
			coreArr = append(coreArr, natsLogCore)
		} else {
			consoleCore := zapcore.NewCore(encoder, zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout)), zapcore.ErrorLevel)
			coreArr = append(coreArr, consoleCore)
		}
	} else {
		encoderConfig := zap.NewProductionEncoderConfig()
		encoder := zapcore.NewJSONEncoder(encoderConfig)
		consoleCore := zapcore.NewCore(encoder, zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout)), zapcore.InfoLevel) // The third and subsequent parameters are the log levels for writing to the file. In ErrorLevel mode, only error - level logs are recorded.

		coreArr = append(coreArr, consoleCore)
	}

	log := zap.New(zapcore.NewTee(coreArr...), zap.AddCaller()) // zap.AddCaller() is used to display the file name and line number and can be omitted.
	defer func() { _ = log.Sync() }()

	log = log.With(zap.String("service", config.Name), zap.String("environment", config.Environment))
	return &Logger{
		serviceName: config.Name,
		Logger:      log,
	}
}

func (l *Logger) GetLogger() *zap.Logger { return l.Logger }

type WrappedLogger struct {
	*zap.Logger
}

func (l *Logger) GetWrappedLogger(ctx context.Context) *WrappedLogger {
	return &WrappedLogger{
		Logger: l.With(l.extractContext(ctx)...),
	}
}

func (l *Logger) extractContext(ctx context.Context) []zap.Field {
	fields := []zap.Field{
		zap.String("service_name", l.serviceName),
		zap.Int("pid", os.Getpid()),
	}

	userID := utils.GetUserIDFromCtx(ctx)
	requestID := utils.GetRequestIDFromCtx(ctx)
	traceID := utils.GetTraceIDFromCtx(ctx)
	spanID := utils.GetSpanIDFromCtx(ctx)

	if userID != "" {
		fields = append(fields, zap.String(string(common.KEY_AUTH_USER_ID), userID))
	}

	if requestID != "" {
		fields = append(fields, zap.String(string(common.KEY_REQUEST_ID), requestID))
	}

	if traceID != "" {
		fields = append(fields, zap.String(string(common.KEY_TRACE_ID), traceID))
	}

	if spanID != "" {
		fields = append(fields, zap.String(string(common.KEY_SPAN_ID), spanID))
	}

	return fields
}
