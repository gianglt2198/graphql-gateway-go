package logging

import (
	"strings"

	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

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
