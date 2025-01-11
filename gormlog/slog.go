package glog

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"gorm.io/gorm/logger"
)

var _ logger.Interface = (*Slog)(nil)

type Slog struct {
	LogLevel      logger.LogLevel
	SlowThreshold time.Duration
}

func (log *Slog) LogMode(level logger.LogLevel) logger.Interface {
	log.LogLevel = level
	return log
}

func (log *Slog) Info(ctx context.Context, msg string, data ...interface{}) {
	if log.LogLevel >= logger.Info {
		slog.InfoContext(ctx, fmt.Sprintf(msg, data...))
	}
}

func (log *Slog) Warn(ctx context.Context, msg string, data ...interface{}) {
	if log.LogLevel >= logger.Warn {
		slog.WarnContext(ctx, fmt.Sprintf(msg, data...))
	}
}

func (log *Slog) Error(ctx context.Context, msg string, data ...interface{}) {
	if log.LogLevel >= logger.Error {
		slog.ErrorContext(ctx, fmt.Sprintf(msg, data...))
	}
}

func (log *Slog) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if log.LogLevel <= logger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()

	elapsedMs := float64(elapsed.Nanoseconds()) / 1e6

	rowsStr := fmt.Sprintf("%d", rows)
	if rows == -1 {
		rowsStr = "-"
	}

	formattedSQL := formatSQL(sql)

	logMessage := func(level logger.LogLevel, msg string, keyVals ...interface{}) {
		switch level {
		case logger.Error:
			slog.ErrorContext(ctx, msg, keyVals...)
		case logger.Warn:
			slog.WarnContext(ctx, msg, keyVals...)
		case logger.Info:
			slog.InfoContext(ctx, msg, keyVals...)
		}
	}

	message := fmt.Sprintf("[%.3fms] [rows:%s] %s", elapsedMs, rowsStr, formattedSQL)

	switch {
	case err != nil && log.LogLevel >= logger.Error:
		logMessage(logger.Error, message, slog.String("error", err.Error()))
	case elapsed > log.SlowThreshold && log.SlowThreshold != 0 && log.LogLevel >= logger.Warn:
		logMessage(logger.Warn, message)
	case log.LogLevel == logger.Info:
		logMessage(logger.Info, message)
	}
}

func formatSQL(sql string) string {
	// Replace all newlines with spaces
	sql = strings.ReplaceAll(sql, "\n", " ")

	// Replace multiple spaces with a single space
	sql = strings.Join(strings.Fields(sql), " ")

	return sql
}
