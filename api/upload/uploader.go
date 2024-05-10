package upload

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/even0306/cloudreve_go_sdk/api/requrl"
)

type UploadFunc interface {
	Upload()
}

type S3FileUploadReq struct {
	Session     string
	UploadURL   string
	CompleteURL string
}

type S3FileUploadResp struct {
	Data string
}

type StorageFileUploadResp struct {
	Data string
}

var s3FileUploadReq S3FileUploadReq

func NewS3FileUploadFunc(req S3FileUploadReq) *S3FileUploadResp {
	s3FileUploadReq = S3FileUploadReq{
		Session:     req.Session,
		UploadURL:   req.UploadURL,
		CompleteURL: req.CompleteURL,
	}

	return &S3FileUploadResp{}
}

func NewStorageFileUploadResp() *StorageFileUploadResp {
	return &StorageFileUploadResp{}
}

// S3上传
func (u *S3FileUploadResp) Upload() {
	var parmas url.Values
	parmas.Add()
	req, err := http.NewRequest("PUT", s3FileUploadReq.UploadURL, nil)
	if err != nil {
		slog.Error(err.Error())
		return
	}

	resp, err := requrl.Client.Do(req)
	if err != nil {
		slog.Error(err.Error())
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.Warn("", "Status", resp.StatusCode)
		return
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error(err.Error())
		return
	}

	if data != nil {
		slog.Error(string(data))
	}

	req, err = http.NewRequest("GET", requrl.ReqHost+"/api/v3/callback/s3"+s3FileUploadReq.Session, nil)
	if err != nil {
		slog.Error(err.Error())
		return
	}

	resp, err = requrl.Client.Do(req)
	if err != nil {
		slog.Error(err.Error())
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.Warn("", "Status", resp.StatusCode)
		return
	}

	err = json.NewDecoder(resp.Body).Decode(&u)
	if err != nil {
		slog.Error(err.Error())
		return
	}
}

// 本地存储上传
func (u *StorageFileUploadResp) Upload() {
	req, err := http.NewRequest("GET", requrl.ReqHost+"/api/v3/callback/s3"+s3FileUploadReq.Session, nil)
	if err != nil {
		slog.Error(err.Error())
		return
	}

	resp, err := requrl.Client.Do(req)
	if err != nil {
		slog.Error(err.Error())
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.Warn("", "Status", resp.StatusCode)
		return
	}

	err = json.NewDecoder(resp.Body).Decode(&u)
	if err != nil {
		slog.Error(err.Error())
		return
	}
}
