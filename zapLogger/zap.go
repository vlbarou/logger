package zapLogger

import (
	"context"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"net/http"
	"os"
	"path/filepath"
	"sync"
)

var (
	internalLogger *zap.Logger
)

type LoggerImpl struct {
	mainLogger         *zap.Logger
	maxSizeMB          int
	maxBackups         int
	maxAge             int
	logServerPort      string
	logRotationEnabled bool
	logFile            string
	doneCh             chan struct{} // signal that logger server terminated
	ctx                context.Context
	cancel             context.CancelFunc
	atomicLevel        zap.AtomicLevel // Create an AtomicLevel to control logging level at runtime
	wg                 sync.WaitGroup
}

func New() *LoggerImpl {

	ctx, cancel := context.WithCancel(context.Background())
	logger := &LoggerImpl{
		maxAge:             MaxAge,
		maxBackups:         MaxBackups,
		maxSizeMB:          MaxSizeMB,
		logServerPort:      LoggerServerPort,
		logFile:            LogFile,
		logRotationEnabled: false,
		atomicLevel:        zap.NewAtomicLevelAt(zap.InfoLevel),
		ctx:                ctx,
		cancel:             cancel,
		doneCh:             make(chan struct{}, 1), // buffered to avoid blocking
	}

	return logger
}

func (logger *LoggerImpl) WithLogRotation(r bool) *LoggerImpl {
	logger.logRotationEnabled = r
	return logger
}

func (logger *LoggerImpl) WithLogfile(l string) *LoggerImpl {
	logger.logFile = l
	return logger
}

func (logger *LoggerImpl) WithMaxSizeMB(m int) *LoggerImpl {
	logger.maxSizeMB = m
	return logger
}

func (logger *LoggerImpl) WithMaxBackups(m int) *LoggerImpl {
	logger.maxBackups = m
	return logger
}

func (logger *LoggerImpl) WithMaxAge(m int) *LoggerImpl {
	logger.maxAge = m
	return logger
}

func (logger *LoggerImpl) WithPort(port string) *LoggerImpl {
	logger.logServerPort = port
	return logger
}

// isIgnorableSyncError safely ignores the error thrown when trying to sync to `os.Stdout`
// In particular, Zap's Sync() flushes buffered logs to the underlying writer.
// When the writer happens to be `os.Stdout` or `os.Stderr`, Zap tries to fsync() (flush to disk).
//
// But `/dev/stdout` is not a real file on disk. It is a pseudo-device, and fsync() on it is not supported.
// Hence, in case of trying to sync with `os.Stdout` the error says “You can't fsync() a device like /dev/stdout.”
// and we can safely ignore it
var isIgnorableSyncError = func(err error) bool {
	var pathErr *os.PathError
	return errors.As(err, &pathErr) && pathErr.Path == "/dev/stdout"
}

func (logger *LoggerImpl) Sync() {
	logger.mainLogger.Sync()
	internalLogger.Sync()
}

func (logger *LoggerImpl) Shutdown() error {
	var err1, err2 error

	// Trigger cancellation to start shutdown process
	logger.cancel()

	// Wait until all goroutines started by Start() have finished
	<-logger.doneCh

	// close internal logger (just flush to disk in-flight data)
	if err := internalLogger.Sync(); err != nil && !isIgnorableSyncError(err) {
		err1 = err
	}
	// close main logger (just flush to disk in-flight data)
	if err := logger.mainLogger.Sync(); err != nil && !isIgnorableSyncError(err) {
		err2 = err
	}

	return errors.Join(err1, err2)
}

func (logger *LoggerImpl) IsShutdown() bool {
	select {
	case <-logger.doneCh:
		return true
	default:
		return false
	}
}

//func (logger *LoggerImpl) Start() *LoggerImpl {
//
//	logger.doneCh = make(chan struct{}, 1)
//	go func(ctx context.Context) {
//		defer close(logger.doneCh)
//
//		mux := http.NewServeMux()
//		mux.HandleFunc(LogServerURI, logger.logLevelHandler)
//
//		server := &http.Server{
//			Addr:    ":" + logger.logServerPort,
//			Handler: mux,
//		}
//
//		internalLogger.Info(fmt.Sprintf("Starting log server on :%s", logger.logServerPort))
//		go func() {
//			if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
//				logger.Error("HTTP server failed", zap.Error(err))
//			}
//		}()
//
//		<-ctx.Done()
//
//		internalLogger.Info("Shutting down log server...")
//		shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), GracefulShutdownTimeout)
//		defer cancelShutdown()
//
//		if err := server.Shutdown(shutdownCtx); err != nil {
//			internalLogger.Error("Server shutdown failed", zap.Error(err))
//		}
//	}(logger.ctx)
//
//	mLogger, iLogger := logger.createLogger()
//	logger.mainLogger = mLogger
//	internalLogger = iLogger
//	return logger
//}

