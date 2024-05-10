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

	"github.com/even0306/cloudreve_go_sdk/api/upload"
	"github.com/even0306/cloudreve_go_sdk/requrl"
)

type Operation interface {
	Download(fileInfo Object, dst string)
	Upload(reqInfo FileUploadReq)
	Move()
	Copy()
	GetDirectoryList(path string)
	DeleteUploadSessionID()
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

type FileUploadData struct {
	// 上传分片大小
	ChunkSize int64 `json:"chunkSize"`
	// 上传会话过期时间
	Expires int64 `json:"expires"`
	// 用于上传完成后确认
	SessionID   string   `json:"sessionID"`
	UploadURLs  []string `json:"uploadURLs"`
	UploadID    string   `json:"uploadID"`
	CompleteURL string   `json:"completeURL"`
}

type FileUploadResp struct {
	// 响应状态
	Code int64          `json:"code"`
	Data FileUploadData `json:"data"`
	// 错误信息
	Msg string `json:"msg"`
}

type FileUploadReq struct {
	// 待上传文件修改日期的毫秒级时间戳
	LastModified int64 `json:"last_modified"`
	// 待上传文件类型，可留空
	MIMEType string `json:"mime_type"`
	// 文件名
	Name string `json:"name"`
	// 文件上传路径（相对网盘根目录）
	Path string `json:"path"`
	// 存储策略ID，可从接口“文件管理/列目录”中获得
	PolicyID string `json:"policy_id"`
	// 文件大小（字节）
	Size int64 `json:"size"`
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

func NewFileUploadFunc() *FileUploadResp {
	return &FileUploadResp{
		Code: 0,
		Data: FileUploadData{},
		Msg:  "",
	}
}

// 获取目录列表
func (list *DirectoryList) GetDirectoryList(path string) {
	resp, err := requrl.Client.Get(requrl.ReqHost + "/api/v3/directory" + path)
	if err != nil {
		slog.Error(err.Error())
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.Warn("", "Status", resp.StatusCode)
		return
	}

	err = json.NewDecoder(resp.Body).Decode(&list)
	if err != nil {
		slog.Error(err.Error())
		return
	}

	if list.Code != 0 {
		slog.Error(fmt.Sprint(list.Code), "Msg", list.Msg, "Data", list.Data)
		return
	}

	slog.Debug(fmt.Sprint(list.Code), "Msg", list.Msg, "Data", list.Data)
}

// 下载文件
func (fileDownloadResp *FileDownloadResp) FileDownload(fileInfo Object, dst string) {
	req, err := http.NewRequest("PUT", requrl.ReqHost+"/api/v3/file/download/"+fileInfo.ID, nil)
	if err != nil {
		slog.Error(err.Error())
		return
	}

	resp, err := requrl.Client.Do(req)
	if err != nil {
		slog.Error(err.Error())
		return
	}

	if resp.StatusCode != http.StatusOK {
		slog.Warn("", "Status", resp.StatusCode)
	}

	err = json.NewDecoder(resp.Body).Decode(&fileDownloadResp)
	if err != nil {
		slog.Error(err.Error())
		return
	}

	slog.Info("", slog.Int("Code:", fileDownloadResp.Code), slog.String("Msg:", fileDownloadResp.Msg), slog.Any("Data:", fileDownloadResp.Data))

	if fileDownloadResp.Code != 0 {
		slog.Error(fmt.Sprint(fileDownloadResp.Code), "Msg", fileDownloadResp.Msg, "Data", fileDownloadResp.Data)
		return
	}

	req, err = http.NewRequest("GET", requrl.ReqHost+fileDownloadResp.Data, nil)
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
	}

	fileData, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error(err.Error())
		return
	}

	err = json.NewDecoder(bytes.NewReader(fileData)).Decode(&fileDownloadResp)
	if err != nil {
		slog.Error(err.Error())
		return
	}

	if fileDownloadResp.Code != 0 {
		slog.Error(fmt.Sprint(fileDownloadResp.Code), "Msg", fileDownloadResp.Msg, "Data", fileDownloadResp.Data)
		return
	}

	f, err := os.OpenFile(filepath.Join(dst, fileInfo.Name), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		slog.Error(err.Error())
		return
	}
	defer f.Close()

	io.Copy(f, bytes.NewReader(fileData))
}

// 上传文件
func (fileUploadResp *FileUploadResp) Upload(storage string, reqInfo FileUploadReq) {
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(reqInfo)
	if err != nil {
		slog.Error(err.Error())
		return
	}

	req, err := http.NewRequest("PUT", requrl.ReqHost+"/api/v3/file/upload", &buf)
	if err != nil {
		slog.Error(err.Error())
		return
	}

	// 创建上传会话
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

	err = json.NewDecoder(resp.Body).Decode(&fileUploadResp)
	if err != nil {
		slog.Error(err.Error())
		return
	}

	slog.Debug(fmt.Sprint(fileUploadResp.Code), "Data", fileUploadResp.Data, "Msg", fileUploadResp.Msg)

	if fileUploadResp.Code != 0 {
		slog.Error(fmt.Sprint(fileUploadResp.Code), "Msg", fileUploadResp.Msg, "Data", fileUploadResp.Data)
		return
	}

	switch storage {
	case "s3":
		fileUploadReq := upload.S3FileUploadReq{
			Session:     fileUploadResp.Data.SessionID,
			UploadURL:   fileUploadResp.Data.UploadID,
			CompleteURL: fileUploadResp.Data.CompleteURL,
		}
		s3 := upload.NewS3FileUploadFunc(fileUploadReq)
		s3.Upload()
	default:
		slog.Warn("不存在的存储类型")
	}
}

// 删除上传会话
func (fileUploadResp *FileUploadResp) DeleteUploadSessionID(fileUploadSessionID string) {
	req, err := http.NewRequest("DELETE", requrl.ReqHost+"/api/v3/file/upload/"+fileUploadSessionID, nil)
	if err != nil {
		slog.Error(err.Error())
		return
	}

	resp, err := requrl.Client.Do(req)
	if err != nil {
		slog.Error(err.Error())
		return
	}

	if resp.StatusCode != http.StatusOK {
		slog.Warn("", "Status", resp.StatusCode)
		return
	}

	err = json.NewDecoder(resp.Body).Decode(&fileUploadResp)
	if err != nil {
		slog.Error(err.Error())
		return
	}

	slog.Info(fmt.Sprint(fileUploadResp.Code), "Msg", fileUploadResp.Msg, "Data", fileUploadResp.Data)

	if fileUploadResp.Code != 0 {
		slog.Error(fmt.Sprint(fileUploadResp.Code), "Msg", fileUploadResp.Msg, "Data", fileUploadResp.Data)
		return
	}
}

func (f *FileMoveResp) Move() {

}

func (f *FileCopyResp) Copy() {

}
