package logging

import "go.uber.org/zap"

type NopLogger struct{}

func NewNopLogger() *NopLogger {
	return &NopLogger{}
}

func (l NopLogger) Error(_ string, _ ...interface{}) {}
func (l NopLogger) Info(_ string, _ ...interface{})  {}
func (l NopLogger) Debug(_ string, _ ...interface{}) {}
func (l NopLogger) With(_ ...interface{}) Logger     { return l }
func (l NopLogger) ZapLogger() *zap.Logger           { return zap.NewNop() }
