package common_logger

import (
	"errors"
	"sync"

	"github.com/vlbarou/logger/default_logger"
)

var (
	// assign a default logger, in case `createLogger` is not explicitly called (e.g. in unit tests)
	loggerInstance Logger
	once           sync.Once
	initError      error
)

func GetLogger(loggerType LoggerType, config ...Config) (Logger, error) {
	once.Do(func() {
		switch loggerType {
		case Zap:
			loggerInstance = startLogger(config)
		default:
			loggerInstance = default_logger.New()
			initError = errors.New("required logger not found. Default logger initialized")
		}
	})
	return loggerInstance, initError
}

func Shutdown() error {
	return loggerInstance.Shutdown()
}

// Info logs an info-level message
func Info(msg string, fields ...any) {
	loggerInstance.Info(msg, fields...)
}

// Debug logs an debug-level message
func Debug(msg string, fields ...any) {
	loggerInstance.Debug(msg, fields...)
}

// Warn logs an error-level message
func Warn(msg string, fields ...any) {
	loggerInstance.Warn(msg, fields...)
}

// Error logs an error-level message
func Error(msg string, fields ...any) {
	loggerInstance.Error(msg, fields...)
}
