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

func (l ZapLogger) Warn(msg string, keysAndValues ...interface{}) {
	l.logger.Warnw(msg, keysAndValues...)
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

type logType int

const (
	logTypeProd logType = iota
	logTypeDev
)

func buildLogger(lvl string, t logType) (*ZapLogger, error) {
	atomicLevel, err := zap.ParseAtomicLevel(lvl)
	if err != nil {
		return nil, fmt.Errorf("error parsing log level: %w", err)
	}

	var cfg zap.Config
	switch t {
	case logTypeProd:
		cfg = zap.NewProductionConfig()
	case logTypeDev:
		fallthrough
	default:
		cfg = zap.NewDevelopmentConfig()
	}

	cfg.Level = atomicLevel
	logger, err := cfg.Build(
		zap.AddCallerSkip(1),
	)
	if err != nil {
		return nil, fmt.Errorf("error building logger: %w", err)
	}

	return &ZapLogger{
		logger: logger.Sugar(),
	}, nil
}

func NewZapProd(lvl string) (*ZapLogger, error) {
	return buildLogger(lvl, logTypeProd)
}

func NewZapDev() (*ZapLogger, error) {
	return buildLogger("debug", logTypeDev)
}

func FromZapLogger(zapLogger *zap.Logger) *ZapLogger {
	return &ZapLogger{
		logger: zapLogger.Sugar(),
	}
}
