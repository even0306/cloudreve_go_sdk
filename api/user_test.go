package api

import (
	"log/slog"
	"testing"

	"github.com/cloudreve_client/v2/client"
	"github.com/cloudreve_client/v2/config"
)

func TestUserGet(T *testing.T) {
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

	userOpt := NewUserOperation()
	userOpt.GetUserProfile(userInfo.Data.Id)
}
