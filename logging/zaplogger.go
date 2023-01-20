package logging

import (
	"fmt"

	"go.uber.org/zap"
)

type ZapLogger struct {
	logger *zap.SugaredLogger
}

func (l ZapLogger) Error(msg string, keysAndValues ...interface{}) {
	l.logger.Errorw(msg, keysAndValues...)
}

func (l ZapLogger) Info(msg string, keysAndValues ...interface{}) {
	l.logger.Infow(msg, keysAndValues...)
}

func (l ZapLogger) Debug(msg string, keysAndValues ...interface{}) {
	l.logger.Debugw(msg, keysAndValues...)
}

func (l ZapLogger) With(args ...interface{}) Logger {
	return ZapLogger{
		logger: l.logger.With(args...),
	}
}

func (l ZapLogger) ZapLogger() *zap.Logger {
	return l.logger.Desugar()
}

func NewZapProd() (*ZapLogger, error) {
	logger, err := zap.NewProduction()
	sugar := logger.Sugar()

	if err != nil {
		return nil, fmt.Errorf("error setting logger: %w", err)
	}

	return &ZapLogger{
		logger: sugar,
	}, nil
}

func NewZapDev() (*ZapLogger, error) {
	logger, err := zap.NewDevelopment()
	sugar := logger.Sugar()

	if err != nil {
		return nil, fmt.Errorf("error setting logger: %w", err)
	}

	return &ZapLogger{
		logger: sugar,
	}, nil
}

func FromZapLogger(zapLogger *zap.Logger) *ZapLogger {
	return &ZapLogger{
		logger: zapLogger.Sugar(),
	}
}
