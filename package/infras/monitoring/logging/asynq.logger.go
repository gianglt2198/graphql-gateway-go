package logging

import (
	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

// AsynqLogger is an Fx event logger that logs events to Zap.
type AsynqLogger struct {
	*zap.SugaredLogger

	// logLevel   zapcore.Level // default: zapcore.InfoLevel
	// errorLevel *zapcore.Level
}

var _ asynq.Logger = (*AsynqLogger)(nil)

func (l *Logger) Asynq() asynq.Logger {
	return &AsynqLogger{
		SugaredLogger: l.Sugar(),
	}
}
