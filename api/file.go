package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/cloudreve_client/v2/client"
)

type Operation interface {
	Download()
	Upload()
	Move()
	Copy()
	GetDirectoryList()
}

type Policy struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	MaxSize  int    `json:"max_size"`
	FileType string `json:"file_type"`
}

type Object struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Path          string    `json:"path"`
	Thumb         bool      `json:"thumb"`
	Size          int       `json:"size"`
	Type          string    `json:"type"`
	Date          time.Time `json:"date"`
	CreateDate    time.Time `json:"create_date"`
	SourceEnabled bool      `json:"source_enabled"`
}

type DirectoryListData struct {
	Parent  string   `json:"parent"`
	Objects []Object `json:"objects"`
	Policy  Policy   `json:"policy"`
}

type DirectoryList struct {
	Code int               `json:"code"`
	Data DirectoryListData `json:"data"`
	Msg  string            `json:"msg"`
}

type FileDownloadResp struct {
	Code int    `json:"code"`
	Data string `json:"data"`
	Msg  string `json:"msg"`
}

type FileUploadResp struct {
}

type FileMoveResp struct {
}

type FileCopyResp struct {
}

func NewDirectoryListFunc() *DirectoryList {
	return &DirectoryList{
		Code: 0,
		Data: DirectoryListData{},
		Msg:  "",
	}
}

func NewFileDownloadFunc() *FileDownloadResp {
	return &FileDownloadResp{
		Code: 0,
		Data: "",
		Msg:  "",
	}
}

// 获取目录列表
func (list *DirectoryList) GetDirectoryList(path string) {
	resp, err := client.Client.Get(client.RespUrl + "/api/v3/directory" + path)
	if err != nil {
		slog.Error(err.Error())
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&list)
	if err != nil {
		slog.Error(err.Error())
	}
	slog.Info("返回结果：", slog.Int("Code:", list.Code), slog.String("Msg:", list.Msg), slog.Any("Data:", list.Data))
}

// 下载文件
func (fileDownloadResp *FileDownloadResp) FileDownload(fileInfo Object, path string) {
	req, err := http.NewRequest("PUT", client.RespUrl+"/api/v3/file/download/"+fileInfo.ID, nil)
	if err != nil {
		slog.Error(err.Error())
	}

	resp, err := client.Client.Do(req)
	if err != nil {
		slog.Error(err.Error())
	}

	if resp.StatusCode != http.StatusOK {
		slog.Warn("", "Status", resp.StatusCode)
	}

	err = json.NewDecoder(resp.Body).Decode(&fileDownloadResp)
	if err != nil {
		slog.Error(err.Error())
		fmt.Print(fileDownloadResp.Data)
	}

	slog.Info("", slog.Int("Code:", fileDownloadResp.Code), slog.String("Msg:", fileDownloadResp.Msg), slog.Any("Data:", fileDownloadResp.Data))

	if fileDownloadResp.Code != 0 {
		slog.Error(fmt.Sprint(fileDownloadResp.Code), "Msg", fileDownloadResp.Msg, "Data", fileDownloadResp.Data)
	}

	req, err = http.NewRequest("GET", client.RespUrl+fileDownloadResp.Data, nil)
	if err != nil {
		slog.Error(err.Error())
	}

	resp, err = client.Client.Do(req)
	if err != nil {
		slog.Error(err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.Warn("", "Status", resp.StatusCode)
	}

	fileData, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error(err.Error())
	}

	err = json.NewDecoder(bytes.NewReader(fileData)).Decode(&fileDownloadResp)
	if err != nil {
		slog.Error(err.Error())
	}

	if fileDownloadResp.Code != 0 {
		slog.Error(fmt.Sprint(fileDownloadResp.Code), "Msg", fileDownloadResp.Msg, "Data", fileDownloadResp.Data)
	}

	f, err := os.OpenFile(filepath.Join(path, fileInfo.Name), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		slog.Error(err.Error())
	}
	defer f.Close()

	io.Copy(f, bytes.NewReader(fileData))
}

// 上传文件
func (f *FileUploadResp) Upload() {

}

func (f *FileMoveResp) Move() {

}

func (f *FileCopyResp) Copy() {

}
