package log

import (
	"os"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var once sync.Once
var logger = build(zap.InfoLevel)

// SetLogger set global logger
func SetLogger(l *zap.Logger) {
	once.Do(func() {
		logger = l
	})
}

func GetLogger() *zap.Logger {
	return logger
}

func NewDefaultLoggerWithLevel(level zapcore.Level) *zap.Logger {
	return build(level)
}

func build(level zapcore.Level) *zap.Logger {
	lv, err := zapcore.ParseLevel(os.Getenv("LOG_LEVEL"))
	if err == nil {
		level = lv
	}

	cfg := zap.Config{
		Level:       zap.NewAtomicLevelAt(level),
		Development: false,
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		Encoding:         "json",
		EncoderConfig:    zap.NewProductionEncoderConfig(),
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stdout"},
	}
	lg, _ := cfg.Build(zap.AddCallerSkip(1), zap.WithCaller(false))
	return lg
}
