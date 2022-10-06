package logger

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
	gormlg "gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
)

// GormLoggerWrapper default gorm logger
type GormLoggerWrapper struct {
	Logger                    Logger
	SlowThreshold             time.Duration
	IgnoreRecordNotFoundError bool
}

// LogMode log mode
func (w *GormLoggerWrapper) LogMode(_ gormlg.LogLevel) gormlg.Interface {
	return w
}

// Info print info
func (w *GormLoggerWrapper) Info(ctx context.Context, format string, args ...interface{}) {
	if w.Logger != nil {
		w.Logger.Infof(ctx, format, args...)
	}
}

// Warn print warn message
func (w *GormLoggerWrapper) Warn(ctx context.Context, format string, args ...interface{}) {
	if w.Logger != nil {
		w.Logger.Warnf(ctx, format, args...)
	}
}

// Error print error message
func (w *GormLoggerWrapper) Error(ctx context.Context, format string, args ...interface{}) {
	if w.Logger != nil {
		w.Logger.Errorf(ctx, format, args...)
	}
}

// Trace print sql message
func (w *GormLoggerWrapper) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if w.Logger == nil {
		return
	}

	sql, rows := fc()
	var rowStr string = "-"
	if rows != -1 {
		rowStr = fmt.Sprint(rows)
	}

	elapsed := time.Since(begin)
	logger := w.Logger.
		WithField("caller", utils.FileWithLineNum()).
		WithField("time", fmt.Sprint(float64(elapsed.Nanoseconds())/1e6)+"ms").
		WithField("rows", rowStr).
		WithField("sql", sql)

	switch {
	case err != nil && (!errors.Is(err, gorm.ErrRecordNotFound) || !w.IgnoreRecordNotFoundError):
		logger.WithField("error_details", err.Error()).Errorf(ctx, "Database Error")
	case elapsed > w.SlowThreshold && w.SlowThreshold != 0:
		slowLog := fmt.Sprintf("SLOW SQL >= %s", w.SlowThreshold)
		logger.WithField("warning", slowLog).Warnf(ctx, "Database Warning")
	default:
		logger.Infof(ctx, "Database Info")
	}
}
