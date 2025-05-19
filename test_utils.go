package common_logger

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"unsafe"
)

const (
	MaxAge             = "maxAge"
	MaxSizeMB          = "maxSizeMB"
	MaxBackups         = "maxBackups"
	LogFile            = "logFile"
	LogRotationEnabled = "logRotationEnabled"
)

// createTempFile create a temp file in the default /tmp directory
func createTempFile() (file *os.File, tempDir string, err error) {
	// Step 1: Create a temporary directory
	tempDir, err = os.MkdirTemp("", "my-temp-dir-*")
	if err != nil {
		return
	}

	// Step 2: Create a file inside the temp directory
	tempFilePath := filepath.Join(tempDir, "app.log")
	file, err = os.Create(tempFilePath)
	return
}

func getLoggerFieldValue(logger any, param string) any {
	v := reflect.ValueOf(logger)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		if fieldType.Name == param {
			// If field is unexported, use unsafe to access its value
			if !field.CanInterface() {
				field = reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem()
			}
			return field.Interface()
		}
	}

	fmt.Printf("Field %q not found\n", param)
	return nil
}

// ResetLogger re-initializes singleton `GetLogger` method
// The trick is to reset the internal state of `Once` struct
func ResetLogger() {
	once = sync.Once{}
	loggerInstance = nil
	initError = nil
}
