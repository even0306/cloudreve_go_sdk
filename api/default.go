package api

import (
	"log/slog"
	"net/http"
	"net/http/cookiejar"
)

var (
	// 包含请求协议，主机地址和端口
	ReqHost string
	Client  *http.Client
)

func Default(requestHost string) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		slog.Error(err.Error())
	}
	Client = &http.Client{Jar: jar}

	ReqHost = requestHost
}
