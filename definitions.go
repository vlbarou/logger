package common_logger

type (
	LoggerType int

	Config struct {
		MaxSizeMB   string
		MaxBackups  string
		MaxAge      string
		LogFile     string
		LogRotation bool
	}
)

const (
	Zap LoggerType = iota
)

func (l LoggerType) String() string {
	switch l {
	case Zap:
		return "Zap"
	default:
		return "Unknown"
	}
}
