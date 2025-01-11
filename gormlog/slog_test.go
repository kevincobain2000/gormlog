package glog

import (
	"context"
	"errors"
	"testing"
	"time"

	"gorm.io/gorm/logger"
)

func TestGormSlog_LogMode(t *testing.T) {
	tests := []struct {
		name          string
		initialLevel  logger.LogLevel
		newLevel      logger.LogLevel
		expectedLevel logger.LogLevel
	}{
		{
			name:          "change from info to error",
			initialLevel:  logger.Info,
			newLevel:      logger.Error,
			expectedLevel: logger.Error,
		},
		{
			name:          "change from error to silent",
			initialLevel:  logger.Error,
			newLevel:      logger.Silent,
			expectedLevel: logger.Silent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log := &Slog{LogLevel: tt.initialLevel}
			result := log.LogMode(tt.newLevel)

			gormLog, ok := result.(*Slog)
			if !ok {
				t.Error("LogMode did not return *GormSlog")
			}

			if gormLog.LogLevel != tt.expectedLevel {
				t.Errorf("LogMode() got = %v, want %v", gormLog.LogLevel, tt.expectedLevel)
			}
		})
	}
}

func TestGormSlog_Trace(t *testing.T) {
	tests := []struct {
		name          string
		logLevel      logger.LogLevel
		slowThreshold time.Duration
		elapsed       time.Duration
		sql           string
		rows          int64
		err           error
		shouldLog     bool
	}{
		{
			name:          "silent mode",
			logLevel:      logger.Silent,
			slowThreshold: time.Second,
			elapsed:       time.Millisecond,
			sql:           "SELECT * FROM users",
			rows:          10,
			err:           nil,
			shouldLog:     false,
		},
		{
			name:          "error logging",
			logLevel:      logger.Error,
			slowThreshold: time.Second,
			elapsed:       time.Millisecond,
			sql:           "SELECT * FROM users",
			rows:          10,
			err:           errors.New("database error"),
			shouldLog:     true,
		},
		{
			name:          "slow query warning",
			logLevel:      logger.Warn,
			slowThreshold: time.Millisecond,
			elapsed:       time.Second,
			sql:           "SELECT * FROM users",
			rows:          10,
			err:           nil,
			shouldLog:     true,
		},
		{
			name:          "info logging",
			logLevel:      logger.Info,
			slowThreshold: time.Second,
			elapsed:       time.Millisecond,
			sql:           "SELECT * FROM users",
			rows:          10,
			err:           nil,
			shouldLog:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(*testing.T) {
			log := &Slog{
				LogLevel:      tt.logLevel,
				SlowThreshold: tt.slowThreshold,
			}

			begin := time.Now().Add(-tt.elapsed)
			fc := func() (string, int64) {
				return tt.sql, tt.rows
			}

			// Create a context with cancel to simulate timeout scenarios
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			log.Trace(ctx, begin, fc, tt.err)
			// Note: In a real implementation, you might want to mock slog
			// and verify the actual logging output
		})
	}
}

func TestFormatSQL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "single line query",
			input:    "SELECT * FROM users",
			expected: "SELECT * FROM users",
		},
		{
			name: "multi line query",
			input: `
				SELECT *
				FROM users
				WHERE id = 1
			`,
			expected: "SELECT * FROM users WHERE id = 1",
		},
		{
			name:     "query with extra spaces",
			input:    "  SELECT  *  FROM  users  ",
			expected: "SELECT * FROM users",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatSQL(tt.input)
			if result != tt.expected {
				t.Errorf("formatSQL() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGormSlog_LogLevels(t *testing.T) {
	tests := []struct {
		name      string
		level     logger.LogLevel
		fn        func(*Slog, context.Context, string, ...interface{})
		message   string
		data      []interface{}
		shouldLog bool
	}{
		{
			name:      "info level enabled",
			level:     logger.Info,
			fn:        (*Slog).Info,
			message:   "test message",
			data:      []interface{}{"value"},
			shouldLog: true,
		},
		{
			name:      "info level disabled",
			level:     logger.Silent,
			fn:        (*Slog).Info,
			message:   "test message",
			data:      []interface{}{"value"},
			shouldLog: false,
		},
		{
			name:      "warn level enabled",
			level:     logger.Warn,
			fn:        (*Slog).Warn,
			message:   "test message",
			data:      []interface{}{"value"},
			shouldLog: true,
		},
		{
			name:      "error level enabled",
			level:     logger.Error,
			fn:        (*Slog).Error,
			message:   "test message",
			data:      []interface{}{"value"},
			shouldLog: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(*testing.T) {
			log := &Slog{LogLevel: tt.level}
			ctx := context.Background()

			// Call the logging function
			tt.fn(log, ctx, tt.message, tt.data...)
			// Note: In a real implementation, you might want to mock slog
			// and verify the actual logging output
		})
	}
}
