package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewSugaredLogger new a zap sugared logger
func NewSugaredLogger(name string, level string) *zap.SugaredLogger {
	cfg := zap.NewProductionConfig()
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	cfg.EncoderConfig.EncodeDuration = zapcore.StringDurationEncoder
	cfg.Encoding = "console"
	if level != "" {
		err := cfg.Level.UnmarshalText([]byte(level))
		if err != nil {
			panic(err)
		}
	}
	logger, _ := cfg.Build()
	sugarLogger := logger.Named(name).Sugar()
	return sugarLogger
}
