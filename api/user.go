package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/cloudreve_client/v2/client"
)

type UserAPI interface {
	GetUserInfo()
}

type User struct {
	Id    string `json:"id"`
	Nick  string `json:"nick"`
	Group string `json:"group"`
	Date  string `json:"date"`
}

type Source struct {
	Name string `json:"name"`
	Size int    `json:"size"`
}

type UserItems struct {
	Key             string    `json:"key"`
	IsDir           bool      `json:"is_dir"`
	Password        string    `json:"password"`
	CreateDate      time.Time `json:"create_date"`
	Downloads       int       `json:"downloads"`
	RemainDownloads int       `json:"remain_downloads"`
	Views           int       `json:"views"`
	Expire          int       `json:"expire"`
	Preview         bool      `json:"preview"`
	Source          Source    `json:"source"`
}

type UserData struct {
	Items []UserItems `json:"items"`
	Total int         `json:"total"`
	User  User        `json:"user"`
}

type UserRespBody struct {
	Code int      `json:"code"`
	Data UserData `json:"data"`
	Msg  string   `json:"msg"`
}

func NewUserOperation() *UserRespBody {
	return &UserRespBody{
		Code: 0,
		Data: UserData{},
		Msg:  "",
	}
}

func (respBody *UserRespBody) GetUserProfile(id string) {
	params := url.Values{}
	params.Add("type", "default")
	params.Add("page", "1")
	req, err := http.NewRequest("GET", client.RespUrl+"/api/v3/user/profile/"+id+"?"+params.Encode(), nil)
	if err != nil {
		slog.Error(err.Error())
	}
	resp, err := client.Client.Do(req)
	if err != nil {
		slog.Error(err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.Error("", "Status", resp.StatusCode)
	}

	err = json.NewDecoder(resp.Body).Decode(&respBody)
	if err != nil {
		slog.Error(err.Error())
	}

	slog.Info("返回结果：", slog.Int("Code:", respBody.Code), slog.String("Msg:", respBody.Msg), slog.Any("Data:", respBody.Data))
}
