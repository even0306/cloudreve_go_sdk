package logging

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
)

func NewLogger(filePath string) {
	logFile, err := os.OpenFile(filepath.Join("server.log"), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		slog.Error(err.Error())
	}
	logger := slog.New(slog.NewTextHandler(io.MultiWriter(os.Stdout, logFile), &slog.HandlerOptions{Level: slog.LevelError}))
	slog.SetDefault(logger)
}
