package monitoring

import (
	"context"
	"os"
	"strings"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/gianglt2198/graphql-gateway-go/pkg/common"
	"github.com/gianglt2198/graphql-gateway-go/pkg/config"
	"github.com/gianglt2198/graphql-gateway-go/pkg/utils"
)

type AppLogger struct {
	serviceName string
	logger      *zap.Logger
}

func NewLogger(cfg config.BaseConfig) *AppLogger {
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
		serviceName: cfg.Name,
		logger:      log,
	}
}

func (l *AppLogger) GetLogger() *zap.Logger { return l.logger }

func (l *AppLogger) Debug(msg string, fields ...zap.Field) {
	l.logger.Debug(msg, fields...)
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
	l.logger.Info(msg, append(fields, l.extractContext(ctx)...)...)
}

func (l *AppLogger) DebugC(ctx context.Context, msg string, fields ...zap.Field) {
	l.logger.Debug(msg, append(fields, l.extractContext(ctx)...)...)
}

func (l *AppLogger) WarnC(ctx context.Context, msg string, fields ...zap.Field) {
	l.logger.Warn(msg, append(fields, l.extractContext(ctx)...)...)
}

func (l *AppLogger) ErrorC(ctx context.Context, msg string, fields ...zap.Field) {
	l.logger.Error(msg, append(fields, l.extractContext(ctx)...)...)
}

func (l *AppLogger) Fx() fxevent.Logger {
	return &FxLogger{
		Logger: l.logger,
	}
}

// FxLogger is an Fx event logger that logs events to Zap.
type FxLogger struct {
	Logger *zap.Logger

	logLevel   zapcore.Level // default: zapcore.InfoLevel
	errorLevel *zapcore.Level
}

var _ fxevent.Logger = (*FxLogger)(nil)

// UseErrorLevel sets the level of error logs emitted by Fx to level.
func (l *FxLogger) UseErrorLevel(level zapcore.Level) {
	l.errorLevel = &level
}

// UseLogLevel sets the level of non-error logs emitted by Fx to level.
func (l *FxLogger) UseLogLevel(level zapcore.Level) {
	l.logLevel = level
}

func (l *FxLogger) logEvent(msg string, fields ...zap.Field) {
	l.Logger.Log(l.logLevel, msg, fields...)
}

func (l *FxLogger) logError(msg string, fields ...zap.Field) {
	lvl := zapcore.ErrorLevel
	if l.errorLevel != nil {
		lvl = *l.errorLevel
	}
	l.Logger.Log(lvl, msg, fields...)
}

