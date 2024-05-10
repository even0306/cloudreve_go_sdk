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

	"github.com/even0306/cloudreve_go_sdk/requrl"
	"github.com/even0306/cloudreve_go_sdk/upload"
)

type Operation interface {
	Download(fileInfo Object, dst string) error
	Upload(reqInfo FileUploadReq) error
	Move() error
	Copy() error
	GetDirectoryList(path string) error
	DeleteUploadSessionID() error
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
func (list *DirectoryList) GetDirectoryList(path string) error {
	resp, err := requrl.Client.Get(requrl.ReqHost + "/api/v3/directory" + path)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf(fmt.Sprint(resp.StatusCode))
	}

	err = json.NewDecoder(resp.Body).Decode(&list)
	if err != nil {
		return err
	}

	if list.Code != 0 {
		slog.Error(fmt.Sprint(list.Code), "Msg", list.Msg, "Data", list.Data)
		return fmt.Errorf(fmt.Sprint(list.Code), list.Msg, list.Data)
	}

	slog.Debug(fmt.Sprint(list.Code), "Msg", list.Msg, "Data", list.Data)

	return nil
}

// 下载文件
func (fileDownloadResp *FileDownloadResp) FileDownload(fileInfo Object, dst string) error {
	req, err := http.NewRequest("PUT", requrl.ReqHost+"/api/v3/file/download/"+fileInfo.ID, nil)
	if err != nil {
		return err
	}

	resp, err := requrl.Client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf(fmt.Sprint(resp.StatusCode))
	}

	err = json.NewDecoder(resp.Body).Decode(&fileDownloadResp)
	if err != nil {
		return err
	}

	slog.Info("", slog.Int("Code:", fileDownloadResp.Code), slog.String("Msg:", fileDownloadResp.Msg), slog.Any("Data:", fileDownloadResp.Data))

	if fileDownloadResp.Code != 0 {
		return fmt.Errorf(fmt.Sprint(fileDownloadResp.Code), fileDownloadResp.Msg, fileDownloadResp.Data)
	}

	req, err = http.NewRequest("GET", requrl.ReqHost+fileDownloadResp.Data, nil)
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

	fileData, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	err = json.NewDecoder(bytes.NewReader(fileData)).Decode(&fileDownloadResp)
	if err != nil {
		return err
	}

	if fileDownloadResp.Code != 0 {
		return fmt.Errorf(fmt.Sprint(fileDownloadResp.Code), fileDownloadResp.Msg, fileDownloadResp.Data)
	}

	f, err := os.OpenFile(filepath.Join(dst, fileInfo.Name), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	defer f.Close()

	io.Copy(f, bytes.NewReader(fileData))

	return nil
}

// 上传文件
func (fileUploadResp *FileUploadResp) Upload(storage string, srcPath string, reqInfo FileUploadReq) error {
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(reqInfo)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PUT", requrl.ReqHost+"/api/v3/file/upload", &buf)
	if err != nil {
		return err
	}

	// 创建上传会话
	resp, err := requrl.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf(fmt.Sprint(resp.StatusCode))
	}

	err = json.NewDecoder(resp.Body).Decode(&fileUploadResp)
	if err != nil {
		return err
	}

	slog.Debug(fmt.Sprint(fileUploadResp.Code), "Data", fileUploadResp.Data, "Msg", fileUploadResp.Msg)

	if fileUploadResp.Code != 0 {
		return fmt.Errorf(fmt.Sprint(fileUploadResp.Code), fileUploadResp.Msg, fileUploadResp.Data)
	}

	// 根据存储类型上传文件
	switch storage {
	case "s3":
		fileUploadReq := upload.S3FileUploadReq{
			Session:     fileUploadResp.Data.SessionID,
			UploadURL:   fileUploadResp.Data.UploadURLs[0],
			CompleteURL: fileUploadResp.Data.CompleteURL,
		}
		s3 := upload.NewS3FileUploadFunc(fileUploadReq)
		err = s3.Upload(srcPath)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("不支持的存储类型")
	}

	return nil
}

// 删除上传会话
func (fileUploadResp *FileUploadResp) DeleteUploadSessionID(fileUploadSessionID string) error {
	req, err := http.NewRequest("DELETE", requrl.ReqHost+"/api/v3/file/upload/"+fileUploadSessionID, nil)
	if err != nil {
		return err
	}

	resp, err := requrl.Client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf(fmt.Sprint(resp.StatusCode))
	}

	err = json.NewDecoder(resp.Body).Decode(&fileUploadResp)
	if err != nil {
		return err
	}

	slog.Info(fmt.Sprint(fileUploadResp.Code), "Msg", fileUploadResp.Msg, "Data", fileUploadResp.Data)

	if fileUploadResp.Code != 0 {
		return fmt.Errorf(fmt.Sprint(fileUploadResp.Code), fileUploadResp.Msg, fileUploadResp.Data)
	}

	return nil
}

func (f *FileMoveResp) Move() error {
	return nil
}

func (f *FileCopyResp) Copy() error {
	return nil
}
