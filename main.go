package main

import (
	"log/slog"
	"os"
	"path/filepath"

	"github.com/cloudreve_client/v2/client"
	"github.com/cloudreve_client/v2/config"
	"github.com/cloudreve_client/v2/logging"
)

var Logging *slog.Logger

func main() {
	fullPath, err := os.Executable()
	if err != nil {
		slog.Error(err.Error())
	}
	filePath := filepath.Dir(fullPath)

	logging.NewLogger(filePath)

	c, err := config.SetConfig(filePath)
	if err != nil {
		slog.Warn(err.Error())
	}

	client.Default(c.GetString("url"))

}
