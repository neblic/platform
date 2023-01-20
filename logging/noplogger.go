package logging

import "go.uber.org/zap"

type NopLogger struct{}

func NewNopLogger() *NopLogger {
	return &NopLogger{}
}

func (l NopLogger) Error(msg string, keysAndValues ...interface{}) {}
func (l NopLogger) Info(msg string, keysAndValues ...interface{})  {}
func (l NopLogger) Debug(msg string, keysAndValues ...interface{}) {}
func (l NopLogger) With(args ...interface{}) Logger                { return l }
func (l NopLogger) ZapLogger() *zap.Logger                         { return zap.NewNop() }
