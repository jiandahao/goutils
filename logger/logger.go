package logger

import (
	"context"
	"fmt"
	"time"

	// unsafe
	_ "unsafe"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	logTmFmtWithMS = "2006-01-02 15:04:05.000"
)

// key is an unexported type for keys defined in this package.
// This prevents collisions with keys defined in other packages.
type key string

// DefaultLogger is the default logger
var DefaultLogger Logger = NewDefaultLogger("debug")

// NilLogger is a logger that implements Logger interface but log nothing.
var NilLogger Logger = &nilLogger{}

// Logger represents a logger interface.
type Logger interface {
	Debugf(ctx context.Context, format string, args ...interface{})
	Infof(ctx context.Context, format string, args ...interface{})
	Warnf(ctx context.Context, format string, args ...interface{})
	Errorf(ctx context.Context, format string, args ...interface{})
	WithField(key string, value interface{}) Logger
	Sync() error
}

type nilLogger struct{}

func (l *nilLogger) Debugf(ctx context.Context, format string, args ...interface{}) {}
func (l *nilLogger) Infof(ctx context.Context, format string, args ...interface{})  {}
func (l *nilLogger) Warnf(ctx context.Context, format string, args ...interface{})  {}
func (l *nilLogger) Errorf(ctx context.Context, format string, args ...interface{}) {}
func (l *nilLogger) WithField(key string, value interface{}) Logger                 { return l }
func (l *nilLogger) Sync() error                                                    { return nil }

type defaultLogger struct {
	*zap.Logger
}

// NewDefaultLogger creates a default logger that based on zap.
func NewDefaultLogger(lvl string) Logger {
	cfg := zap.Config{
		Development: false,
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		Encoding: "console",
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:       "ts",
			LevelKey:      "level",
			NameKey:       "logger",
			CallerKey:     "caller",
			FunctionKey:   zapcore.OmitKey,
			MessageKey:    "msg",
			StacktraceKey: "stacktrace",
			LineEnding:    zapcore.DefaultLineEnding,
			EncodeLevel: func(level zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
				enc.AppendString(level.CapitalString())
			},
			EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
				enc.AppendString(t.Format(logTmFmtWithMS))
			},
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller: func(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
				enc.AppendString(caller.TrimmedPath())
			},
			ConsoleSeparator: "|",
		},
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
	}

	if lvl != "" {
		err := cfg.Level.UnmarshalText([]byte(lvl))
		if err != nil {
			panic(err)
		}
	}

	logger, _ := cfg.Build()
	return &defaultLogger{
		Logger: logger,
	}
}

func (l *defaultLogger) Debugf(ctx context.Context, format string, args ...interface{}) {
	l.withMetadataFromContext(ctx).Debug(fmt.Sprintf(format, args...))
}

func (l *defaultLogger) Infof(ctx context.Context, format string, args ...interface{}) {
	l.withMetadataFromContext(ctx).Info(fmt.Sprintf(format, args...))
}

func (l *defaultLogger) Warnf(ctx context.Context, format string, args ...interface{}) {
	l.withMetadataFromContext(ctx).Warn(fmt.Sprintf(format, args...))
}

func (l *defaultLogger) Errorf(ctx context.Context, format string, args ...interface{}) {
	l.withMetadataFromContext(ctx).Error(fmt.Sprintf(format, args...))
}

func (l *defaultLogger) WithField(key string, value interface{}) Logger {
	return &defaultLogger{Logger: l.Logger.With(zap.Reflect(key, value))}
}

func (l *defaultLogger) withMetadataFromContext(ctx context.Context) *defaultLogger {
	md := MetadataFromContext(ctx)

	var logger Logger = l
	for key, values := range md.md {
		if len(values) == 1 {
			logger = logger.WithField(key, values[0])
		} else {
			logger = logger.WithField(key, values)
		}
	}

	return logger.(*defaultLogger)
}

func (l *defaultLogger) Sync() error {
	return l.Logger.Sync()
}

// traceContextKey is the key for metadata in Contexts. It is unexported;
var mdContextKey key

// AppendMetadata appends metadata into context and return a copy of parent context in which a new metadata is added.
func AppendMetadata(ctx context.Context, md MD) context.Context {
	metadata, _ := ctx.Value(mdContextKey).(map[string][]string)
	newmd := make(map[string][]string)

	for key, value := range metadata {
		newmd[key] = append(newmd[key], value...)
	}

	md.Range(func(key string, value []string) {
		newmd[key] = append(newmd[key], value...)
	})

	return context.WithValue(ctx, mdContextKey, newmd)
}

// MetadataFromContext returns metadata within context
func MetadataFromContext(ctx context.Context) MD {
	metadata, _ := ctx.Value(mdContextKey).(map[string][]string)
	return MD{md: metadata}
}

// MD represetns a append-only log metadata.
type MD struct {
	md map[string][]string
}

func NewMetadata() MD {
	return MD{
		md: make(map[string][]string),
	}
}

func (md MD) Append(key string, value string) MD {
	md.md[key] = append(md.md[key], value)
	return md
}

func (md MD) Value(key string) ([]string, bool) {
	v, ok := md.md[key]
	if !ok {
		return nil, false
	}

	vv := make([]string, len(v))

	copy(vv, v)

	return vv, true
}

func (md MD) Range(walkFn func(key string, value []string)) {
	for key := range md.md {
		value, _ := md.Value(key)
		walkFn(key, value)
	}
}
