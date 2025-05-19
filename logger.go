package common_logger

type Logger interface {
	Info(message string, args ...any)
	Debug(message string, args ...any)
	Error(message string, args ...any)
	Warn(message string, args ...any)
	Shutdown() error
	Sync()
}
