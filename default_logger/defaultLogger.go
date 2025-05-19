package default_logger

import (
	"fmt"
	"log"
	"os"
	"time"
)

type DefaultLogger struct {
	logger *log.Logger
	file   *os.File
}

func New() *DefaultLogger {
	return &DefaultLogger{
		logger: log.New(os.Stdout, "", log.LstdFlags),
	}
}

// Debug logs an info-level message with structured key-value pairs.
func (d *DefaultLogger) Debug(msg string, args ...any) {
	d.logger.Println(createLog(msg, "DEBUG", args...))
}

// Info logs an info-level message with structured key-value pairs.
func (d *DefaultLogger) Info(msg string, args ...any) {
	d.logger.Println(createLog(msg, "INFO", args...))
}

// Warn logs an info-level message with structured key-value pairs.
func (d *DefaultLogger) Warn(msg string, args ...any) {
	d.logger.Println(createLog(msg, "WARN", args...))
}

// Error logs an info-level message with structured key-value pairs.
func (d *DefaultLogger) Error(msg string, args ...any) {
	d.logger.Println(createLog(msg, "ERROR", args...))
}

func (d *DefaultLogger) Shutdown() error {
	d.logger.Println(createLog("default logger shutdown", "ERROR"))
	return nil
}

func (d *DefaultLogger) Sync() {
	// not needed
}

func createLog(msg string, level string, args ...any) string {
	timestamp := time.Now().Format(time.RFC3339)
	logMsg := fmt.Sprintf("level=%s time=%s msg=%q", level, timestamp, msg)

	if len(args)%2 != 0 {
		logMsg += "[invalid key-value pairs]"
	} else {
		for i := 0; i < len(args); i += 2 {
			key := fmt.Sprintf("%v", args[i])
			val := fmt.Sprintf("%v", args[i+1])
			logMsg += fmt.Sprintf(" %s=%s", key, val)
		}
	}
	return logMsg
}
