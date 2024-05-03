package common

import (
	"log/slog"
	"os"
	"path/filepath"
)

type OsFileOps interface {
	GetFileProperties()
}

type Properties struct {
	LastModified int64
	MIMEType     string
	Size         int64
	Name         string
	Path         string
}

func NewOsFileOpsFunc() *Properties {
	return &Properties{
		LastModified: 0,
		MIMEType:     "",
		Size:         0,
	}
}

// 获取文件属性
func (proper *Properties) GetFileProperties(file string) {
	fileInfo, err := os.Stat(file)
	if err != nil {
		slog.Error(err.Error())
	}

	filePath, _ := filepath.Split(file)

	proper.Name = fileInfo.Name()
	proper.LastModified = fileInfo.ModTime().UnixMilli()
	proper.Size = fileInfo.Size()
	proper.Path = filePath
}

// 获取执行文件所在路径
func GetExecPath() string {
	dir, err := os.Executable()
	if err != nil {
		slog.Error(err.Error())
	}

	dir, _ = filepath.Split(dir)
	slog.Info(dir)
	return dir
}
