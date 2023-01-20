package logging

import "go.uber.org/zap"

type Logger interface {
	Error(msg string, keysAndValues ...interface{})
	Info(msg string, keysAndValues ...interface{})
	Debug(msg string, keysAndValues ...interface{})
	With(args ...interface{}) Logger
	ZapLogger() *zap.Logger
}
