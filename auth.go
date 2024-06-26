package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/even0306/cloudreve_go_sdk/requrl"
)

type Auth interface {
	Login() error
}

type Group struct {
	Id                   int    `json:"id"`
	Name                 string `json:"name"`
	AllowShare           bool   `json:"allowShare"`
	AllowRemoteDownload  bool   `json:"allowRemoteDownload"`
	AllowArchiveDownload bool   `json:"allowArchiveDownload"`
	ShareDownload        bool   `json:"shareDownload"`
	Compress             bool   `json:"compress"`
	Webdav               bool   `json:"webdav"`
	SourceBatch          int    `json:"sourceBatch"`
	AdvanceDelete        bool   `json:"advanceDelete"`
	AllowWebDAVProxy     bool   `json:"allowWebDAVProxy"`
}

type AuthData struct {
	Id              string   `json:"id"`
	User_name       string   `json:"user_name"`
	Nickname        string   `json:"nickname"`
	Status          int      `json:"status"`
	Avatar          string   `json:"avatar"`
	Created_at      string   `json:"created_at"`
	Preferred_theme string   `json:"preferred_theme"`
	Anonymous       bool     `json:"anonymous"`
	Group           Group    `json:"group"`
	Tags            []string `json:"tags"`
}

type AuthRespBody struct {
	Code int      `json:"code"`
	Data AuthData `json:"data"`
	Msg  string   `json:"msg"`
}

type AuthUserInfo struct {
	UserName    string
	Password    string
	CaptchaCode string
}

func NewAuthFunc() *AuthRespBody {
	return &AuthRespBody{
		Code: 0,
		Data: AuthData{},
		Msg:  "",
	}
}

func (respBody *AuthRespBody) Login(loginData AuthUserInfo) error {
	b, err := json.Marshal(&loginData)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", requrl.ReqHost+"/api/v3/user/session", bytes.NewReader(b))
	if err != nil {
		return err
	}

	resp, err := requrl.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf(fmt.Sprint(resp.StatusCode))
	}

	err = json.NewDecoder(resp.Body).Decode(&respBody)
	if err != nil {
		return err
	}
	slog.Info(respBody.Msg, "Code", respBody.Code, "Data", respBody.Data)

	return nil
}
