package common_logger

import (
	"strconv"

	"github.com/vlbarou/logger/zapLogger"
)

func startLogger(config []Config) Logger {
	loggerInstance = zapLogger.New()
	l, _ := loggerInstance.(*zapLogger.LoggerImpl)

	if len(config) > 0 {
		if config[0].LogFile != "" {
			l.WithLogfile(config[0].LogFile)
		}

		if config[0].MaxSizeMB != "" {
			l.WithMaxSizeMB(toInt(config[0].MaxSizeMB))
		}

		if config[0].MaxBackups != "" {
			l.WithMaxBackups(toInt(config[0].MaxBackups))
		}

		if config[0].MaxAge != "" {
			l.WithMaxAge(toInt(config[0].MaxAge))
		}

		l.WithLogRotation(config[0].LogRotation)
	}

	l.Start()
	return l
}

func toInt(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}
