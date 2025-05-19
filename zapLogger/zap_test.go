package zapLogger

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"os"
	"strings"
	"testing"
	"time"
)

type ZapLogTestSuite struct {
	suite.Suite
	logger      *LoggerImpl
	tempDir     string
	tempLogFile *os.File
}

func (suite *ZapLogTestSuite) TearDownTest() {
	assert.Nil(suite.T(), suite.logger.Shutdown())
	assert.Nil(suite.T(), os.RemoveAll(suite.tempDir))
}

func (suite *ZapLogTestSuite) TestLoggerServerTerminates() {

	var err error
	suite.tempLogFile, suite.tempDir, err = createTempFile()

	suite.logger = New().WithLogfile(suite.tempLogFile.Name()).Start()

	// Allow some time for the server to start
	time.Sleep(100 * time.Millisecond)

	// Cancel context to trigger shutdown
	suite.logger.Shutdown()

	// Wait for the goroutine to finish
	select {
	case <-suite.logger.doneCh:
		// this is expected
	case <-time.After(2 * time.Second):
		suite.T().Fatal("log server goroutine did not terminate in time")
	}

	assert.Nil(suite.T(), err)
}

func (suite *ZapLogTestSuite) TestIsShutDown() {

	var err error
	suite.tempLogFile, suite.tempDir, err = createTempFile()

	suite.logger = New().WithLogfile(suite.tempLogFile.Name()).Start()

	// Allow some time for the server to start
	time.Sleep(100 * time.Millisecond)

	assert.Nil(suite.T(), err)
	assert.False(suite.T(), suite.logger.IsShutdown())

	suite.logger.Shutdown()
	assert.True(suite.T(), suite.logger.IsShutdown())
}

func (suite *ZapLogTestSuite) TestStartWithConfig() {

	var err error

	// arrange
	port := GetFreePort()
	suite.tempLogFile, suite.tempDir, err = createTempFile()

	suite.logger = New().
		WithLogfile(suite.tempLogFile.Name()).
		WithPort(port).
		WithMaxAge(10).
		WithMaxBackups(20).
		WithMaxSizeMB(30).
		Start()

	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), suite.logger)
	assert.Equal(suite.T(), 10, suite.logger.maxAge)
	assert.Equal(suite.T(), 20, suite.logger.maxBackups)
	assert.Equal(suite.T(), 30, suite.logger.maxSizeMB)
	assert.Equal(suite.T(), port, suite.logger.logServerPort)
	assert.Equal(suite.T(), suite.tempLogFile.Name(), suite.logger.logFile)
}

func (suite *ZapLogTestSuite) TestLogRotation() {
	var err error
	suite.tempLogFile, suite.tempDir, err = createTempFile()

	// Initialize logger with small max size
	suite.logger = New().
		WithLogfile(suite.tempLogFile.Name()).
		WithLogRotation(true).
		WithMaxSizeMB(1).
		Start()

	// Write logs repeatedly to trigger logRotationEnabled
	largeMsg := strings.Repeat("A", 100*1024) // 100 KB per log line

	for i := 0; i < 30; i++ { // Write ~ 2000 KB total
		suite.logger.Info("test log logRotationEnabled", "line", i, "payload", largeMsg)
	}

	// Sync writes
	suite.logger.Sync()

	// Give lumberjack time to compress rotated files
	time.Sleep(2 * time.Second)

	// Check rotated files
	files, err := os.ReadDir(suite.tempDir)

	// assert
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), 3, len(files))
}

func TestRouterTestSuite(t *testing.T) {
	suite.Run(t, new(ZapLogTestSuite))
}