// LogEvent logs the given event to the provided Zap logger.
func (l *FxLogger) LogEvent(event fxevent.Event) {
	switch e := event.(type) {
	case *fxevent.OnStartExecuting:
		l.logEvent("OnStart hook executing",
			zap.String("callee", e.FunctionName),
			zap.String("caller", e.CallerName),
		)
	case *fxevent.OnStartExecuted:
		if e.Err != nil {
			l.logError("OnStart hook failed",
				zap.String("callee", e.FunctionName),
				zap.String("caller", e.CallerName),
				zap.Error(e.Err),
			)
		} else {
			l.logEvent("OnStart hook executed",
				zap.String("callee", e.FunctionName),
				zap.String("caller", e.CallerName),
				zap.String("runtime", e.Runtime.String()),
			)
		}
	case *fxevent.OnStopExecuting:
		l.logEvent("OnStop hook executing",
			zap.String("callee", e.FunctionName),
			zap.String("caller", e.CallerName),
		)
	case *fxevent.OnStopExecuted:
		if e.Err != nil {
			l.logError("OnStop hook failed",
				zap.String("callee", e.FunctionName),
				zap.String("caller", e.CallerName),
				zap.Error(e.Err),
			)
		} else {
			l.logEvent("OnStop hook executed",
				zap.String("callee", e.FunctionName),
				zap.String("caller", e.CallerName),
				zap.String("runtime", e.Runtime.String()),
			)
		}
	case *fxevent.Supplied:
		if e.Err != nil {
			l.logError("error encountered while applying options",
				zap.String("type", e.TypeName),
				zap.Strings("stacktrace", e.StackTrace),
				zap.Strings("moduletrace", e.ModuleTrace),
				moduleField(e.ModuleName),
				zap.Error(e.Err))
		} else {
			l.logEvent("supplied",
				zap.String("type", e.TypeName),
				moduleField(e.ModuleName),
			)
		}
	case *fxevent.Provided:
		for _, rtype := range e.OutputTypeNames {
			l.logEvent("provided",
				zap.String("type", rtype),
				moduleField(e.ModuleName),
				maybeBool("private", e.Private),
			)
		}
		if e.Err != nil {
			l.logError("error encountered while applying options",
				moduleField(e.ModuleName),
				zap.Strings("stacktrace", e.StackTrace),
				zap.Strings("moduletrace", e.ModuleTrace),
				zap.Error(e.Err))
		}
	case *fxevent.Replaced:
		for _, rtype := range e.OutputTypeNames {
			l.logEvent("replaced",
				moduleField(e.ModuleName),
				zap.String("type", rtype),
			)
		}
		if e.Err != nil {
			l.logError("error encountered while replacing",
				zap.Strings("stacktrace", e.StackTrace),
				zap.Strings("moduletrace", e.ModuleTrace),
				moduleField(e.ModuleName),
				zap.Error(e.Err))
		}
	case *fxevent.Decorated:
		for _, rtype := range e.OutputTypeNames {
			l.logEvent("decorated",
				zap.String("decorator", e.DecoratorName),
				moduleField(e.ModuleName),
				zap.String("type", rtype),
			)
		}
		if e.Err != nil {
			l.logError("error encountered while applying options",
				zap.Strings("stacktrace", e.StackTrace),
				zap.Strings("moduletrace", e.ModuleTrace),
				moduleField(e.ModuleName),
				zap.Error(e.Err))
		}
	case *fxevent.Run:
		if e.Err != nil {
			l.logError("error returned",
				zap.String("name", e.Name),
				zap.String("kind", e.Kind),
				moduleField(e.ModuleName),
				zap.Error(e.Err),
			)
		} else {
			l.logEvent("run",
				zap.String("name", e.Name),
				zap.String("kind", e.Kind),
				zap.String("runtime", e.Runtime.String()),
				moduleField(e.ModuleName),
			)
		}
	case *fxevent.Invoking:
		// Do not log stack as it will make logs hard to read.
		l.logEvent("invoking",
			zap.String("function", e.FunctionName),
			moduleField(e.ModuleName),
		)
	case *fxevent.Invoked:
		if e.Err != nil {
			l.logError("invoke failed",
				zap.Error(e.Err),
				zap.String("stack", e.Trace),
				zap.String("function", e.FunctionName),
				moduleField(e.ModuleName),
			)
		}
	case *fxevent.Stopping:
		l.logEvent("received signal",
			zap.String("signal", strings.ToUpper(e.Signal.String())))
	case *fxevent.Stopped:
		if e.Err != nil {
			l.logError("stop failed", zap.Error(e.Err))
		}
	case *fxevent.RollingBack:
		l.logError("start failed, rolling back", zap.Error(e.StartErr))
	case *fxevent.RolledBack:
		if e.Err != nil {
			l.logError("rollback failed", zap.Error(e.Err))
		}
	case *fxevent.Started:
		if e.Err != nil {
			l.logError("start failed", zap.Error(e.Err))
		} else {
			l.logEvent("started")
		}
	case *fxevent.LoggerInitialized:
		if e.Err != nil {
			l.logError("custom logger initialization failed", zap.Error(e.Err))
		} else {
			l.logEvent("initialized custom fxevent.Logger", zap.String("function", e.ConstructorName))
		}
	}
}

func moduleField(name string) zap.Field {
	if len(name) == 0 {
		return zap.Skip()
	}
	return zap.String("module", name)
}

func maybeBool(name string, b bool) zap.Field {
	if b {
		return zap.Bool(name, true)
	}
	return zap.Skip()
}