func (logger *LoggerImpl) Start() *LoggerImpl {
	logger.createLogger()

	mux := http.NewServeMux()
	mux.HandleFunc(LogServerURI, logger.logLevelHandler)

	server := &http.Server{
		Addr:    ":" + logger.logServerPort,
		Handler: mux,
	}

	internalLogger.Info(fmt.Sprintf("Starting log server on :%s", logger.logServerPort))

	logger.wg.Add(1)
	go func() {
		defer logger.wg.Done()
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("HTTP server failed", zap.Error(err))
		}
	}()

	logger.wg.Add(1)
	go func() {
		defer logger.wg.Done()
		<-logger.ctx.Done() // When context is canceled this unblocks and the shutdown process continues

		internalLogger.Info("Shutting down log server...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), GracefulShutdownTimeout)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Error("Server shutdown failed", zap.Error(err))
		}
	}()

	// Close doneCh after all goroutines finish
	go func() {
		logger.wg.Wait()
		close(logger.doneCh)
	}()

	return logger
}

func (logger *LoggerImpl) logLevelHandler(w http.ResponseWriter, r *http.Request) {
	level := r.URL.Query().Get("level")
	if level == "" {
		http.Error(w, "level is required", http.StatusBadRequest)
		return
	}

	var newLevel zapcore.Level
	if err := newLevel.UnmarshalText([]byte(level)); err != nil {
		http.Error(w, fmt.Sprintf("invalid level: %v", err), http.StatusBadRequest)
		return
	}

	logger.atomicLevel.SetLevel(newLevel)
	internalLogger.Error("Log level changed", zap.String("new_level", newLevel.String()))
	fmt.Fprintf(w, "Log level set to %s\n", newLevel.String())
}

func (logger *LoggerImpl) Info(message string, args ...any) {
	fields := toZapFields(args...)
	logger.mainLogger.Info(message, fields...)
}

func (logger *LoggerImpl) Debug(message string, args ...any) {
	fields := toZapFields(args...)
	logger.mainLogger.Debug(message, fields...)
}

func (logger *LoggerImpl) Error(message string, args ...any) {
	fields := toZapFields(args...)
	logger.mainLogger.Error(message, fields...)
}

func (logger *LoggerImpl) Warn(message string, args ...any) {
	fields := toZapFields(args...)
	logger.mainLogger.Warn(message, fields...)
}

func (logger *LoggerImpl) createLogger() {
	var file zapcore.WriteSyncer
	stdout := zapcore.AddSync(os.Stdout)

	if logger.logRotationEnabled {
		/*
			lumberjack.Logger doesn't have a built-in Shutdown or Close method.
			So once started, it's a zombie goroutine unless the process exits.
		*/
		file = zapcore.AddSync(&lumberjack.Logger{
			Filename:   logger.logFile,
			MaxSize:    logger.maxSizeMB,  // Maximum size (in MB) of a single log file before it gets rotated (e.g., 10MB)
			MaxBackups: logger.maxBackups, // Number of old log files to keep (e.g., 3 old logs)
			MaxAge:     logger.maxAge,     // Maximum age (in days) to retain old log files (e.g., 7 days)
			Compress:   Compress,          // ✅ compress rotated files (.gz), hardcoded in the format "<filename>-<timestamp>.gz"
		})
	} else {
		fileHandle, err := OpenOrCreateFile(logger.logFile)
		if err != nil {
			panic(fmt.Sprintf("failed to open log file: %v", err))
		}
		file = zapcore.AddSync(fileHandle)
	}

	// Initialize AtomicLevel globally so it can be updated
	logger.atomicLevel = zap.NewAtomicLevelAt(zap.InfoLevel)

	productionCfg := zap.NewProductionEncoderConfig()
	productionCfg.TimeKey = TimeKey
	productionCfg.EncodeTime = zapcore.ISO8601TimeEncoder

	developmentCfg := zap.NewDevelopmentEncoderConfig()
	developmentCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder

	consoleEncoder := zapcore.NewConsoleEncoder(developmentCfg)
	fileEncoder := zapcore.NewJSONEncoder(productionCfg)

	core := zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, stdout, logger.atomicLevel),
		zapcore.NewCore(fileEncoder, file, logger.atomicLevel),
	)

	/*
		Since we use wrapper, we don't want just the "AddCaller". This would invoke the IMMEDIATE caller, which is the
		wrapper itself not the true origin of the log call.
		To fix this, you need to adjust the stack frame depth using zap.AddCallerSkip(n), where n is how many stack frames to skip beyond the default.
		Hence, zap.AddCallerSkip(N) skips N frames in the call stack to find the "real" caller.

		If you have multiple layers of wrapper functions (e.g., structured logger → generic logger → zap.Logger), you may need to bump AddCallerSkip(2) or more.

		With only 1 frame skipped, zap is more conservative, so even if the call stack is "flattened" or modified during shutdown, it can still locate a valid caller frame.
		Instead, if set to 2 it logs: "Logger.check error: failed to get caller"
	*/

	internalLogger = zap.New(core, zap.AddCaller())                          // use this logger to log in the wrapper
	logger.mainLogger = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(2)) // use this logger for your main app

	return
}

// OpenOrCreateFile ensures the directory exists, and opens the file for appending.
// If the file doesn't exist, it is created.
func OpenOrCreateFile(filePath string) (*os.File, error) {
	// Ensure the directory exists
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Open the file in append mode, create it if it doesn't exist
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, fmt.Errorf("failed to open or create file %s: %w", filePath, err)
	}

	return file, nil
}
