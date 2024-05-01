package api

import (
	"log/slog"
	"testing"

	"github.com/cloudreve_client/v2/client"
	"github.com/cloudreve_client/v2/config"
)

func TestGetDirectoryList(t *testing.T) {
	c, err := config.SetConfig("..")
	if err != nil {
		slog.Error(err.Error())
	}

	client.Default(c.GetString("address"))

	loginData := make(map[string]any)
	loginData["userName"] = c.GetString("login.user")
	loginData["password"] = c.GetString("login.password")

	userInfo := NewAuthFunc()
	userInfo.Login(loginData)

	listFunc := NewDirectoryListFunc()
	listFunc.GetDirectoryList("/")
}

func TestDownload(t *testing.T) {
	c, err := config.SetConfig("..")
	if err != nil {
		slog.Error(err.Error())
	}

	client.Default(c.GetString("address"))

	loginData := make(map[string]any)
	loginData["userName"] = c.GetString("login.user")
	loginData["password"] = c.GetString("login.password")

	userInfo := NewAuthFunc()
	userInfo.Login(loginData)

	listFunc := NewDirectoryListFunc()
	fileDownloadFunc := NewFileDownloadFunc()

	listFunc.GetDirectoryList("/")
	for _, v := range listFunc.Data.Objects {
		if v.Name == "传输助手.md" {
			fileDownloadFunc.FileDownload(v, c.GetStringSlice("path")[0])
		}
	}
}
