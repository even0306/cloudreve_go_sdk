package upload

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"

	"github.com/even0306/cloudreve_go_sdk/requrl"
)

type UploadFunc interface {
	Upload(srcPath string) error
}

type S3FileUploadReq struct {
	Session     string
	UploadURL   string
	CompleteURL string
}

type S3FileUploadResp struct {
	Code int
	Data []byte
	Etag string
	Msg  string
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
func (u *S3FileUploadResp) Upload(srcPath string) error {
	file, err := os.Open(srcPath)
	if err != nil {
		slog.Error(err.Error())
	}
	defer file.Close()

	body := &bytes.Buffer{}
	_, err = io.Copy(body, file)
	if err != nil {
		slog.Error(err.Error())
	}

	// 开始上传
	req, err := http.NewRequest("PUT", s3FileUploadReq.UploadURL, body)
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

	u.Etag = resp.Header.Get("Etag")

	u.Data, err = io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if u.Data != nil && string(u.Data) != "" {
		return fmt.Errorf(string(u.Data))
	}

	xmlBody := "<CompleteMultipartUpload><Part><PartNumber>1</PartNumber><ETag>" + u.Etag + "</ETag></Part></CompleteMultipartUpload>"

	// 完成上传
	req, err = http.NewRequest("POST", s3FileUploadReq.CompleteURL, bytes.NewReader([]byte(xmlBody)))
	if err != nil {
		return err
	}

	resp, err = requrl.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf(fmt.Sprint(resp.StatusCode))
	}

	u.Data, err = io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if string(u.Data) == "" {
		return fmt.Errorf(string(u.Data))
	}

	slog.Info(string(u.Data))

	// 验证上传完
	req, err = http.NewRequest("GET", requrl.ReqHost+"/api/v3/callback/s3/"+s3FileUploadReq.Session, nil)
	if err != nil {
		return err
	}

	resp, err = requrl.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf(fmt.Sprint(resp.StatusCode))
	}

	err = json.NewDecoder(resp.Body).Decode(&u)
	if err != nil {
		return err
	}

	if u.Code != 0 {
		return fmt.Errorf(fmt.Sprint(u.Code), "Msg", u.Msg)
	}

	return nil
}

// 本地存储上传
func (u *StorageFileUploadResp) Upload(srcPath string) error {
	req, err := http.NewRequest("GET", requrl.ReqHost+"/api/v3/callback/s3"+s3FileUploadReq.Session, nil)
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

	err = json.NewDecoder(resp.Body).Decode(&u)
	if err != nil {
		return err
	}

	return nil
}
