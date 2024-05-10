package api

import (
	"log/slog"
	"os"
	"testing"

	"github.com/even0306/cloudreve_go_sdk/internal/config"
	"github.com/even0306/cloudreve_go_sdk/internal/logging"
	"github.com/even0306/cloudreve_go_sdk/requrl"
)

func TestMain(t *testing.T) {
	logging.NewLogger("debug")

	c, err := config.SetConfig("./config.yml")
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	requrl.Default(c.GetString("address"))

	srcPath := "D:\\Users\\even0\\Pictures\\壁纸\\1.jpg"
	f, err := os.Stat(srcPath)
	if err != nil {
		slog.Error(err.Error())
	}

	auth := NewAuthFunc()

	var userInfo = AuthUserInfo{
		UserName:    c.GetString("login.user"),
		Password:    c.GetString("login.password"),
		CaptchaCode: "",
	}

	err = auth.Login(userInfo)
	if err != nil {
		slog.Error(err.Error())
	}

	list := NewDirectoryListFunc()
	err = list.GetDirectoryList("/")
	if err != nil {
		slog.Error(err.Error())
	}

	var reqInfo = FileUploadReq{
		LastModified: f.ModTime().UnixMilli(),
		MIMEType:     "",
		Name:         f.Name(),
		Path:         list.Data.Objects[0].Path,
		PolicyID:     list.Data.Policy.ID,
		Size:         f.Size(),
	}

	up := NewFileUploadFunc()
	err = up.CreateUpload("storage", srcPath, reqInfo)
	if err != nil {
		slog.Error(err.Error())
		err = up.DeleteUploadSessionID(up.Data.SessionID)
		if err != nil {
			slog.Error(err.Error())
		}
	}
}
