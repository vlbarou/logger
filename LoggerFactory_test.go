package common_logger

import (
	"os"
	"testing"
	"time"

	"github.com/vlbarou/logger/zapLogger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/vlbarou/logger/mocks"
)

type LoggerTestSuite struct {
	suite.Suite
	tempDir     string
	tempLogFile *os.File
}

func (suite *LoggerTestSuite) TearDownTest() {
	ResetLogger()
	assert.Nil(suite.T(), os.RemoveAll(suite.tempDir))
}

func (suite *LoggerTestSuite) TestCreateLoggerWithDefaultConfig() {
	var err error

	// arrange
	suite.tempLogFile, suite.tempDir, err = createTempFile()

	// act
	err = GetLogger(Zap, Config{LogFile: suite.tempLogFile.Name()})

	// assert
	l, ok := loggerInstance.(*zapLogger.LoggerImpl)

	assert.Nil(suite.T(), err)
	assert.True(suite.T(), ok)
	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), loggerInstance)
	assert.NotNil(suite.T(), l)

	maxAge, _ := getLoggerFieldValue(l, MaxAge).(int)
	assert.Equal(suite.T(), 7, maxAge)

	maxSizeMB, _ := getLoggerFieldValue(l, MaxSizeMB).(int)
	assert.Equal(suite.T(), 10, maxSizeMB)

	maxBackups, _ := getLoggerFieldValue(l, MaxBackups).(int)
	assert.Equal(suite.T(), 3, maxBackups)

	logFile, _ := getLoggerFieldValue(l, LogFile).(string)
	assert.Equal(suite.T(), suite.tempLogFile.Name(), logFile)
}

func (suite *LoggerTestSuite) TestCreateLoggerWithCustomConfig() {

	var err error

	// arrange
	suite.tempLogFile, suite.tempDir, err = createTempFile()

	// act
	config := Config{
		MaxSizeMB:  "1",
		MaxBackups: "2",
		MaxAge:     "3",
		LogFile:    suite.tempLogFile.Name(),
	}
	err = GetLogger(Zap, config)

	// assert
	l, ok := loggerInstance.(*zapLogger.LoggerImpl)

	assert.Nil(suite.T(), err)
	assert.True(suite.T(), ok)
	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), loggerInstance)
	assert.NotNil(suite.T(), l)

	logRotationEnabled, _ := getLoggerFieldValue(l, LogRotationEnabled).(bool)
	assert.False(suite.T(), logRotationEnabled)

	maxAge, _ := getLoggerFieldValue(l, MaxAge).(int)
	assert.Equal(suite.T(), 3, maxAge)

	maxSizeMB, _ := getLoggerFieldValue(l, MaxSizeMB).(int)
	assert.Equal(suite.T(), 1, maxSizeMB)

	maxBackups, _ := getLoggerFieldValue(l, MaxBackups).(int)
	assert.Equal(suite.T(), 2, maxBackups)

	logFile, _ := getLoggerFieldValue(l, LogFile).(string)
	assert.Equal(suite.T(), suite.tempLogFile.Name(), logFile)
}

func (suite *LoggerTestSuite) TestCallGetLoggerMultipleTimes() {

	var err error

	// arrange
	suite.tempLogFile, suite.tempDir, err = createTempFile()
	err = GetLogger(Zap, Config{LogFile: suite.tempLogFile.Name()})

	// act (try to create a new logger with a new configuration)
	config := Config{
		MaxSizeMB:   "1",
		MaxBackups:  "2",
		MaxAge:      "3",
		LogRotation: true,
	}
	err = GetLogger(Zap, config)

	// assert
	l, ok := loggerInstance.(*zapLogger.LoggerImpl)

	assert.Nil(suite.T(), err)
	assert.True(suite.T(), ok)
	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), loggerInstance)
	assert.NotNil(suite.T(), l)

	logRotationEnabled, _ := getLoggerFieldValue(l, LogRotationEnabled).(bool)
	assert.False(suite.T(), logRotationEnabled)
	maxAge, _ := getLoggerFieldValue(l, MaxAge).(int)
	assert.Equal(suite.T(), 7, maxAge)

	maxSizeMB, _ := getLoggerFieldValue(l, MaxSizeMB).(int)
	assert.Equal(suite.T(), 10, maxSizeMB)

	maxBackups, _ := getLoggerFieldValue(l, MaxBackups).(int)
	assert.Equal(suite.T(), 3, maxBackups)
}

func (suite *LoggerTestSuite) TestShutdown() {

	var err error

	// arrange
	suite.tempLogFile, suite.tempDir, err = createTempFile()

	err = GetLogger(Zap, Config{LogFile: suite.tempLogFile.Name()})
	time.Sleep(time.Second)

	// act
	err = Shutdown()

	// assert
	assert.Nil(suite.T(), err)
}

func (suite *LoggerTestSuite) TestLogEvents() {

	var err error

	// arrange
	suite.tempLogFile, suite.tempDir, err = createTempFile()
	loggerInstance = new(mocks.Logger)
	l, ok := loggerInstance.(*mocks.Logger)

	l.On("Info", mock.Anything, mock.Anything)
	l.On("Debug", mock.Anything, mock.Anything)
	l.On("Warn", mock.Anything, mock.Anything)
	l.On("Error", mock.Anything, mock.Anything)

	Info("test")
	Debug("test")
	Warn("test")
	Error("test")

	assert.Nil(suite.T(), err)
	assert.True(suite.T(), ok)
	l.AssertExpectations(suite.T())

}

func TestRouterTestSuite(t *testing.T) {
	suite.Run(t, new(LoggerTestSuite))
}
