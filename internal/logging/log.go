package logging

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/even0306/cloudreve_go_sdk/internal/common"
)

func NewLogger(level string) {
	dir := common.GetExecPath()
	logFile, err := os.OpenFile(filepath.Join(dir), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		slog.Error(err.Error())
	}
	defer logFile.Close()

	var slogLevel slog.Level
	switch strings.ToLower(level) {
	case "debug":
		slogLevel = slog.LevelDebug
	case "info":
		slogLevel = slog.LevelInfo
	case "warnning":
		slogLevel = slog.LevelWarn
	case "error":
		slogLevel = slog.LevelError
	default:
		slog.Error("", "Panic", "不支持的日志等级")
	}

	logger := slog.New(slog.NewTextHandler(io.MultiWriter(os.Stdout, logFile), &slog.HandlerOptions{Level: slogLevel}))
	slog.SetDefault(logger)
}
