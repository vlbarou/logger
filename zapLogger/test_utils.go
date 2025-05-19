package zapLogger

import (
	"net"
	"os"
	"path/filepath"
	"strconv"
)

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

func GetFreePort() string {

	// Bind to a random available port to find a free port
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return "0"
	}
	defer listener.Close()

	// Get the port from the listener's address
	address := listener.Addr().(*net.TCPAddr)
	return strconv.Itoa(address.Port)
}
