package client

import (
	"log/slog"
	"net/http"
	"net/http/cookiejar"
)

var RespUrl string
var Client *http.Client

func Default(requestURL string) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		slog.Error(err.Error())
	}
	Client = &http.Client{Jar: jar}

	RespUrl = requestURL
}
