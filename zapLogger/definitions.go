package zapLogger

import "time"

const (
	MaxSizeMB               = 10
	MaxBackups              = 3
	MaxAge                  = 7
	Compress                = true
	LogFile                 = "/tmp/logs/app.log"
	TimeKey                 = "timestamp"
	LoggerServerPort        = "8081"
	GracefulShutdownTimeout = 5 * time.Second
	LogServerURI            = "/loglevel"
)
