package upload

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/even0306/cloudreve_go_sdk/requrl"
)

type UploadFunc interface {
	// 传入待上传文件的路径和请求结构体
	Upload(srcPath string, reqInfo any) error
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

type StorageFileUploadReq struct {
	Session string
}

type StorageFileUploadResp struct {
	Code int
	Data []byte
	Msg  string
}

func NewS3FileUploadFunc() *S3FileUploadResp {
	return &S3FileUploadResp{}
}

func NewStorageFileUploadFunc() *StorageFileUploadResp {
	return &StorageFileUploadResp{}
}

// S3上传
func (u *S3FileUploadResp) Upload(srcPath string, reqInfo any) error {
	info := reqInfo.(S3FileUploadReq)
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
	req, err := http.NewRequest("PUT", info.UploadURL, body)
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
	req, err = http.NewRequest("POST", info.CompleteURL, bytes.NewReader([]byte(xmlBody)))
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

	spXml := strings.Split(string(u.Data), "<Code>")
	Code := strings.Split(spXml[1], "</Code>")
	if Code[0] == "AccessDenied" {
		return fmt.Errorf(string(u.Data))
	}

	slog.Info(string(u.Data))

	// 验证上传完
	req, err = http.NewRequest("GET", requrl.ReqHost+"/api/v3/callback/s3/"+info.Session, nil)
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
		return fmt.Errorf(u.Msg, fmt.Sprint(u.Code))
	}

	return nil
}

// 本地存储上传
func (u *StorageFileUploadResp) Upload(srcPath string, reqInfo any) error {
	info := reqInfo.(StorageFileUploadReq)
	file, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	_, err = io.Copy(body, file)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", requrl.ReqHost+"/api/v3/file/upload/"+info.Session+"/0", body)
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

	if u.Code != 0 {
		return fmt.Errorf(u.Msg, fmt.Sprint(u.Code))
	}

	return nil
}
